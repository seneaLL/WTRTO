package native

import (
	"strconv"

	"github.com/seneaLL/WTRTO/internal/clipboard"
	"github.com/seneaLL/WTRTO/internal/native/icons"
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

	label, _, _ = fitText(c, label, fontSize, r.W-12)
	c.TextCentered(r, textCol, fontSize, label)

	return hover && in.Released
}

var vectorIcons = map[string]bool{
	"check": true, "arrow-left": true, "arrow-right": true,
	"arrow-up": true, "arrow-down": true, "center-h": true, "center-v": true,
}

func drawIcon(c *Canvas, r Rect, kind string, col Color, thickness int) {
	cx, cy := float64(r.X)+float64(r.W)/2, float64(r.Y)+float64(r.H)/2
	s := float64(r.W)
	if r.H < r.W {
		s = float64(r.H)
	}
	s *= 0.5

	switch kind {
	case "check":
		c.Line([]Point{
			{X: cx - s*0.5, Y: cy},
			{X: cx - s*0.15, Y: cy + s*0.4},
			{X: cx + s*0.5, Y: cy - s*0.4},
		}, col, thickness)
	case "arrow-right":
		c.Line([]Point{{X: cx - s*0.5, Y: cy}, {X: cx + s*0.15, Y: cy}}, col, thickness)
		c.Line([]Point{
			{X: cx - s*0.15, Y: cy - s*0.35},
			{X: cx + s*0.5, Y: cy},
			{X: cx - s*0.15, Y: cy + s*0.35},
		}, col, thickness)
	case "arrow-left":
		c.Line([]Point{{X: cx + s*0.5, Y: cy}, {X: cx - s*0.15, Y: cy}}, col, thickness)
		c.Line([]Point{
			{X: cx + s*0.15, Y: cy - s*0.35},
			{X: cx - s*0.5, Y: cy},
			{X: cx + s*0.15, Y: cy + s*0.35},
		}, col, thickness)
	case "arrow-up":
		c.Line([]Point{{X: cx, Y: cy + s*0.5}, {X: cx, Y: cy - s*0.15}}, col, thickness)
		c.Line([]Point{
			{X: cx - s*0.35, Y: cy + s*0.15},
			{X: cx, Y: cy - s*0.5},
			{X: cx + s*0.35, Y: cy + s*0.15},
		}, col, thickness)
	case "arrow-down":
		c.Line([]Point{{X: cx, Y: cy - s*0.5}, {X: cx, Y: cy + s*0.15}}, col, thickness)
		c.Line([]Point{
			{X: cx - s*0.35, Y: cy - s*0.15},
			{X: cx, Y: cy + s*0.5},
			{X: cx + s*0.35, Y: cy - s*0.15},
		}, col, thickness)
	case "center-h":
		c.Line([]Point{{X: cx - s*0.2, Y: cy}, {X: cx + s*0.2, Y: cy}}, col, thickness)
		c.Line([]Point{
			{X: cx - s*0.2, Y: cy - s*0.25},
			{X: cx - s*0.5, Y: cy},
			{X: cx - s*0.2, Y: cy + s*0.25},
		}, col, thickness)
		c.Line([]Point{
			{X: cx + s*0.2, Y: cy - s*0.25},
			{X: cx + s*0.5, Y: cy},
			{X: cx + s*0.2, Y: cy + s*0.25},
		}, col, thickness)
	case "center-v":
		c.Line([]Point{{X: cx, Y: cy - s*0.2}, {X: cx, Y: cy + s*0.2}}, col, thickness)
		c.Line([]Point{
			{X: cx - s*0.25, Y: cy - s*0.2},
			{X: cx, Y: cy - s*0.5},
			{X: cx + s*0.25, Y: cy - s*0.2},
		}, col, thickness)
		c.Line([]Point{
			{X: cx - s*0.25, Y: cy + s*0.2},
			{X: cx, Y: cy + s*0.5},
			{X: cx + s*0.25, Y: cy + s*0.2},
		}, col, thickness)
	}
}

func IconButton(c *Canvas, in *Input, r Rect, icon string, bg, bgHover, textCol Color, fontSize int) (clicked, hover bool) {
	return iconButtonInset(c, in, r, icon, bg, bgHover, textCol, fontSize, 1.0/6)
}

func iconButtonInset(c *Canvas, in *Input, r Rect, icon string, bg, bgHover, textCol Color, fontSize int, insetFrac float64) (clicked, hover bool) {
	return iconButtonCorners(c, in, r, icon, bg, bgHover, textCol, fontSize, insetFrac, true, true, true, true)
}

func iconButtonCorners(c *Canvas, in *Input, r Rect, icon string, bg, bgHover, textCol Color, fontSize int, insetFrac float64, tl, tr, br, bl bool) (clicked, hover bool) {
	hover = r.Contains(in.MouseX, in.MouseY)
	col := bg
	if hover {
		col = bgHover
	}
	c.FillRoundedRectCorners(r, RadiusSmall, tl, tr, br, bl, col)

	switch {
	case svgIcons[icon] != nil:
		insetX, insetY := int(float64(r.W)*insetFrac), int(float64(r.H)*insetFrac)
		pad := Rect{X: r.X + insetX, Y: r.Y + insetY, W: r.W - 2*insetX, H: r.H - 2*insetY}
		drawSVGIcon(c, pad, svgIcons[icon], textCol)
	case vectorIcons[icon]:
		drawIcon(c, r, icon, textCol, 2)
	default:
		c.TextCentered(r, textCol, fontSize, icon)
	}

	return hover && in.Released, hover
}

func Checkbox(c *Canvas, in *Input, r Rect, checked bool, boxCol, checkCol, textCol Color, fontSize int, label string) bool {
	hover := r.Contains(in.MouseX, in.MouseY)
	box := Rect{X: r.X, Y: r.Y, W: r.H, H: r.H}
	c.StrokeRoundedRect(box, RadiusSmall-2, boxCol, 1)
	if checked {
		inset := Rect{X: box.X + 4, Y: box.Y + 4, W: box.W - 8, H: box.H - 8}
		c.FillRoundedRect(inset, RadiusSmall-4, checkCol)
	}
	c.TextVCentered(box.X+box.W+12, box, textCol, fontSize, label)

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

	if clicked, _ := iconButtonCorners(c, in, minusRect, "minus", bg, bgHover, textCol, fontSize, 0.3, true, false, false, true); clicked {
		newValue -= step * mult
		stepped = true
	}

	if clicked, _ := iconButtonCorners(c, in, plusRect, "plus", bg, bgHover, textCol, fontSize, 0.3, false, true, true, false); clicked {
		newValue += step * mult
		stepped = true
	}

	midHover := midRect.Contains(in.MouseX, in.MouseY)
	midCol := bg
	if midHover && !editing {
		midCol = bgHover
	}
	c.FillRect(midRect, midCol)
	if editing {
		c.StrokeRect(midRect, focusCol, 1)
	}

	buf = editBuf
	c.ClipRect(Rect{X: midRect.X + 2, Y: midRect.Y, W: midRect.W - 4, H: midRect.H})
	if editing && !stepped {
		submitted, clipErr = applyTextEditKeys(in, &buf, cursor, selectedAll)

		runes := []rune(buf)
		tw, _ := c.TextSize(buf, fontSize)
		tx := midRect.X + (midRect.W-tw)/2

		if *selectedAll && buf != "" {
			drawSelectionHighlight(c, Rect{X: tx, Y: midRect.Y, W: tw, H: midRect.H}, tw)
		}

		c.TextVCentered(tx, midRect, textCol, fontSize, buf)

		if !*selectedAll {
			cw, _ := c.TextSize(string(runes[:*cursor]), fontSize)
			cx := tx + cw
			c.Line([]Point{{X: float64(cx), Y: float64(midRect.Y + 4)}, {X: float64(cx), Y: float64(midRect.Y + midRect.H - 4)}}, focusCol, 2)
		}
	} else {
		label := strconv.FormatFloat(value, 'f', precision, 64)
		tw, _ := c.TextSize(label, fontSize)
		c.TextVCentered(midRect.X+(midRect.W-tw)/2, midRect, textCol, fontSize, label)
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

	tw, _ := c.TextSize(newValue, fontSize)
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
	c.TextVCentered(tx, r, textCol, fontSize, newValue)

	if focused && !*selectedAll {
		cx := tx + cursorW
		c.Line([]Point{{X: float64(cx), Y: float64(r.Y + 4)}, {X: float64(cx), Y: float64(r.Y + r.H - 4)}}, focusCol, 2)
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

const SelectRowHeight = 26

func SelectBox(c *Canvas, in *Input, r Rect, current string, open bool, bg, bgHover, textCol Color, fontSize int) bool {
	hover := r.Contains(in.MouseX, in.MouseY)
	col := bg
	if hover || open {
		col = bgHover
	}
	c.FillRoundedRect(r, RadiusSmall, col)

	caretW := 20

	label, _, _ := fitText(c, current, fontSize, r.W-20-caretW-10)
	c.TextVCentered(r.X+10, r, textCol, fontSize, label)

	caretRect := Rect{X: r.X + r.W - caretW - 10, Y: r.Y, W: caretW, H: r.H}
	caret := icons.IconCaretDown
	if open {
		caret = flipY(caret)
	}
	drawSVGIcon(c, caretRect, caret, textCol)

	return hover && in.Released
}

const SelectMaxVisibleRows = 8

const selectListBottomMargin = 12

func SelectListBounds(headerRect Rect, optionCount, containerH int) Rect {
	rows := optionCount
	if rows > SelectMaxVisibleRows {
		rows = SelectMaxVisibleRows
	}
	h := rows * SelectRowHeight

	limit := containerH - selectListBottomMargin

	y := headerRect.Y + headerRect.H
	if limit > 0 && y+h > limit {
		if fit := limit - h; fit >= headerRect.Y+headerRect.H {
			y = fit
		} else {
			h = (limit - y) / SelectRowHeight * SelectRowHeight
			if h < SelectRowHeight {
				h = SelectRowHeight
			}
		}
	}

	return Rect{X: headerRect.X, Y: y, W: headerRect.W, H: h}
}

func SelectList(c *Canvas, in *Input, headerRect Rect, options []string, current int, scroll *int, containerH int, bg, hoverBg, textCol, borderCol Color, fontSize int) (newIdx int, selected bool) {
	listRect := SelectListBounds(headerRect, len(options), containerH)

	visibleRows := listRect.H / SelectRowHeight
	if visibleRows < 1 {
		visibleRows = 1
	}

	maxScroll := len(options) - visibleRows
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

	shadowRect := Rect{X: listRect.X - 2, Y: listRect.Y - 2, W: listRect.W + 4, H: listRect.H + 4}
	c.FillRoundedRect(shadowRect, RadiusSmall, Color{R: 0, G: 0, B: 0, A: 140})
	c.FillRoundedRect(listRect, RadiusSmall, bg)

	newIdx = current
	visible := len(options) - *scroll
	if visible > visibleRows {
		visible = visibleRows
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
		label, _, _ := fitText(c, options[optIdx], fontSize, rowRect.W-20)
		c.TextVCentered(rowRect.X+10, rowRect, textCol, fontSize, label)
		if hover && in.Released {
			newIdx = optIdx
			selected = true
		}
	}
	c.Unclip()

	if maxScroll > 0 {
		trackH := listRect.H - 6
		thumbH := trackH * visibleRows / len(options)
		if thumbH < 14 {
			thumbH = 14
		}
		thumbY := listRect.Y + 3 + (trackH-thumbH)*(*scroll)/maxScroll
		c.FillRoundedRect(Rect{X: listRect.X + listRect.W - 5, Y: thumbY, W: 3, H: thumbH}, 1, borderCol)
	}

	c.StrokeRoundedRect(listRect, RadiusSmall, borderCol, 2)

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
		x := float64(r.X) + float64(i)*float64(r.W)/float64(len(values)-1)
		norm := (v - min) / span
		y := float64(r.Y+r.H) - float64(norm)*float64(r.H)
		pts[i] = Point{X: x, Y: y}
	}
	c.Line(pts, col, 2)
}
