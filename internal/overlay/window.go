package overlay

import (
	"github.com/seneal/wtrto/internal/native"
	"github.com/seneal/wtrto/internal/platform"
)

type Window struct {
	nw     *native.Window
	bounds platform.Rect
}

func New(title string, fpsLimit int) (*Window, error) {
	bounds := platform.ScreenBounds()

	nw, err := native.NewWindow(native.WindowOptions{
		Title:        title,
		X:            bounds.X,
		Y:            bounds.Y,
		W:            bounds.W,
		H:            bounds.H,
		Transparent:  true,
		Decorated:    false,
		AlwaysOnTop:  true,
		ClickThrough: true,
	})
	if err != nil {
		return nil, err
	}

	w := &Window{nw: nw, bounds: bounds}
	w.SetTargetFPS(fpsLimit)

	return w, nil
}

func (w *Window) Size() (int, int) {
	return w.bounds.W, w.bounds.H
}

func (w *Window) SetTargetFPS(fps int) {
	if fps <= 0 {
		w.nw.SetVSync(true)

		return
	}

	w.nw.SetVSync(false)
	w.nw.SetFPS(fps)
}

func (w *Window) SetAlpha(a float64) {
	w.nw.SetAlpha(a)
}

func (w *Window) SetClickThrough(enable bool) {
	w.nw.SetClickThrough(enable)
	if !enable {
		w.nw.Focus()
	}
}

func (w *Window) SetShouldClose(v bool) {
	if v {
		w.nw.Close()
	}
}

func (w *Window) Hide() {
	w.nw.Hide()
}

func (w *Window) Show(refocus bool) {
	w.nw.Show()

	if refocus {
		w.nw.Focus()
	}
}

func (w *Window) ActiveWindowPID() int32 {
	return w.nw.ActiveWindowPID()
}

func (w *Window) Run(loop native.FrameFunc) {
	w.nw.Run(loop)
}
