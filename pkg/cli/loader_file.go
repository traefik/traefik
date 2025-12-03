package cli

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/file"
	"github.com/traefik/paerser/flag"
)

// FileLoader loads a configuration from a file.
type FileLoader struct {
	ConfigFileFlag string
	filename       string
}

// GetFilename returns the configuration file if any.
func (f *FileLoader) GetFilename() string {
	return f.filename
}

// Load loads the command's configuration from a file either specified with the -baqup.configfile flag, or from default locations.
func (f *FileLoader) Load(args []string, cmd *cli.Command) (bool, error) {
	ref, err := flag.Parse(args, cmd.Configuration)
	if err != nil {
		_ = cmd.PrintHelp(os.Stdout)
		return false, err
	}

	configFileFlag := "baqup.configfile"
	if _, ok := ref["baqup.configFile"]; ok {
		configFileFlag = "baqup.configFile"
	}

	if f.ConfigFileFlag != "" {
		configFileFlag = "baqup." + f.ConfigFileFlag
		if _, ok := ref[strings.ToLower(configFileFlag)]; ok {
			configFileFlag = "baqup." + strings.ToLower(f.ConfigFileFlag)
		}
	}

	configFile, err := loadConfigFiles(ref[configFileFlag], cmd.Configuration)
	if err != nil {
		return false, err
	}

	f.filename = configFile

	if configFile == "" {
		return false, nil
	}

	log.Printf("Configuration loaded from file: %s", configFile)

	content, _ := os.ReadFile(configFile)
	log.Debug().Str("configFile", configFile).Bytes("content", content).Send()

	return true, nil
}

// loadConfigFiles tries to decode the given configuration file and all default locations for the configuration file.
// It stops as soon as decoding one of them is successful.
func loadConfigFiles(configFile string, element interface{}) (string, error) {
	finder := cli.Finder{
		BasePaths:  []string{"/etc/baqup/baqup", "$XDG_CONFIG_HOME/baqup", "$HOME/.config/baqup", "./baqup"},
		Extensions: []string{"toml", "yaml", "yml"},
	}

	filePath, err := finder.Find(configFile)
	if err != nil {
		return "", err
	}

	if len(filePath) == 0 {
		return "", nil
	}

	if err := file.Decode(filePath, element); err != nil {
		return "", err
	}
	return filePath, nil
}
