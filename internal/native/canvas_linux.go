//go:build linux

package native

/*
#cgo pkg-config: freetype2 cairo
#include <X11/Xlib.h>
#include <X11/Xft/Xft.h>
#include <cairo/cairo.h>
#include <math.h>

static void wtrto_xft_set_clip(XftDraw *draw, int x, int y, int w, int h) {
    XRectangle rect;
    rect.x = 0;
    rect.y = 0;
    rect.width = w;
    rect.height = h;
    XftDrawSetClipRectangles(draw, x, y, &rect, 1);
}

static void wtrto_xft_clear_clip(XftDraw *draw) {
    XftDrawSetClip(draw, None);
}
*/
import "C"

import "unsafe"

type Canvas struct {
	win *Window
}

func (c *Canvas) Size() (int, int) {
	return c.win.Size()
}

func setSource(cr *C.cairo_t, col Color) {
	C.cairo_set_source_rgba(cr,
		C.double(col.R)/255,
		C.double(col.G)/255,
		C.double(col.B)/255,
		C.double(col.A)/255,
	)
}

func (c *Canvas) ClipRect(r Rect) {
	C.wtrto_xft_set_clip(c.win.xftDraw, C.int(r.X), C.int(r.Y), C.int(r.W), C.int(r.H))
}

func (c *Canvas) Unclip() {
	C.wtrto_xft_clear_clip(c.win.xftDraw)
}

func (c *Canvas) FillRect(r Rect, col Color) {
	cr := c.win.cairoCtx
	setSource(cr, col)
	C.cairo_rectangle(cr, C.double(r.X), C.double(r.Y), C.double(r.W), C.double(r.H))
	C.cairo_fill(cr)
}

func (c *Canvas) StrokeRect(r Rect, col Color, width int) {
	cr := c.win.cairoCtx
	setSource(cr, col)
	C.cairo_set_line_width(cr, C.double(width))
	half := float64(width) / 2
	C.cairo_rectangle(cr, C.double(float64(r.X)+half), C.double(float64(r.Y)+half), C.double(float64(r.W)-2*half), C.double(float64(r.H)-2*half))
	C.cairo_stroke(cr)
}

func roundedPath(cr *C.cairo_t, r Rect, radius int) {
	rad := float64(radius)
	if max := float64(r.W); rad > max/2 {
		rad = max / 2
	}
	if max := float64(r.H); rad > max/2 {
		rad = max / 2
	}
	x, y, w, h := float64(r.X), float64(r.Y), float64(r.W), float64(r.H)
	deg := C.double(3.14159265358979 / 180.0)

	C.cairo_new_sub_path(cr)
	C.cairo_arc(cr, C.double(x+w-rad), C.double(y+rad), C.double(rad), -90*deg, 0*deg)
	C.cairo_arc(cr, C.double(x+w-rad), C.double(y+h-rad), C.double(rad), 0*deg, 90*deg)
	C.cairo_arc(cr, C.double(x+rad), C.double(y+h-rad), C.double(rad), 90*deg, 180*deg)
	C.cairo_arc(cr, C.double(x+rad), C.double(y+rad), C.double(rad), 180*deg, 270*deg)
	C.cairo_close_path(cr)
}

func (c *Canvas) FillRoundedRect(r Rect, radius int, col Color) {
	cr := c.win.cairoCtx
	setSource(cr, col)
	roundedPath(cr, r, radius)
	C.cairo_fill(cr)
}

func (c *Canvas) StrokeRoundedRect(r Rect, radius int, col Color, width int) {
	cr := c.win.cairoCtx
	setSource(cr, col)
	C.cairo_set_line_width(cr, C.double(width))
	inset := float64(width) / 2
	roundedPath(cr, Rect{X: r.X + int(inset), Y: r.Y + int(inset), W: r.W - int(2*inset), H: r.H - int(2*inset)}, radius)
	C.cairo_stroke(cr)
}

func (c *Canvas) FillCircle(cx, cy, radius int, col Color) {
	cr := c.win.cairoCtx
	setSource(cr, col)
	C.cairo_arc(cr, C.double(cx), C.double(cy), C.double(radius), 0, 2*3.14159265358979)
	C.cairo_fill(cr)
}

func (c *Canvas) Line(points []Point, col Color, width int) {
	if len(points) < 2 {
		return
	}
	cr := c.win.cairoCtx
	setSource(cr, col)
	C.cairo_set_line_width(cr, C.double(width))
	C.cairo_set_line_cap(cr, C.CAIRO_LINE_CAP_ROUND)
	C.cairo_set_line_join(cr, C.CAIRO_LINE_JOIN_ROUND)

	C.cairo_move_to(cr, C.double(points[0].X), C.double(points[0].Y))
	for _, p := range points[1:] {
		C.cairo_line_to(cr, C.double(p.X), C.double(p.Y))
	}
	C.cairo_stroke(cr)
}

func (c *Canvas) text(x, y int, col Color, size int, bold bool, s string) {
	w := c.win
	font := w.font(size, bold)
	if font == nil {
		return
	}
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	var xc C.XRenderColor
	xc.red = C.ushort(uint16(col.R) * 257)
	xc.green = C.ushort(uint16(col.G) * 257)
	xc.blue = C.ushort(uint16(col.B) * 257)
	xc.alpha = C.ushort(uint16(col.A) * 257)

	var xftColor C.XftColor
	C.XftColorAllocValue(w.display, w.visual, w.colormap, &xc, &xftColor)
	defer C.XftColorFree(w.display, w.visual, w.colormap, &xftColor)

	C.cairo_surface_flush(w.cairoSurface)
	C.XftDrawStringUtf8(w.xftDraw, &xftColor, font, C.int(x), C.int(y), (*C.XftChar8)(unsafe.Pointer(cs)), C.int(len(s)))
	C.cairo_surface_mark_dirty(w.cairoSurface)
}

func (c *Canvas) Text(x, y int, col Color, size int, s string) {
	c.text(x, y, col, size, false, s)
}

func (c *Canvas) TextBold(x, y int, col Color, size int, s string) {
	c.text(x, y, col, size, true, s)
}

func (c *Canvas) textSize(s string, size int, bold bool) (int, int) {
	w := c.win
	font := w.font(size, bold)
	if font == nil {
		return 0, 0
	}
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	var extents C.XGlyphInfo
	C.XftTextExtentsUtf8(w.display, font, (*C.XftChar8)(unsafe.Pointer(cs)), C.int(len(s)), &extents)
	return int(extents.xOff), int(font.height)
}

func (c *Canvas) TextSize(s string, size int) (int, int) {
	return c.textSize(s, size, false)
}

func (c *Canvas) TextSizeBold(s string, size int) (int, int) {
	return c.textSize(s, size, true)
}

func (c *Canvas) DrawArtificialHorizon(cx, cy, radius int, pitchDeg, rollDeg float64, sky, ground, line, border Color) {
	cr := c.win.cairoCtx
	tau := 2 * 3.14159265358979

	C.cairo_save(cr)
	C.cairo_arc(cr, C.double(cx), C.double(cy), C.double(radius), 0, C.double(tau))
	C.cairo_clip(cr)

	C.cairo_translate(cr, C.double(cx), C.double(cy))
	C.cairo_rotate(cr, C.double(-rollDeg*3.14159265358979/180))

	big := float64(radius) * 3
	pitchOffset := pitchDeg / 90.0 * float64(radius)

	setSource(cr, sky)
	C.cairo_rectangle(cr, C.double(-big), C.double(-big+pitchOffset), C.double(2*big), C.double(big))
	C.cairo_fill(cr)

	setSource(cr, ground)
	C.cairo_rectangle(cr, C.double(-big), C.double(pitchOffset), C.double(2*big), C.double(big))
	C.cairo_fill(cr)

	setSource(cr, line)
	C.cairo_set_line_width(cr, 2)
	C.cairo_move_to(cr, C.double(-big), C.double(pitchOffset))
	C.cairo_line_to(cr, C.double(big), C.double(pitchOffset))
	C.cairo_stroke(cr)

	C.cairo_restore(cr)

	setSource(cr, border)
	C.cairo_set_line_width(cr, 2)
	C.cairo_arc(cr, C.double(cx), C.double(cy), C.double(radius), 0, C.double(tau))
	C.cairo_stroke(cr)

	setSource(cr, line)
	C.cairo_set_line_width(cr, 2)
	C.cairo_move_to(cr, C.double(cx-14), C.double(cy))
	C.cairo_line_to(cr, C.double(cx-4), C.double(cy))
	C.cairo_stroke(cr)
	C.cairo_move_to(cr, C.double(cx+4), C.double(cy))
	C.cairo_line_to(cr, C.double(cx+14), C.double(cy))
	C.cairo_stroke(cr)
}
