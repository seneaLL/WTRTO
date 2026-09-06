package hud

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/seneaLL/WTRTO/internal/i18n"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slug(name string) string {
	s := slugRe.ReplaceAllString(strings.ToLower(name), "_")

	return strings.Trim(s, "_")
}

func templatesDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(exe), "configs"), nil
}

func templatePath(name string) (string, error) {
	dir, err := templatesDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, slug(name)+".json"), nil
}

func Load(name string) (Template, error) {
	p, err := templatePath(name)
	if err != nil {
		return Template{}, err
	}

	return loadFile(p)
}

func loadFile(p string) (Template, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return Template{}, err
	}
	var t Template
	if err := json.Unmarshal(data, &t); err != nil {
		return Template{}, err
	}

	return t, nil
}

func Save(t Template) error {
	p, err := templatePath(t.Name)
	if err != nil {
		return err
	}

	return saveFile(p, t)
}

func saveFile(p string, t Template) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p, data, 0o644)
}

func EnsureDefault(def Template) Template {
	if t, err := Load(def.Name); err == nil {
		return t
	}
	Save(def)

	return def
}

func EnsureBuiltins() {
	for _, t := range BuiltinTemplates() {
		p, err := templatePath(t.Name)
		if err != nil {
			continue
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			saveFile(p, t)
		}
	}
}

func List() ([]string, error) {
	dir, err := templatesDir()
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
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		t, err := loadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		names = append(names, t.Name)
	}
	sort.Strings(names)

	return names, nil
}

func Import(srcPath string) (Template, error) {
	t, err := loadFile(srcPath)
	if err != nil {
		return Template{}, err
	}
	if t.Name == "" {
		t.Name = strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	}
	if err := Save(t); err != nil {
		return Template{}, err
	}

	return t, nil
}

func Export(t Template, destPath string) error {
	return saveFile(destPath, t)
}

type builtinTemplateErr struct{}

func (builtinTemplateErr) Error() string { return i18n.T("error.builtin_template_delete") }

var errBuiltinTemplate error = builtinTemplateErr{}

func Delete(name string) error {
	if IsBuiltin(name) {
		return errBuiltinTemplate
	}
	p, err := templatePath(name)
	if err != nil {
		return err
	}

	return os.Remove(p)
}
