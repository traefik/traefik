//go:build !windows
// +build !windows

package acme

import (
	"fmt"
	"os"
)

// CheckFile checks file permissions and content size.
func CheckFile(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil && os.IsNotExist(err) {
		nf, err := os.Create(name)
		if err != nil {
			return false, err
		}
		defer nf.Close()
		return false, nf.Chmod(0o600)
	}
	if err != nil {
		return false, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false, err
	}

	if fi.Mode().Perm()&0o077 != 0 {
		return false, fmt.Errorf("permissions %o for %s are too open, please use 600", fi.Mode().Perm(), name)
	}

	return fi.Size() > 0, nil
}
