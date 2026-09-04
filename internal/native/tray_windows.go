package native

import (
	"syscall"
	"unsafe"
)

var (
	shell32 = syscall.NewLazyDLL("shell32.dll")

	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")

	procCreatePopupMenu = user32.NewProc("CreatePopupMenu")
	procAppendMenuW     = user32.NewProc("AppendMenuW")
	procTrackPopupMenu  = user32.NewProc("TrackPopupMenuEx")
	procDestroyMenu     = user32.NewProc("DestroyMenu")
	procGetCursorPos    = user32.NewProc("GetCursorPos")
	procLoadIconW       = user32.NewProc("LoadIconW")
	procPostMessageW    = user32.NewProc("PostMessageW")
)

const (
	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	mfString       = 0x00000000
	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100

	idiApplication = 32512

	trayCmdShow = 1
	trayCmdExit = 2
)

type notifyIconDataW struct {
	CbSize            uint32
	Hwnd              uintptr
	UID               uint32
	Flags             uint32
	CallbackMessage   uint32
	HIcon             uintptr
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          [16]byte
	HBalloonIcon      uintptr
}

type pointT struct{ X, Y int32 }

func utf16Copy(dst []uint16, s string) {
	src, err := syscall.UTF16FromString(s)
	if err != nil {
		return
	}

	n := copy(dst, src)
	if n == len(dst) {
		dst[len(dst)-1] = 0
	}
}

func (w *Window) EnableTray(tooltip, showLabel, exitLabel string, onShow, onExit func()) {
	w.trayShowLabel = showLabel
	w.trayExitLabel = exitLabel
	w.onTrayShow = onShow
	w.onTrayExit = onExit

	icon, _, _ := procLoadIconW.Call(0, idiApplication)

	var nid notifyIconDataW
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Hwnd = w.hwnd
	nid.UID = 1
	nid.Flags = nifMessage | nifIcon | nifTip
	nid.CallbackMessage = wmTrayIcon
	nid.HIcon = icon
	utf16Copy(nid.SzTip[:], tooltip)

	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))

	w.trayNID = nid
	w.trayActive = true
}

func (w *Window) DisableTray() {
	if !w.trayActive {
		return
	}

	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&w.trayNID)))
	w.trayActive = false
}

func (w *Window) showTrayMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	showText, err := syscall.UTF16PtrFromString(w.trayShowLabel)
	if err != nil {
		return
	}

	exitText, err := syscall.UTF16PtrFromString(w.trayExitLabel)
	if err != nil {
		return
	}

	procAppendMenuW.Call(menu, mfString, trayCmdShow, uintptr(unsafe.Pointer(showText)))
	procAppendMenuW.Call(menu, mfString, trayCmdExit, uintptr(unsafe.Pointer(exitText)))

	var pt pointT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	procSetForegroundWindow.Call(w.hwnd)
	cmd, _, _ := procTrackPopupMenu.Call(menu, tpmRightButton|tpmReturnCmd, uintptr(pt.X), uintptr(pt.Y), w.hwnd, 0)
	procPostMessageW.Call(w.hwnd, wmNull, 0, 0)

	switch cmd {
	case trayCmdShow:
		if w.onTrayShow != nil {
			w.onTrayShow()
		}
	case trayCmdExit:
		if w.onTrayExit != nil {
			w.onTrayExit()
		}
	}
}
