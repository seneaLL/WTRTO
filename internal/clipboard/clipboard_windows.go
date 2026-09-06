package clipboard

import (
	"syscall"
	"unsafe"

	"github.com/seneaLL/WTRTO/internal/i18n"
)

type unavailableErr struct{}

func (unavailableErr) Error() string { return i18n.T("error.clipboard_unavailable_win") }

var ErrUnavailable error = unavailableErr{}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalFree       = kernel32.NewProc("GlobalFree")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

func Available() bool {
	return true
}

func Copy(text string) error {
	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return ErrUnavailable
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	utf16Text, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	size := len(utf16Text) * 2
	hMem, _, _ := procGlobalAlloc.Call(gmemMoveable, uintptr(size))
	if hMem == 0 {
		return ErrUnavailable
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		procGlobalFree.Call(hMem)

		return ErrUnavailable
	}

	dst := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(utf16Text))
	copy(dst, utf16Text)
	procGlobalUnlock.Call(hMem)

	if r, _, _ := procSetClipboardData.Call(cfUnicodeText, hMem); r == 0 {
		procGlobalFree.Call(hMem)

		return ErrUnavailable
	}

	return nil
}

func Paste() (string, error) {
	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return "", ErrUnavailable
	}
	defer procCloseClipboard.Call()

	h, _, _ := procGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return "", ErrUnavailable
	}

	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return "", ErrUnavailable
	}
	defer procGlobalUnlock.Call(h)

	var length int
	for {
		u := *(*uint16)(unsafe.Pointer(ptr + uintptr(length*2)))
		if u == 0 {
			break
		}

		length++
	}

	slice := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), length)

	return syscall.UTF16ToString(slice), nil
}
