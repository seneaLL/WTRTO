//go:build !linux && !windows

package platform

func ScreenBounds() Rect {
	return Rect{X: 0, Y: 0, W: fallbackW, H: fallbackH}
}
