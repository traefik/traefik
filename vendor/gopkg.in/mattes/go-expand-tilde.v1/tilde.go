package tilde

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ErrNoHome = errors.New("no home found")
)

func Expand(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := Home()
	if err != nil {
		return "", err
	}

	return home + path[1:], nil
}

func Home() (string, error) {
	home := ""

	switch runtime.GOOS {
	case "windows":
		home = filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
		if home == "" {
			home = os.Getenv("UserProfile")
		}

	default:
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", ErrNoHome
	}
	return home, nil
}
