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
		"debug.section_app":     "APPLICATION",
		"debug.section_wt":      "WAR THUNDER",
		"debug.section_system":  "SYSTEM",
		"debug.cpu":             "CPU",
		"debug.ram":             "RAM",
		"debug.gpu_short":       "GPU",
		"debug.gpu_mem":         "Memory",
		"debug.gpu_unavailable": "GPU: no data (nvidia-smi not found)",
		"debug.wt_not_running":  "War Thunder is not running",
		"settings.fps_limit":    "Overlay FPS limit",
		"fps.vsync":             "VSync",
		"hud.edit_mode":         "Edit HUD layout",
		"hud.edit_hint":         "Drag elements in the overlay while enabled (aircraft only)",
		"hud.edit_done":         "Done editing",
		"tray.show":             "Show",
		"tray.exit":             "Exit",

		"hotkey.label":     "Overlay toggle hotkey",
		"hotkey.capturing": "Press a key combination…",
		"update.available": "Update available",
		"limits.updated":   "Aircraft limits data updated",

		"error.invalid_share_code":        "invalid code: corrupted or not a WTRTO code",
		"error.dialog_unavailable":        "file dialog unavailable",
		"error.dialog_unavailable_linux":  "file dialog unavailable: zenity or kdialog not found",
		"error.clipboard_unavailable":     "clipboard unavailable",
		"error.clipboard_unavailable_win": "clipboard unavailable",
		"error.clipboard_unavailable_nix": "clipboard unavailable: xclip or xsel not found",
		"error.builtin_template_delete":   "the built-in template cannot be deleted, copy it first",

		"editor.copy_suffix":                    " (copy)",
		"editor.clipboard_unavailable":          "Clipboard unavailable (no xclip/xsel)",
		"editor.tab_template":                   "Template",
		"editor.tab_element":                    "Element",
		"editor.active_template":                "Active template",
		"editor.load_failed":                    "Failed to load template",
		"editor.loaded_prefix":                  "Loaded: ",
		"editor.fill_background":                "Fill background",
		"editor.readonly_notice":                "Built-in template - view only. Save a copy below to edit.",
		"editor.save_as":                        "Save as",
		"editor.save_as_copy_tip":               "Save as a copy of the template",
		"editor.enter_template_name":            "Enter a template name",
		"editor.name_reserved":                  "This name is reserved for a built-in template, choose another",
		"editor.save_error_prefix":              "Save error: ",
		"editor.saved_as_prefix":                "Saved as: ",
		"editor.new_blank_template_tip":         "New blank template",
		"editor.new_template_name_fmt":          "New template %d",
		"editor.created_prefix":                 "Created: ",
		"editor.export_tip":                     "Export template to file",
		"editor.import_tip":                     "Import template from file",
		"editor.delete_template_tip":            "Delete template",
		"editor.delete_error_prefix":            "Delete error: ",
		"editor.deleted_prefix":                 "Deleted: ",
		"editor.template_code":                  "Template code",
		"editor.copy_code_tip":                  "Copy template code to clipboard",
		"editor.generate_code_error_prefix":     "Code generation error: ",
		"editor.code_ready_no_clipboard":        "Code is ready (copy it manually from the field above - clipboard unavailable)",
		"editor.code_copied":                    "Code copied to clipboard",
		"editor.load_from_code_tip":             "Load template from the code above",
		"editor.paste_code_first":               "Paste the code in the field above",
		"editor.loaded_from_code_prefix":        "Loaded from code: ",
		"editor.add_element":                    "+ Add element",
		"editor.select_element_hint":            "Select an element on screen to edit",
		"editor.readonly_no_elements":           "Elements cannot be selected in a built-in template",
		"editor.align_h_tip":                    "Align center horizontally",
		"editor.align_v_tip":                    "Align center vertically",
		"editor.type":                           "Type",
		"editor.value":                          "Value",
		"editor.auto_color":                     "Auto color by limit",
		"editor.label":                          "Label",
		"editor.units":                          "Units",
		"editor.precision":                      "Precision",
		"editor.bold":                           "Bold",
		"editor.font_size":                      "Font size",
		"editor.thickness":                      "Thickness",
		"editor.style":                          "Style",
		"editor.direction":                      "Direction",
		"editor.side":                           "Side",
		"editor.length":                         "Length",
		"editor.range":                          "Range",
		"editor.minor_step":                     "Minor step",
		"editor.major_step":                     "Major step",
		"editor.zones":                          "Zones",
		"editor.delete_zone_tip":                "Delete zone",
		"editor.add_zone_tip":                   "Add zone",
		"editor.color":                          "Color",
		"editor.glow":                           "Glow",
		"editor.glow_use_own_color":             "Element color",
		"editor.glow_color":                     "Glow color",
		"editor.glow_intensity":                 "Intensity, %",
		"editor.delete":                         "Delete",
		"editor.export_dialog_title":            "Export template",
		"editor.save_dialog_unavailable_prefix": "Save dialog unavailable: ",
		"editor.export_error_prefix":            "Export error: ",
		"editor.exported_to_prefix":             "Exported to ",
		"editor.import_dialog_title":            "Import template",
		"editor.open_dialog_unavailable_prefix": "File dialog unavailable: ",
		"editor.import_error_prefix":            "Import error: ",
		"editor.imported_prefix":                "Imported: ",
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
		"debug.section_app":     "ПРИЛОЖЕНИЕ",
		"debug.section_wt":      "WAR THUNDER",
		"debug.section_system":  "СИСТЕМА",
		"debug.cpu":             "CPU",
		"debug.ram":             "RAM",
		"debug.gpu_short":       "GPU",
		"debug.gpu_mem":         "Память",
		"debug.gpu_unavailable": "GPU: нет данных (nvidia-smi не найден)",
		"debug.wt_not_running":  "War Thunder не запущен",
		"settings.fps_limit":    "Лимит FPS оверлея",
		"fps.vsync":             "VSync",
		"hud.edit_mode":         "Редактировать раскладку HUD",
		"hud.edit_hint":         "Перетаскивай элементы в оверлее (только для самолётов)",
		"hud.edit_done":         "Готово",
		"tray.show":             "Показать",
		"tray.exit":             "Выход",

		"hotkey.label":     "Хоткей вкл/выкл оверлея",
		"hotkey.capturing": "Нажмите комбинацию клавиш…",
		"update.available": "Доступно обновление",
		"limits.updated":   "Обновлены данные о лимитах самолётов",

		"error.invalid_share_code":        "некорректный код: повреждён или это не код WTRTO",
		"error.dialog_unavailable":        "диалог выбора файла недоступен",
		"error.dialog_unavailable_linux":  "диалог выбора файла недоступен: не найден zenity или kdialog",
		"error.clipboard_unavailable":     "буфер обмена недоступен",
		"error.clipboard_unavailable_win": "буфер обмена недоступен",
		"error.clipboard_unavailable_nix": "буфер обмена недоступен: не найден xclip или xsel",
		"error.builtin_template_delete":   "стандартный шаблон нельзя удалить, сначала скопируйте его",

		"editor.copy_suffix":                    " (копия)",
		"editor.clipboard_unavailable":          "Буфер обмена недоступен (нет xclip/xsel)",
		"editor.tab_template":                   "Шаблон",
		"editor.tab_element":                    "Элемент",
		"editor.active_template":                "Активный шаблон",
		"editor.load_failed":                    "Не удалось загрузить шаблон",
		"editor.loaded_prefix":                  "Загружен: ",
		"editor.fill_background":                "Закрашивать фон",
		"editor.readonly_notice":                "Стандартный шаблон - только просмотр. Сохраните копию ниже, чтобы редактировать.",
		"editor.save_as":                        "Сохранить как",
		"editor.save_as_copy_tip":               "Сохранить как копию шаблона",
		"editor.enter_template_name":            "Введите имя шаблона",
		"editor.name_reserved":                  "Это имя зарезервировано под стандартный шаблон, выберите другое",
		"editor.save_error_prefix":              "Ошибка сохранения: ",
		"editor.saved_as_prefix":                "Сохранено как: ",
		"editor.new_blank_template_tip":         "Новый пустой шаблон",
		"editor.new_template_name_fmt":          "Новый шаблон %d",
		"editor.created_prefix":                 "Создан: ",
		"editor.export_tip":                     "Экспорт шаблона в файл",
		"editor.import_tip":                     "Импорт шаблона из файла",
		"editor.delete_template_tip":            "Удалить шаблон",
		"editor.delete_error_prefix":            "Ошибка удаления: ",
		"editor.deleted_prefix":                 "Удалён: ",
		"editor.template_code":                  "Код шаблона",
		"editor.copy_code_tip":                  "Скопировать код шаблона в буфер обмена",
		"editor.generate_code_error_prefix":     "Ошибка генерации кода: ",
		"editor.code_ready_no_clipboard":        "Код готов (скопируйте вручную из поля выше - буфер обмена недоступен)",
		"editor.code_copied":                    "Код скопирован в буфер обмена",
		"editor.load_from_code_tip":             "Загрузить шаблон из кода выше",
		"editor.paste_code_first":               "Вставьте код в поле выше",
		"editor.loaded_from_code_prefix":        "Загружено по коду: ",
		"editor.add_element":                    "+ Добавить элемент",
		"editor.select_element_hint":            "Выберите элемент на экране для редактирования",
		"editor.readonly_no_elements":           "Элементы недоступны для выбора в стандартном шаблоне",
		"editor.align_h_tip":                    "Выровнять по центру по горизонтали",
		"editor.align_v_tip":                    "Выровнять по центру по вертикали",
		"editor.type":                           "Тип",
		"editor.value":                          "Величина",
		"editor.auto_color":                     "Авто-цвет по лимиту",
		"editor.label":                          "Подпись",
		"editor.units":                          "Единицы",
		"editor.precision":                      "Точность",
		"editor.bold":                           "Жирный",
		"editor.font_size":                      "Размер шрифта",
		"editor.thickness":                      "Толщина",
		"editor.style":                          "Стиль",
		"editor.direction":                      "Направление",
		"editor.side":                           "Сторона",
		"editor.length":                         "Длина",
		"editor.range":                          "Диапазон",
		"editor.minor_step":                     "Малый шаг",
		"editor.major_step":                     "Большой шаг",
		"editor.zones":                          "Зоны",
		"editor.delete_zone_tip":                "Удалить зону",
		"editor.add_zone_tip":                   "Добавить зону",
		"editor.color":                          "Цвет",
		"editor.glow":                           "Свечение",
		"editor.glow_use_own_color":             "Цвет элемента",
		"editor.glow_color":                     "Цвет свечения",
		"editor.glow_intensity":                 "Интенсивность, %",
		"editor.delete":                         "Удалить",
		"editor.export_dialog_title":            "Экспорт шаблона",
		"editor.save_dialog_unavailable_prefix": "Диалог сохранения недоступен: ",
		"editor.export_error_prefix":            "Ошибка экспорта: ",
		"editor.exported_to_prefix":             "Экспортировано в ",
		"editor.import_dialog_title":            "Импорт шаблона",
		"editor.open_dialog_unavailable_prefix": "Диалог выбора файла недоступен: ",
		"editor.import_error_prefix":            "Ошибка импорта: ",
		"editor.imported_prefix":                "Импортировано: ",
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
