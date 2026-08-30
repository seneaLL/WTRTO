package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	DebugInfo   bool   `json:"debug_info"`
	Language    string `json:"language,omitempty"`
	FPSLimit    int    `json:"fps_limit,omitempty"`
	HUDEditMode bool   `json:"hud_edit_mode,omitempty"`
	HideFPS     bool   `json:"hide_fps,omitempty"`

	HotkeyKeysym uint   `json:"hotkey_keysym,omitempty"`
	HotkeyMods   uint   `json:"hotkey_mods,omitempty"`
	HotkeyLabel  string `json:"hotkey_label,omitempty"`

	ActiveTemplate string `json:"active_template,omitempty"`
}

const DefaultFPSLimit = 60

func (s State) EffectiveFPSLimit() int {
	if s.FPSLimit == 0 {
		return DefaultFPSLimit
	}

	return s.FPSLimit
}

func path() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(exe), "state.json"), nil
}

func Load() State {
	p, err := path()
	if err != nil {
		return State{}
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return State{}
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}
	}

	return s
}

func Save(s State) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p, data, 0o644)
}
