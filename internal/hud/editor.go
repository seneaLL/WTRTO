package hud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/seneaLL/WTRTO/internal/clipboard"
	"github.com/seneaLL/WTRTO/internal/config"
	"github.com/seneaLL/WTRTO/internal/dialog"
	"github.com/seneaLL/WTRTO/internal/native"
)

var (
	panelBg      = native.Color{R: 18, G: 20, B: 24, A: 235}
	panelBorder  = native.Color{R: 90, G: 96, B: 108, A: 255}
	fieldBg      = native.Color{R: 30, G: 32, B: 38, A: 255}
	fieldBorder  = native.Color{R: 70, G: 74, B: 84, A: 255}
	fieldFocus   = native.Color{R: 68, G: 224, B: 140, A: 255}
	btnBg        = native.Color{R: 46, G: 50, B: 58, A: 255}
	btnHover     = native.Color{R: 62, G: 66, B: 76, A: 255}
	dangerBg     = native.Color{R: 140, G: 40, B: 40, A: 255}
	dangerHover  = native.Color{R: 170, G: 55, B: 55, A: 255}
	labelCol     = native.Color{R: 200, G: 204, B: 210, A: 255}
	textCol      = native.Color{R: 235, G: 238, B: 242, A: 255}
	okCol        = native.Color{R: 120, G: 230, B: 140, A: 255}
	errCol       = native.Color{R: 235, G: 110, B: 100, A: 255}
	tabActiveCol = native.Color{R: 68, G: 224, B: 140, A: 255}
	tooltipBg    = native.Color{R: 10, G: 12, B: 15, A: 235}
)

func newElement() Element {
	return Element{
		ID:        fmt.Sprintf("el_%d", time.Now().UnixNano()),
		Kind:      KindText,
		Binding:   BindIAS,
		Label:     "IAS",
		Unit:      "km/h",
		X:         0.5,
		Y:         0.5,
		FontSize:  15,
		Precision: 0,
		Color:     Color{R: 80, G: 255, B: 90, A: 255},
	}
}

func kindOptions() []string { return []string{"text", "tape_v", "tape_h"} }

func bindingOptions() []string {
	opts := make([]string, len(AllBindings))
	for i, b := range AllBindings {
		opts[i] = string(b)
	}

	return opts
}

func indexOf(opts []string, v string) int {
	for i, o := range opts {
		if o == v {
			return i
		}
	}

	return 0
}

func deconflictBuiltinName(t Template) Template {
	if IsBuiltin(t.Name) {
		t.Name = t.Name + " (копия)"
	}

	return t
}

func wrapText(c *native.Canvas, s string, fontSize, maxWidth int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, 2)
	cur := words[0]
	for _, w := range words[1:] {
		trial := cur + " " + w
		tw, _ := c.TextSize(trial, fontSize)
		if tw > maxWidth {
			lines = append(lines, cur)
			cur = w
		} else {
			cur = trial
		}
	}

	return append(lines, cur)
}

const panelGroupGap = 10

func row(panelX, y int) native.Rect {
	return native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: 28}
}

type dropdownField struct {
	rect    native.Rect
	options []string
	current int
	apply   func(int)
}

type tooltipInfo struct {
	rect native.Rect
	text string
}

func drawTooltip(c *native.Canvas, t tooltipInfo, screenW, screenH int) {
	fontSize := 12
	tw, th := c.TextSize(t.text, fontSize)
	w, h := tw+16, th+10
	x := t.rect.X + t.rect.W/2 - w/2
	y := t.rect.Y - h - 6
	if y < 0 {
		y = t.rect.Y + t.rect.H + 6
	}
	if y+h > screenH {
		y = screenH - h
	}
	if x < 0 {
		x = 0
	}
	if x+w > screenW {
		x = screenW - w
	}
	box := native.Rect{X: x, Y: y, W: w, H: h}
	c.FillRoundedRect(box, native.RadiusSmall, tooltipBg)
	c.StrokeRoundedRect(box, native.RadiusSmall, fieldBorder, 1)
	c.TextCentered(box, textCol, fontSize, t.text)
}

func DrawPropertiesPanel(c *native.Canvas, in *native.Input, screenW, screenH int, tmpl *Template, edit *EditState) bool {
	originalIn := in
	readOnly := IsBuiltin(tmpl.Name)
	changed := false

	panelX := screenW - PanelWidth
	c.FillRect(native.Rect{X: panelX, Y: 0, W: PanelWidth, H: screenH}, panelBg)
	c.StrokeRect(native.Rect{X: panelX, Y: 0, W: PanelWidth, H: screenH}, panelBorder, 1)

	var pendingDropdown *dropdownField
	var tooltip *tooltipInfo

	iconButton := func(r native.Rect, icon, tip string, bg, bgHover native.Color) bool {
		clicked, hover := native.IconButton(c, in, r, icon, bg, bgHover, textCol, 15)
		if hover {
			tooltip = &tooltipInfo{rect: r, text: tip}
		}

		return clicked
	}

	selectField := func(key string, r native.Rect, options []string, current int, apply func(int)) {
		cur := "-"
		if current >= 0 && current < len(options) {
			cur = options[current]
		}
		if native.SelectBox(c, in, r, cur, edit.OpenDropdown == key, btnBg, btnHover, textCol, 13) {
			if edit.OpenDropdown == key {
				edit.OpenDropdown = ""
			} else {
				edit.OpenDropdown = key
				edit.DropdownScroll = 0
			}
		}
		if edit.OpenDropdown == key {
			pendingDropdown = &dropdownField{rect: r, options: options, current: current, apply: apply}

			listBounds := native.SelectListBounds(r, len(options), screenH)
			if listBounds.Contains(in.MouseX, in.MouseY) {
				masked := *in
				masked.MouseDown = false
				masked.Pressed = false
				masked.Released = false
				masked.ScrollDelta = 0
				in = &masked
			}
		}
	}

	numberField := func(key string, r native.Rect, value, step float64, precision int) float64 {
		editing := edit.FocusField == key
		editBuf, cursor, selAll := "", 0, false

		if editing {
			editBuf = edit.NumberEditBuf
			cursor = edit.TextCursor
			selAll = edit.TextSelectedAll
		}

		nv, buf, clicked, submitted, stepped, clipErr := native.NumberStepper(c, in, r, value, step, precision, btnBg, btnHover, textCol, fieldFocus, 13, editing, editBuf, &cursor, &selAll)

		if clipErr != nil {
			edit.StatusMsg, edit.StatusOK = "Буфер обмена недоступен (нет xclip/xsel)", false
		}

		switch {
		case stepped:
			edit.FocusField = ""
			edit.NumberEditBuf = ""

		case clicked:
			edit.FocusField = key
			seed := strconv.FormatFloat(value, 'f', precision, 64)
			edit.NumberEditBuf = seed
			edit.TextCursor = len([]rune(seed))
			edit.TextSelectedAll = false

		case editing && submitted:
			if parsed, err := strconv.ParseFloat(strings.ReplaceAll(buf, ",", "."), 64); err == nil {
				nv = parsed
			}

			edit.FocusField = ""
			edit.NumberEditBuf = ""

		case editing:
			edit.NumberEditBuf = buf
			edit.TextCursor = cursor
			edit.TextSelectedAll = selAll
		}

		return nv
	}

	textField := func(key string, r native.Rect, value string, fontSize int) (string, bool) {
		focused := edit.FocusField == key
		selAll := focused && edit.TextSelectedAll
		cursor := len([]rune(value))

		if focused {
			cursor = edit.TextCursor
		}

		nv, submitted, clicked, clipErr := native.TextInput(c, in, r, value, focused, &cursor, &selAll, fieldBorder, fieldFocus, fieldBg, textCol, fontSize)

		if clipErr != nil {
			edit.StatusMsg, edit.StatusOK = "Буфер обмена недоступен (нет xclip/xsel)", false
		}

		if clicked {
			edit.FocusField = key
		}

		if focused || clicked {
			edit.TextCursor = cursor
			edit.TextSelectedAll = selAll
		}

		return nv, submitted
	}

	y := 16

	tabW := (PanelWidth - 32) / 2
	tmplTabRect := native.Rect{X: panelX + 16, Y: y, W: tabW, H: 32}
	elemTabRect := native.Rect{X: tmplTabRect.X + tabW, Y: y, W: PanelWidth - 32 - tabW, H: 32}
	drawTab := func(r native.Rect, label string, active bool) bool {
		hover := r.Contains(in.MouseX, in.MouseY)
		col := labelCol
		if active || hover {
			col = textCol
		}
		c.TextCentered(r, col, 14, label)
		underline := native.Rect{X: r.X, Y: r.Y + r.H - 2, W: r.W, H: 2}
		underlineCol := panelBorder
		if active {
			underlineCol = tabActiveCol
		}
		c.FillRect(underline, underlineCol)

		return hover && in.Released
	}
	if drawTab(tmplTabRect, "Шаблон", edit.PanelTab != "element") {
		edit.PanelTab = "template"
	}
	if drawTab(elemTabRect, "Элемент", edit.PanelTab == "element") {
		edit.PanelTab = "element"
	}
	y += 44

	textWidth := PanelWidth - 32
	wrapped := func(s string, col native.Color, fontSize int) {
		for _, line := range wrapText(c, s, fontSize, textWidth) {
			c.Text(panelX+16, y+12, col, fontSize, line)
			y += fontSize + 5
		}
	}
	label := func(s string) {
		wrapped(s, labelCol, 12)
	}

	pairRow := func(leftLabel, rightLabel string, leftFrac float64) (native.Rect, native.Rect) {
		full := row(panelX, y)
		leftW := int(float64(full.W-8) * leftFrac)
		leftR := native.Rect{X: full.X, Y: y + 17, W: leftW, H: 28}
		rightR := native.Rect{X: full.X + leftW + 8, Y: y + 17, W: full.W - leftW - 8, H: 28}
		if leftLabel != "" {
			c.Text(leftR.X, y+12, labelCol, 12, leftLabel)
		}
		if rightLabel != "" {
			c.Text(rightR.X, y+12, labelCol, 12, rightLabel)
		}
		y += 17

		return leftR, rightR
	}

	finish := func() bool {
		drawPendingDropdown(c, originalIn, edit, pendingDropdown, screenH)
		if tooltip != nil {
			drawTooltip(c, *tooltip, screenW, screenH)
		}

		return changed
	}

	if edit.PanelTab != "element" {
		label("Активный шаблон")
		names, _ := List()
		tidx := indexOf(names, tmpl.Name)
		selectField("template", row(panelX, y), names, tidx, func(i int) {
			loaded, err := Load(names[i])
			if err != nil {
				edit.StatusMsg, edit.StatusOK = "Не удалось загрузить шаблон", false

				return
			}
			*tmpl = loaded
			edit.Selected = ""
			edit.FocusField = ""
			edit.ShareCode = ""
			s := config.Load()
			s.ActiveTemplate = loaded.Name
			config.Save(s)
			edit.StatusMsg, edit.StatusOK = "Загружен: "+loaded.Name, true
		})
		y += 34

		bgRect := native.Rect{X: panelX + 16, Y: y, W: 20, H: 20}
		if native.Checkbox(c, in, bgRect, !edit.HideBackground, fieldBorder, fieldFocus, textCol, 13, "Закрашивать фон") {
			edit.HideBackground = !edit.HideBackground
		}
		y += 34 + panelGroupGap

		if readOnly {
			wrapped("Стандартный шаблон - только просмотр. Сохраните копию ниже, чтобы редактировать.", errCol, 11)
			y += 6
		}

		label("Сохранить как")
		nameRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32 - 44, H: 28}
		nv, submitted := textField("newname", nameRect, edit.NewTemplateName, 13)
		if nv != edit.NewTemplateName {
			edit.NewTemplateName = nv
		}
		saveAsRect := native.Rect{X: nameRect.X + nameRect.W + 8, Y: y, W: 36, H: 28}
		saveAsClicked := iconButton(saveAsRect, "save", "Сохранить как копию шаблона", btnBg, btnHover)
		if submitted || saveAsClicked {
			name := edit.NewTemplateName
			if name == "" {
				edit.StatusMsg, edit.StatusOK = "Введите имя шаблона", false
			} else if IsBuiltin(name) {
				edit.StatusMsg, edit.StatusOK = "Это имя зарезервировано под стандартный шаблон, выберите другое", false
			} else {
				fork := Template{Name: name, Army: tmpl.Army, Elements: append([]Element(nil), tmpl.Elements...)}
				if err := Save(fork); err != nil {
					edit.StatusMsg, edit.StatusOK = "Ошибка сохранения: "+err.Error(), false
				} else {
					*tmpl = fork
					s := config.Load()
					s.ActiveTemplate = fork.Name
					config.Save(s)
					edit.StatusMsg, edit.StatusOK = "Сохранено как: "+fork.Name, true
					edit.NewTemplateName = ""
					edit.FocusField = ""
					edit.ShareCode = ""
				}
			}
		}
		y += 40 + panelGroupGap

		tmplBtnCount := 3
		if !readOnly {
			tmplBtnCount = 4
		}
		tmplBtnH := 40
		tmplBtnGap := 8
		tmplBtnW := (PanelWidth - 32 - (tmplBtnCount-1)*tmplBtnGap) / tmplBtnCount
		tmplBtnX := func(i int) int { return panelX + 16 + i*(tmplBtnW+tmplBtnGap) }

		newBlankRect := native.Rect{X: tmplBtnX(0), Y: y, W: tmplBtnW, H: tmplBtnH}
		if iconButton(newBlankRect, "plus", "Новый пустой шаблон", btnBg, btnHover) {
			blank := Template{Name: fmt.Sprintf("Новый шаблон %d", time.Now().Unix()%100000), Army: "air"}
			Save(blank)
			*tmpl = blank
			edit.Selected = ""
			edit.FocusField = ""
			edit.ShareCode = ""
			s := config.Load()
			s.ActiveTemplate = blank.Name
			config.Save(s)
			edit.StatusMsg, edit.StatusOK = "Создан: "+blank.Name, true
		}
		exportRect := native.Rect{X: tmplBtnX(1), Y: y, W: tmplBtnW, H: tmplBtnH}
		if iconButton(exportRect, "export", "Экспорт шаблона в файл", btnBg, btnHover) {
			edit.PendingDialog = "export"
		}
		importRect := native.Rect{X: tmplBtnX(2), Y: y, W: tmplBtnW, H: tmplBtnH}
		if iconButton(importRect, "import", "Импорт шаблона из файла", btnBg, btnHover) {
			edit.PendingDialog = "import"
		}
		if !readOnly {
			deleteTmplRect := native.Rect{X: tmplBtnX(3), Y: y, W: tmplBtnW, H: tmplBtnH}
			if iconButton(deleteTmplRect, "trash", "Удалить шаблон", dangerBg, dangerHover) {
				deletedName := tmpl.Name
				if err := Delete(deletedName); err != nil {
					edit.StatusMsg, edit.StatusOK = "Ошибка удаления: "+err.Error(), false
				} else {
					fallback, ferr := Load(DefaultAirTemplate().Name)
					if ferr != nil {
						fallback = EnsureDefault(DefaultAirTemplate())
					}
					*tmpl = fallback
					edit.Selected = ""
					edit.FocusField = ""
					edit.OpenDropdown = ""
					edit.ShareCode = ""
					s := config.Load()
					s.ActiveTemplate = fallback.Name
					config.Save(s)
					edit.StatusMsg, edit.StatusOK = "Удалён: "+deletedName, true
				}
			}
		}
		y += 50 + panelGroupGap

		label("Код шаблона")
		codeRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32 - 92, H: 28}
		cv, _ := textField("sharecode", codeRect, edit.ShareCode, 11)
		if cv != edit.ShareCode {
			edit.ShareCode = cv
		}
		shareRect := native.Rect{X: codeRect.X + codeRect.W + 8, Y: y, W: 40, H: 28}
		if iconButton(shareRect, "clipboard", "Скопировать код шаблона в буфер обмена", btnBg, btnHover) {
			code, err := EncodeShareCode(*tmpl)
			if err != nil {
				edit.StatusMsg, edit.StatusOK = "Ошибка генерации кода: "+err.Error(), false
			} else {
				edit.ShareCode = code
				if cerr := clipboard.Copy(code); cerr != nil {
					edit.StatusMsg, edit.StatusOK = "Код готов (скопируйте вручную из поля выше - буфер обмена недоступен)", true
				} else {
					edit.StatusMsg, edit.StatusOK = "Код скопирован в буфер обмена", true
				}
			}
		}
		loadRect := native.Rect{X: shareRect.X + shareRect.W + 4, Y: y, W: 40, H: 28}
		if iconButton(loadRect, "download", "Загрузить шаблон из кода выше", btnBg, btnHover) {
			if edit.ShareCode == "" {
				edit.StatusMsg, edit.StatusOK = "Вставьте код в поле выше", false
			} else if loaded, err := DecodeShareCode(edit.ShareCode); err != nil {
				edit.StatusMsg, edit.StatusOK = err.Error(), false
			} else {
				loaded = deconflictBuiltinName(loaded)
				if err := Save(loaded); err != nil {
					edit.StatusMsg, edit.StatusOK = "Ошибка сохранения: "+err.Error(), false
				} else {
					*tmpl = loaded
					edit.Selected = ""
					edit.FocusField = ""
					s := config.Load()
					s.ActiveTemplate = loaded.Name
					config.Save(s)
					edit.StatusMsg, edit.StatusOK = "Загружено по коду: "+loaded.Name, true
				}
			}
		}
		y += 38

		if edit.StatusMsg != "" {
			col := okCol
			if !edit.StatusOK {
				col = errCol
			}
			wrapped(edit.StatusMsg, col, 12)
		}

		return finish()
	}

	if !readOnly {
		addRect := row(panelX, y)
		if native.Button(c, in, addRect, "+ Добавить элемент", btnBg, btnHover, textCol, 14) {
			e := newElement()
			tmpl.Elements = append(tmpl.Elements, e)
			edit.Selected = e.ID
			edit.FocusField = ""
			edit.OpenDropdown = ""
			changed = true
		}
		y += 40
		c.Line([]native.Point{{X: float64(panelX + 16), Y: float64(y)}, {X: float64(panelX + PanelWidth - 16), Y: float64(y)}}, panelBorder, 1)
		y += 16
	}

	if edit.Selected == "" {
		msg := "Выберите элемент на экране для редактирования"
		if readOnly {
			msg = "Элементы недоступны для выбора в стандартном шаблоне"
		}
		wrapped(msg, labelCol, 12)

		return finish()
	}

	idx := -1
	for i := range tmpl.Elements {
		if tmpl.Elements[i].ID == edit.Selected {
			idx = i
			break
		}
	}
	if idx == -1 {
		edit.Selected = ""

		return finish()
	}
	e := &tmpl.Elements[idx]

	if !readOnly {
		centerHRect := native.Rect{X: panelX + 16, Y: y, W: 40, H: 28}
		centerVRect := native.Rect{X: centerHRect.X + centerHRect.W + 8, Y: y, W: 40, H: 28}
		if iconButton(centerHRect, "align-h", "Выровнять по центру по горизонтали", btnBg, btnHover) {
			e.X = 0.5
			changed = true
		}
		if iconButton(centerVRect, "align-v", "Выровнять по центру по вертикали", btnBg, btnHover) {
			e.Y = 0.5
			changed = true
		}
		y += 34 + panelGroupGap
	}

	label("Тип")
	opts := kindOptions()
	selectField("kind", row(panelX, y), opts, indexOf(opts, string(e.Kind)), func(i int) {
		newKind := ElementKind(opts[i])
		if newKind != e.Kind {
			e.Kind = newKind
			changed = true
		}
	})
	y += 36 + panelGroupGap

	if e.Kind != KindHorizon {
		label("Величина")
		bopts := bindingOptions()
		selectField("binding", row(panelX, y), bopts, indexOf(bopts, string(e.Binding)), func(i int) {
			newBind := Binding(bopts[i])
			if newBind != e.Binding {
				e.Binding = newBind
				changed = true
			}
		})
		y += 36

		if supportsAutoColor(e.Binding) {
			autoColorRect := row(panelX, y)
			if native.Checkbox(c, in, native.Rect{X: autoColorRect.X, Y: autoColorRect.Y, W: 20, H: 20}, e.AutoColor, fieldBorder, fieldFocus, textCol, 13, "Авто-цвет по лимиту") {
				e.AutoColor = !e.AutoColor
				changed = true
			}
			y += 34
		}
		y += panelGroupGap
	}

	if e.Kind == KindText {
		labelR, unitR := pairRow("Подпись", "Единицы", 0.62)
		nv, submitted := textField("label", labelR, e.Label, 13)
		if submitted {
			edit.FocusField = ""
		}
		if nv != e.Label {
			e.Label = nv
			changed = true
		}
		nv, submitted = textField("unit", unitR, e.Unit, 13)
		if submitted {
			edit.FocusField = ""
		}
		if nv != e.Unit {
			e.Unit = nv
			changed = true
		}
		y += 36

		precR, boldR := pairRow("Точность", "", 0.4)
		nv2 := numberField("el_precision", precR, float64(e.Precision), 1, 0)
		if int(nv2) != e.Precision {
			e.Precision = int(nv2)
			changed = true
		}
		if native.Checkbox(c, in, native.Rect{X: boldR.X, Y: boldR.Y, W: 20, H: 20}, e.Bold, fieldBorder, fieldFocus, textCol, 13, "Жирный") {
			e.Bold = !e.Bold
			changed = true
		}
		y += 36 + panelGroupGap
	}

	if e.Kind != KindHorizon {
		if e.Kind == KindTapeV || e.Kind == KindTapeH {
			fsRect, thickRect := pairRow("Размер шрифта", "Толщина", 0.5)
			fontSize := e.FontSize
			if fontSize == 0 {
				fontSize = 15
			}
			nv := numberField("el_fontsize", fsRect, float64(fontSize), 1, 0)
			if int(nv) != fontSize {
				e.FontSize = int(nv)
				changed = true
			}
			thickness := e.Thickness
			if thickness == 0 {
				thickness = 1
			}
			nvt := numberField("el_thickness", thickRect, float64(thickness), 1, 0)
			if int(nvt) != thickness {
				e.Thickness = int(nvt)
				if e.Thickness < 1 {
					e.Thickness = 1
				}
				changed = true
			}
			y += 36
		} else {
			label("Размер шрифта")
			fontSize := e.FontSize
			if fontSize == 0 {
				fontSize = 15
			}
			nv := numberField("el_fontsize", row(panelX, y), float64(fontSize), 1, 0)
			if int(nv) != fontSize {
				e.FontSize = int(nv)
				changed = true
			}
			y += 36
		}
		y += panelGroupGap
	}

	if e.Kind == KindTapeV || e.Kind == KindTapeH {
		styleRect, dirRect := pairRow("Стиль", "Направление", 0.5)
		sopts := []string{string(StyleStraight), string(StyleArc)}
		si := 0
		if e.Style == StyleArc {
			si = 1
		}
		selectField("style", styleRect, sopts, si, func(i int) {
			newStyle := Style(sopts[i])
			if newStyle != e.Style {
				e.Style = newStyle
				changed = true
			}
		})
		var dopts []string
		if e.Kind == KindTapeV {
			dopts = []string{string(DirUp), string(DirDown)}
		} else {
			dopts = []string{string(DirCW), string(DirCCW)}
		}
		selectField("direction", dirRect, dopts, indexOf(dopts, string(e.Direction)), func(i int) {
			newDir := Direction(dopts[i])
			if newDir != e.Direction {
				e.Direction = newDir
				changed = true
			}
		})
		y += 36

		if e.Kind == KindTapeV {
			label("Сторона")
			lopts := []string{string(SideAuto), string(SideLeft), string(SideRight)}
			selectField("labelside", row(panelX, y), lopts, indexOf(lopts, string(e.LabelSide)), func(i int) {
				newSide := LabelSide(lopts[i])
				if newSide != e.LabelSide {
					e.LabelSide = newSide
					changed = true
				}
			})
			y += 36
		}

		lenRect, rangeRect := pairRow("Длина", "Диапазон", 0.5)
		nv := numberField("el_length", lenRect, e.Length, 0.01, 3)
		if nv != e.Length {
			e.Length = nv
			changed = true
		}
		nv = numberField("el_range", rangeRect, e.Range, 10, 0)
		if nv != e.Range {
			e.Range = nv
			changed = true
		}
		y += 36

		minorRect, majorRect := pairRow("Малый шаг", "Большой шаг", 0.5)
		nv = numberField("el_minorstep", minorRect, e.MinorStep, 1, 0)
		if nv != e.MinorStep {
			e.MinorStep = nv
			changed = true
		}
		nv = numberField("el_majorstep", majorRect, e.MajorStep, 5, 0)
		if nv != e.MajorStep {
			e.MajorStep = nv
			changed = true
		}
		y += 36 + panelGroupGap

		zonesLocked := e.AutoColor
		zonesRealIn := in
		if zonesLocked {
			masked := *in
			masked.MouseDown, masked.Pressed, masked.Released, masked.ScrollDelta = false, false, false, 0
			in = &masked
		}
		zonesTop := y

		label("Зоны")
		rowH := 36
		thW := PanelWidth - 32 - rowH - rowH - 16
		for zi := 0; zi < len(e.Zones); zi++ {
			swatchRect := native.Rect{X: panelX + 16, Y: y, W: rowH, H: rowH}
			thRect := native.Rect{X: swatchRect.X + rowH + 8, Y: y, W: thW, H: rowH}
			delRect := native.Rect{X: panelX + 16 + PanelWidth - 32 - rowH, Y: y, W: rowH, H: rowH}

			nv := numberField(fmt.Sprintf("el_zone_%d", zi), thRect, e.Zones[zi].Threshold, 10, 0)
			if nv != e.Zones[zi].Threshold {
				e.Zones[zi].Threshold = nv
				changed = true
			}

			c.FillRoundedRect(swatchRect, native.RadiusSmall, toNativeColor(e.Zones[zi].Color))
			swatchBorder := fieldBorder
			if edit.ColorPickerFor == e.ID && edit.ColorPickerZone == zi {
				swatchBorder = fieldFocus
			}
			c.StrokeRoundedRect(swatchRect, native.RadiusSmall, swatchBorder, 2)
			if !zonesLocked && swatchRect.Contains(in.MouseX, in.MouseY) && in.Released {
				if edit.ColorPickerFor == e.ID && edit.ColorPickerZone == zi {
					edit.ColorPickerFor = ""
				} else {
					edit.ColorPickerFor = e.ID
					edit.ColorPickerZone = zi
				}
			}

			if iconButton(delRect, "trash", "Удалить зону", dangerBg, dangerHover) {
				e.Zones = append(e.Zones[:zi], e.Zones[zi+1:]...)
				if edit.ColorPickerFor == e.ID && edit.ColorPickerZone == zi {
					edit.ColorPickerFor = ""
				}
				changed = true
				y += rowH + 10
				continue
			}
			y += rowH + 10

			if edit.ColorPickerFor == e.ID && edit.ColorPickerZone == zi {
				pickerH := native.ColorPickerHeight(PanelWidth - 32)
				pickerRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: pickerH}
				newZoneCol := native.ColorPicker(c, in, pickerRect, toNativeColor(e.Zones[zi].Color))
				if newZoneCol != toNativeColor(e.Zones[zi].Color) {
					e.Zones[zi].Color = Color{R: newZoneCol.R, G: newZoneCol.G, B: newZoneCol.B, A: newZoneCol.A}
					changed = true
				}
				y += pickerH + 10
			}
		}
		addZoneRect := native.Rect{X: panelX + 16, Y: y, W: 40, H: 32}
		if iconButton(addZoneRect, "plus", "Добавить зону", btnBg, btnHover) {
			e.Zones = append(e.Zones, Zone{Threshold: 0, Color: Color{R: 80, G: 255, B: 90, A: 255}})
			changed = true
		}
		y += 40

		if zonesLocked {
			in = zonesRealIn
			c.FillRect(native.Rect{X: panelX + 16, Y: zonesTop, W: PanelWidth - 32, H: y - zonesTop}, native.Color{R: 10, G: 12, B: 15, A: 150})
		}
		y += panelGroupGap
	}

	if !e.AutoColor && len(e.Zones) == 0 {
		label("Цвет")
		colorPickerH := native.ColorPickerHeight(PanelWidth - 32)
		newCol := native.ColorPicker(c, in, native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: colorPickerH}, toNativeColor(e.Color))
		if newCol != toNativeColor(e.Color) {
			e.Color = Color{R: newCol.R, G: newCol.G, B: newCol.B, A: newCol.A}
			changed = true
		}
		y += colorPickerH + 16
		y += panelGroupGap
	}

	glowRect := row(panelX, y)
	if native.Checkbox(c, in, native.Rect{X: glowRect.X, Y: glowRect.Y, W: 20, H: 20}, e.GlowEnabled, fieldBorder, fieldFocus, textCol, 13, "Свечение") {
		e.GlowEnabled = !e.GlowEnabled
		if e.GlowEnabled {
			if e.GlowIntensity == 0 {
				e.GlowIntensity = 0.6
			}
			if !e.GlowUseOwn && e.GlowColor.A == 0 {
				e.GlowUseOwn = true
			}
		}
		changed = true
	}
	y += 34

	if e.GlowEnabled {
		ownRect := row(panelX, y)
		if native.Checkbox(c, in, native.Rect{X: ownRect.X, Y: ownRect.Y, W: 20, H: 20}, e.GlowUseOwn, fieldBorder, fieldFocus, textCol, 13, "Цвет элемента") {
			e.GlowUseOwn = !e.GlowUseOwn
			if !e.GlowUseOwn && e.GlowColor.A == 0 {
				e.GlowColor = Color{R: 255, G: 255, B: 255, A: 255}
			}
			changed = true
		}
		y += 34

		if !e.GlowUseOwn {
			label("Цвет свечения")
			glowPickerH := native.ColorPickerHeight(PanelWidth - 32)
			newGlowCol := native.ColorPicker(c, in, native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: glowPickerH}, toNativeColor(e.GlowColor))
			if newGlowCol != toNativeColor(e.GlowColor) {
				e.GlowColor = Color{R: newGlowCol.R, G: newGlowCol.G, B: newGlowCol.B, A: newGlowCol.A}
				changed = true
			}
			y += glowPickerH + 16
		}

		label("Интенсивность, %")
		intensityPct := numberField("el_glow_intensity", row(panelX, y), e.GlowIntensity*100, 5, 0)
		nv := intensityPct / 100
		if nv < 0 {
			nv = 0
		}
		if nv > 1 {
			nv = 1
		}
		if nv != e.GlowIntensity {
			e.GlowIntensity = nv
			changed = true
		}
		y += 36
	}

	delRect := native.Rect{X: panelX + 16, Y: screenH - 60, W: PanelWidth - 32, H: 36}
	if native.Button(c, in, delRect, "Удалить", dangerBg, dangerHover, textCol, 14) {
		tmpl.Elements = append(tmpl.Elements[:idx], tmpl.Elements[idx+1:]...)
		edit.Selected = ""
		edit.FocusField = ""
		edit.OpenDropdown = ""
		changed = true
	}

	return finish()
}

func drawPendingDropdown(c *native.Canvas, in *native.Input, edit *EditState, pending *dropdownField, screenH int) {
	if pending == nil {
		return
	}
	newIdx, selected := native.SelectList(c, in, pending.rect, pending.options, pending.current, &edit.DropdownScroll, screenH, fieldBg, btnHover, textCol, fieldFocus, 13)
	if selected {
		pending.apply(newIdx)
		edit.OpenDropdown = ""

		return
	}

	if in.Released && !pending.rect.Contains(in.MouseX, in.MouseY) {
		edit.OpenDropdown = ""
	}
}

func ResolvePendingDialog(tmpl *Template, edit *EditState) {
	action := edit.PendingDialog
	edit.PendingDialog = ""

	switch action {
	case "export":
		path, err := dialog.SaveFile("Экспорт шаблона", slug(tmpl.Name)+".json")

		switch {
		case err != nil:
			edit.StatusMsg, edit.StatusOK = "Диалог сохранения недоступен: "+err.Error(), false

		case path == "":

		default:
			if exportErr := Export(*tmpl, path); exportErr != nil {
				edit.StatusMsg, edit.StatusOK = "Ошибка экспорта: "+exportErr.Error(), false
			} else {
				edit.StatusMsg, edit.StatusOK = "Экспортировано в "+path, true
			}
		}

	case "import":
		path, err := dialog.OpenFile("Импорт шаблона")

		switch {
		case err != nil:
			edit.StatusMsg, edit.StatusOK = "Диалог выбора файла недоступен: "+err.Error(), false

		case path == "":

		default:
			loaded, lerr := Import(path)
			if lerr != nil {
				edit.StatusMsg, edit.StatusOK = "Ошибка импорта: "+lerr.Error(), false
				break
			}

			loaded = deconflictBuiltinName(loaded)
			*tmpl = loaded
			edit.Selected = ""
			edit.FocusField = ""
			edit.ShareCode = ""
			s := config.Load()
			s.ActiveTemplate = loaded.Name
			config.Save(s)
			edit.StatusMsg, edit.StatusOK = "Импортировано: "+loaded.Name, true
		}
	}
}
