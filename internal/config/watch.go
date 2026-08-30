package config

import "time"

func Watch(interval time.Duration, stop <-chan struct{}, onChange func(State)) {
	last := Load()
	onChange(last)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			cur := Load()
			if cur != last {
				last = cur
				onChange(cur)
			}
		}
	}
}
