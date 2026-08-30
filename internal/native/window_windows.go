package native

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procPeekMessageW             = user32.NewProc("PeekMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procUpdateLayeredWindow      = user32.NewProc("UpdateLayeredWindow")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetDC                    = user32.NewProc("GetDC")
	procReleaseDC                = user32.NewProc("ReleaseDC")
	procRegisterHotKey           = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey         = user32.NewProc("UnregisterHotKey")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetKeyState              = user32.NewProc("GetKeyState")

	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procDeleteDC           = gdi32.NewProc("DeleteDC")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

var gwlExStyle = -20

const (
	wsPopup      = 0x80000000
	wsExLayered  = 0x00080000
	wsExTopMost  = 0x00000008
	wsExToolWin  = 0x00000080
	wsExTransp   = 0x00000020
	swHide       = 0
	swShow       = 5
	ulwAlpha     = 2
	acSrcOver    = 0
	acSrcAlpha   = 1
	pmRemove     = 0x0001
	wmDestroy    = 0x0002
	wmClose      = 0x0010
	wmMouseMove  = 0x0200
	wmLButtonDwn = 0x0201
	wmLButtonUp  = 0x0202
	wmKeyDown    = 0x0100
	wmChar       = 0x0102
	wmHotkey     = 0x0312
	vkBack       = 0x08
	vkReturn     = 0x0D
	vkEscape     = 0x1B
	vkDelete     = 0x2E
	vkLeft       = 0x25
	vkRight      = 0x27
	vkShift      = 0x10
	vkControl    = 0x11
	vkMenu       = 0x12
	vkLWin       = 0x5B
	vkRWin       = 0x5C
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type msgT struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	ptX     int32
	ptY     int32
}

type point32 struct{ X, Y int32 }

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

type blendFunction struct {
	BlendOp             byte
	BlendFlags          byte
	SourceConstantAlpha byte
	AlphaFormat         byte
}

func isKeyDown(vk int) bool {
	r, _, _ := procGetKeyState.Call(uintptr(vk))

	return int16(r) < 0
}

func currentKeyMods() uint {
	var mods uint

	if isKeyDown(vkShift) {
		mods |= ModShift
	}

	if isKeyDown(vkControl) {
		mods |= ModCtrl
	}

	if isKeyDown(vkMenu) {
		mods |= ModAlt
	}

	if isKeyDown(vkLWin) || isKeyDown(vkRWin) {
		mods |= ModSuper
	}

	return mods
}

func IsModifierKeySym(vk uint) bool {
	switch vk {
	case vkShift, vkControl, vkMenu, vkLWin, vkRWin:
		return true
	}

	return false
}

func lparamToXY(lParam uintptr) (int, int) {
	x := int(int16(uint16(lParam & 0xFFFF)))
	y := int(int16(uint16((lParam >> 16) & 0xFFFF)))

	return x, y
}

type Window struct {
	hwnd     uintptr
	hInst    uintptr
	memDC    uintptr
	hBitmap  uintptr
	graphics uintptr

	fonts map[fontKey]uintptr

	w, h        int
	fps         int
	resizable   bool
	shouldClose bool
	input       *Input

	hotkeyRegistered bool
}

var currentWindow *Window

func wndProcCallback(hwnd, msg, wParam, lParam uintptr) uintptr {
	if currentWindow != nil && currentWindow.hwnd == hwnd {
		if handled := currentWindow.handleMessage(uint32(msg), wParam, lParam); handled {
			return 0
		}
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)

	return ret
}

func (w *Window) handleMessage(msg uint32, wParam, lParam uintptr) bool {
	in := w.input
	if in == nil {
		return false
	}

	switch msg {
	case wmDestroy:
		w.shouldClose = true
		procPostQuitMessage.Call(0)

		return true

	case wmClose:
		w.shouldClose = true

		return true

	case wmMouseMove:
		in.MouseX, in.MouseY = lparamToXY(lParam)
		in.KeyMods = currentKeyMods()

		return true

	case wmLButtonDwn:
		in.MouseDown = true
		in.Pressed = true
		in.MouseX, in.MouseY = lparamToXY(lParam)
		in.KeyMods = currentKeyMods()

		return true

	case wmLButtonUp:
		in.MouseDown = false
		in.Released = true
		in.MouseX, in.MouseY = lparamToXY(lParam)
		in.KeyMods = currentKeyMods()

		return true

	case wmKeyDown:
		vk := uint(wParam)
		in.KeyEvent = true
		in.KeySym = vk
		in.KeyMods = currentKeyMods()

		switch vk {
		case vkBack:
			in.KeyBackspace = true
		case vkReturn:
			in.KeyEnter = true
		case vkEscape:
			in.KeyEscape = true
		case vkDelete:
			in.KeyDelete = true
		case vkLeft:
			in.KeyLeft = true
		case vkRight:
			in.KeyRight = true
		}

		if in.KeyMods&ModCtrl != 0 {
			switch vk {
			case 'A':
				in.KeyCtrlLetter = 'a'
			case 'C':
				in.KeyCtrlLetter = 'c'
			case 'V':
				in.KeyCtrlLetter = 'v'
			}
		}

		return true

	case wmChar:
		code := uint16(wParam)
		if code >= 32 && in.KeyCtrlLetter == 0 {
			in.KeyRune = rune(code)
		}

		return true

	case wmHotkey:
		if wParam == 1 {
			in.HotkeyTriggered = true
		}

		return true
	}

	return false
}

func NewWindow(opts WindowOptions) (*Window, error) {
	gdipInit()

	hInst, _, _ := procGetModuleHandleW.Call(0)

	className, _ := syscall.UTF16PtrFromString("WTRTOWindowClass")
	title, _ := syscall.UTF16PtrFromString(opts.Title)

	wc := wndClassEx{
		style:         0,
		lpfnWndProc:   syscall.NewCallback(wndProcCallback),
		hInstance:     hInst,
		lpszClassName: className,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	exStyle := uintptr(wsExLayered | wsExToolWin)
	if opts.AlwaysOnTop {
		exStyle |= wsExTopMost
	}
	if opts.ClickThrough {
		exStyle |= wsExTransp
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		uintptr(wsPopup),
		uintptr(opts.X), uintptr(opts.Y), uintptr(opts.W), uintptr(opts.H),
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		return nil, &nativeError{"CreateWindowExW failed"}
	}

	w := &Window{
		hwnd:      hwnd,
		hInst:     hInst,
		w:         opts.W,
		h:         opts.H,
		resizable: opts.Decorated,
		fps:       30,
		fonts:     make(map[fontKey]uintptr),
	}
	currentWindow = w

	w.createBuffers(opts.W, opts.H)

	procShowWindow.Call(hwnd, swShow)

	return w, nil
}

func (w *Window) createBuffers(width, height int) {
	if width < 1 {
		width = 1
	}

	if height < 1 {
		height = 1
	}

	screenDC, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, screenDC)

	w.memDC, _, _ = procCreateCompatibleDC.Call(screenDC)

	bmi := bitmapInfo{
		Header: bitmapInfoHeader{
			biWidth:       int32(width),
			biHeight:      -int32(height),
			biPlanes:      1,
			biBitCount:    32,
			biCompression: 0,
		},
	}
	bmi.Header.biSize = uint32(unsafe.Sizeof(bmi.Header))

	var bits uintptr
	hBitmap, _, _ := procCreateDIBSection.Call(w.memDC, uintptr(unsafe.Pointer(&bmi)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	w.hBitmap = hBitmap

	procSelectObject.Call(w.memDC, w.hBitmap)

	w.graphics = gdipCreateFromHDC(w.memDC)

	w.w, w.h = width, height
}

func (w *Window) destroyBuffers() {
	if w.graphics != 0 {
		gdipDeleteGraphics(w.graphics)
		w.graphics = 0
	}

	if w.hBitmap != 0 {
		procDeleteObject.Call(w.hBitmap)
		w.hBitmap = 0
	}

	if w.memDC != 0 {
		procDeleteDC.Call(w.memDC)
		w.memDC = 0
	}
}

func (w *Window) Size() (int, int) {
	return w.w, w.h
}

func (w *Window) Pos() (int, int) {
	return 0, 0
}

func (w *Window) SetPos(x, y int) {
	procSetWindowPos.Call(w.hwnd, 0, uintptr(x), uintptr(y), 0, 0, 0x0001|0x0004)
}

func (w *Window) SetClickThrough(enable bool) {
	cur, _, _ := procGetWindowLongPtrW.Call(w.hwnd, uintptr(gwlExStyle))
	if enable {
		cur |= wsExTransp
	} else {
		cur &^= wsExTransp
	}

	procSetWindowLongPtrW.Call(w.hwnd, uintptr(gwlExStyle), cur)
}

func (w *Window) Close() {
	w.shouldClose = true
}

func (w *Window) Hide() {
	procShowWindow.Call(w.hwnd, swHide)
}

func (w *Window) Show() {
	procShowWindow.Call(w.hwnd, swShow)
}

func (w *Window) ActiveWindowPID() int32 {
	fg, _, _ := procGetForegroundWindow.Call()
	if fg == 0 {
		return 0
	}

	var pid uint32
	procGetWindowThreadProcessId.Call(fg, uintptr(unsafe.Pointer(&pid)))

	return int32(pid)
}

func (w *Window) Focus() {
	procSetForegroundWindow.Call(w.hwnd)
}

func (w *Window) GrabHotkey(keysym uint, mods uint) bool {
	w.UngrabHotkey()

	var winMods uintptr
	if mods&ModShift != 0 {
		winMods |= 0x0004
	}

	if mods&ModCtrl != 0 {
		winMods |= 0x0002
	}

	if mods&ModAlt != 0 {
		winMods |= 0x0001
	}

	if mods&ModSuper != 0 {
		winMods |= 0x0008
	}

	r, _, _ := procRegisterHotKey.Call(w.hwnd, 1, winMods, uintptr(keysym))
	w.hotkeyRegistered = r != 0

	return w.hotkeyRegistered
}

func (w *Window) UngrabHotkey() {
	if !w.hotkeyRegistered {
		return
	}

	procUnregisterHotKey.Call(w.hwnd, 1)
	w.hotkeyRegistered = false
}

func (w *Window) pollEvents(in *Input) {
	in.Pressed = false
	in.Released = false
	in.KeyEvent = false
	in.KeyRune = 0
	in.KeyCtrlLetter = 0
	in.KeyBackspace = false
	in.KeyEnter = false
	in.KeyEscape = false
	in.KeyDelete = false
	in.KeyLeft = false
	in.KeyRight = false
	in.HotkeyTriggered = false

	w.input = in

	var m msgT
	for {
		r, _, _ := procPeekMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
		if r == 0 {
			break
		}

		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (w *Window) SetFPS(fps int) {
	if fps <= 0 {
		fps = 30
	}

	w.fps = fps
}

func (w *Window) Run(frame FrameFunc) {
	var in Input
	canvas := &Canvas{win: w}
	w.input = &in

	screenDC, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, screenDC)

	srcPt := point32{0, 0}
	blend := blendFunction{BlendOp: acSrcOver, SourceConstantAlpha: 255, AlphaFormat: acSrcAlpha}

	for !w.shouldClose {
		start := time.Now()

		w.pollEvents(&in)

		gdipClearGraphics(w.graphics)

		if !frame(canvas, &in) {
			break
		}

		procUpdateLayeredWindow.Call(
			w.hwnd, screenDC, 0, 0, w.memDC,
			uintptr(unsafe.Pointer(&srcPt)), 0,
			uintptr(unsafe.Pointer(&blend)), ulwAlpha,
		)

		frameDur := time.Second / time.Duration(w.fps)
		if elapsed := time.Since(start); elapsed < frameDur {
			time.Sleep(frameDur - elapsed)
		}
	}

	w.destroyBuffers()
	procDestroyWindow.Call(w.hwnd)
}

type fontKey struct {
	size int
	bold bool
}

type nativeError struct{ msg string }

func (e *nativeError) Error() string { return e.msg }
