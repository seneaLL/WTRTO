package native

import (
	"syscall"
	"unsafe"
)

var (
	gdiplus = syscall.NewLazyDLL("gdiplus.dll")

	procGdiplusStartup                        = gdiplus.NewProc("GdiplusStartup")
	procGdipCreateFromHDC                     = gdiplus.NewProc("GdipCreateFromHDC")
	procGdipCreateBitmapFromScan0             = gdiplus.NewProc("GdipCreateBitmapFromScan0")
	procGdipGetImageGraphicsContext           = gdiplus.NewProc("GdipGetImageGraphicsContext")
	procGdipDisposeImage                      = gdiplus.NewProc("GdipDisposeImage")
	procGdipDeleteGraphics                    = gdiplus.NewProc("GdipDeleteGraphics")
	procGdipGraphicsClear                     = gdiplus.NewProc("GdipGraphicsClear")
	procGdipSetSmoothingMode                  = gdiplus.NewProc("GdipSetSmoothingMode")
	procGdipSetTextRenderingHint              = gdiplus.NewProc("GdipSetTextRenderingHint")
	procGdipSetClipRectI                      = gdiplus.NewProc("GdipSetClipRectI")
	procGdipResetClip                         = gdiplus.NewProc("GdipResetClip")
	procGdipCreateSolidFill                   = gdiplus.NewProc("GdipCreateSolidFill")
	procGdipDeleteBrush                       = gdiplus.NewProc("GdipDeleteBrush")
	procGdipFillRectangleI                    = gdiplus.NewProc("GdipFillRectangleI")
	procGdipCreatePen1                        = gdiplus.NewProc("GdipCreatePen1")
	procGdipDeletePen                         = gdiplus.NewProc("GdipDeletePen")
	procGdipDrawRectangleI                    = gdiplus.NewProc("GdipDrawRectangleI")
	procGdipSetPenLineCap197819               = gdiplus.NewProc("GdipSetPenLineCap197819")
	procGdipSetPenLineJoin                    = gdiplus.NewProc("GdipSetPenLineJoin")
	procGdipDrawLinesI                        = gdiplus.NewProc("GdipDrawLinesI")
	procGdipDrawLines                         = gdiplus.NewProc("GdipDrawLines")
	procGdipFillRectangle                     = gdiplus.NewProc("GdipFillRectangle")
	procGdipFillEllipseI                      = gdiplus.NewProc("GdipFillEllipseI")
	procGdipCreatePath                        = gdiplus.NewProc("GdipCreatePath")
	procGdipDeletePath                        = gdiplus.NewProc("GdipDeletePath")
	procGdipAddPathArcI                       = gdiplus.NewProc("GdipAddPathArcI")
	procGdipAddPathLine2                      = gdiplus.NewProc("GdipAddPathLine2")
	procGdipStartPathFigure                   = gdiplus.NewProc("GdipStartPathFigure")
	procGdipClosePathFigure                   = gdiplus.NewProc("GdipClosePathFigure")
	procGdipFillPath                          = gdiplus.NewProc("GdipFillPath")
	procGdipDrawPath                          = gdiplus.NewProc("GdipDrawPath")
	procGdipSetClipPath                       = gdiplus.NewProc("GdipSetClipPath")
	procGdipCreateFontFamilyFromName          = gdiplus.NewProc("GdipCreateFontFamilyFromName")
	procGdipCreateFont                        = gdiplus.NewProc("GdipCreateFont")
	procGdipDeleteFont                        = gdiplus.NewProc("GdipDeleteFont")
	procGdipDrawString                        = gdiplus.NewProc("GdipDrawString")
	procGdipMeasureString                     = gdiplus.NewProc("GdipMeasureString")
	procGdipStringFormatGetGenericTypographic = gdiplus.NewProc("GdipStringFormatGetGenericTypographic")
	procGdipRotateWorldTransform              = gdiplus.NewProc("GdipRotateWorldTransform")
	procGdipTranslateWorldTransform           = gdiplus.NewProc("GdipTranslateWorldTransform")
	procGdipResetWorldTransform               = gdiplus.NewProc("GdipResetWorldTransform")
)

const (
	smoothingModeAntiAlias = 4

	textRenderingHintAA   = 3
	combineModeReplace    = 0
	lineCapRound          = 2
	lineJoinRound         = 2
	unitPixel             = 2
	fontStyleRegular      = 0
	fontStyleBold         = 1
	matrixOrderPrepend    = 0
	fillModeAlternate     = 0
	fillModeWinding       = 1
	pixelFormat32bppPARGB = 0x000E200B
)

type gdiplusStartupInput struct {
	GdiplusVersion           uint32
	_                        uint32
	DebugEventCallback       uintptr
	SuppressBackgroundThread int32
	SuppressExternalCodecs   int32
}

type gpRectF struct{ X, Y, W, H float32 }

type gpPointI struct{ X, Y int32 }

type gpPointF struct{ X, Y float32 }

type Canvas struct {
	win *Window
}

type textSizeKey struct {
	s    string
	size int
	bold bool
}

type textSizeVal struct{ w, h int }

var gdipToken uintptr

func gdipInit() {
	if gdipToken != 0 {
		return
	}

	input := gdiplusStartupInput{GdiplusVersion: 1}
	var token uintptr
	procGdiplusStartup.Call(uintptr(unsafe.Pointer(&token)), uintptr(unsafe.Pointer(&input)), 0)
	gdipToken = token
}

func gdipCreateBitmapGraphics(width, height int, bits uintptr) (bitmap, graphics uintptr) {
	stride := width * 4
	procGdipCreateBitmapFromScan0.Call(
		uintptr(width), uintptr(height), uintptr(stride),
		uintptr(pixelFormat32bppPARGB), bits,
		uintptr(unsafe.Pointer(&bitmap)),
	)
	procGdipGetImageGraphicsContext.Call(bitmap, uintptr(unsafe.Pointer(&graphics)))
	procGdipSetSmoothingMode.Call(graphics, smoothingModeAntiAlias)
	procGdipSetTextRenderingHint.Call(graphics, textRenderingHintAA)

	return bitmap, graphics
}

func gdipDisposeImage(bitmap uintptr) {
	procGdipDisposeImage.Call(bitmap)
}

func gdipDeleteGraphics(g uintptr) {
	procGdipDeleteGraphics.Call(g)
}

func gdipClearGraphics(g uintptr) {
	procGdipGraphicsClear.Call(g, 0)
}

func argb(col Color) uintptr {
	return uintptr(uint32(col.A)<<24 | uint32(col.R)<<16 | uint32(col.G)<<8 | uint32(col.B))
}

func argbKey(col Color) uint32 {
	return uint32(col.A)<<24 | uint32(col.R)<<16 | uint32(col.G)<<8 | uint32(col.B)
}

type penKey struct {
	col   uint32
	width float32
}

func (c *Canvas) brushFor(col Color) uintptr {
	key := argbKey(col)
	if b, ok := c.win.brushes[key]; ok {
		return b
	}

	var brush uintptr
	procGdipCreateSolidFill.Call(argb(col), uintptr(unsafe.Pointer(&brush)))
	c.win.brushes[key] = brush

	return brush
}

func (c *Canvas) penFor(col Color, width float32) uintptr {
	key := penKey{col: argbKey(col), width: width}
	if p, ok := c.win.pens[key]; ok {
		return p
	}

	var pen uintptr
	procGdipCreatePen1.Call(argb(col), floatBits(width), unitPixel, uintptr(unsafe.Pointer(&pen)))
	c.win.pens[key] = pen

	return pen
}

func (c *Canvas) Size() (int, int) {
	return c.win.Size()
}

func (c *Canvas) ClipRect(r Rect) {
	procGdipSetClipRectI.Call(c.win.graphics, uintptr(r.X), uintptr(r.Y), uintptr(r.W), uintptr(r.H), combineModeReplace)
}

func (c *Canvas) Unclip() {
	procGdipResetClip.Call(c.win.graphics)
}

func (c *Canvas) FillRect(r Rect, col Color) {
	brush := c.brushFor(col)

	procGdipFillRectangleI.Call(c.win.graphics, brush, uintptr(r.X), uintptr(r.Y), uintptr(r.W), uintptr(r.H))
}

func (c *Canvas) StrokeRect(r Rect, col Color, width int) {
	pen := c.penFor(col, float32(width))

	half := width / 2
	procGdipDrawRectangleI.Call(c.win.graphics, pen, uintptr(r.X+half), uintptr(r.Y+half), uintptr(r.W-2*half), uintptr(r.H-2*half))
}

func roundedRectPath(r Rect, radius int) uintptr {
	if maxR := r.W / 2; radius > maxR {
		radius = maxR
	}

	if maxR := r.H / 2; radius > maxR {
		radius = maxR
	}

	d := radius * 2
	var path uintptr
	procGdipCreatePath.Call(fillModeAlternate, uintptr(unsafe.Pointer(&path)))

	procGdipAddPathArcI.Call(path, uintptr(r.X+r.W-d), uintptr(r.Y), uintptr(d), uintptr(d), floatBits(-90), floatBits(90))
	procGdipAddPathArcI.Call(path, uintptr(r.X+r.W-d), uintptr(r.Y+r.H-d), uintptr(d), uintptr(d), floatBits(0), floatBits(90))
	procGdipAddPathArcI.Call(path, uintptr(r.X), uintptr(r.Y+r.H-d), uintptr(d), uintptr(d), floatBits(90), floatBits(90))
	procGdipAddPathArcI.Call(path, uintptr(r.X), uintptr(r.Y), uintptr(d), uintptr(d), floatBits(180), floatBits(90))
	procGdipClosePathFigure.Call(path)

	return path
}

func (c *Canvas) FillRoundedRect(r Rect, radius int, col Color) {
	path := roundedRectPath(r, radius)
	defer procGdipDeletePath.Call(path)

	brush := c.brushFor(col)

	procGdipFillPath.Call(c.win.graphics, brush, path)
}

func (c *Canvas) StrokeRoundedRect(r Rect, radius int, col Color, width int) {
	half := width / 2
	inset := Rect{X: r.X + half, Y: r.Y + half, W: r.W - 2*half, H: r.H - 2*half}
	path := roundedRectPath(inset, radius)
	defer procGdipDeletePath.Call(path)

	pen := c.penFor(col, float32(width))

	procGdipDrawPath.Call(c.win.graphics, pen, path)
}

func (c *Canvas) FillCircle(cx, cy, radius int, col Color) {
	brush := c.brushFor(col)

	procGdipFillEllipseI.Call(c.win.graphics, brush, uintptr(cx-radius), uintptr(cy-radius), uintptr(radius*2), uintptr(radius*2))
}

func (c *Canvas) FillPath(subpaths [][]Point, col Color) {
	var path uintptr
	procGdipCreatePath.Call(fillModeWinding, uintptr(unsafe.Pointer(&path)))
	defer procGdipDeletePath.Call(path)

	for _, sp := range subpaths {
		if len(sp) < 2 {
			continue
		}
		pts := make([]gpPointF, len(sp))
		for i, p := range sp {
			pts[i] = gpPointF{X: float32(p.X), Y: float32(p.Y)}
		}
		procGdipStartPathFigure.Call(path)
		procGdipAddPathLine2.Call(path, uintptr(unsafe.Pointer(&pts[0])), uintptr(len(pts)))
		procGdipClosePathFigure.Call(path)
	}

	brush := c.brushFor(col)
	procGdipFillPath.Call(c.win.graphics, brush, path)
}

func (c *Canvas) Line(points []Point, col Color, width int) {
	if len(points) < 2 {
		return
	}

	pen := c.penFor(col, float32(width))

	procGdipSetPenLineCap197819.Call(pen, lineCapRound, lineCapRound, lineCapRound)
	procGdipSetPenLineJoin.Call(pen, lineJoinRound)

	pts := make([]gpPointF, len(points))
	for i, p := range points {
		pts[i] = gpPointF{X: float32(p.X), Y: float32(p.Y)}
	}

	procGdipDrawLines.Call(c.win.graphics, pen, uintptr(unsafe.Pointer(&pts[0])), uintptr(len(pts)))
}

func (c *Canvas) fontFor(size int, bold bool) uintptr {
	key := fontKey{size: size, bold: bold}
	if f, ok := c.win.fonts[key]; ok {
		return f
	}

	familyName, _ := syscall.UTF16PtrFromString("Segoe UI")

	var family uintptr
	procGdipCreateFontFamilyFromName.Call(uintptr(unsafe.Pointer(familyName)), 0, uintptr(unsafe.Pointer(&family)))

	style := uintptr(fontStyleRegular)
	if bold {
		style = fontStyleBold
	}

	var font uintptr
	procGdipCreateFont.Call(family, floatBits(float32(size)), style, unitPixel, uintptr(unsafe.Pointer(&font)))

	c.win.fonts[key] = font

	return font
}

func (c *Canvas) text(x, y int, col Color, size int, bold bool, s string) {
	font := c.fontFor(size, bold)
	if font == 0 {
		return
	}

	utf16Str, err := syscall.UTF16FromString(s)
	if err != nil {
		return
	}
	utf16Str = utf16Str[:len(utf16Str)-1]

	brush := c.brushFor(col)

	rc := gpRectF{X: float32(x), Y: float32(y - size), W: 4000, H: float32(size) * 2}

	if len(utf16Str) == 0 {
		return
	}

	procGdipDrawString.Call(
		c.win.graphics,
		uintptr(unsafe.Pointer(&utf16Str[0])), uintptr(len(utf16Str)),
		font, uintptr(unsafe.Pointer(&rc)), 0, brush,
	)
}

func (c *Canvas) Text(x, y int, col Color, size int, s string) {
	c.text(x, y, col, size, false, s)
}

func (c *Canvas) TextBold(x, y int, col Color, size int, s string) {
	c.text(x, y, col, size, true, s)
}

func (c *Canvas) textSize(s string, size int, bold bool) (int, int) {
	if s == "" {
		return 0, size + 4
	}

	key := textSizeKey{s: s, size: size, bold: bold}
	if v, ok := c.win.textSizes[key]; ok {
		return v.w, v.h
	}

	font := c.fontFor(size, bold)
	if font == 0 {
		return 0, size + 4
	}

	utf16Str, err := syscall.UTF16FromString(s)
	if err != nil {
		return 0, size + 4
	}
	utf16Str = utf16Str[:len(utf16Str)-1]

	if len(utf16Str) == 0 {
		return 0, size + 4
	}

	rc := gpRectF{X: 0, Y: 0, W: 4000, H: float32(size) * 3}
	var bbox gpRectF
	var charsFitted, linesFilled int32

	procGdipMeasureString.Call(
		c.win.graphics,
		uintptr(unsafe.Pointer(&utf16Str[0])), uintptr(len(utf16Str)),
		font, uintptr(unsafe.Pointer(&rc)), 0,
		uintptr(unsafe.Pointer(&bbox)), uintptr(unsafe.Pointer(&charsFitted)), uintptr(unsafe.Pointer(&linesFilled)),
	)

	w, h := int(bbox.W), size+4
	c.win.textSizes[key] = textSizeVal{w: w, h: h}

	return w, h
}

func (c *Canvas) TextSize(s string, size int) (int, int) {
	return c.textSize(s, size, false)
}

var typographicFormat uintptr

func genericTypographicFormat() uintptr {
	if typographicFormat == 0 {
		procGdipStringFormatGetGenericTypographic.Call(uintptr(unsafe.Pointer(&typographicFormat)))
	}

	return typographicFormat
}

func (c *Canvas) TextCentered(r Rect, col Color, size int, s string) {
	font := c.fontFor(size, false)
	if font == 0 {
		return
	}

	utf16Str, err := syscall.UTF16FromString(s)
	if err != nil {
		return
	}
	utf16Str = utf16Str[:len(utf16Str)-1]
	if len(utf16Str) == 0 {
		return
	}

	format := genericTypographicFormat()

	measureRc := gpRectF{X: 0, Y: 0, W: 4000, H: float32(size) * 3}
	var bbox gpRectF
	var charsFitted, linesFilled int32
	procGdipMeasureString.Call(
		c.win.graphics,
		uintptr(unsafe.Pointer(&utf16Str[0])), uintptr(len(utf16Str)),
		font, uintptr(unsafe.Pointer(&measureRc)), format,
		uintptr(unsafe.Pointer(&bbox)), uintptr(unsafe.Pointer(&charsFitted)), uintptr(unsafe.Pointer(&linesFilled)),
	)

	brush := c.brushFor(col)
	drawRc := gpRectF{
		X: float32(r.X) + (float32(r.W)-bbox.W)/2 - bbox.X,
		Y: float32(r.Y) + (float32(r.H)-bbox.H)/2 - bbox.Y,
		W: 4000,
		H: float32(size) * 3,
	}
	procGdipDrawString.Call(
		c.win.graphics,
		uintptr(unsafe.Pointer(&utf16Str[0])), uintptr(len(utf16Str)),
		font, uintptr(unsafe.Pointer(&drawRc)), format, brush,
	)
}

func (c *Canvas) TextVCentered(x int, r Rect, col Color, size int, s string) {
	font := c.fontFor(size, false)
	if font == 0 {
		return
	}

	utf16Str, err := syscall.UTF16FromString(s)
	if err != nil {
		return
	}
	utf16Str = utf16Str[:len(utf16Str)-1]
	if len(utf16Str) == 0 {
		return
	}

	format := genericTypographicFormat()

	measureRc := gpRectF{X: 0, Y: 0, W: 4000, H: float32(size) * 3}
	var bbox gpRectF
	var charsFitted, linesFilled int32
	procGdipMeasureString.Call(
		c.win.graphics,
		uintptr(unsafe.Pointer(&utf16Str[0])), uintptr(len(utf16Str)),
		font, uintptr(unsafe.Pointer(&measureRc)), format,
		uintptr(unsafe.Pointer(&bbox)), uintptr(unsafe.Pointer(&charsFitted)), uintptr(unsafe.Pointer(&linesFilled)),
	)

	brush := c.brushFor(col)
	drawRc := gpRectF{
		X: float32(x) - bbox.X,
		Y: float32(r.Y) + (float32(r.H)-bbox.H)/2 - bbox.Y,
		W: 4000,
		H: float32(size) * 3,
	}
	procGdipDrawString.Call(
		c.win.graphics,
		uintptr(unsafe.Pointer(&utf16Str[0])), uintptr(len(utf16Str)),
		font, uintptr(unsafe.Pointer(&drawRc)), format, brush,
	)
}

func (c *Canvas) TextSizeBold(s string, size int) (int, int) {
	return c.textSize(s, size, true)
}

func (c *Canvas) DrawArtificialHorizon(cx, cy, radius int, pitchDeg, rollDeg float64, sky, ground, line, border Color) {
	g := c.win.graphics

	ellipse := Rect{X: cx - radius, Y: cy - radius, W: radius * 2, H: radius * 2}
	var clipPath uintptr
	procGdipCreatePath.Call(fillModeAlternate, uintptr(unsafe.Pointer(&clipPath)))
	procGdipAddPathArcI.Call(clipPath, uintptr(ellipse.X), uintptr(ellipse.Y), uintptr(ellipse.W), uintptr(ellipse.H), floatBits(0), floatBits(360))
	procGdipSetClipPath.Call(g, clipPath, combineModeReplace)
	procGdipDeletePath.Call(clipPath)

	procGdipTranslateWorldTransform.Call(g, floatBits(float32(cx)), floatBits(float32(cy)), matrixOrderPrepend)
	procGdipRotateWorldTransform.Call(g, floatBits(float32(-rollDeg)), matrixOrderPrepend)

	big := float64(radius * 3)
	pitchOffset := pitchDeg / 90.0 * float64(radius)

	skyBrush := c.brushFor(sky)
	groundBrush := c.brushFor(ground)

	procGdipFillRectangle.Call(g, skyBrush, floatBits(float32(-big)), floatBits(float32(-big+pitchOffset)), floatBits(float32(2*big)), floatBits(float32(big)))
	procGdipFillRectangle.Call(g, groundBrush, floatBits(float32(-big)), floatBits(float32(pitchOffset)), floatBits(float32(2*big)), floatBits(float32(big)))

	linePen := c.penFor(line, 2)
	horizonPts := []gpPointF{{X: float32(-big), Y: float32(pitchOffset)}, {X: float32(big), Y: float32(pitchOffset)}}
	procGdipDrawLines.Call(g, linePen, uintptr(unsafe.Pointer(&horizonPts[0])), 2)

	procGdipResetWorldTransform.Call(g)
	procGdipResetClip.Call(g)

	borderPen := c.penFor(border, 2)
	procGdipDrawRectangleI.Call(g, borderPen, uintptr(cx-radius), uintptr(cy-radius), uintptr(radius*2), uintptr(radius*2))

	tickPen := c.penFor(line, 2)
	left := []gpPointI{{X: int32(cx - 14), Y: int32(cy)}, {X: int32(cx - 4), Y: int32(cy)}}
	right := []gpPointI{{X: int32(cx + 4), Y: int32(cy)}, {X: int32(cx + 14), Y: int32(cy)}}
	procGdipDrawLinesI.Call(g, tickPen, uintptr(unsafe.Pointer(&left[0])), 2)
	procGdipDrawLinesI.Call(g, tickPen, uintptr(unsafe.Pointer(&right[0])), 2)
}

func floatBits(f float32) uintptr {
	return uintptr(*(*uint32)(unsafe.Pointer(&f)))
}
