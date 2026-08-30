package platform

import (
	"os/exec"
	"regexp"
	"strconv"
)

var screenRe = regexp.MustCompile(`Screen \d+:.*current (\d+) x (\d+)`)

func ScreenBounds() Rect {
	out, err := exec.Command("xrandr", "--query").Output()
	if err == nil {
		if m := screenRe.FindSubmatch(out); m != nil {
			w, err1 := strconv.Atoi(string(m[1]))
			h, err2 := strconv.Atoi(string(m[2]))
			if err1 == nil && err2 == nil && w > 0 && h > 0 {
				return Rect{X: 0, Y: 0, W: w, H: h}
			}
		}
	}

	return Rect{X: 0, Y: 0, W: fallbackW, H: fallbackH}
}
