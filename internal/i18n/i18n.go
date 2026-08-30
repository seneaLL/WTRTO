package i18n

import (
	"os"
	"strings"
	"sync"
)

type Lang string

const (
	EN Lang = "en"
	RU Lang = "ru"
)

var (
	mu      sync.RWMutex
	current = detectDefault()
)

var strings_ = map[Lang]map[string]string{
	EN: {
		"app.title":             "WTRTO",
		"app.subtitle":          "War Thunder Real-Time Overlay",
		"status.on":             "Overlay is running",
		"status.off":            "Overlay is off",
		"button.enable":         "Enable overlay",
		"button.disable":        "Disable overlay",
		"debug.checkbox":        "Show debug info in overlay",
		"menu.settings":         "Settings",
		"menu.language":         "Language",
		"overlay.title":         "WTRTO overlay",
		"overlay.show_fps":      "Show FPS in the top-left corner",
		"debug.title":           "DEBUG",
		"debug.fps":             "Overlay FPS",
		"debug.cpu_app":         "App CPU",
		"debug.mem_app":         "App MEM",
		"debug.cpu_wt":          "War Thunder CPU",
		"debug.mem_wt":          "War Thunder MEM",
		"debug.gpu":             "GPU load (system)",
		"debug.gpu_mem":         "GPU memory",
		"debug.gpu_unavailable": "GPU: no data (nvidia-smi not found)",
		"debug.wt_not_running":  "War Thunder: not running",
		"settings.fps_limit":    "Overlay FPS limit",
		"fps.unlimited":         "Unlimited",
		"hud.edit_mode":         "Edit HUD layout",
		"hud.edit_hint":         "Drag elements in the overlay while enabled (aircraft only)",
		"hud.edit_done":         "Done editing",
	},
	RU: {
		"app.title":             "WTRTO",
		"app.subtitle":          "War Thunder Real-Time Overlay",
		"status.on":             "Оверлей включён",
		"status.off":            "Оверлей выключён",
		"button.enable":         "Включить оверлей",
		"button.disable":        "Выключить оверлей",
		"debug.checkbox":        "Показывать отладочную информацию в оверлее",
		"menu.settings":         "Настройки",
		"menu.language":         "Язык",
		"overlay.title":         "WTRTO overlay",
		"overlay.show_fps":      "Показывать FPS в левом верхнем углу",
		"debug.title":           "ОТЛАДКА",
		"debug.fps":             "FPS оверлея",
		"debug.cpu_app":         "CPU приложения",
		"debug.mem_app":         "MEM приложения",
		"debug.cpu_wt":          "CPU War Thunder",
		"debug.mem_wt":          "MEM War Thunder",
		"debug.gpu":             "Загрузка GPU (система)",
		"debug.gpu_mem":         "Память GPU",
		"debug.gpu_unavailable": "GPU: нет данных (nvidia-smi не найден)",
		"debug.wt_not_running":  "War Thunder: не запущен",
		"settings.fps_limit":    "Лимит FPS оверлея",
		"fps.unlimited":         "Без лимита",
		"hud.edit_mode":         "Редактировать раскладку HUD",
		"hud.edit_hint":         "Перетаскивай элементы в оверлее (только для самолётов)",
		"hud.edit_done":         "Готово",
	},
}

func detectDefault() Lang {
	env := os.Getenv("LC_ALL")
	if env == "" {
		env = os.Getenv("LANG")
	}
	if strings.HasPrefix(strings.ToLower(env), "ru") {
		return RU
	}

	return EN
}

func Set(l Lang) {
	mu.Lock()
	defer mu.Unlock()
	if l != EN && l != RU {
		return
	}
	current = l
}

func Current() Lang {
	mu.RLock()
	defer mu.RUnlock()

	return current
}

func T(key string) string {
	mu.RLock()
	l := current
	mu.RUnlock()

	if m, ok := strings_[l]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}

	return strings_[EN][key]
}
