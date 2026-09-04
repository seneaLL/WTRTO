package native

func (c *Canvas) Glow(r Rect, col Color, intensity float64) {
	if intensity <= 0 {
		return
	}
	if intensity > 1 {
		intensity = 1
	}

	const layers = 5
	maxExpand := 6 + int(26*intensity)

	for i := layers; i >= 1; i-- {
		t := float64(i) / float64(layers)
		expand := int(float64(maxExpand)*t + 0.5)
		fall := (1 - t) * (1 - t)

		a := int(float64(col.A) * intensity * fall * 0.9)
		if a <= 0 {
			continue
		}

		gr := Rect{X: r.X - expand, Y: r.Y - expand, W: r.W + 2*expand, H: r.H + 2*expand}
		radius := gr.W / 2
		if gr.H/2 < radius {
			radius = gr.H / 2
		}

		c.FillRoundedRect(gr, radius, Color{R: col.R, G: col.G, B: col.B, A: uint8(a)})
	}
}
