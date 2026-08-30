package hud

import (
	"fmt"
	"math"
	"sort"

	"github.com/seneal/wtrto/internal/native"
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

func arcPointV(cx, cy, radius, sign int, theta float64, extra int) (int, int) {
	r := float64(radius + extra)
	cxF := float64(cx) - float64(sign*radius)
	x := cxF + r*math.Cos(theta)*float64(sign)
	y := float64(cy) + r*math.Sin(theta)

	return int(x), int(y)
}

func arcPointH(cx, cy, radius int, theta float64, extra int) (int, int) {
	r := float64(radius + extra)
	cyF := float64(cy) + float64(radius)
	x := float64(cx) + r*math.Sin(theta)
	y := cyF - r*math.Cos(theta)

	return int(x), int(y)
}

func drawTapeV(c *native.Canvas, cx, cy, lengthPx int, value float64, e Element, baseCol native.Color) {
	if e.Range <= 0 || e.MinorStep <= 0 {
		return
	}
	col := zoneColor(e.Zones, value, baseCol)
	thickness := elementThickness(e)
	half := e.Range / 2
	pxPerUnit := float64(lengthPx) / e.Range
	fontSize := e.FontSize
	if fontSize == 0 {
		fontSize = 13
	}

	sign := 1
	switch e.LabelSide {
	case SideLeft:
		sign = 1
	case SideRight:
		sign = -1
	default:
		if e.X >= 0.5 {
			sign = -1
		}
	}

	dirMul := 1.0
	if e.Direction == DirDown {
		dirMul = -1
	}

	arc := e.Style == StyleArc
	radius := lengthPx / 2

	if arc {
		prevX, prevY, havePrev := 0, 0, false
		for t := -float64(lengthPx / 2); t <= float64(lengthPx/2); t += spineSegmentPx {
			x, y := arcPointV(cx, cy, radius, sign, t/float64(radius), 0)
			if havePrev {
				segCol := zoneColor(e.Zones, tickValueAt(t-spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
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
			segCol := zoneColor(e.Zones, tickValueAt(float64(py)+spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
			c.Line([]native.Point{{X: cx, Y: cy + py}, {X: cx, Y: cy + py2}}, segCol, thickness)
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
		tickCol := zoneColor(e.Zones, tick, baseCol)

		tickLen := 6
		if major {
			tickLen = 14
		}

		var tx1, ty1, tx2, ty2, lx, ly int
		if arc {
			theta := offset / float64(radius)
			tx1, ty1 = arcPointV(cx, cy, radius, sign, theta, 0)
			tx2, ty2 = arcPointV(cx, cy, radius, sign, theta, tickLen)
			lx, ly = arcPointV(cx, cy, radius, sign, theta, tickLen+6)
		} else {
			ty := cy + int(offset)
			tx1, ty1 = cx, ty
			tx2, ty2 = cx+sign*tickLen, ty
			lx, ly = cx+sign*(tickLen+6), ty
		}
		c.Line([]native.Point{{X: tx1, Y: ty1}, {X: tx2, Y: ty2}}, tickCol, thickness)

		if major {
			label := fmt.Sprintf("%.0f", display)
			w, h := c.TextSize(label, fontSize)
			if sign < 0 {
				lx -= w
			}
			c.Text(lx, ly+h/2-2, tickCol, fontSize, label)
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
	c.Text(boxRect.X+8, cy+bh/2-2, col, fontSize+2, boxLabel)
}

func drawTapeH(c *native.Canvas, cx, cy, lengthPx int, value float64, e Element, baseCol native.Color) {
	if e.Range <= 0 || e.MinorStep <= 0 {
		return
	}
	col := zoneColor(e.Zones, value, baseCol)
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
		prevX, prevY, havePrev := 0, 0, false
		for t := -float64(lengthPx / 2); t <= float64(lengthPx/2); t += spineSegmentPx {
			x, y := arcPointH(cx, cy, radius, t/float64(radius), 0)
			if havePrev {
				segCol := zoneColor(e.Zones, tickValueAt(t-spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
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
			segCol := zoneColor(e.Zones, tickValueAt(float64(px)+spineSegmentPx/2, value, pxPerUnit, dirMul), baseCol)
			c.Line([]native.Point{{X: cx + px, Y: cy}, {X: cx + px2, Y: cy}}, segCol, thickness)
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
		tickCol := zoneColor(e.Zones, tick, baseCol)

		tickLen := 6
		if major {
			tickLen = 14
		}

		var tx1, ty1, tx2, ty2, lx, ly int
		if arc {
			theta := offset / float64(radius)
			tx1, ty1 = arcPointH(cx, cy, radius, theta, 0)
			tx2, ty2 = arcPointH(cx, cy, radius, theta, tickLen)
			lx, ly = arcPointH(cx, cy, radius, theta, tickLen+6)
		} else {
			tx := cx + int(offset)
			tx1, ty1 = tx, cy
			tx2, ty2 = tx, cy-tickLen
			lx, ly = tx, cy-tickLen-6
		}
		c.Line([]native.Point{{X: tx1, Y: ty1}, {X: tx2, Y: ty2}}, tickCol, thickness)

		if major {
			label := fmt.Sprintf("%.0f", display)
			w, _ := c.TextSize(label, fontSize)
			c.Text(lx-w/2, ly, tickCol, fontSize, label)
		}
	}

	c.Line([]native.Point{{X: cx, Y: cy + 4}, {X: cx - 8, Y: cy + 16}, {X: cx + 8, Y: cy + 16}, {X: cx, Y: cy + 4}}, col, thickness+1)

	boxLabel := fmt.Sprintf("%.0f", wrapValue(value, e.Wrap))
	bw, bh := c.TextSize(boxLabel, fontSize+2)
	boxRect := native.Rect{X: cx - bw/2 - 8, Y: cy - bh - 22, W: bw + 16, H: bh + 10}
	c.FillRect(boxRect, native.Color{R: 10, G: 12, B: 15, A: 220})
	c.StrokeRect(boxRect, col, thickness+1)
	c.Text(boxRect.X+8, boxRect.Y+bh+2, col, fontSize+2, boxLabel)
}
