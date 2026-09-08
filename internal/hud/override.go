package hud

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AircraftOverride struct {
	BaseTemplate string    `json:"base_template"`
	Aircraft     string    `json:"aircraft"`
	Elements     []Element `json:"elements"`
}

func aircraftDir() (string, error) {
	dir, err := templatesDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "aircraft"), nil
}

func overridesDir() (string, error) {
	dir, err := templatesDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "overrides"), nil
}

func aircraftTemplatePath(aircraft string) (string, error) {
	dir, err := aircraftDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, slug(aircraft)+".json"), nil
}

func overridePath(base, aircraft string) (string, error) {
	dir, err := overridesDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, slug(base)+"__"+slug(aircraft)+".json"), nil
}

func LoadStandalone(aircraft string) (Template, bool) {
	p, err := aircraftTemplatePath(aircraft)
	if err != nil {
		return Template{}, false
	}
	t, err := loadFile(p)
	if err != nil {
		return Template{}, false
	}

	return t, true
}

func ListStandalone() ([]Template, error) {
	dir, err := aircraftDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	var out []Template
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		t, err := loadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Aircraft < out[j].Aircraft })

	return out, nil
}

func DeleteStandalone(aircraft string) error {
	p, err := aircraftTemplatePath(aircraft)
	if err != nil {
		return err
	}

	return os.Remove(p)
}

func LoadOverride(base, aircraft string) (AircraftOverride, bool) {
	p, err := overridePath(base, aircraft)
	if err != nil {
		return AircraftOverride{}, false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return AircraftOverride{}, false
	}
	var o AircraftOverride
	if err := json.Unmarshal(data, &o); err != nil {
		return AircraftOverride{}, false
	}

	return o, true
}

func SaveOverride(o AircraftOverride) error {
	p, err := overridePath(o.BaseTemplate, o.Aircraft)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p, data, 0o644)
}

func DeleteOverride(base, aircraft string) error {
	p, err := overridePath(base, aircraft)
	if err != nil {
		return err
	}

	return os.Remove(p)
}

func ListOverrides(base string) ([]string, error) {
	dir, err := overridesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	prefix := slug(base) + "__"
	var aircraft []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var o AircraftOverride
		if err := json.Unmarshal(data, &o); err != nil {
			continue
		}
		aircraft = append(aircraft, o.Aircraft)
	}
	sort.Strings(aircraft)

	return aircraft, nil
}

func Resolve(base Template, aircraft string) (elements []Element, locked int) {
	if aircraft == "" {
		return base.Elements, 0
	}
	if t, ok := LoadStandalone(aircraft); ok {
		return t.Elements, 0
	}
	if o, ok := LoadOverride(base.Name, aircraft); ok {
		combined := make([]Element, 0, len(base.Elements)+len(o.Elements))
		combined = append(combined, base.Elements...)
		combined = append(combined, o.Elements...)

		return combined, len(base.Elements)
	}

	return base.Elements, 0
}
