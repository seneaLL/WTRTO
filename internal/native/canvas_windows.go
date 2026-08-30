package native

import (
	"syscall"
	"unsafe"
)

var (
	gdiplus = syscall.NewLazyDLL("gdiplus.dll")

	procGdiplusStartup               = gdiplus.NewProc("GdiplusStartup")
	procGdipCreateFromHDC            = gdiplus.NewProc("GdipCreateFromHDC")
	procGdipDeleteGraphics           = gdiplus.NewProc("GdipDeleteGraphics")
	procGdipGraphicsClear            = gdiplus.NewProc("GdipGraphicsClear")
	procGdipSetSmoothingMode         = gdiplus.NewProc("GdipSetSmoothingMode")
	procGdipSetTextRenderingHint     = gdiplus.NewProc("GdipSetTextRenderingHint")
	procGdipSetClipRectI             = gdiplus.NewProc("GdipSetClipRectI")
	procGdipResetClip                = gdiplus.NewProc("GdipResetClip")
	procGdipCreateSolidFill          = gdiplus.NewProc("GdipCreateSolidFill")
	procGdipDeleteBrush              = gdiplus.NewProc("GdipDeleteBrush")
	procGdipFillRectangleI           = gdiplus.NewProc("GdipFillRectangleI")
	procGdipCreatePen1               = gdiplus.NewProc("GdipCreatePen1")
	procGdipDeletePen                = gdiplus.NewProc("GdipDeletePen")
	procGdipDrawRectangleI           = gdiplus.NewProc("GdipDrawRectangleI")
	procGdipSetPenLineCap197819      = gdiplus.NewProc("GdipSetPenLineCap197819")
	procGdipSetPenLineJoin           = gdiplus.NewProc("GdipSetPenLineJoin")
	procGdipDrawLinesI               = gdiplus.NewProc("GdipDrawLinesI")
	procGdipFillEllipseI             = gdiplus.NewProc("GdipFillEllipseI")
	procGdipCreatePath               = gdiplus.NewProc("GdipCreatePath")
	procGdipDeletePath               = gdiplus.NewProc("GdipDeletePath")
	procGdipAddPathArcI              = gdiplus.NewProc("GdipAddPathArcI")
	procGdipClosePathFigure          = gdiplus.NewProc("GdipClosePathFigure")
	procGdipFillPath                 = gdiplus.NewProc("GdipFillPath")
	procGdipDrawPath                 = gdiplus.NewProc("GdipDrawPath")
	procGdipSetClipPath              = gdiplus.NewProc("GdipSetClipPath")
	procGdipCreateFontFamilyFromName = gdiplus.NewProc("GdipCreateFontFamilyFromName")
	procGdipCreateFont               = gdiplus.NewProc("GdipCreateFont")
	procGdipDeleteFont               = gdiplus.NewProc("GdipDeleteFont")
	procGdipDrawString               = gdiplus.NewProc("GdipDrawString")
	procGdipMeasureString            = gdiplus.NewProc("GdipMeasureString")
	procGdipRotateWorldTransform     = gdiplus.NewProc("GdipRotateWorldTransform")
	procGdipTranslateWorldTransform  = gdiplus.NewProc("GdipTranslateWorldTransform")
	procGdipResetWorldTransform      = gdiplus.NewProc("GdipResetWorldTransform")
)

const (
	smoothingModeAntiAlias = 4
	textRenderingHintAA    = 4
	combineModeReplace     = 0
	lineCapRound           = 2
	lineJoinRound          = 2
	unitPixel              = 2
	fontStyleRegular       = 0
	fontStyleBold          = 1
	matrixOrderPrepend     = 0
	fillModeAlternate      = 0
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

type Canvas struct {
	win *Window
}

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

func gdipCreateFromHDC(hdc uintptr) uintptr {
	var graphics uintptr
	procGdipCreateFromHDC.Call(hdc, uintptr(unsafe.Pointer(&graphics)))
	procGdipSetSmoothingMode.Call(graphics, smoothingModeAntiAlias)
	procGdipSetTextRenderingHint.Call(graphics, textRenderingHintAA)

	return graphics
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
	var brush uintptr
	procGdipCreateSolidFill.Call(argb(col), uintptr(unsafe.Pointer(&brush)))
	defer procGdipDeleteBrush.Call(brush)

	procGdipFillRectangleI.Call(c.win.graphics, brush, uintptr(r.X), uintptr(r.Y), uintptr(r.W), uintptr(r.H))
}

func (c *Canvas) StrokeRect(r Rect, col Color, width int) {
	var pen uintptr
	procGdipCreatePen1.Call(argb(col), floatBits(float32(width)), unitPixel, uintptr(unsafe.Pointer(&pen)))
	defer procGdipDeletePen.Call(pen)

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

	var brush uintptr
	procGdipCreateSolidFill.Call(argb(col), uintptr(unsafe.Pointer(&brush)))
	defer procGdipDeleteBrush.Call(brush)

	procGdipFillPath.Call(c.win.graphics, brush, path)
}

func (c *Canvas) StrokeRoundedRect(r Rect, radius int, col Color, width int) {
	half := width / 2
	inset := Rect{X: r.X + half, Y: r.Y + half, W: r.W - 2*half, H: r.H - 2*half}
	path := roundedRectPath(inset, radius)
	defer procGdipDeletePath.Call(path)

	var pen uintptr
	procGdipCreatePen1.Call(argb(col), floatBits(float32(width)), unitPixel, uintptr(unsafe.Pointer(&pen)))
	defer procGdipDeletePen.Call(pen)

	procGdipDrawPath.Call(c.win.graphics, pen, path)
}

func (c *Canvas) FillCircle(cx, cy, radius int, col Color) {
	var brush uintptr
	procGdipCreateSolidFill.Call(argb(col), uintptr(unsafe.Pointer(&brush)))
	defer procGdipDeleteBrush.Call(brush)

	procGdipFillEllipseI.Call(c.win.graphics, brush, uintptr(cx-radius), uintptr(cy-radius), uintptr(radius*2), uintptr(radius*2))
}

func (c *Canvas) Line(points []Point, col Color, width int) {
	if len(points) < 2 {
		return
	}

	var pen uintptr
	procGdipCreatePen1.Call(argb(col), floatBits(float32(width)), unitPixel, uintptr(unsafe.Pointer(&pen)))
	defer procGdipDeletePen.Call(pen)

	procGdipSetPenLineCap197819.Call(pen, lineCapRound, lineCapRound, lineCapRound)
	procGdipSetPenLineJoin.Call(pen, lineJoinRound)

	pts := make([]gpPointI, len(points))
	for i, p := range points {
		pts[i] = gpPointI{X: int32(p.X), Y: int32(p.Y)}
	}

	procGdipDrawLinesI.Call(c.win.graphics, pen, uintptr(unsafe.Pointer(&pts[0])), uintptr(len(pts)))
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

	var brush uintptr
	procGdipCreateSolidFill.Call(argb(col), uintptr(unsafe.Pointer(&brush)))
	defer procGdipDeleteBrush.Call(brush)

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
	font := c.fontFor(size, bold)
	if font == 0 || s == "" {
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

	return int(bbox.W), size + 4
}

func (c *Canvas) TextSize(s string, size int) (int, int) {
	return c.textSize(s, size, false)
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

	big := radius * 3
	pitchOffset := int(pitchDeg / 90.0 * float64(radius))

	var skyBrush, groundBrush uintptr
	procGdipCreateSolidFill.Call(argb(sky), uintptr(unsafe.Pointer(&skyBrush)))
	procGdipCreateSolidFill.Call(argb(ground), uintptr(unsafe.Pointer(&groundBrush)))

	procGdipFillRectangleI.Call(g, skyBrush, uintptr(-big), uintptr(-big+pitchOffset), uintptr(2*big), uintptr(big))
	procGdipFillRectangleI.Call(g, groundBrush, uintptr(-big), uintptr(pitchOffset), uintptr(2*big), uintptr(big))

	procGdipDeleteBrush.Call(skyBrush)
	procGdipDeleteBrush.Call(groundBrush)

	var linePen uintptr
	procGdipCreatePen1.Call(argb(line), floatBits(2), unitPixel, uintptr(unsafe.Pointer(&linePen)))
	horizonPts := []gpPointI{{X: int32(-big), Y: int32(pitchOffset)}, {X: int32(big), Y: int32(pitchOffset)}}
	procGdipDrawLinesI.Call(g, linePen, uintptr(unsafe.Pointer(&horizonPts[0])), 2)
	procGdipDeletePen.Call(linePen)

	procGdipResetWorldTransform.Call(g)
	procGdipResetClip.Call(g)

	var borderPen uintptr
	procGdipCreatePen1.Call(argb(border), floatBits(2), unitPixel, uintptr(unsafe.Pointer(&borderPen)))
	procGdipDrawRectangleI.Call(g, borderPen, uintptr(cx-radius), uintptr(cy-radius), uintptr(radius*2), uintptr(radius*2))
	procGdipDeletePen.Call(borderPen)

	var tickPen uintptr
	procGdipCreatePen1.Call(argb(line), floatBits(2), unitPixel, uintptr(unsafe.Pointer(&tickPen)))
	left := []gpPointI{{X: int32(cx - 14), Y: int32(cy)}, {X: int32(cx - 4), Y: int32(cy)}}
	right := []gpPointI{{X: int32(cx + 4), Y: int32(cy)}, {X: int32(cx + 14), Y: int32(cy)}}
	procGdipDrawLinesI.Call(g, tickPen, uintptr(unsafe.Pointer(&left[0])), 2)
	procGdipDrawLinesI.Call(g, tickPen, uintptr(unsafe.Pointer(&right[0])), 2)
	procGdipDeletePen.Call(tickPen)
}

func floatBits(f float32) uintptr {
	return uintptr(*(*uint32)(unsafe.Pointer(&f)))
}
