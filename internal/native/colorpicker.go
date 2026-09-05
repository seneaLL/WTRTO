package native

import "math"

func hsvToRGB(h, s, v float64) (r, g, b float64) {
	hp := math.Mod(h, 360)
	if hp < 0 {
		hp += 360
	}
	ch := v * s
	x := ch * (1 - math.Abs(math.Mod(hp/60, 2)-1))
	m := v - ch

	var r1, g1, b1 float64
	switch {
	case hp < 60:
		r1, g1, b1 = ch, x, 0
	case hp < 120:
		r1, g1, b1 = x, ch, 0
	case hp < 180:
		r1, g1, b1 = 0, ch, x
	case hp < 240:
		r1, g1, b1 = 0, x, ch
	case hp < 300:
		r1, g1, b1 = x, 0, ch
	default:
		r1, g1, b1 = ch, 0, x
	}

	return r1 + m, g1 + m, b1 + m
}

func rgbToHSV(r, g, b float64) (h, s, v float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	v = max
	d := max - min

	if max == 0 {
		s = 0
	} else {
		s = d / max
	}

	if d == 0 {
		h = 0
	} else {
		switch max {
		case r:
			h = 60 * math.Mod((g-b)/d, 6)
		case g:
			h = 60 * ((b-r)/d + 2)
		default:
			h = 60 * ((r-g)/d + 4)
		}
		if h < 0 {
			h += 360
		}
	}

	return h, s, v
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

const (
	colorPickerHueH   = 18
	colorPickerAlphaH = 18
	colorPickerGap    = 8
)

func ColorPickerHeight(w int) int {
	svH := w * 3 / 5
	if svH < 90 {
		svH = 90
	}

	return svH + colorPickerGap + colorPickerHueH + colorPickerGap + colorPickerAlphaH
}

func ColorPicker(c *Canvas, in *Input, r Rect, col Color) Color {
	svH := r.H - colorPickerHueH - colorPickerAlphaH - 2*colorPickerGap
	if svH < 40 {
		svH = 40
	}

	svRect := Rect{X: r.X, Y: r.Y, W: r.W, H: svH}
	hueRect := Rect{X: r.X, Y: svRect.Y + svH + colorPickerGap, W: r.W, H: colorPickerHueH}
	alphaRect := Rect{X: r.X, Y: hueRect.Y + colorPickerHueH + colorPickerGap, W: r.W, H: colorPickerAlphaH}

	h, s, v := rgbToHSV(float64(col.R)/255, float64(col.G)/255, float64(col.B)/255)
	alpha := float64(col.A) / 255

	const svStrips = 64
	for i := 0; i < svStrips; i++ {
		fs := float64(i) / float64(svStrips-1)
		rr, gg, bb := hsvToRGB(h, fs, 1)
		x0 := svRect.X + i*svRect.W/svStrips
		w := (i+1)*svRect.W/svStrips - i*svRect.W/svStrips
		c.FillRect(Rect{X: x0, Y: svRect.Y, W: w, H: svRect.H}, Color{R: uint8(rr*255 + 0.5), G: uint8(gg*255 + 0.5), B: uint8(bb*255 + 0.5), A: 255})
	}
	for j := 0; j < svStrips; j++ {
		fv := 1 - float64(j)/float64(svStrips-1)
		y0 := svRect.Y + j*svRect.H/svStrips
		hgt := (j+1)*svRect.H/svStrips - j*svRect.H/svStrips
		alpha := uint8((1-fv)*255 + 0.5)
		c.FillRect(Rect{X: svRect.X, Y: y0, W: svRect.W, H: hgt}, Color{R: 0, G: 0, B: 0, A: alpha})
	}
	if in.MouseDown && svRect.Contains(in.MouseX, in.MouseY) {
		s = clamp01(float64(in.MouseX-svRect.X) / float64(svRect.W))
		v = 1 - clamp01(float64(in.MouseY-svRect.Y)/float64(svRect.H))
	}
	hx := svRect.X + int(s*float64(svRect.W))
	hy := svRect.Y + int((1-v)*float64(svRect.H))
	c.FillCircle(hx, hy, 7, Color{R: 255, G: 255, B: 255, A: 255})
	rNow, gNow, bNow := hsvToRGB(h, s, v)
	c.FillCircle(hx, hy, 5, Color{R: uint8(rNow*255 + 0.5), G: uint8(gNow*255 + 0.5), B: uint8(bNow*255 + 0.5), A: 255})

	const hueStrips = 72
	for i := 0; i < hueStrips; i++ {
		hh := float64(i) / float64(hueStrips) * 360
		rr, gg, bb := hsvToRGB(hh, 1, 1)
		x0 := hueRect.X + i*hueRect.W/hueStrips
		w := (i+1)*hueRect.W/hueStrips - i*hueRect.W/hueStrips
		c.FillRect(Rect{X: x0, Y: hueRect.Y, W: w, H: hueRect.H}, Color{R: uint8(rr*255 + 0.5), G: uint8(gg*255 + 0.5), B: uint8(bb*255 + 0.5), A: 255})
	}
	if in.MouseDown && hueRect.Contains(in.MouseX, in.MouseY) {
		h = clamp01(float64(in.MouseX-hueRect.X)/float64(hueRect.W)) * 360
	}
	hueX := hueRect.X + int(h/360*float64(hueRect.W))
	c.Line([]Point{{X: float64(hueX), Y: float64(hueRect.Y - 2)}, {X: float64(hueX), Y: float64(hueRect.Y + hueRect.H + 2)}}, Color{R: 255, G: 255, B: 255, A: 255}, 2)

	rf, gf, bf := hsvToRGB(h, s, v)
	R, G, B := rf*255, gf*255, bf*255
	const alphaStrips = 48
	const alphaBg = 50.0
	for i := 0; i < alphaStrips; i++ {
		a := float64(i) / float64(alphaStrips-1)
		x0 := alphaRect.X + i*alphaRect.W/alphaStrips
		w := (i+1)*alphaRect.W/alphaStrips - i*alphaRect.W/alphaStrips
		rr := R*a + alphaBg*(1-a)
		gg := G*a + alphaBg*(1-a)
		bb := B*a + alphaBg*(1-a)
		c.FillRect(Rect{X: x0, Y: alphaRect.Y, W: w, H: alphaRect.H}, Color{R: uint8(rr + 0.5), G: uint8(gg + 0.5), B: uint8(bb + 0.5), A: 255})
	}
	if in.MouseDown && alphaRect.Contains(in.MouseX, in.MouseY) {
		alpha = clamp01(float64(in.MouseX-alphaRect.X) / float64(alphaRect.W))
	}
	alphaX := alphaRect.X + int(alpha*float64(alphaRect.W))
	c.Line([]Point{{X: float64(alphaX), Y: float64(alphaRect.Y - 2)}, {X: float64(alphaX), Y: float64(alphaRect.Y + alphaRect.H + 2)}}, Color{R: 255, G: 255, B: 255, A: 255}, 2)

	return Color{R: uint8(rf*255 + 0.5), G: uint8(gf*255 + 0.5), B: uint8(bf*255 + 0.5), A: uint8(alpha*255 + 0.5)}
}
