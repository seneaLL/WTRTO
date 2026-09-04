package hud

import (
	"fmt"

	"github.com/seneal/wtrto/internal/native"
)

type EditState struct {
	Dragging   string
	OffsetX    int
	OffsetY    int
	StartX     int
	StartY     int
	Selected   string
	FocusField string

	OpenDropdown   string
	DropdownScroll int

	HideBackground bool

	NumberEditBuf   string
	TextSelectedAll bool
	TextCursor      int

	NewTemplateName string
	ShareCode       string
	StatusMsg       string
	StatusOK        bool

	PendingDialog string
}

const PanelWidth = 340
const snapThreshold = 6

var snapGuideColor = native.Color{R: 255, G: 60, B: 220, A: 220}

func toNativeColor(c Color) native.Color {
	return native.Color{R: c.R, G: c.G, B: c.B, A: c.A}
}

func bindingValue(b Binding, v Values) float64 {
	switch b {
	case BindThrottlePct:
		return v.ThrottlePct
	case BindIAS:
		return v.IAS
	case BindTAS:
		return v.TAS
	case BindMach:
		return v.Mach
	case BindAltitude:
		return v.Altitude
	case BindFuelKg:
		return v.FuelKg
	case BindFuelTime:
		return v.FuelTimeMin
	case BindFuelRate:
		return v.FuelRateKgM
	case BindOilTemp1:
		return v.OilTemp1
	case BindOilTemp2:
		return v.OilTemp2
	case BindCompass:
		return v.Compass
	case BindAoA:
		return v.AoA
	case BindAoS:
		return v.AoS
	case BindGLoad:
		return v.GLoad
	case BindVSpeed:
		return v.VSpeed
	case BindIASRate:
		return v.IASRate
	}

	return 0
}

func formatValue(e Element, v Values) string {
	numStr := fmt.Sprintf(fmt.Sprintf("%%.%df", e.Precision), bindingValue(e.Binding, v))
	switch {
	case e.Label != "" && e.Unit != "":
		return fmt.Sprintf("%s  %s %s", e.Label, numStr, e.Unit)
	case e.Label != "":
		return fmt.Sprintf("%s  %s", e.Label, numStr)
	case e.Unit != "":
		return fmt.Sprintf("%s %s", numStr, e.Unit)
	default:
		return numStr
	}
}

func elementBounds(e Element, c *native.Canvas, screenW, screenH, x, y int) native.Rect {
	switch e.Kind {
	case KindHorizon:
		r := int(e.Size * float64(screenH))

		return native.Rect{X: x - r, Y: y - r, W: r * 2, H: r * 2}
	case KindTapeV:
		if e.Style == StyleArc {
			r := int(e.Length * float64(screenH) / 2)

			return native.Rect{X: x - r - 40, Y: y - r, W: r*2 + 40, H: r * 2}
		}

		return native.Rect{X: x - 30, Y: y - 30, W: 60, H: 60}
	case KindTapeH:
		if e.Style == StyleArc {
			r := int(e.Length * float64(screenW) / 2)

			return native.Rect{X: x - r, Y: y - r - 40, W: r * 2, H: r*2 + 40}
		}

		return native.Rect{X: x - 30, Y: y - 30, W: 60, H: 60}
	}
	fs := e.FontSize
	if fs == 0 {
		fs = 15
	}
	w, h := c.TextSize("XXXXXXXXXXXXXXXXXXXX", fs)

	return native.Rect{X: x - 4, Y: y - h, W: w, H: h + 8}
}

var (
	horizonSky    = native.Color{R: 40, G: 95, B: 165, A: 230}
	horizonGround = native.Color{R: 95, G: 65, B: 35, A: 230}
	editBgColor   = native.Color{R: 14, G: 16, B: 20, A: 235}
)

func clampToScreen(b native.Rect, screenW, screenH, x, y int) (int, int) {
	if b.X < 0 {
		x -= b.X
	} else if right := b.X + b.W; right > screenW {
		x -= right - screenW
	}

	if b.Y < 0 {
		y -= b.Y
	} else if bottom := b.Y + b.H; bottom > screenH {
		y -= bottom - screenH
	}

	return x, y
}

func Draw(c *native.Canvas, screenW, screenH int, tmpl *Template, v Values, editMode bool, edit *EditState, in *native.Input) {
	if !v.Valid && !editMode {
		return
	}

	if editMode && !edit.HideBackground {
		c.FillRect(native.Rect{X: 0, Y: 0, W: screenW, H: screenH}, editBgColor)
	}

	readOnly := IsBuiltin(tmpl.Name)

	for i := range tmpl.Elements {
		e := &tmpl.Elements[i]
		x := int(e.X * float64(screenW))
		y := int(e.Y * float64(screenH))

		var text string
		fs := e.FontSize
		if fs == 0 {
			fs = 15
		}

		switch e.Kind {
		case KindHorizon:
			r := int(e.Size * float64(screenH))
			x, y = clampToScreen(native.Rect{X: x - r, Y: y - r, W: r * 2, H: r * 2}, screenW, screenH, x, y)
		case KindTapeV, KindTapeH:
			x, y = clampToScreen(elementBounds(*e, c, screenW, screenH, x, y), screenW, screenH, x, y)
		default:
			text = formatValue(*e, v)
			tw, th := c.TextSize(text, fs)
			x, y = clampToScreen(native.Rect{X: x, Y: y - th, W: tw, H: th + 4}, screenW, screenH, x, y)
		}

		if e.GlowEnabled {
			var bound native.Rect
			switch e.Kind {
			case KindHorizon:
				r := int(e.Size * float64(screenH))
				bound = native.Rect{X: x - r, Y: y - r, W: r * 2, H: r * 2}
			case KindTapeV, KindTapeH:
				bound = elementBounds(*e, c, screenW, screenH, x, y)
			default:
				tw, th := c.TextSize(text, fs)
				bound = native.Rect{X: x, Y: y - th, W: tw, H: th + 4}
			}
			c.Glow(bound, glowColor(*e, v), e.GlowIntensity)
		}

		switch e.Kind {
		case KindHorizon:
			r := int(e.Size * float64(screenH))
			col := toNativeColor(e.Color)
			c.DrawArtificialHorizon(x, y, r, v.Pitch, v.Roll, horizonSky, horizonGround, col, col)
		case KindTapeV:
			length := int(e.Length * float64(screenH))
			drawTapeV(c, x, y, length, bindingValue(e.Binding, v), *e, toNativeColor(e.Color))
		case KindTapeH:
			length := int(e.Length * float64(screenW))
			drawTapeH(c, x, y, length, bindingValue(e.Binding, v), *e, toNativeColor(e.Color))
		default:
			if e.Bold {
				c.TextBold(x, y, toNativeColor(e.Color), fs, text)
			} else {
				c.Text(x, y, toNativeColor(e.Color), fs, text)
			}
		}

		if editMode {
			b := elementBounds(*e, c, screenW, screenH, x, y)
			borderCol := native.Color{R: 255, G: 255, B: 255, A: 130}
			if edit.Dragging == e.ID {
				borderCol = native.Color{R: 255, G: 210, B: 60, A: 230}
			}
			c.StrokeRect(b, borderCol, 1)

			if !readOnly && b.Contains(in.MouseX, in.MouseY) && in.Pressed && edit.Dragging == "" {
				edit.Dragging = e.ID
				edit.OffsetX = in.MouseX - x
				edit.OffsetY = in.MouseY - y
				edit.StartX = x
				edit.StartY = y
				if edit.Selected != e.ID {
					edit.Selected = e.ID
					edit.FocusField = ""
					edit.OpenDropdown = ""
				}
			}
		}
	}

	if editMode && edit.Dragging != "" {
		if in.MouseDown {
			rawX := in.MouseX - edit.OffsetX
			rawY := in.MouseY - edit.OffsetY

			if in.KeyMods&native.ModShift != 0 {
				dx := rawX - edit.StartX
				dy := rawY - edit.StartY
				if abs(dx) > abs(dy) {
					rawY = edit.StartY
				} else {
					rawX = edit.StartX
				}
			}

			guideX, haveGuideX := 0, false
			guideY, haveGuideY := 0, false

			if in.KeyMods&native.ModAlt == 0 {
				var dragged *Element
				for i := range tmpl.Elements {
					if tmpl.Elements[i].ID == edit.Dragging {
						dragged = &tmpl.Elements[i]
						break
					}
				}

				xs := []int{screenW / 2}
				ys := []int{screenH / 2}

				var leftOff, rightOff, topOff, bottomOff int
				if dragged != nil {
					anchorX0 := int(dragged.X * float64(screenW))
					anchorY0 := int(dragged.Y * float64(screenH))
					db := elementBounds(*dragged, c, screenW, screenH, anchorX0, anchorY0)
					leftOff = db.X - anchorX0
					rightOff = db.X + db.W - anchorX0
					topOff = db.Y - anchorY0
					bottomOff = db.Y + db.H - anchorY0
				}

				var others []*Element
				for i := range tmpl.Elements {
					o := &tmpl.Elements[i]
					if o.ID == edit.Dragging {
						continue
					}
					others = append(others, o)

					ox := int(o.X * float64(screenW))
					oy := int(o.Y * float64(screenH))
					xs = append(xs, ox)
					ys = append(ys, oy)

					xs = append(xs, screenW-ox)
					ys = append(ys, screenH-oy)

					if dragged != nil {
						ob := elementBounds(*o, c, screenW, screenH, ox, oy)
						oLeft, oRight := ob.X, ob.X+ob.W
						oTop, oBottom := ob.Y, ob.Y+ob.H
						xs = append(xs, oLeft-leftOff, oLeft-rightOff, oRight-leftOff, oRight-rightOff)
						ys = append(ys, oTop-topOff, oTop-bottomOff, oBottom-topOff, oBottom-bottomOff)
					}
				}

				if dragged != nil {
					for i := 0; i < len(others); i++ {
						for j := i + 1; j < len(others); j++ {
							ai, bj := others[i], others[j]
							ab := elementBounds(*ai, c, screenW, screenH, int(ai.X*float64(screenW)), int(ai.Y*float64(screenH)))
							bb := elementBounds(*bj, c, screenW, screenH, int(bj.X*float64(screenW)), int(bj.Y*float64(screenH)))
							if ab.X > bb.X {
								ab, bb = bb, ab
							}
							if ab.X+ab.W < bb.X {
								px := (bb.X + ab.X + ab.W - leftOff - rightOff) / 2
								xs = append(xs, px)
							}
							if ab.Y > bb.Y {
								ab, bb = bb, ab
							}
							if ab.Y+ab.H < bb.Y {
								py := (bb.Y + ab.Y + ab.H - topOff - bottomOff) / 2
								ys = append(ys, py)
							}
						}
					}
				}

				if sx, ok := nearestSnap(rawX, xs); ok {
					rawX = sx
					guideX, haveGuideX = sx, true
				}
				if sy, ok := nearestSnap(rawY, ys); ok {
					rawY = sy
					guideY, haveGuideY = sy, true
				}
			}

			for i := range tmpl.Elements {
				e := &tmpl.Elements[i]
				if e.ID != edit.Dragging {
					continue
				}
				e.X = clamp01(float64(rawX) / float64(screenW))
				e.Y = clamp01(float64(rawY) / float64(screenH))
			}

			if haveGuideX {
				c.Line([]native.Point{{X: guideX, Y: 0}, {X: guideX, Y: screenH}}, snapGuideColor, 1)
			}
			if haveGuideY {
				c.Line([]native.Point{{X: 0, Y: guideY}, {X: screenW, Y: guideY}}, snapGuideColor, 1)
			}
		} else {
			edit.Dragging = ""
		}
	}

	if editMode && in.Pressed && edit.Dragging == "" && edit.Selected != "" && in.MouseX < screenW-PanelWidth {
		edit.Selected = ""
		edit.FocusField = ""
		edit.OpenDropdown = ""
	}
}

func glowColor(e Element, v Values) native.Color {
	if !e.GlowUseOwn {
		return toNativeColor(e.GlowColor)
	}

	switch e.Kind {
	case KindTapeV, KindTapeH:
		return zoneColor(e.Zones, bindingValue(e.Binding, v), toNativeColor(e.Color))
	default:
		return toNativeColor(e.Color)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}

func nearestSnap(target int, candidates []int) (int, bool) {
	best := 0
	bestDist := snapThreshold + 1
	found := false
	for _, cand := range candidates {
		d := abs(target - cand)
		if d < bestDist {
			bestDist = d
			best = cand
			found = true
		}
	}

	return best, found
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}

	return v
}
