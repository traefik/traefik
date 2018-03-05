package acme

import "os"

// Check file content size
// Do not check file permissions on Windows right now
func checkFile(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false, err
	}

	return fi.Size() > 0, nil
}
