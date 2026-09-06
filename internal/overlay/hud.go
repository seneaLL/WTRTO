package overlay

import (
	"fmt"
	"time"

	"github.com/seneaLL/WTRTO/internal/i18n"
	"github.com/seneaLL/WTRTO/internal/metrics"
	"github.com/seneaLL/WTRTO/internal/native"
	"github.com/seneaLL/WTRTO/internal/version"
)

var (
	hudAccent    = native.Color{R: 68, G: 224, B: 140, A: 255}
	hudTextDim   = native.Color{R: 148, G: 155, B: 166, A: 220}
	hudText      = native.Color{R: 230, G: 232, B: 236, A: 255}
	hudPanelBg   = native.Color{R: 16, G: 18, B: 22, A: 235}
	hudPanelBrd  = native.Color{R: 46, G: 51, B: 62, A: 255}
	hudDivider   = native.Color{R: 36, G: 40, B: 49, A: 255}
	hudTrack     = native.Color{R: 40, G: 43, B: 51, A: 255}
	hudAccentApp = native.Color{R: 108, G: 186, B: 255, A: 255}
	hudAccentWT  = native.Color{R: 255, G: 176, B: 90, A: 255}
	hudAccentGPU = native.Color{R: 210, G: 138, B: 255, A: 255}
)

const (
	debugPanelWidth      = 292
	debugPanelPad        = 16
	debugPanelMargin     = 20
	debugBarRowHeight    = 32
	debugTextRowHeight   = 26
	debugSectionGap      = 12
	debugSectionTitleGap = 28
	debugRowFontSize     = 11
)

func NewHUD(screenW, screenH int, sampler *metrics.Sampler, debugEnabled func() bool, showFPS func() bool, active func() bool, rightInset func() int) native.FrameFunc {
	var lastFrame time.Time
	var fps float64
	barAnim := map[string]float64{}

	return func(c *native.Canvas, in *native.Input) bool {
		now := time.Now()
		dt := 0.0
		if !lastFrame.IsZero() {
			if d := now.Sub(lastFrame).Seconds(); d > 0 {
				dt = d
				fps = 1 / d
			}
		}
		lastFrame = now
		sampler.RecordFPS(fps)

		if !active() {
			return true
		}

		if showFPS() {
			c.TextBold(20, 32, hudAccent, 16, fmt.Sprintf("%s - %.0f FPS", i18n.T("overlay.title"), fps))
		}

		buildText := version.String()
		bw, bh := c.TextSize(buildText, 12)
		c.Text(screenW-bw-16, screenH-bh-10, hudTextDim, 12, buildText)

		if debugEnabled() {
			inset := 0
			if rightInset != nil {
				inset = rightInset()
			}
			drawDebugPanel(c, screenW, screenH, inset, sampler, barAnim, dt)
		}

		return true
	}
}

type debugRow struct {
	key    string
	label  string
	value  string
	pct    float64
	accent native.Color
	note   bool
}

func animatedPct(state map[string]float64, key string, target, dt float64) float64 {
	const speed = 10.0

	cur, ok := state[key]
	if !ok || dt <= 0 {
		state[key] = target

		return target
	}

	t := dt * speed
	if t > 1 {
		t = 1
	}
	cur += (target - cur) * t
	state[key] = cur

	return cur
}

type debugSection struct {
	title string
	rows  []debugRow
}

func rowHeight(r debugRow) int {
	if r.note {
		return debugTextRowHeight + 8
	}
	if r.pct >= 0 {
		return debugBarRowHeight
	}

	return debugTextRowHeight
}

func sectionHeight(s debugSection) int {
	h := debugSectionTitleGap
	for _, r := range s.rows {
		h += rowHeight(r)
	}

	return h
}

func drawDebugPanel(c *native.Canvas, screenW, screenH, rightInset int, sampler *metrics.Sampler, barAnim map[string]float64, dt float64) {
	last := func(values []float32) float32 {
		if len(values) == 0 {
			return 0
		}

		return values[len(values)-1]
	}
	pct := func(v float32) float64 { return float64(v) / 100 }

	appRows := []debugRow{
		{key: "app.cpu", label: i18n.T("debug.cpu"), value: fmt.Sprintf("%.0f%%", last(sampler.AppCPU.Values())), pct: pct(last(sampler.AppCPU.Values())), accent: hudAccentApp},
	}
	if av := sampler.AppGPU.Values(); len(av) > 0 {
		appRows = append(appRows, debugRow{key: "app.gpu", label: i18n.T("debug.gpu_short"), value: fmt.Sprintf("%.0f%%", last(av)), pct: pct(last(av)), accent: hudAccentApp})
	}
	appRows = append(appRows, debugRow{label: i18n.T("debug.ram"), value: fmt.Sprintf("%.0f MB", last(sampler.AppMem.Values())), pct: -1})

	sections := []debugSection{
		{title: i18n.T("debug.section_app"), rows: appRows},
	}

	if sampler.WTRunning() {
		wtRows := []debugRow{
			{key: "wt.cpu", label: i18n.T("debug.cpu"), value: fmt.Sprintf("%.0f%%", last(sampler.WTCPU.Values())), pct: pct(last(sampler.WTCPU.Values())), accent: hudAccentWT},
		}
		if wv := sampler.WTGPU.Values(); len(wv) > 0 {
			wtRows = append(wtRows, debugRow{key: "wt.gpu", label: i18n.T("debug.gpu_short"), value: fmt.Sprintf("%.0f%%", last(wv)), pct: pct(last(wv)), accent: hudAccentWT})
		}
		wtRows = append(wtRows, debugRow{label: i18n.T("debug.ram"), value: fmt.Sprintf("%.0f MB", last(sampler.WTMem.Values())), pct: -1})
		sections = append(sections, debugSection{title: i18n.T("debug.section_wt"), rows: wtRows})
	} else {
		sections = append(sections, debugSection{title: i18n.T("debug.section_wt"), rows: []debugRow{{value: i18n.T("debug.wt_not_running"), pct: -1, note: true}}})
	}

	sysRows := []debugRow{
		{key: "sys.cpu", label: i18n.T("debug.cpu"), value: fmt.Sprintf("%.0f%%", last(sampler.SysCPU.Values())), pct: pct(last(sampler.SysCPU.Values())), accent: hudAccentGPU},
	}
	if sampler.GPUAvailable() {
		used, total := sampler.GPUMem()
		memPct := 0.0
		if total > 0 {
			memPct = used / total
		}
		sysRows = append(sysRows,
			debugRow{key: "sys.gpu", label: i18n.T("debug.gpu_short"), value: fmt.Sprintf("%.0f%%", last(sampler.GPUUtil.Values())), pct: pct(last(sampler.GPUUtil.Values())), accent: hudAccentGPU},
			debugRow{key: "sys.gpu_mem", label: i18n.T("debug.gpu_mem"), value: fmt.Sprintf("%.0f%%", memPct*100), pct: memPct, accent: hudAccentGPU},
		)
	} else {
		sysRows = append(sysRows, debugRow{value: i18n.T("debug.gpu_unavailable"), pct: -1, note: true})
	}
	ramUsed, ramTotal := sampler.SysRAMBytes()
	sysRows = append(sysRows, debugRow{label: i18n.T("debug.ram"), value: fmt.Sprintf("%.0f / %.0f MB", ramUsed, ramTotal), pct: -1})

	sections = append(sections, debugSection{title: i18n.T("debug.section_system"), rows: sysRows})

	headerH := 30 + 14 + 22
	height := headerH
	for i, s := range sections {
		height += sectionHeight(s)
		if i < len(sections)-1 {
			height += debugSectionGap
		}
	}
	height += 16

	x := screenW - debugPanelWidth - debugPanelMargin - rightInset
	y := screenH - height - debugPanelMargin - 28

	panel := native.Rect{X: x, Y: y, W: debugPanelWidth, H: height}
	c.FillRoundedRect(panel, native.RadiusLarge, hudPanelBg)
	c.StrokeRoundedRect(panel, native.RadiusLarge, hudPanelBrd, 1)

	cy := y + 30
	c.TextBold(x+debugPanelPad, cy, hudText, 13, i18n.T("debug.title"))
	cy += 14
	c.Line([]native.Point{{X: float64(x + debugPanelPad), Y: float64(cy)}, {X: float64(x + debugPanelWidth - debugPanelPad), Y: float64(cy)}}, hudDivider, 1)
	cy += 22

	rowW := debugPanelWidth - 2*debugPanelPad

	for si, s := range sections {
		c.Text(x+debugPanelPad, cy+10, hudTextDim, 11, s.title)
		cy += debugSectionTitleGap

		for _, r := range s.rows {
			if r.note {
				c.Text(x+debugPanelPad, cy+11, hudTextDim, debugRowFontSize, r.value)
				cy += rowHeight(r)

				continue
			}

			c.Text(x+debugPanelPad, cy+10, hudTextDim, debugRowFontSize, r.label)
			vw, _ := c.TextSize(r.value, debugRowFontSize)
			c.TextBold(x+debugPanelPad+rowW-vw, cy+10, hudText, debugRowFontSize, r.value)

			if r.pct >= 0 {
				barY := cy + 18
				shown := animatedPct(barAnim, r.key, r.pct, dt)
				drawProgressBar(c, x+debugPanelPad, barY, rowW, shown, hudTrack, r.accent)
			}

			cy += rowHeight(r)
		}

		if si < len(sections)-1 {
			cy += debugSectionGap
		}
	}
}

func drawProgressBar(c *native.Canvas, x, y, w int, pct float64, trackCol, fillCol native.Color) {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	const barH = 6

	c.FillRoundedRect(native.Rect{X: x, Y: y, W: w, H: barH}, 3, trackCol)

	fillW := int(float64(w) * pct)
	if fillW <= 0 {
		return
	}
	if fillW < barH {
		fillW = barH
	}

	c.FillRoundedRect(native.Rect{X: x, Y: y, W: fillW, H: barH}, 3, fillCol)
}
