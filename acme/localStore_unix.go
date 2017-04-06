// +build !windows

package acme

import (
	"fmt"
	"os"
)

// Check file permissions
func checkPermissions(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("permissions %o for %s are too open, please use 600", fi.Mode().Perm(), name)
	}
	return nil
}
