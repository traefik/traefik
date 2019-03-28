package staert

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/containous/flaeg"
)

var _ Source = (*TomlSource)(nil)

// TomlSource implement staert.Source
type TomlSource struct {
	filename     string
	dirNFullPath []string
	fullPath     string
}

// NewTomlSource creates and return a pointer on Source.
// Parameter filename is the file name (without extension type, ".toml" will be added)
// dirNFullPath may contain directories or fullPath to the file.
func NewTomlSource(filename string, dirNFullPath []string) *TomlSource {
	return &TomlSource{filename, dirNFullPath, ""}
}

// ConfigFileUsed return config file used
func (ts *TomlSource) ConfigFileUsed() string {
	return ts.fullPath
}

// Parse calls toml.DecodeFile() func
func (ts *TomlSource) Parse(cmd *flaeg.Command) (*flaeg.Command, error) {
	ts.fullPath = findFile(ts.filename, ts.dirNFullPath)
	if len(ts.fullPath) < 2 {
		return cmd, nil
	}

	metadata, err := toml.DecodeFile(ts.fullPath, cmd.Config)
	if err != nil {
		return nil, err
	}

	boolFlags, err := flaeg.GetBoolFlags(cmd.Config)
	if err != nil {
		return nil, err
	}

	flgArgs, hasUnderField := generateArgs(metadata, boolFlags)

	err = flaeg.Load(cmd.Config, cmd.DefaultPointersConfig, flgArgs)
	if err != nil && err != flaeg.ErrParserNotFound {
		return nil, err
	}

	if hasUnderField {
		_, err := toml.DecodeFile(ts.fullPath, cmd.Config)
		if err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

func preProcessDir(dirIn string) (string, error) {
	expanded := os.ExpandEnv(dirIn)
	return filepath.Abs(expanded)
}

func findFile(filename string, dirNFile []string) string {
	for _, df := range dirNFile {
		if df != "" {
			fullPath, _ := preProcessDir(df)
			if fileInfo, err := os.Stat(fullPath); err == nil && !fileInfo.IsDir() {
				return fullPath
			}

			fullPath = filepath.Join(fullPath, filename+".toml")
			if fileInfo, err := os.Stat(fullPath); err == nil && !fileInfo.IsDir() {
				return fullPath
			}
		}
	}
	return ""
}

func generateArgs(metadata toml.MetaData, flags []string) ([]string, bool) {
	var flgArgs []string
	keys := metadata.Keys()
	hasUnderField := false

	for i, key := range keys {
		if metadata.Type(key.String()) == "Hash" {
			// TOML hashes correspond to Go structs or maps.
			for j := i; j < len(keys); j++ {
				if strings.Contains(keys[j].String(), key.String()+".") {
					hasUnderField = true
					break
				}
			}

			match := false
			for _, flag := range flags {
				if flag == strings.ToLower(key.String()) {
					match = true
					break
				}
			}
			if match {
				flgArgs = append(flgArgs, "--"+strings.ToLower(key.String()))
			}
		}
	}

	return flgArgs, hasUnderField
}
