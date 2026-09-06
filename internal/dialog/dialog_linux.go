package dialog

import (
	"os/exec"
	"strings"

	"github.com/seneaLL/WTRTO/internal/i18n"
)

type unavailableErr struct{}

func (unavailableErr) Error() string { return i18n.T("error.dialog_unavailable_linux") }

var ErrUnavailable error = unavailableErr{}

func hasZenity() bool {
	_, err := exec.LookPath("zenity")

	return err == nil
}

func hasKdialog() bool {
	_, err := exec.LookPath("kdialog")

	return err == nil
}

func run(cmd *exec.Cmd) (string, error) {
	out, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(string(out)), nil
}

func OpenFile(title string) (string, error) {
	switch {
	case hasZenity():
		return run(exec.Command("zenity", "--file-selection", "--title="+title))

	case hasKdialog():
		return run(exec.Command("kdialog", "--getopenfilename", ".", "*.json", "--title", title))
	}

	return "", ErrUnavailable
}

func SaveFile(title, defaultName string) (string, error) {
	switch {
	case hasZenity():
		return run(exec.Command("zenity", "--file-selection", "--save", "--confirm-overwrite", "--title="+title, "--filename="+defaultName))

	case hasKdialog():
		return run(exec.Command("kdialog", "--getsavefilename", defaultName, "--title", title))
	}

	return "", ErrUnavailable
}
