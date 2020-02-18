package cli

import (
	"os"
	"path/filepath"
	"strings"

	config "github.com/containous/traefik/v2/pkg/config/file"
	"github.com/containous/traefik/v2/pkg/config/flag"
	"github.com/containous/traefik/v2/pkg/log"
)

// DirLoader loads configuration files from a directory, merging them into a single config.
type DirLoader struct{}

// Load loads the command's configuration from the config directory.
func (e *DirLoader) Load(args []string, cmd *Command) (bool, error) {
	ref, err := flag.Parse(args, cmd.Configuration)
	if err != nil {
		_ = cmd.PrintHelp(os.Stdout)
		return false, err
	}

	configFileFlag := "traefik.configfile"
	if _, ok := ref["traefik.configFile"]; ok {
		configFileFlag = "traefik.configFile"
	}

	if configDir, ok := ref[configFileFlag]; ok {
		if fileInfo, err := os.Stat(configDir); err != nil || !fileInfo.IsDir() {
			return false, err
		}

		logger := log.WithoutContext()
		logger.Printf("Loading configuration from directory: %s", configDir)

		if err := filepath.Walk(configDir, walkFunc(cmd.Configuration, logger)); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func walkFunc(cf interface{}, logger log.Logger) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Warnf("Failed to access path %q: %v", path, err)
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		currFileExt := strings.ToLower(filepath.Ext(path))
		switch currFileExt {
		case ".toml", ".yml", ".yaml":
			break
		default:
			if !info.IsDir() {
				logger.Infof("Ignoring file %s", file.Name())
			}
			return nil
		}

		logger.Infof("Loading config file %s", file.Name())
		if err := config.Decode(file.Name(), cf); err != nil {
			return err
		}

		return nil
	}
}
