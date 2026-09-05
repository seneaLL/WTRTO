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
	hudAccent   = native.Color{R: 68, G: 224, B: 140, A: 255}
	hudTextDim  = native.Color{R: 158, G: 165, B: 176, A: 220}
	hudText     = native.Color{R: 230, G: 232, B: 236, A: 255}
	hudPanelBg  = native.Color{R: 15, G: 17, B: 21, A: 235}
	hudPanelBrd = native.Color{R: 46, G: 51, B: 62, A: 255}
	hudDivider  = native.Color{R: 36, G: 40, B: 49, A: 255}
)

const (
	debugPanelWidth  = 360
	debugPanelPad    = 20
	debugPanelMargin = 20
	debugRowHeight   = 68
)

func NewHUD(screenW, screenH int, sampler *metrics.Sampler, debugEnabled func() bool, showFPS func() bool, active func() bool) native.FrameFunc {
	var lastFrame time.Time
	var fps float64

	return func(c *native.Canvas, in *native.Input) bool {
		now := time.Now()
		if !lastFrame.IsZero() {
			if dt := now.Sub(lastFrame).Seconds(); dt > 0 {
				fps = 1 / dt
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
			drawDebugPanel(c, screenW, screenH, sampler)
		}

		return true
	}
}

func drawDebugPanel(c *native.Canvas, screenW, screenH int, sampler *metrics.Sampler) {
	rows := 7
	height := rows*debugRowHeight + 56
	x := screenW - debugPanelWidth - debugPanelMargin
	y := screenH - height - debugPanelMargin - 28

	panel := native.Rect{X: x, Y: y, W: debugPanelWidth, H: height}
	c.FillRoundedRect(panel, native.RadiusLarge, hudPanelBg)
	c.StrokeRoundedRect(panel, native.RadiusLarge, hudPanelBrd, 1)

	cy := y + 30
	c.TextBold(x+debugPanelPad, cy, hudText, 13, i18n.T("debug.title"))
	cy += 14
	c.Line([]native.Point{{X: x + debugPanelPad, Y: cy}, {X: x + debugPanelWidth - debugPanelPad, Y: cy}}, hudDivider, 1)
	cy += 22

	cy = metricRow(c, x+debugPanelPad, cy, i18n.T("debug.fps"), "", sampler.FPS.Values(), hudAccent, true)
	cy = metricRow(c, x+debugPanelPad, cy, i18n.T("debug.cpu_app"), "%", sampler.AppCPU.Values(), native.Color{R: 108, G: 186, B: 255, A: 255}, true)
	cy = metricRow(c, x+debugPanelPad, cy, i18n.T("debug.mem_app"), "MB", sampler.AppMem.Values(), native.Color{R: 108, G: 186, B: 255, A: 255}, true)

	if sampler.WTRunning() {
		cy = metricRow(c, x+debugPanelPad, cy, i18n.T("debug.cpu_wt"), "%", sampler.WTCPU.Values(), native.Color{R: 255, G: 176, B: 90, A: 255}, true)
		cy = metricRow(c, x+debugPanelPad, cy, i18n.T("debug.mem_wt"), "MB", sampler.WTMem.Values(), native.Color{R: 255, G: 176, B: 90, A: 255}, true)
	} else {
		c.Text(x+debugPanelPad, cy+16, hudTextDim, 12, i18n.T("debug.wt_not_running"))
		cy += debugRowHeight
	}

	if sampler.GPUAvailable() {
		used, total := sampler.GPUMem()
		c.Text(x+debugPanelPad, cy+16, hudTextDim, 12, fmt.Sprintf("%s: %.0f / %.0f MB", i18n.T("debug.gpu_mem"), used, total))
		cy += 28
		cy = metricRow(c, x+debugPanelPad, cy, i18n.T("debug.gpu"), "%", sampler.GPUUtil.Values(), native.Color{R: 210, G: 138, B: 255, A: 255}, true)
	} else {
		c.Text(x+debugPanelPad, cy+16, hudTextDim, 12, i18n.T("debug.gpu_unavailable"))
	}
}

func metricRow(c *native.Canvas, x, y int, label, unit string, values []float32, col native.Color, show bool) int {
	if !show {
		return y
	}

	var last, min, max, avg float32
	if len(values) > 0 {
		last = values[len(values)-1]
		min, max, avg = metrics.Stats(values)
	}

	c.Text(x, y+13, hudTextDim, 12, label)
	c.Text(x, y+30, hudText, 12, fmt.Sprintf("%.0f%s  (min %.0f / avg %.0f / max %.0f)", last, unit, min, avg, max))

	if len(values) > 1 {
		native.Sparkline(c, native.Rect{X: x, Y: y + 38, W: debugPanelWidth - 2*debugPanelPad, H: 22}, values, col)
	}

	return y + debugRowHeight
}
