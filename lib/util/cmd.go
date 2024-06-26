package util

import (
	"fmt"
	"log/slog"
	"os/exec"
)

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func CheckExecutableExists(executable string) bool {
	path, err := exec.LookPath(executable)
	if err != nil {
		slog.Warn(fmt.Sprintf("`%s` executable not found", executable))
		return false
	} else {
		slog.Debug(fmt.Sprintf("`%s` executable found at `%s`", executable, path))
		return true
	}
}

func RunCommand(name string, args ...string) ([]byte, error) {
	exists := CheckExecutableExists(name)
	if !exists {
		return nil, nil
	}
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		ErrAttr(err)
		return nil, err
	} else {
		return output, nil
	}
}
