package cli

import (
	"io/ioutil"
	"os"

	"github.com/containous/traefik/pkg/config/file"
	"github.com/containous/traefik/pkg/config/flag"
	"github.com/containous/traefik/pkg/log"
)

// FileLoader loads configuration from file.
type FileLoader struct {
	filename string
}

// GetFilename returns the configuration file if any.
func (f *FileLoader) GetFilename() string {
	return f.filename
}

// Load loads the configuration.
func (f *FileLoader) Load(args []string, cmd *Command) (bool, error) {
	ref, err := flag.Parse(args, cmd.Configuration)
	if err != nil {
		_ = PrintHelp(os.Stdout, cmd)
		return false, err
	}

	// FIXME
	configFile, err := loadConfigFiles(ref["traefik.configfile"], cmd.Configuration)
	if err != nil {
		return false, err
	}

	f.filename = configFile

	if configFile == "" {
		return false, nil
	}

	logger := log.WithoutContext()
	logger.Printf("Configuration loaded from file: %s", configFile)

	content, _ := ioutil.ReadFile(configFile)
	logger.Debug(string(content))

	return true, nil
}

// loadConfigFiles tries to decode the given configuration file and all default locations for the configuration file.
// It stops as soon as decoding one of them is successful.
func loadConfigFiles(configFile string, element interface{}) (string, error) {
	finder := Finder{
		BasePaths:  []string{"/etc/traefik/traefik", "$XDG_CONFIG_HOME/traefik", "$HOME/.config/traefik", "./traefik"},
		Extensions: []string{"toml", "yaml", "yml"},
	}

	filePath, err := finder.Find(configFile)
	if err != nil {
		return "", err
	}

	if len(filePath) == 0 {
		return "", nil
	}

	if err = file.Decode(filePath, element, "http", "tcp", "tls", "TLSOptions", "TLSStores"); err != nil {
		return "", err
	}
	return filePath, nil
}
