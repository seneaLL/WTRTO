package platform

import "syscall"

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
)

const (
	smXVirtualScreen  = 76
	smYVirtualScreen  = 77
	smCXVirtualScreen = 78
	smCYVirtualScreen = 79
)

func getMetric(index int) int {
	r, _, _ := procGetSystemMetrics.Call(uintptr(index))

	return int(r)
}

func ScreenBounds() Rect {
	w := getMetric(smCXVirtualScreen)
	h := getMetric(smCYVirtualScreen)
	if w <= 0 || h <= 0 {
		return Rect{X: 0, Y: 0, W: fallbackW, H: fallbackH}
	}

	return Rect{
		X: getMetric(smXVirtualScreen),
		Y: getMetric(smYVirtualScreen),
		W: w,
		H: h,
	}
}
