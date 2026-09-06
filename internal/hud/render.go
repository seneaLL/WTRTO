package hud

import (
	"fmt"

	"github.com/seneaLL/WTRTO/internal/native"
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

	PanelTab string

	ColorPickerFor  string
	ColorPickerZone int

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

var (
	speedAutoGreen  = native.Color{R: 80, G: 255, B: 90, A: 255}
	speedAutoYellow = native.Color{R: 255, G: 220, B: 60, A: 255}
	speedAutoRed    = native.Color{R: 255, G: 70, B: 60, A: 255}
)

const (
	speedWarnStart = 0.85
	speedWarnFull  = 0.95
)

func ratioColor(r float64) native.Color {
	switch {
	case r >= speedWarnFull:
		return speedAutoRed
	case r >= speedWarnStart:
		t := (r - speedWarnStart) / (speedWarnFull - speedWarnStart)

		return lerpColor(speedAutoYellow, speedAutoRed, t)
	default:
		return speedAutoGreen
	}
}

var AutoColorBindings = []Binding{BindIAS, BindMach, BindGLoad}

func supportsAutoColor(b Binding) bool {
	for _, ab := range AutoColorBindings {
		if ab == b {
			return true
		}
	}

	return false
}

type limitRef struct {
	known  bool
	posMax float64
	negMax float64
}

func limitFor(b Binding, v Values) limitRef {
	switch b {
	case BindIAS:
		return limitRef{known: v.SpeedLimitKnown, posMax: v.SpeedLimitMaxKmh}
	case BindMach:
		return limitRef{known: v.MachLimitKnown, posMax: v.MachLimitMax}
	case BindGLoad:
		return limitRef{known: v.GLoadLimitKnown, posMax: v.GLoadLimitPos, negMax: v.GLoadLimitNeg}
	}

	return limitRef{}
}

func (lr limitRef) ratio(value float64) (float64, bool) {
	if !lr.known {
		return 0, false
	}
	if value >= 0 {
		if lr.posMax <= 0 {
			return 0, false
		}

		return value / lr.posMax, true
	}
	if lr.negMax >= 0 {
		return 0, false
	}

	return value / lr.negMax, true
}

func autoColorForBinding(b Binding, v Values, value float64) native.Color {
	if b == BindIAS && v.SpeedWarning {
		return speedAutoRed
	}

	r, ok := limitFor(b, v).ratio(value)
	if !ok {
		return speedAutoGreen
	}
	if r < 0 {
		r = -r
	}

	return ratioColor(r)
}

func elementColor(e Element, v Values) native.Color {
	if e.AutoColor && supportsAutoColor(e.Binding) {
		return autoColorForBinding(e.Binding, v, bindingValue(e.Binding, v))
	}

	return toNativeColor(e.Color)
}

func tapeZoneColor(e Element, v Values, value float64, fallback native.Color) native.Color {
	if e.AutoColor && supportsAutoColor(e.Binding) {
		return autoColorForBinding(e.Binding, v, value)
	}

	return zoneColor(e.Zones, value, fallback)
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

	case BindAileron:
		return v.Aileron
	case BindElevator:
		return v.Elevator
	case BindRudder:
		return v.Rudder
	case BindFlaps:
		return v.Flaps
	case BindGearPct:
		return v.GearPct
	case BindRollRate:
		return v.RollRate
	case BindFuelPct:
		return v.FuelPct
	case BindTrimmer:
		return v.Trimmer
	case BindRadioAlt:
		return v.RadioAlt
	case BindTurn:
		return v.Turn
	case BindWingSweep:
		return v.WingSweep

	case BindThrottle1:
		return v.EngineThrottle[0]
	case BindThrottle2:
		return v.EngineThrottle[1]
	case BindThrottle3:
		return v.EngineThrottle[2]
	case BindThrottle4:
		return v.EngineThrottle[3]

	case BindRPM1:
		return v.EngineRPM[0]
	case BindRPM2:
		return v.EngineRPM[1]
	case BindRPM3:
		return v.EngineRPM[2]
	case BindRPM4:
		return v.EngineRPM[3]

	case BindManifold1:
		return v.EngineManifold[0]
	case BindManifold2:
		return v.EngineManifold[1]
	case BindManifold3:
		return v.EngineManifold[2]
	case BindManifold4:
		return v.EngineManifold[3]

	case BindOilTemp3:
		return v.EngineOilTemp[2]
	case BindOilTemp4:
		return v.EngineOilTemp[3]

	case BindWaterTemp1:
		return v.EngineWaterTemp[0]
	case BindWaterTemp2:
		return v.EngineWaterTemp[1]
	case BindWaterTemp3:
		return v.EngineWaterTemp[2]
	case BindWaterTemp4:
		return v.EngineWaterTemp[3]

	case BindPower1:
		return v.EnginePower[0]
	case BindPower2:
		return v.EnginePower[1]
	case BindPower3:
		return v.EnginePower[2]
	case BindPower4:
		return v.EnginePower[3]

	case BindThrust1:
		return v.EngineThrust[0]
	case BindThrust2:
		return v.EngineThrust[1]
	case BindThrust3:
		return v.EngineThrust[2]
	case BindThrust4:
		return v.EngineThrust[3]

	case BindEfficiency1:
		return v.EngineEfficiency[0]
	case BindEfficiency2:
		return v.EngineEfficiency[1]
	case BindEfficiency3:
		return v.EngineEfficiency[2]
	case BindEfficiency4:
		return v.EngineEfficiency[3]

	case BindPropPitch1:
		return v.EnginePropPitch[0]
	case BindPropPitch2:
		return v.EnginePropPitch[1]
	case BindPropPitch3:
		return v.EnginePropPitch[2]
	case BindPropPitch4:
		return v.EnginePropPitch[3]
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
		length := int(e.Length * float64(screenH))
		if e.Style == StyleArc {
			r := length / 2

			return native.Rect{X: x - r - 40, Y: y - r, W: r*2 + 40, H: r * 2}
		}

		fs := e.FontSize
		if fs == 0 {
			fs = 13
		}
		bw, _ := c.TextSize("9999", fs+2)
		protrude := bw + 16 + 20
		half := length / 2
		if tapeVSign(e) > 0 {
			return native.Rect{X: x - 4, Y: y - half, W: protrude, H: length}
		}

		return native.Rect{X: x - protrude, Y: y - half, W: protrude, H: length}
	case KindTapeH:
		length := int(e.Length * float64(screenW))
		if e.Style == StyleArc {
			r := length / 2

			return native.Rect{X: x - r, Y: y - r - 40, W: r * 2, H: r*2 + 40}
		}

		fs := e.FontSize
		if fs == 0 {
			fs = 13
		}
		_, bh := c.TextSize("9999", fs+2)
		above := bh + 22 + 10
		half := length / 2

		return native.Rect{X: x - half, Y: y - above, W: length, H: above + 16}
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

	elemBounds := make([]native.Rect, len(tmpl.Elements))
	elemPos := make([][2]int, len(tmpl.Elements))

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
			drawTapeV(c, x, y, length, bindingValue(e.Binding, v), *e, v, toNativeColor(e.Color))
		case KindTapeH:
			length := int(e.Length * float64(screenW))
			drawTapeH(c, x, y, length, bindingValue(e.Binding, v), *e, v, toNativeColor(e.Color))
		default:
			col := elementColor(*e, v)
			if e.Bold {
				c.TextBold(x, y, col, fs, text)
			} else {
				c.Text(x, y, col, fs, text)
			}
		}

		if editMode {
			b := elementBounds(*e, c, screenW, screenH, x, y)
			elemBounds[i] = b
			elemPos[i] = [2]int{x, y}

			borderCol := native.Color{R: 255, G: 255, B: 255, A: 130}
			borderWidth := 1
			if edit.Selected == e.ID {
				borderCol = native.Color{R: 90, G: 180, B: 255, A: 220}
				borderWidth = 2
			}
			if edit.Dragging == e.ID {
				borderCol = native.Color{R: 255, G: 210, B: 60, A: 230}
				borderWidth = 2
			}
			c.StrokeRect(b, borderCol, borderWidth)
		}
	}

	if editMode && !readOnly && in.Pressed && edit.Dragging == "" {
		var hits []int
		for i := range tmpl.Elements {
			if elemBounds[i].Contains(in.MouseX, in.MouseY) {
				hits = append(hits, i)
			}
		}

		if len(hits) > 0 {
			pick := hits[len(hits)-1]
			if in.KeyMods&native.ModCtrl != 0 && edit.Selected != "" {
				for pos := len(hits) - 1; pos >= 0; pos-- {
					if tmpl.Elements[hits[pos]].ID == edit.Selected {
						next := pos - 1
						if next < 0 {
							next = len(hits) - 1
						}
						pick = hits[next]

						break
					}
				}
			}

			e := &tmpl.Elements[pick]
			edit.Dragging = e.ID
			edit.OffsetX = in.MouseX - elemPos[pick][0]
			edit.OffsetY = in.MouseY - elemPos[pick][1]
			edit.StartX = elemPos[pick][0]
			edit.StartY = elemPos[pick][1]
			if edit.Selected != e.ID {
				edit.Selected = e.ID
				edit.FocusField = ""
				edit.OpenDropdown = ""
			}
			edit.PanelTab = "element"
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
				c.Line([]native.Point{{X: float64(guideX), Y: 0}, {X: float64(guideX), Y: float64(screenH)}}, snapGuideColor, 1)
			}
			if haveGuideY {
				c.Line([]native.Point{{X: 0, Y: float64(guideY)}, {X: float64(screenW), Y: float64(guideY)}}, snapGuideColor, 1)
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
		return tapeZoneColor(e, v, bindingValue(e.Binding, v), toNativeColor(e.Color))
	default:
		return elementColor(e, v)
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
