package hud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/seneaLL/WTRTO/internal/clipboard"
	"github.com/seneaLL/WTRTO/internal/config"
	"github.com/seneaLL/WTRTO/internal/dialog"
	"github.com/seneaLL/WTRTO/internal/i18n"
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

type bgPreset struct {
	color  Color
	border bool
}

func bgPresets() []bgPreset {
	return []bgPreset{
		{color: defaultTextBg},
		{color: Color{R: 0, G: 0, B: 0, A: 235}},
		{color: Color{R: 255, G: 255, B: 255, A: 55}},
		{color: Color{R: 0, G: 0, B: 0, A: 140}, border: true},
	}
}

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

type bindingOption struct {
	Value Binding
	Label string
	Group string
}

var staticBindingMeta = []struct {
	value      Binding
	labelKey   string
	groupKey   string
	shortLabel string
	unit       string
}{
	{BindIAS, "editor.metric.ias", "editor.group_speed", "IAS", "km/h"},
	{BindTAS, "editor.metric.tas", "editor.group_speed", "TAS", "km/h"},
	{BindMach, "editor.metric.mach", "editor.group_speed", "MACH", ""},
	{BindThrottlePct, "editor.metric.throttle_pct", "editor.group_speed", "THR", "%"},

	{BindAltitude, "editor.metric.altitude", "editor.group_altitude", "ALT", "m"},
	{BindRadioAlt, "editor.metric.radio_altitude", "editor.group_altitude", "RALT", "m"},
	{BindVSpeed, "editor.metric.vspeed", "editor.group_altitude", "VS", "m/s"},

	{BindCompass, "editor.metric.compass", "editor.group_navigation", "HDG", "°"},
	{BindTurn, "editor.metric.turn", "editor.group_navigation", "TURN", ""},

	{BindAoA, "editor.metric.aoa", "editor.group_aero", "AoA", "°"},
	{BindAoS, "editor.metric.aos", "editor.group_aero", "AoS", "°"},
	{BindGLoad, "editor.metric.gload", "editor.group_aero", "G", ""},
	{BindIASRate, "editor.metric.ias_rate", "editor.group_aero", "ACC", "km/h/s"},
	{BindWingSweep, "editor.metric.wing_sweep", "editor.group_aero", "SWEEP", "%"},

	{BindAileron, "editor.metric.aileron", "editor.group_controls", "AIL", "%"},
	{BindElevator, "editor.metric.elevator", "editor.group_controls", "ELEV", "%"},
	{BindRudder, "editor.metric.rudder", "editor.group_controls", "RUD", "%"},
	{BindFlaps, "editor.metric.flaps", "editor.group_controls", "FLAP", "%"},
	{BindGearPct, "editor.metric.gear", "editor.group_controls", "GEAR", "%"},
	{BindAirbrake, "editor.metric.airbrake", "editor.group_controls", "AIRBRK", "%"},
	{BindRollRate, "editor.metric.roll_rate", "editor.group_controls", "ROLL", "°/s"},
	{BindTrimmer, "editor.metric.trimmer", "editor.group_controls", "TRIM", "%"},

	{BindFuelKg, "editor.metric.fuel_kg", "editor.group_fuel", "FUEL", "kg"},
	{BindFuelPct, "editor.metric.fuel_pct", "editor.group_fuel", "FUEL", "%"},
	{BindFuelTime, "editor.metric.fuel_time", "editor.group_fuel", "FUEL", "min"},
	{BindFuelRate, "editor.metric.fuel_rate", "editor.group_fuel", "FUEL", "kg/min"},
}

func bindingOptions() []bindingOption {
	opts := make([]bindingOption, 0, len(staticBindingMeta)+MaxEngines*len(engineMetrics))
	for _, s := range staticBindingMeta {
		opts = append(opts, bindingOption{Value: s.value, Label: i18n.T(s.labelKey), Group: i18n.T(s.groupKey)})
	}
	for n := 1; n <= MaxEngines; n++ {
		group := fmt.Sprintf(i18n.T("editor.group_engine_fmt"), n)
		for _, m := range engineMetrics {
			opts = append(opts, bindingOption{Value: engineBinding(m, n), Label: i18n.T(m.labelKey), Group: group})
		}
	}

	return opts
}

func defaultLabelUnit(b Binding) (string, string) {
	if ref, ok := engineBindingIndex[b]; ok {
		return fmt.Sprintf("%s%d", ref.metric.shortLabel, ref.n), ref.metric.unit
	}
	for _, s := range staticBindingMeta {
		if s.value == b {
			return s.shortLabel, s.unit
		}
	}

	return "", ""
}

func indexOfBindingOption(opts []bindingOption, v Binding) int {
	for i, o := range opts {
		if o.Value == v {
			return i
		}
	}

	return 0
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
		t.Name = t.Name + i18n.T("editor.copy_suffix")
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
	labels  []string
	groups  []string
	current int
	apply   func(int)
}

type tooltipInfo struct {
	rect native.Rect
	text string
}

type selectExtra struct {
	Labels []string
	Groups []string
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

func DrawPropertiesPanel(c *native.Canvas, in *native.Input, screenW, screenH int, tmpl *Template, edit *EditState, currentAircraft string) bool {
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

	selectField := func(key string, r native.Rect, values []string, current int, apply func(int), extra ...selectExtra) {
		labels := values
		var groups []string
		if len(extra) > 0 {
			if extra[0].Labels != nil {
				labels = extra[0].Labels
			}
			groups = extra[0].Groups
		}

		cur := "-"
		if current >= 0 && current < len(labels) {
			cur = labels[current]
		}
		if native.SelectBox(c, in, r, cur, edit.OpenDropdown == key, btnBg, btnHover, textCol, 13) {
			if edit.OpenDropdown == key {
				edit.OpenDropdown = ""
			} else {
				edit.OpenDropdown = key
				row := native.SelectRowForOption(current, groups)
				scroll := row - native.SelectMaxVisibleRows/2
				if scroll < 0 {
					scroll = 0
				}
				edit.DropdownScroll = scroll
			}
		}
		if edit.OpenDropdown == key {
			pendingDropdown = &dropdownField{rect: r, labels: labels, groups: groups, current: current, apply: apply}

			rows := native.SelectDisplayRows(len(labels), groups)
			listBounds := native.SelectListBounds(r, rows, screenH)
			if listBounds.Contains(in.MouseX, in.MouseY) {
				masked := *in
				masked.MouseX = -1
				masked.MouseY = -1
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
			edit.StatusMsg, edit.StatusOK = i18n.T("editor.clipboard_unavailable"), false
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
			edit.StatusMsg, edit.StatusOK = i18n.T("editor.clipboard_unavailable"), false
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
	if drawTab(tmplTabRect, i18n.T("editor.tab_template"), edit.PanelTab != "element") {
		edit.PanelTab = "template"
	}
	if drawTab(elemTabRect, i18n.T("editor.tab_element"), edit.PanelTab == "element") {
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

	if edit.PanelTab != "element" && edit.OverrideAircraft != "" {
		wrapped(fmt.Sprintf(i18n.T("editor.override_editing_fmt"), edit.OverrideAircraft, tmpl.Name), labelCol, 12)
		y += 8

		backRect := row(panelX, y)
		if native.Button(c, in, backRect, i18n.T("editor.override_back"), btnBg, btnHover, textCol, 13) {
			edit.OverrideAircraft = ""
			edit.OverrideElements = nil
			edit.Selected = ""
			edit.FocusField = ""
		}
		y += 40 + panelGroupGap

		return finish()
	}

	if edit.PanelTab != "element" {
		label(i18n.T("editor.active_template"))
		names, _ := List()
		tidx := indexOf(names, tmpl.Name)
		selectField("template", row(panelX, y), names, tidx, func(i int) {
			loaded, err := Load(names[i])
			if err != nil {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.load_failed"), false

				return
			}
			*tmpl = loaded
			edit.Selected = ""
			edit.FocusField = ""
			edit.ShareCode = ""
			s := config.Load()
			s.ActiveTemplate = loaded.Name
			config.Save(s)
			edit.StatusMsg, edit.StatusOK = i18n.T("editor.loaded_prefix")+loaded.Name, true
		})
		y += 34

		bgRect := native.Rect{X: panelX + 16, Y: y, W: 20, H: 20}
		if native.Checkbox(c, in, bgRect, !edit.HideBackground, fieldBorder, fieldFocus, textCol, 13, i18n.T("editor.fill_background")) {
			edit.HideBackground = !edit.HideBackground
		}
		y += 34 + panelGroupGap

		if readOnly {
			wrapped(i18n.T("editor.readonly_notice"), errCol, 11)
			y += 6
		}

		label(i18n.T("editor.save_as"))
		nameRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32 - 44, H: 28}
		nv, submitted := textField("newname", nameRect, edit.NewTemplateName, 13)
		if nv != edit.NewTemplateName {
			edit.NewTemplateName = nv
		}
		saveAsRect := native.Rect{X: nameRect.X + nameRect.W + 8, Y: y, W: 36, H: 28}
		saveAsClicked := iconButton(saveAsRect, "save", i18n.T("editor.save_as_copy_tip"), btnBg, btnHover)
		if submitted || saveAsClicked {
			name := edit.NewTemplateName
			if name == "" {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.enter_template_name"), false
			} else if IsBuiltin(name) {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.name_reserved"), false
			} else {
				fork := Template{Name: name, Army: tmpl.Army, Aircraft: tmpl.Aircraft, Elements: append([]Element(nil), tmpl.Elements...)}
				if err := Save(fork); err != nil {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.save_error_prefix")+err.Error(), false
				} else {
					*tmpl = fork
					s := config.Load()
					s.ActiveTemplate = fork.Name
					config.Save(s)
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.saved_as_prefix")+fork.Name, true
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
		if iconButton(newBlankRect, "plus", i18n.T("editor.new_blank_template_tip"), btnBg, btnHover) {
			blank := Template{Name: fmt.Sprintf(i18n.T("editor.new_template_name_fmt"), time.Now().Unix()%100000), Army: "air"}
			Save(blank)
			*tmpl = blank
			edit.Selected = ""
			edit.FocusField = ""
			edit.ShareCode = ""
			s := config.Load()
			s.ActiveTemplate = blank.Name
			config.Save(s)
			edit.StatusMsg, edit.StatusOK = i18n.T("editor.created_prefix")+blank.Name, true
		}
		exportRect := native.Rect{X: tmplBtnX(1), Y: y, W: tmplBtnW, H: tmplBtnH}
		if iconButton(exportRect, "export", i18n.T("editor.export_tip"), btnBg, btnHover) {
			edit.PendingDialog = "export"
		}
		importRect := native.Rect{X: tmplBtnX(2), Y: y, W: tmplBtnW, H: tmplBtnH}
		if iconButton(importRect, "import", i18n.T("editor.import_tip"), btnBg, btnHover) {
			edit.PendingDialog = "import"
		}
		if !readOnly {
			deleteTmplRect := native.Rect{X: tmplBtnX(3), Y: y, W: tmplBtnW, H: tmplBtnH}
			if iconButton(deleteTmplRect, "trash", i18n.T("editor.delete_template_tip"), dangerBg, dangerHover) {
				deletedName := tmpl.Name
				var delErr error
				if tmpl.Aircraft != "" {
					delErr = DeleteStandalone(tmpl.Aircraft)
				} else {
					delErr = Delete(deletedName)
				}
				if delErr != nil {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.delete_error_prefix")+delErr.Error(), false
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
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.deleted_prefix")+deletedName, true
				}
			}
		}
		y += 50 + panelGroupGap

		if tmpl.Aircraft != "" {
			wrapped(fmt.Sprintf(i18n.T("editor.aircraft_only_notice_fmt"), tmpl.Aircraft), labelCol, 12)
			y += 6 + panelGroupGap
		} else {
			label(i18n.T("editor.aircraft_id"))
			aircraftRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32 - 44, H: 28}
			av, asub := textField("aircraftfield", aircraftRect, edit.AircraftFieldBuf, 13)
			if av != edit.AircraftFieldBuf {
				edit.AircraftFieldBuf = av
			}
			if asub {
				edit.FocusField = ""
			}
			useCurRect := native.Rect{X: aircraftRect.X + aircraftRect.W + 8, Y: y, W: 36, H: 28}
			if iconButton(useCurRect, "download", i18n.T("editor.use_current_aircraft_tip"), btnBg, btnHover) {
				if currentAircraft == "" {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.no_current_aircraft"), false
				} else {
					edit.AircraftFieldBuf = currentAircraft
				}
			}
			y += 36

			aircraftBtnW := PanelWidth - 32
			newAircraftRect := native.Rect{X: panelX + 16, Y: y, W: aircraftBtnW, H: 36}
			if native.Button(c, in, newAircraftRect, i18n.T("editor.new_aircraft_template"), btnBg, btnHover, textCol, 13) {
				if edit.AircraftFieldBuf == "" {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.aircraft_id_required"), false
				} else {
					blank := Template{Name: fmt.Sprintf(i18n.T("editor.new_template_name_fmt"), time.Now().Unix()%100000), Army: "air", Aircraft: edit.AircraftFieldBuf}
					Save(blank)
					*tmpl = blank
					edit.Selected = ""
					edit.FocusField = ""
					edit.ShareCode = ""
					edit.AircraftFieldBuf = ""
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.created_prefix")+blank.Name, true
				}
			}
			y += 44

			overrideRect := native.Rect{X: panelX + 16, Y: y, W: aircraftBtnW, H: 36}
			if native.Button(c, in, overrideRect, i18n.T("editor.override_open"), btnBg, btnHover, textCol, 13) {
				if edit.AircraftFieldBuf == "" {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.aircraft_id_required"), false
				} else {
					existing, _ := LoadOverride(tmpl.Name, edit.AircraftFieldBuf)
					edit.OverrideAircraft = edit.AircraftFieldBuf
					edit.OverrideElements = append([]Element(nil), existing.Elements...)
					edit.Selected = ""
					edit.FocusField = ""
					edit.OpenDropdown = ""
				}
			}
			y += 44 + panelGroupGap

			overrideAircraft, _ := ListOverrides(tmpl.Name)
			if len(overrideAircraft) > 0 {
				label(i18n.T("editor.override_list"))
				ovIdx := indexOf(overrideAircraft, edit.AircraftFieldBuf)
				selectField("override_list", row(panelX, y), overrideAircraft, ovIdx, func(i int) {
					edit.AircraftFieldBuf = overrideAircraft[i]
				})
				y += 34

				delOverrideRect := row(panelX, y)
				if edit.AircraftFieldBuf != "" && native.Button(c, in, delOverrideRect, i18n.T("editor.override_delete"), dangerBg, dangerHover, textCol, 13) {
					DeleteOverride(tmpl.Name, edit.AircraftFieldBuf)
					edit.AircraftFieldBuf = ""
				}
				y += 36 + panelGroupGap
			}
		}

		if standalone, _ := ListStandalone(); len(standalone) > 0 {
			label(i18n.T("editor.standalone_list"))
			names := make([]string, len(standalone))
			for i, t := range standalone {
				names[i] = t.Name + " (" + t.Aircraft + ")"
			}
			selectField("standalone_list", row(panelX, y), names, -1, func(i int) {
				*tmpl = standalone[i]
				edit.Selected = ""
				edit.FocusField = ""
				edit.ShareCode = ""
			})
			y += 34 + panelGroupGap
		}

		label(i18n.T("editor.template_code"))
		codeRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32 - 92, H: 28}
		cv, _ := textField("sharecode", codeRect, edit.ShareCode, 11)
		if cv != edit.ShareCode {
			edit.ShareCode = cv
		}
		shareRect := native.Rect{X: codeRect.X + codeRect.W + 8, Y: y, W: 40, H: 28}
		if iconButton(shareRect, "clipboard", i18n.T("editor.copy_code_tip"), btnBg, btnHover) {
			code, err := EncodeShareCode(*tmpl)
			if err != nil {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.generate_code_error_prefix")+err.Error(), false
			} else {
				edit.ShareCode = code
				if cerr := clipboard.Copy(code); cerr != nil {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.code_ready_no_clipboard"), true
				} else {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.code_copied"), true
				}
			}
		}
		loadRect := native.Rect{X: shareRect.X + shareRect.W + 4, Y: y, W: 40, H: 28}
		if iconButton(loadRect, "download", i18n.T("editor.load_from_code_tip"), btnBg, btnHover) {
			if edit.ShareCode == "" {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.paste_code_first"), false
			} else if loaded, err := DecodeShareCode(edit.ShareCode); err != nil {
				edit.StatusMsg, edit.StatusOK = err.Error(), false
			} else {
				loaded = deconflictBuiltinName(loaded)
				if err := Save(loaded); err != nil {
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.save_error_prefix")+err.Error(), false
				} else {
					*tmpl = loaded
					edit.Selected = ""
					edit.FocusField = ""
					s := config.Load()
					s.ActiveTemplate = loaded.Name
					config.Save(s)
					edit.StatusMsg, edit.StatusOK = i18n.T("editor.loaded_from_code_prefix")+loaded.Name, true
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
		if native.Button(c, in, addRect, i18n.T("editor.add_element"), btnBg, btnHover, textCol, 14) {
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
		msg := i18n.T("editor.select_element_hint")
		if readOnly {
			msg = i18n.T("editor.readonly_no_elements")
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
		if iconButton(centerHRect, "align-h", i18n.T("editor.align_h_tip"), btnBg, btnHover) {
			e.X = 0.5
			changed = true
		}
		if iconButton(centerVRect, "align-v", i18n.T("editor.align_v_tip"), btnBg, btnHover) {
			e.Y = 0.5
			changed = true
		}
		y += 34 + panelGroupGap
	}

	label(i18n.T("editor.type"))
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
		label(i18n.T("editor.value"))
		bopts := bindingOptions()
		values := make([]string, len(bopts))
		labels := make([]string, len(bopts))
		groups := make([]string, len(bopts))
		for i, o := range bopts {
			values[i] = string(o.Value)
			labels[i] = o.Label
			groups[i] = o.Group
		}
		selectField("binding", row(panelX, y), values, indexOfBindingOption(bopts, e.Binding), func(i int) {
			newBind := Binding(values[i])
			if newBind != e.Binding {
				e.Binding = newBind
				e.Label, e.Unit = defaultLabelUnit(newBind)
				changed = true
			}
		}, selectExtra{Labels: labels, Groups: groups})
		y += 36

		if supportsAutoColor(e.Binding) {
			autoColorRect := row(panelX, y)
			if native.Checkbox(c, in, native.Rect{X: autoColorRect.X, Y: autoColorRect.Y, W: 20, H: 20}, e.AutoColor, fieldBorder, fieldFocus, textCol, 13, i18n.T("editor.auto_color")) {
				e.AutoColor = !e.AutoColor
				changed = true
			}
			y += 34
		}
		y += panelGroupGap
	}

	if e.Kind == KindText {
		labelR, unitR := pairRow(i18n.T("editor.label"), i18n.T("editor.units"), 0.62)
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

		label(i18n.T("editor.precision"))
		nv2 := numberField("el_precision", row(panelX, y), float64(e.Precision), 1, 0)
		if int(nv2) != e.Precision {
			e.Precision = int(nv2)
			changed = true
		}
		y += 36

		boldR, glowR := pairRow("", "", 0.5)
		if native.Checkbox(c, in, native.Rect{X: boldR.X, Y: boldR.Y, W: 20, H: 20}, e.Bold, fieldBorder, fieldFocus, textCol, 13, i18n.T("editor.bold")) {
			e.Bold = !e.Bold
			changed = true
		}
		if native.Checkbox(c, in, native.Rect{X: glowR.X, Y: glowR.Y, W: 20, H: 20}, e.Glow, fieldBorder, fieldFocus, textCol, 13, i18n.T("editor.glow")) {
			e.Glow = !e.Glow
			changed = true
		}
		y += 34 + panelGroupGap

		bgRowRect := row(panelX, y)
		if native.Checkbox(c, in, native.Rect{X: bgRowRect.X, Y: bgRowRect.Y, W: 20, H: 20}, e.BgEnabled, fieldBorder, fieldFocus, textCol, 13, i18n.T("editor.background")) {
			e.BgEnabled = !e.BgEnabled
			if e.BgEnabled && e.BgColor.A == 0 {
				e.BgColor = defaultTextBg
			}
			changed = true
		}
		y += 34

		if e.BgEnabled {
			label(i18n.T("editor.background_presets"))
			presets := bgPresets()
			presetGap := 8
			presetW := (PanelWidth - 32 - presetGap*(len(presets)-1)) / len(presets)
			presetRowRect := native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: 28}
			for pi, p := range presets {
				pr := native.Rect{X: presetRowRect.X + pi*(presetW+presetGap), Y: y, W: presetW, H: 28}
				c.FillRoundedRect(pr, native.RadiusSmall, toNativeColor(p.color))
				borderCol := fieldBorder
				borderWidth := 1
				if p.color == e.BgColor && p.border == e.BgBorder {
					borderCol = fieldFocus
					borderWidth = 2
				}
				c.StrokeRoundedRect(pr, native.RadiusSmall, borderCol, borderWidth)
				if p.border {
					inset := native.Rect{X: pr.X + 4, Y: pr.Y + 4, W: pr.W - 8, H: pr.H - 8}
					c.StrokeRoundedRect(inset, native.RadiusSmall, toNativeColor(e.Color), 1)
				}
				if pr.Contains(in.MouseX, in.MouseY) && in.Released {
					e.BgColor = p.color
					e.BgBorder = p.border
					changed = true
				}
			}
			y += 28 + panelGroupGap

			label(i18n.T("editor.background_color"))
			bgPickerH := native.ColorPickerHeight(PanelWidth - 32)
			newBgCol := native.ColorPicker(c, in, native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: bgPickerH}, toNativeColor(e.BgColor))
			if newBgCol != toNativeColor(e.BgColor) {
				e.BgColor = Color{R: newBgCol.R, G: newBgCol.G, B: newBgCol.B, A: newBgCol.A}
				changed = true
			}
			y += bgPickerH + 16
		}
		y += panelGroupGap
	}

	if e.Kind != KindHorizon {
		if e.Kind == KindTapeV || e.Kind == KindTapeH {
			fsRect, thickRect := pairRow(i18n.T("editor.font_size"), i18n.T("editor.thickness"), 0.5)
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
			label(i18n.T("editor.font_size"))
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

	if e.Kind != KindText {
		boldRect, glowRect := pairRow("", "", 0.5)
		if native.Checkbox(c, in, native.Rect{X: boldRect.X, Y: boldRect.Y, W: 20, H: 20}, e.Bold, fieldBorder, fieldFocus, textCol, 13, i18n.T("editor.bold")) {
			e.Bold = !e.Bold
			changed = true
		}
		if native.Checkbox(c, in, native.Rect{X: glowRect.X, Y: glowRect.Y, W: 20, H: 20}, e.Glow, fieldBorder, fieldFocus, textCol, 13, i18n.T("editor.glow")) {
			e.Glow = !e.Glow
			changed = true
		}
		y += 34 + panelGroupGap
	}

	if e.Kind == KindTapeV || e.Kind == KindTapeH {
		styleRect, dirRect := pairRow(i18n.T("editor.style"), i18n.T("editor.direction"), 0.5)
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
			label(i18n.T("editor.side"))
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

		lenRect, rangeRect := pairRow(i18n.T("editor.length"), i18n.T("editor.range"), 0.5)
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

		minorRect, majorRect := pairRow(i18n.T("editor.minor_step"), i18n.T("editor.major_step"), 0.5)
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
			masked.MouseX, masked.MouseY = -1, -1
			masked.MouseDown, masked.Pressed, masked.Released, masked.ScrollDelta = false, false, false, 0
			in = &masked
		}
		zonesTop := y

		label(i18n.T("editor.zones"))
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

			if iconButton(delRect, "trash", i18n.T("editor.delete_zone_tip"), dangerBg, dangerHover) {
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
		if iconButton(addZoneRect, "plus", i18n.T("editor.add_zone_tip"), btnBg, btnHover) {
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
		label(i18n.T("editor.color"))
		colorPickerH := native.ColorPickerHeight(PanelWidth - 32)
		newCol := native.ColorPicker(c, in, native.Rect{X: panelX + 16, Y: y, W: PanelWidth - 32, H: colorPickerH}, toNativeColor(e.Color))
		if newCol != toNativeColor(e.Color) {
			e.Color = Color{R: newCol.R, G: newCol.G, B: newCol.B, A: newCol.A}
			changed = true
		}
		y += colorPickerH + 16
		y += panelGroupGap
	}

	delRect := native.Rect{X: panelX + 16, Y: screenH - 60, W: PanelWidth - 32, H: 36}
	if native.Button(c, in, delRect, i18n.T("editor.delete"), dangerBg, dangerHover, textCol, 14) {
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
	newIdx, selected := native.SelectList(c, in, pending.rect, pending.labels, pending.groups, pending.current, &edit.DropdownScroll, screenH, fieldBg, btnHover, textCol, labelCol, fieldFocus, 13)
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
		path, err := dialog.SaveFile(i18n.T("editor.export_dialog_title"), slug(tmpl.Name)+".json")

		switch {
		case err != nil:
			edit.StatusMsg, edit.StatusOK = i18n.T("editor.save_dialog_unavailable_prefix")+err.Error(), false

		case path == "":

		default:
			if exportErr := Export(*tmpl, path); exportErr != nil {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.export_error_prefix")+exportErr.Error(), false
			} else {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.exported_to_prefix")+path, true
			}
		}

	case "import":
		path, err := dialog.OpenFile(i18n.T("editor.import_dialog_title"))

		switch {
		case err != nil:
			edit.StatusMsg, edit.StatusOK = i18n.T("editor.open_dialog_unavailable_prefix")+err.Error(), false

		case path == "":

		default:
			loaded, lerr := Import(path)
			if lerr != nil {
				edit.StatusMsg, edit.StatusOK = i18n.T("editor.import_error_prefix")+lerr.Error(), false
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
			edit.StatusMsg, edit.StatusOK = i18n.T("editor.imported_prefix")+loaded.Name, true
		}
	}
}
