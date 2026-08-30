package dialog

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"
)

var ErrUnavailable = errors.New("диалог выбора файла недоступен")

var (
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")

	procGetOpenFileName = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileName = comdlg32.NewProc("GetSaveFileNameW")
)

const (
	ofnPathMustExist   = 0x00000800
	ofnFileMustExist   = 0x00001000
	ofnOverwritePrompt = 0x00000002
)

type openFileName struct {
	StructSize    uint32
	Owner         uintptr
	Instance      uintptr
	Filter        *uint16
	CustomFilter  *uint16
	MaxCustFilter uint32
	FilterIndex   uint32
	File          *uint16
	MaxFile       uint32
	FileTitle     *uint16
	MaxFileTitle  uint32
	InitialDir    *uint16
	Title         *uint16
	Flags         uint32
	FileOffset    uint16
	FileExtension uint16
	DefExt        *uint16
	CustData      uintptr
	FnHook        uintptr
	TemplateName  *uint16
	PvReserved    uintptr
	DwReserved    uint32
	FlagsEx       uint32
}

func OpenFile(title string) (string, error) {
	buf := make([]uint16, 4096)

	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return "", err
	}

	filterPtr, err := syscall.UTF16PtrFromString("JSON\x00*.json\x00\x00")
	if err != nil {
		return "", err
	}

	ofn := openFileName{
		Owner:   0,
		Filter:  filterPtr,
		File:    &buf[0],
		MaxFile: uint32(len(buf)),
		Title:   titlePtr,
		Flags:   ofnPathMustExist | ofnFileMustExist,
	}
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))

	r, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", nil
	}

	return syscall.UTF16ToString(buf), nil
}

func SaveFile(title, defaultName string) (string, error) {
	buf := make([]uint16, 4096)

	defNameUTF16, err := syscall.UTF16FromString(defaultName)
	if err != nil {
		return "", err
	}
	copy(buf, defNameUTF16)

	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return "", err
	}

	filterPtr, err := syscall.UTF16PtrFromString("JSON\x00*.json\x00\x00")
	if err != nil {
		return "", err
	}

	ofn := openFileName{
		Owner:   0,
		Filter:  filterPtr,
		File:    &buf[0],
		MaxFile: uint32(len(buf)),
		Title:   titlePtr,
		Flags:   ofnOverwritePrompt,
	}
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))

	r, _, _ := procGetSaveFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", nil
	}

	path := syscall.UTF16ToString(buf)
	if !strings.Contains(path, ".") {
		path += ".json"
	}

	return path, nil
}
