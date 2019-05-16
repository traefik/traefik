package cli

import (
	"os"
	"path/filepath"
	"strings"
)

// Finder finds a first available file path.
type Finder struct {
	BasePaths  []string
	Extensions []string
}

// Find finds a first available file path.
func (f Finder) Find(configFile string) (string, error) {
	paths := f.getPaths(configFile)

	for _, filePath := range paths {
		fp := os.ExpandEnv(filePath)

		_, err := os.Stat(fp)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}

		return filepath.Abs(fp)
	}

	return "", nil
}

func (f Finder) getPaths(configFile string) []string {
	var paths []string
	if strings.TrimSpace(configFile) != "" {
		paths = append(paths, configFile)
	}

	for _, basePath := range f.BasePaths {
		for _, ext := range f.Extensions {
			paths = append(paths, basePath+"."+ext)
		}
	}

	return paths
}
