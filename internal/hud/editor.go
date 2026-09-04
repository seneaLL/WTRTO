package hud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/seneal/wtrto/internal/clipboard"
	"github.com/seneal/wtrto/internal/config"
	"github.com/seneal/wtrto/internal/dialog"
	"github.com/seneal/wtrto/internal/native"
)

var (
	panelBg     = native.Color{R: 18, G: 20, B: 24, A: 235}
	panelBorder = native.Color{R: 90, G: 96, B: 108, A: 255}
	fieldBg     = native.Color{R: 30, G: 32, B: 38, A: 255}
	fieldBorder = native.Color{R: 70, G: 74, B: 84, A: 255}
	fieldFocus  = native.Color{R: 68, G: 224, B: 140, A: 255}
	btnBg       = native.Color{R: 46, G: 50, B: 58, A: 255}
	btnHover    = native.Color{R: 62, G: 66, B: 76, A: 255}
	dangerBg    = native.Color{R: 140, G: 40, B: 40, A: 255}
	dangerHover = native.Color{R: 170, G: 55, B: 55, A: 255}
	labelCol    = native.Color{R: 200, G: 204, B: 210, A: 255}
	textCol     = native.Color{R: 235, G: 238, B: 242, A: 255}
	okCol       = native.Color{R: 120, G: 230, B: 140, A: 255}
	errCol      = native.Color{R: 235, G: 110, B: 100, A: 255}
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

func row(panelX, y int) native.Rect {
	return native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: 28}
}

type dropdownField struct {
	rect    native.Rect
	options []string
	current int
	apply   func(int)
}

func DrawPropertiesPanel(c *native.Canvas, in *native.Input, screenW, screenH int, tmpl *Template, edit *EditState) bool {

	originalIn := in
	readOnly := IsBuiltin(tmpl.Name)
	changed := false

	if !readOnly {
		addRect := native.Rect{X: 16, Y: 16, W: 150, H: 34}
		if native.Button(c, in, addRect, "+ Добавить элемент", btnBg, btnHover, textCol, 14) {
			e := newElement()
			tmpl.Elements = append(tmpl.Elements, e)
			edit.Selected = e.ID
			edit.FocusField = ""
			edit.OpenDropdown = ""
			changed = true
		}
	}

	panelX := screenW - PanelWidth
	c.FillRect(native.Rect{X: panelX, Y: 0, W: PanelWidth, H: screenH}, panelBg)
	c.StrokeRect(native.Rect{X: panelX, Y: 0, W: PanelWidth, H: screenH}, panelBorder, 1)

	var pending *dropdownField
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
			pending = &dropdownField{rect: r, options: options, current: current, apply: apply}

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

	y := 20
	c.TextBold(panelX+16, y+18, textCol, 15, "Шаблоны")
	y += 34

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
	y += 34

	if readOnly {
		wrapped("Стандартный шаблон - только просмотр. Сохраните копию ниже, чтобы редактировать.", errCol, 11)
		y += 6
	}

	nameRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32 - 92, H: 28}
	nv, submitted := textField("newname", nameRect, edit.NewTemplateName, 13)
	if nv != edit.NewTemplateName {
		edit.NewTemplateName = nv
	}
	saveAsRect := native.Rect{X: nameRect.X + nameRect.W + 8, Y: y, W: 84, H: 28}
	saveAsClicked := native.Button(c, in, saveAsRect, "Сохранить", btnBg, btnHover, textCol, 12)
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
	y += 36

	if readOnly {
		newBlankRect := row(panelX, y)
		if native.Button(c, in, newBlankRect, "Новый пустой шаблон", btnBg, btnHover, textCol, 13) {
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
	} else {
		newBlankRect := native.Rect{X: panelX + 16, Y: y, W: (PanelWidth-32)/2 - 4, H: 30}
		deleteRect := native.Rect{X: newBlankRect.X + newBlankRect.W + 8, Y: y, W: (PanelWidth-32)/2 - 4, H: 30}
		if native.Button(c, in, newBlankRect, "Новый пустой", btnBg, btnHover, textCol, 13) {
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
		if native.Button(c, in, deleteRect, "Удалить шаблон", dangerBg, dangerHover, textCol, 13) {
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
	y += 36

	exportRect := native.Rect{X: panelX + 16, Y: y, W: (PanelWidth-32)/2 - 4, H: 30}
	importRect := native.Rect{X: exportRect.X + exportRect.W + 8, Y: y, W: (PanelWidth-32)/2 - 4, H: 30}

	if native.Button(c, in, exportRect, "Экспорт", btnBg, btnHover, textCol, 13) {
		edit.PendingDialog = "export"
	}

	if native.Button(c, in, importRect, "Импорт", btnBg, btnHover, textCol, 13) {
		edit.PendingDialog = "import"
	}
	y += 38

	label("Код шаблона")
	codeRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: 28}
	cv, _ := textField("sharecode", codeRect, edit.ShareCode, 11)
	if cv != edit.ShareCode {
		edit.ShareCode = cv
	}
	y += 34

	shareRect := native.Rect{X: panelX + 16, Y: y, W: (PanelWidth-32)/2 - 4, H: 30}
	loadRect := native.Rect{X: shareRect.X + shareRect.W + 8, Y: y, W: (PanelWidth-32)/2 - 4, H: 30}
	if native.Button(c, in, shareRect, "Поделиться", btnBg, btnHover, textCol, 13) {
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
	if native.Button(c, in, loadRect, "Загрузить", btnBg, btnHover, textCol, 13) {
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
		y += 8
	}

	y += 8
	c.Line([]native.Point{{X: panelX + 16, Y: y}, {X: panelX + PanelWidth - 16, Y: y}}, panelBorder, 1)
	y += 16

	if edit.Selected == "" {
		msg := "Выберите элемент на экране для редактирования"
		if readOnly {
			msg = "Элементы недоступны для выбора в стандартном шаблоне"
		}
		wrapped(msg, labelCol, 12)
		drawPendingDropdown(c, originalIn, edit, pending, screenH)

		return changed
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
		drawPendingDropdown(c, originalIn, edit, pending, screenH)

		return changed
	}
	e := &tmpl.Elements[idx]

	c.TextBold(panelX+16, y+18, textCol, 15, "Свойства элемента")
	y += 36

	if !readOnly {
		centerHRect := native.Rect{X: panelX + 16, Y: y, W: (PanelWidth-32)/2 - 4, H: 28}
		centerVRect := native.Rect{X: centerHRect.X + centerHRect.W + 8, Y: y, W: (PanelWidth-32)/2 - 4, H: 28}
		if native.Button(c, in, centerHRect, "По центру X", btnBg, btnHover, textCol, 12) {
			e.X = 0.5
			changed = true
		}
		if native.Button(c, in, centerVRect, "По центру Y", btnBg, btnHover, textCol, 12) {
			e.Y = 0.5
			changed = true
		}
		y += 34
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
	y += 36

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
	}

	if e.Kind == KindText {
		label("Подпись")
		r := row(panelX, y)
		nv, submitted := textField("label", r, e.Label, 13)
		if submitted {
			edit.FocusField = ""
		}
		if nv != e.Label {
			e.Label = nv
			changed = true
		}
		y += 36

		label("Единицы")
		r = row(panelX, y)
		nv, submitted = textField("unit", r, e.Unit, 13)
		if submitted {
			edit.FocusField = ""
		}
		if nv != e.Unit {
			e.Unit = nv
			changed = true
		}
		y += 36

		label("Точность (знаков)")
		nv2 := numberField("el_precision", row(panelX, y), float64(e.Precision), 1, 0)
		if int(nv2) != e.Precision {
			e.Precision = int(nv2)
			changed = true
		}
		y += 36

		boldRect := row(panelX, y)
		if native.Checkbox(c, in, native.Rect{X: boldRect.X, Y: boldRect.Y, W: 20, H: 20}, e.Bold, fieldBorder, fieldFocus, textCol, 13, "Жирный") {
			e.Bold = !e.Bold
			changed = true
		}
		y += 36
	}

	if e.Kind != KindHorizon {
		label("Размер шрифта")
		nv := numberField("el_fontsize", row(panelX, y), float64(e.FontSize), 1, 0)
		if int(nv) != e.FontSize {
			e.FontSize = int(nv)
			changed = true
		}
		y += 36
	}

	if e.Kind == KindTapeV || e.Kind == KindTapeH {
		label("Стиль")
		sopts := []string{string(StyleStraight), string(StyleArc)}
		si := 0
		if e.Style == StyleArc {
			si = 1
		}
		selectField("style", row(panelX, y), sopts, si, func(i int) {
			newStyle := Style(sopts[i])
			if newStyle != e.Style {
				e.Style = newStyle
				changed = true
			}
		})
		y += 36

		label("Направление")
		var dopts []string
		if e.Kind == KindTapeV {
			dopts = []string{string(DirUp), string(DirDown)}
		} else {
			dopts = []string{string(DirCW), string(DirCCW)}
		}
		selectField("direction", row(panelX, y), dopts, indexOf(dopts, string(e.Direction)), func(i int) {
			newDir := Direction(dopts[i])
			if newDir != e.Direction {
				e.Direction = newDir
				changed = true
			}
		})
		y += 36

		if e.Kind == KindTapeV {
			label("Сторона подписи")
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

		label("Длина (доля экрана)")
		nv := numberField("el_length", row(panelX, y), e.Length, 0.01, 3)
		if nv != e.Length {
			e.Length = nv
			changed = true
		}
		y += 36

		label("Диапазон")
		nv = numberField("el_range", row(panelX, y), e.Range, 10, 0)
		if nv != e.Range {
			e.Range = nv
			changed = true
		}
		y += 36

		label("Малый шаг")
		nv = numberField("el_minorstep", row(panelX, y), e.MinorStep, 1, 0)
		if nv != e.MinorStep {
			e.MinorStep = nv
			changed = true
		}
		y += 36

		label("Большой шаг")
		nv = numberField("el_majorstep", row(panelX, y), e.MajorStep, 5, 0)
		if nv != e.MajorStep {
			e.MajorStep = nv
			changed = true
		}
		y += 36

		label("Толщина линий")
		thickness := e.Thickness
		if thickness == 0 {
			thickness = 1
		}
		nvt := numberField("el_thickness", row(panelX, y), float64(thickness), 1, 0)
		if int(nvt) != thickness {
			e.Thickness = int(nvt)
			if e.Thickness < 1 {
				e.Thickness = 1
			}
			changed = true
		}
		y += 36

		label("Зоны (порог → цвет)")
		for zi := 0; zi < len(e.Zones); zi++ {
			zr := row(panelX, y)
			thRect := native.Rect{X: zr.X, Y: zr.Y, W: 90, H: zr.H}
			nv := numberField(fmt.Sprintf("el_zone_%d", zi), thRect, e.Zones[zi].Threshold, 10, 0)
			if nv != e.Zones[zi].Threshold {
				e.Zones[zi].Threshold = nv
				changed = true
			}
			swatch := native.Rect{X: zr.X + 96, Y: zr.Y, W: 24, H: zr.H}
			c.FillRect(swatch, toNativeColor(e.Zones[zi].Color))
			c.StrokeRect(swatch, fieldBorder, 1)
			delRect := native.Rect{X: zr.X + zr.W - 28, Y: zr.Y, W: 28, H: zr.H}
			if native.Button(c, in, delRect, "×", dangerBg, dangerHover, textCol, 13) {
				e.Zones = append(e.Zones[:zi], e.Zones[zi+1:]...)
				changed = true
				y += 34
				continue
			}
			y += 34
		}
		addZoneRect := row(panelX, y)
		if native.Button(c, in, addZoneRect, "+ Зона", btnBg, btnHover, textCol, 13) {
			e.Zones = append(e.Zones, Zone{Threshold: 0, Color: Color{R: 80, G: 255, B: 90, A: 255}})
			changed = true
		}
		y += 36
	}

	label("Цвет")
	newCol := native.ColorSliders(c, in, native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: 72}, toNativeColor(e.Color))
	if newCol != toNativeColor(e.Color) {
		e.Color = Color{R: newCol.R, G: newCol.G, B: newCol.B, A: newCol.A}
		changed = true
	}
	y += 88

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
			newGlowCol := native.ColorSliders(c, in, native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: 72}, toNativeColor(e.GlowColor))
			if newGlowCol != toNativeColor(e.GlowColor) {
				e.GlowColor = Color{R: newGlowCol.R, G: newGlowCol.G, B: newGlowCol.B, A: newGlowCol.A}
				changed = true
			}
			y += 88
		}

		label("Интенсивность свечения, %")
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
	if native.Button(c, in, delRect, "Удалить элемент", dangerBg, dangerHover, textCol, 14) {
		tmpl.Elements = append(tmpl.Elements[:idx], tmpl.Elements[idx+1:]...)
		edit.Selected = ""
		edit.FocusField = ""
		edit.OpenDropdown = ""
		changed = true
	}

	drawPendingDropdown(c, originalIn, edit, pending, screenH)

	return changed
}

func drawPendingDropdown(c *native.Canvas, in *native.Input, edit *EditState, pending *dropdownField, screenH int) {
	if pending == nil {
		return
	}
	newIdx, selected := native.SelectList(c, in, pending.rect, pending.options, pending.current, &edit.DropdownScroll, screenH, fieldBg, btnHover, textCol, fieldBorder, 13)
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
