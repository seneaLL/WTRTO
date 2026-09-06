package hud

import (
	"fmt"
	"math"
	"sort"

	"github.com/seneaLL/WTRTO/internal/native"
)

func wrapValue(v, wrap float64) float64 {
	if wrap <= 0 {
		return v
	}
	v = math.Mod(v, wrap)
	if v < 0 {
		v += wrap
	}

	return v
}

func isMajorTick(tick, minorStep, majorStep float64) bool {
	if minorStep <= 0 || majorStep <= 0 {
		return false
	}
	n := math.Round(tick / minorStep)
	m := math.Round(majorStep / minorStep)
	if m <= 0 {
		return false
	}

	return math.Mod(math.Mod(n, m)+m, m) == 0
}

func zoneColor(zones []Zone, value float64, fallback native.Color) native.Color {
	if len(zones) == 0 {
		return fallback
	}
	sorted := append([]Zone(nil), zones...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Threshold < sorted[j].Threshold })

	if value <= sorted[0].Threshold {
		return toNativeColor(sorted[0].Color)
	}
	last := sorted[len(sorted)-1]
	if value >= last.Threshold {
		return toNativeColor(last.Color)
	}
	for i := 0; i < len(sorted)-1; i++ {
		a, b := sorted[i], sorted[i+1]
		if value >= a.Threshold && value <= b.Threshold {
			t := (value - a.Threshold) / (b.Threshold - a.Threshold)

			return lerpColor(toNativeColor(a.Color), toNativeColor(b.Color), t)
		}
	}

	return fallback
}

func lerpColor(a, b native.Color, t float64) native.Color {
	lerp := func(x, y uint8) uint8 {
		return uint8(float64(x) + (float64(y)-float64(x))*t)
	}

	return native.Color{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B), A: lerp(a.A, b.A)}
}

func elementThickness(e Element) int {
	if e.Thickness > 0 {
		return e.Thickness
	}

	return 1
}

func tickValueAt(offsetPx, value, pxPerUnit, dirMul float64) float64 {
	return value + offsetPx/(pxPerUnit*dirMul)
}

const spineSegmentPx = 4

func arcPointV(cx, cy, radius, sign int, theta float64, extra int) (float64, float64) {
	r := float64(radius + extra)
	cxF := float64(cx) - float64(sign*radius)
	x := cxF + r*math.Cos(theta)*float64(sign)
	y := float64(cy) + r*math.Sin(theta)

	return x, y
}

func arcPointH(cx, cy, radius int, theta float64, extra int) (float64, float64) {
	r := float64(radius + extra)
	cyF := float64(cy) + float64(radius)
	x := float64(cx) + r*math.Sin(theta)
	y := cyF - r*math.Cos(theta)

	return x, y
}

func tapeVSign(e Element) int {
	switch e.LabelSide {
	case SideLeft:
		return 1
	case SideRight:
		return -1
	default:
		if e.X >= 0.5 {
			return -1
		}

		return 1
	}
}

func drawTapeV(c *native.Canvas, cx, cy, lengthPx int, value float64, e Element, v Values, baseCol native.Color) {
	if e.Range <= 0 || e.MinorStep <= 0 {
		return
	}
	col := tapeZoneColor(e, v, value, baseCol)
	thickness := elementThickness(e)
	half := e.Range / 2
	pxPerUnit := float64(lengthPx) / e.Range
	fontSize := e.FontSize
	if fontSize == 0 {
		fontSize = 13
	}

	sign := tapeVSign(e)

	dirMul := 1.0
	if e.Direction == DirDown {
		dirMul = -1
	}

	arc := e.Style == StyleArc
	radius := lengthPx / 2

	if arc {
		prevX, prevY, havePrev := 0.0, 0.0, false
		for t := -float64(lengthPx / 2); t <= float64(lengthPx/2); t += spineSegmentPx {
			x, y := arcPointV(cx, cy, radius, sign, t/float64(radius), 0)
			if havePrev {
				segCol := tapeZoneColor(e, v, tickValueAt(t-spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
				c.Line([]native.Point{{X: prevX, Y: prevY}, {X: x, Y: y}}, segCol, thickness)
			}
			prevX, prevY, havePrev = x, y, true
		}
	} else {
		for py := -lengthPx / 2; py < lengthPx/2; py += spineSegmentPx {
			py2 := py + spineSegmentPx
			if py2 > lengthPx/2 {
				py2 = lengthPx / 2
			}
			segCol := tapeZoneColor(e, v, tickValueAt(float64(py)+spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
			c.Line([]native.Point{{X: float64(cx), Y: float64(cy + py)}, {X: float64(cx), Y: float64(cy + py2)}}, segCol, thickness)
		}
	}

	start := math.Floor((value-half)/e.MinorStep) * e.MinorStep
	end := value + half

	for tick := start; tick <= end+e.MinorStep; tick += e.MinorStep {
		offset := (tick - value) * pxPerUnit * dirMul
		if offset < -float64(lengthPx/2) || offset > float64(lengthPx/2) {
			continue
		}

		display := wrapValue(tick, e.Wrap)
		major := isMajorTick(tick, e.MinorStep, e.MajorStep)
		tickCol := tapeZoneColor(e, v, tick, baseCol)

		tickLen := 6
		if major {
			tickLen = 14
		}

		var tx1, ty1, tx2, ty2, lx, ly float64
		if arc {
			theta := offset / float64(radius)
			tx1, ty1 = arcPointV(cx, cy, radius, sign, theta, 0)
			tx2, ty2 = arcPointV(cx, cy, radius, sign, theta, tickLen)
			lx, ly = arcPointV(cx, cy, radius, sign, theta, tickLen+6)
		} else {
			ty := float64(cy) + offset
			tx1, ty1 = float64(cx), ty
			tx2, ty2 = float64(cx+sign*tickLen), ty
			lx, ly = float64(cx+sign*(tickLen+6)), ty
		}
		c.Line([]native.Point{{X: tx1, Y: ty1}, {X: tx2, Y: ty2}}, tickCol, thickness)

		if major {
			label := fmt.Sprintf("%.0f", display)
			w, h := c.TextSize(label, fontSize)
			lxi := int(math.Round(lx))
			if sign < 0 {
				lxi -= w
			}
			c.Text(lxi, int(math.Round(ly))+h/2-2, tickCol, fontSize, label)
		}
	}

	boxLabel := fmt.Sprintf("%.0f", value)
	bw, bh := c.TextSize(boxLabel, fontSize+2)
	boxW := bw + 16
	boxH := bh + 10
	boxX := cx
	if sign <= 0 {
		boxX = cx - boxW
	}
	boxRect := native.Rect{X: boxX, Y: cy - boxH/2, W: boxW, H: boxH}
	c.FillRect(boxRect, native.Color{R: 10, G: 12, B: 15, A: 220})
	c.StrokeRect(boxRect, col, thickness+1)
	c.TextCentered(boxRect, col, fontSize+2, boxLabel)
}

func drawTapeH(c *native.Canvas, cx, cy, lengthPx int, value float64, e Element, v Values, baseCol native.Color) {
	if e.Range <= 0 || e.MinorStep <= 0 {
		return
	}
	col := tapeZoneColor(e, v, value, baseCol)
	thickness := elementThickness(e)
	half := e.Range / 2
	pxPerUnit := float64(lengthPx) / e.Range
	fontSize := e.FontSize
	if fontSize == 0 {
		fontSize = 13
	}

	dirMul := 1.0
	if e.Direction == DirCCW {
		dirMul = -1
	}

	arc := e.Style == StyleArc
	radius := lengthPx / 2

	if arc {
		prevX, prevY, havePrev := 0.0, 0.0, false
		for t := -float64(lengthPx / 2); t <= float64(lengthPx/2); t += spineSegmentPx {
			x, y := arcPointH(cx, cy, radius, t/float64(radius), 0)
			if havePrev {
				segCol := tapeZoneColor(e, v, tickValueAt(t-spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
				c.Line([]native.Point{{X: prevX, Y: prevY}, {X: x, Y: y}}, segCol, thickness)
			}
			prevX, prevY, havePrev = x, y, true
		}
	} else {
		for px := -lengthPx / 2; px < lengthPx/2; px += spineSegmentPx {
			px2 := px + spineSegmentPx
			if px2 > lengthPx/2 {
				px2 = lengthPx / 2
			}
			segCol := tapeZoneColor(e, v, tickValueAt(float64(px)+spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
			c.Line([]native.Point{{X: float64(cx + px), Y: float64(cy)}, {X: float64(cx + px2), Y: float64(cy)}}, segCol, thickness)
		}
	}

	start := math.Floor((value-half)/e.MinorStep) * e.MinorStep
	end := value + half

	for tick := start; tick <= end+e.MinorStep; tick += e.MinorStep {
		offset := (tick - value) * pxPerUnit * dirMul
		if offset < -float64(lengthPx/2) || offset > float64(lengthPx/2) {
			continue
		}

		display := wrapValue(tick, e.Wrap)
		major := isMajorTick(tick, e.MinorStep, e.MajorStep)
		tickCol := tapeZoneColor(e, v, tick, baseCol)

		tickLen := 6
		if major {
			tickLen = 14
		}

		var tx1, ty1, tx2, ty2, lx, ly float64
		if arc {
			theta := offset / float64(radius)
			tx1, ty1 = arcPointH(cx, cy, radius, theta, 0)
			tx2, ty2 = arcPointH(cx, cy, radius, theta, tickLen)
			lx, ly = arcPointH(cx, cy, radius, theta, tickLen+6)
		} else {
			tx := float64(cx) + offset
			tx1, ty1 = tx, float64(cy)
			tx2, ty2 = tx, float64(cy-tickLen)
			lx, ly = tx, float64(cy-tickLen-6)
		}
		c.Line([]native.Point{{X: tx1, Y: ty1}, {X: tx2, Y: ty2}}, tickCol, thickness)

		if major {
			label := fmt.Sprintf("%.0f", display)
			w, _ := c.TextSize(label, fontSize)
			c.Text(int(math.Round(lx))-w/2, int(math.Round(ly)), tickCol, fontSize, label)
		}
	}

	c.Line([]native.Point{{X: float64(cx), Y: float64(cy + 4)}, {X: float64(cx - 8), Y: float64(cy + 16)}, {X: float64(cx + 8), Y: float64(cy + 16)}, {X: float64(cx), Y: float64(cy + 4)}}, col, thickness+1)

	boxLabel := fmt.Sprintf("%.0f", wrapValue(value, e.Wrap))
	bw, bh := c.TextSize(boxLabel, fontSize+2)
	boxRect := native.Rect{X: cx - bw/2 - 8, Y: cy - bh - 22, W: bw + 16, H: bh + 10}
	c.FillRect(boxRect, native.Color{R: 10, G: 12, B: 15, A: 220})
	c.StrokeRect(boxRect, col, thickness+1)
	c.TextCentered(boxRect, col, fontSize+2, boxLabel)
}
