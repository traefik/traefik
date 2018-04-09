package acme

import "os"

// CheckFile checks file content size
// Do not check file permissions on Windows right now
func CheckFile(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(name)
			if err != nil {
				return false, err
			}
			return false, f.Chmod(0600)
		}
		return false, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false, err
	}

	return fi.Size() > 0, nil
}
