package native

type Color struct {
	R, G, B, A uint8
}

type Rect struct {
	X, Y, W, H int
}

type Point struct {
	X, Y float64
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type WindowOptions struct {
	Title        string
	X, Y         int
	W, H         int
	Transparent  bool
	Decorated    bool
	AlwaysOnTop  bool
	ClickThrough bool
}

type Input struct {
	MouseX, MouseY int
	MouseDown      bool
	Pressed        bool
	Released       bool
	ScrollDelta    int

	KeyEvent      bool
	KeyRune       rune
	KeyCtrlLetter rune
	KeySym        uint
	KeyMods       uint
	KeyBackspace  bool
	KeyEnter      bool
	KeyEscape     bool
	KeyDelete     bool
	KeyLeft       bool
	KeyRight      bool

	HotkeyTriggered bool
}

const (
	ModShift = 1 << 0
	ModLock  = 1 << 1
	ModCtrl  = 1 << 2
	ModAlt   = 1 << 3
	ModSuper = 1 << 6
)

type FrameFunc func(c *Canvas, in *Input) bool
