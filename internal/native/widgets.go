package native

import (
	"strconv"

	"github.com/seneaLL/WTRTO/internal/clipboard"
)

const (
	RadiusSmall  = 6
	RadiusMedium = 10
	RadiusLarge  = 14
)

func fitText(c *Canvas, s string, fontSize, maxWidth int) (string, int, int) {
	w, h := c.TextSize(s, fontSize)
	if maxWidth <= 0 || w <= maxWidth {
		return s, w, h
	}

	const ellipsis = "…"
	ew, _ := c.TextSize(ellipsis, fontSize)

	runes := []rune(s)
	tw := w
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		if tw, _ = c.TextSize(string(runes), fontSize); tw+ew <= maxWidth {
			break
		}
	}

	if len(runes) == 0 {
		return ellipsis, ew, h
	}

	return string(runes) + ellipsis, tw + ew, h
}

func Button(c *Canvas, in *Input, r Rect, label string, bg, bgHover, textCol Color, fontSize int) bool {
	hover := r.Contains(in.MouseX, in.MouseY)
	col := bg
	if hover {
		col = bgHover
	}
	c.FillRoundedRect(r, RadiusSmall, col)

	label, tw, th := fitText(c, label, fontSize, r.W-12)
	c.Text(r.X+(r.W-tw)/2, r.Y+(r.H+th)/2-2, textCol, fontSize, label)

	return hover && in.Released
}

func Checkbox(c *Canvas, in *Input, r Rect, checked bool, boxCol, checkCol, textCol Color, fontSize int, label string) bool {
	hover := r.Contains(in.MouseX, in.MouseY)
	box := Rect{X: r.X, Y: r.Y, W: r.H, H: r.H}
	c.StrokeRoundedRect(box, RadiusSmall-2, boxCol, 1)
	if checked {
		inset := Rect{X: box.X + 4, Y: box.Y + 4, W: box.W - 8, H: box.H - 8}
		c.FillRoundedRect(inset, RadiusSmall-4, checkCol)
	}
	_, th := c.TextSize(label, fontSize)
	c.Text(box.X+box.W+12, box.Y+(box.H+th)/2-2, textCol, fontSize, label)

	return hover && in.Released
}

func CycleButton(c *Canvas, in *Input, r Rect, options []string, idx int, bg, bgHover, textCol Color, fontSize int) int {
	if len(options) == 0 {
		return idx
	}
	idx = ((idx % len(options)) + len(options)) % len(options)
	if Button(c, in, r, options[idx], bg, bgHover, textCol, fontSize) {
		idx = (idx + 1) % len(options)
	}

	return idx
}

func StepMultiplier(in *Input) float64 {
	mult := 1.0
	if in.KeyMods&ModShift != 0 {
		mult *= 10
	}
	if in.KeyMods&ModCtrl != 0 {
		mult *= 100
	}

	return mult
}

func applyTextEditKeys(in *Input, value *string, cursor *int, selectedAll *bool) (submitted bool, clipErr error) {
	runes := []rune(*value)

	if *cursor > len(runes) {
		*cursor = len(runes)
	}

	if *cursor < 0 {
		*cursor = 0
	}

	switch in.KeyCtrlLetter {
	case 'a':
		*selectedAll = true

		return false, nil

	case 'c':
		if len(runes) > 0 {
			clipErr = clipboard.Copy(*value)
		}

		return false, clipErr

	case 'v':
		txt, err := clipboard.Paste()
		if err != nil {
			return false, err
		}

		pasted := []rune(txt)
		if *selectedAll {
			*value = txt
			*cursor = len(pasted)
		} else {
			merged := make([]rune, 0, len(runes)+len(pasted))
			merged = append(merged, runes[:*cursor]...)
			merged = append(merged, pasted...)
			merged = append(merged, runes[*cursor:]...)
			*value = string(merged)
			*cursor += len(pasted)
		}

		*selectedAll = false

		return false, nil
	}

	switch {
	case in.KeyLeft:
		if *cursor > 0 {
			*cursor--
		}

		*selectedAll = false

	case in.KeyRight:
		if *cursor < len(runes) {
			*cursor++
		}

		*selectedAll = false

	case in.KeyBackspace:
		if *selectedAll {
			*value = ""
			*cursor = 0
		} else if *cursor > 0 {
			merged := make([]rune, 0, len(runes)-1)
			merged = append(merged, runes[:*cursor-1]...)
			merged = append(merged, runes[*cursor:]...)
			*value = string(merged)
			*cursor--
		}

		*selectedAll = false

	case in.KeyRune != 0:
		if *selectedAll {
			*value = string(in.KeyRune)
			*cursor = 1
		} else {
			merged := make([]rune, 0, len(runes)+1)
			merged = append(merged, runes[:*cursor]...)
			merged = append(merged, in.KeyRune)
			merged = append(merged, runes[*cursor:]...)
			*value = string(merged)
			*cursor++
		}

		*selectedAll = false
	}

	return in.KeyEnter, nil
}

func drawSelectionHighlight(c *Canvas, r Rect, textW int) {
	hlW := textW + 4
	if maxW := r.W - 4; hlW > maxW {
		hlW = maxW
	}

	c.FillRect(Rect{X: r.X + 2, Y: r.Y + 3, W: hlW, H: r.H - 6}, Color{R: 70, G: 120, B: 210, A: 130})
}

func NumberStepper(c *Canvas, in *Input, r Rect, value, step float64, precision int, bg, bgHover, textCol, focusCol Color, fontSize int, editing bool, editBuf string, cursor *int, selectedAll *bool) (newValue float64, buf string, clicked bool, submitted bool, stepped bool, clipErr error) {
	btnW := r.H
	minusRect := Rect{X: r.X, Y: r.Y, W: btnW, H: r.H}
	plusRect := Rect{X: r.X + r.W - btnW, Y: r.Y, W: btnW, H: r.H}
	midRect := Rect{X: r.X + btnW, Y: r.Y, W: r.W - 2*btnW, H: r.H}

	newValue = value
	mult := StepMultiplier(in)

	if Button(c, in, minusRect, "-", bg, bgHover, textCol, fontSize) {
		newValue -= step * mult
		stepped = true
	}

	if Button(c, in, plusRect, "+", bg, bgHover, textCol, fontSize) {
		newValue += step * mult
		stepped = true
	}

	midHover := midRect.Contains(in.MouseX, in.MouseY)
	midCol := bg
	if midHover && !editing {
		midCol = bgHover
	}
	c.FillRect(midRect, midCol)
	borderCol := textCol
	if editing {
		borderCol = focusCol
	}
	c.StrokeRect(midRect, borderCol, 1)

	buf = editBuf
	c.ClipRect(Rect{X: midRect.X + 2, Y: midRect.Y, W: midRect.W - 4, H: midRect.H})
	if editing && !stepped {
		submitted, clipErr = applyTextEditKeys(in, &buf, cursor, selectedAll)

		runes := []rune(buf)
		tw, th := c.TextSize(buf, fontSize)
		tx := midRect.X + (midRect.W-tw)/2

		if *selectedAll && buf != "" {
			drawSelectionHighlight(c, Rect{X: tx, Y: midRect.Y, W: tw, H: midRect.H}, tw)
		}

		c.Text(tx, midRect.Y+(midRect.H+th)/2-2, textCol, fontSize, buf)

		if !*selectedAll {
			cw, _ := c.TextSize(string(runes[:*cursor]), fontSize)
			cx := tx + cw
			c.Line([]Point{{X: cx, Y: midRect.Y + 4}, {X: cx, Y: midRect.Y + midRect.H - 4}}, focusCol, 2)
		}
	} else {
		label := strconv.FormatFloat(value, 'f', precision, 64)
		tw, th := c.TextSize(label, fontSize)
		c.Text(midRect.X+(midRect.W-tw)/2, midRect.Y+(midRect.H+th)/2-2, textCol, fontSize, label)
	}
	c.Unclip()

	clicked = midHover && in.Released && !editing && !stepped

	return newValue, buf, clicked, submitted, stepped, clipErr
}

func TextInput(c *Canvas, in *Input, r Rect, value string, focused bool, cursor *int, selectedAll *bool, borderCol, focusCol, bgCol, textCol Color, fontSize int) (newValue string, submitted bool, clicked bool, clipErr error) {
	hover := r.Contains(in.MouseX, in.MouseY)
	col := borderCol
	if focused {
		col = focusCol
	}
	c.FillRoundedRect(r, RadiusSmall-2, bgCol)
	c.StrokeRoundedRect(r, RadiusSmall-2, col, 1)

	newValue = value

	if focused {
		submitted, clipErr = applyTextEditKeys(in, &newValue, cursor, selectedAll)
	}

	runes := []rune(newValue)
	if *cursor > len(runes) {
		*cursor = len(runes)
	}

	if *cursor < 0 {
		*cursor = 0
	}

	tw, th := c.TextSize(newValue, fontSize)
	avail := r.W - 16
	cursorW, _ := c.TextSize(string(runes[:*cursor]), fontSize)
	tx := r.X + 8

	if tw > avail {
		scroll := cursorW - avail
		if scroll < 0 {
			scroll = 0
		}

		if maxScroll := tw - avail; scroll > maxScroll {
			scroll = maxScroll
		}

		tx = r.X + 8 - scroll
	}

	clicked = hover && in.Released
	if clicked {
		*cursor = cursorIndexForClick(c, newValue, fontSize, in.MouseX-tx)
		*selectedAll = false

		runes = []rune(newValue)
		cursorW, _ = c.TextSize(string(runes[:*cursor]), fontSize)
	}

	if focused && *selectedAll && newValue != "" {
		drawSelectionHighlight(c, r, tw)
	}

	c.ClipRect(Rect{X: r.X + 2, Y: r.Y, W: r.W - 4, H: r.H})
	c.Text(tx, r.Y+(r.H+th)/2-2, textCol, fontSize, newValue)

	if focused && !*selectedAll {
		cx := tx + cursorW
		c.Line([]Point{{X: cx, Y: r.Y + 4}, {X: cx, Y: r.Y + r.H - 4}}, focusCol, 2)
	}

	c.Unclip()

	return newValue, submitted, clicked, clipErr
}

func cursorIndexForClick(c *Canvas, value string, fontSize int, offsetX int) int {
	runes := []rune(value)
	accum := 0

	for i, rn := range runes {
		w, _ := c.TextSize(string(rn), fontSize)
		if accum+w/2 > offsetX {
			return i
		}

		accum += w
	}

	return len(runes)
}

func ColorSliders(c *Canvas, in *Input, r Rect, col Color) Color {
	rowH := r.H / 4
	col.R = uint8(colorSlider(c, in, Rect{X: r.X, Y: r.Y, W: r.W, H: rowH - 2}, float64(col.R), Color{R: 220, G: 80, B: 80, A: 255}))
	col.G = uint8(colorSlider(c, in, Rect{X: r.X, Y: r.Y + rowH, W: r.W, H: rowH - 2}, float64(col.G), Color{R: 80, G: 200, B: 100, A: 255}))
	col.B = uint8(colorSlider(c, in, Rect{X: r.X, Y: r.Y + 2*rowH, W: r.W, H: rowH - 2}, float64(col.B), Color{R: 90, G: 140, B: 230, A: 255}))
	col.A = uint8(colorSlider(c, in, Rect{X: r.X, Y: r.Y + 3*rowH, W: r.W, H: rowH - 2}, float64(col.A), Color{R: 200, G: 200, B: 200, A: 255}))

	return col
}

func colorSlider(c *Canvas, in *Input, r Rect, value float64, trackCol Color) float64 {
	c.FillRoundedRect(r, 3, Color{R: 40, G: 42, B: 48, A: 255})
	if value < 0 {
		value = 0
	}
	if value > 255 {
		value = 255
	}
	frac := value / 255
	handleX := r.X + int(frac*float64(r.W))
	if handleX > r.X {
		c.FillRect(Rect{X: r.X, Y: r.Y, W: handleX - r.X, H: r.H}, trackCol)
	}
	if in.MouseDown && r.Contains(in.MouseX, in.MouseY) {
		nf := float64(in.MouseX-r.X) / float64(r.W)
		if nf < 0 {
			nf = 0
		}
		if nf > 1 {
			nf = 1
		}
		value = nf * 255
	}

	return value
}

const SelectRowHeight = 26

func SelectBox(c *Canvas, in *Input, r Rect, current string, open bool, bg, bgHover, textCol Color, fontSize int) bool {
	hover := r.Contains(in.MouseX, in.MouseY)
	col := bg
	if hover || open {
		col = bgHover
	}
	c.FillRoundedRect(r, RadiusSmall, col)

	arrow := "v"
	if open {
		arrow = "^"
	}
	aw, th := c.TextSize(arrow, fontSize)

	label, _, _ := fitText(c, current, fontSize, r.W-20-aw-10)
	c.Text(r.X+10, r.Y+(r.H+th)/2-2, textCol, fontSize, label)
	c.Text(r.X+r.W-aw-10, r.Y+(r.H+th)/2-2, textCol, fontSize, arrow)

	return hover && in.Released
}

const SelectMaxVisibleRows = 8

func SelectListBounds(headerRect Rect, optionCount, containerH int) Rect {
	rows := optionCount
	if rows > SelectMaxVisibleRows {
		rows = SelectMaxVisibleRows
	}
	h := rows * SelectRowHeight

	y := headerRect.Y + headerRect.H
	if containerH > 0 && y+h > containerH {
		if up := headerRect.Y - h; up >= 0 {
			y = up
		} else if fit := containerH - h; fit >= 0 {
			y = fit
		}
	}

	return Rect{X: headerRect.X, Y: y, W: headerRect.W, H: h}
}

func SelectList(c *Canvas, in *Input, headerRect Rect, options []string, current int, scroll *int, containerH int, bg, hoverBg, textCol, borderCol Color, fontSize int) (newIdx int, selected bool) {
	listRect := SelectListBounds(headerRect, len(options), containerH)

	maxScroll := len(options) - SelectMaxVisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}

	if listRect.Contains(in.MouseX, in.MouseY) && in.ScrollDelta != 0 {
		*scroll -= in.ScrollDelta
	}
	if *scroll < 0 {
		*scroll = 0
	}
	if *scroll > maxScroll {
		*scroll = maxScroll
	}

	c.FillRoundedRect(listRect, RadiusSmall, bg)
	c.StrokeRoundedRect(listRect, RadiusSmall, borderCol, 1)

	newIdx = current
	visible := len(options) - *scroll
	if visible > SelectMaxVisibleRows {
		visible = SelectMaxVisibleRows
	}

	c.ClipRect(listRect)
	for i := 0; i < visible; i++ {
		optIdx := i + *scroll
		rowRect := Rect{X: listRect.X, Y: listRect.Y + i*SelectRowHeight, W: listRect.W, H: SelectRowHeight}
		hover := rowRect.Contains(in.MouseX, in.MouseY)
		if hover {
			c.FillRect(rowRect, hoverBg)
		}
		if optIdx == current {
			c.FillRect(Rect{X: rowRect.X, Y: rowRect.Y, W: 3, H: rowRect.H}, textCol)
		}
		label, _, th := fitText(c, options[optIdx], fontSize, rowRect.W-20)
		c.Text(rowRect.X+10, rowRect.Y+(rowRect.H+th)/2-2, textCol, fontSize, label)
		if hover && in.Released {
			newIdx = optIdx
			selected = true
		}
	}
	c.Unclip()

	if maxScroll > 0 {
		trackH := listRect.H - 6
		thumbH := trackH * SelectMaxVisibleRows / len(options)
		if thumbH < 14 {
			thumbH = 14
		}
		thumbY := listRect.Y + 3 + (trackH-thumbH)*(*scroll)/maxScroll
		c.FillRoundedRect(Rect{X: listRect.X + listRect.W - 5, Y: thumbY, W: 3, H: thumbH}, 1, borderCol)
	}

	return newIdx, selected
}

func Sparkline(c *Canvas, r Rect, values []float32, col Color) {
	if len(values) < 2 {
		return
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	span := max - min
	if span == 0 {
		span = 1
	}

	pts := make([]Point, len(values))
	for i, v := range values {
		x := r.X + i*r.W/(len(values)-1)
		norm := (v - min) / span
		y := r.Y + r.H - int(norm*float32(r.H))
		pts[i] = Point{X: x, Y: y}
	}
	c.Line(pts, col, 2)
}
