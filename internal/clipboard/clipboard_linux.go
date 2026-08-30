package clipboard

import (
	"bytes"
	"errors"
	"os/exec"
)

var ErrUnavailable = errors.New("буфер обмена недоступен: не найден xclip или xsel")

func copyArgs() []string {
	if _, err := exec.LookPath("xclip"); err == nil {
		return []string{"xclip", "-selection", "clipboard"}
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		return []string{"xsel", "--clipboard", "--input"}
	}

	return nil
}

func pasteArgs() []string {
	if _, err := exec.LookPath("xclip"); err == nil {
		return []string{"xclip", "-selection", "clipboard", "-o"}
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		return []string{"xsel", "--clipboard", "--output"}
	}

	return nil
}

func Available() bool {
	return copyArgs() != nil
}

func Copy(text string) error {
	args := copyArgs()
	if args == nil {
		return ErrUnavailable
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = bytes.NewReader([]byte(text))

	return cmd.Run()
}

func Paste() (string, error) {
	args := pasteArgs()
	if args == nil {
		return "", ErrUnavailable
	}
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
