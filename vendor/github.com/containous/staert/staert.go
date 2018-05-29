package staert

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/containous/flaeg"
)

// Source interface must be satisfy to Add any kink of Source to Staert as like as TomlFile or Flaeg
type Source interface {
	Parse(cmd *flaeg.Command) (*flaeg.Command, error)
}

// Staert contains the struct to configure, thee default values inside structs and the sources
type Staert struct {
	command *flaeg.Command
	sources []Source
}

// NewStaert creates and return a pointer on Staert. Need defaultConfig and defaultPointersConfig given by references
func NewStaert(rootCommand *flaeg.Command) *Staert {
	s := Staert{
		command: rootCommand,
	}
	return &s
}

// AddSource adds new Source to Staert, give it by reference
func (s *Staert) AddSource(src Source) {
	s.sources = append(s.sources, src)
}

// getConfig for a flaeg.Command run sources Parse func in the raw
func (s *Staert) parseConfigAllSources(cmd *flaeg.Command) error {
	for _, src := range s.sources {
		var err error
		_, err = src.Parse(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig check which command is called and parses config
// It returns the the parsed config or an error if it fails
func (s *Staert) LoadConfig() (interface{}, error) {
	for _, src := range s.sources {
		//Type assertion
		f, ok := src.(*flaeg.Flaeg)
		if ok {
			if fCmd, err := f.GetCommand(); err != nil {
				return nil, err
			} else if s.command != fCmd {
				//IF fleag sub-command
				if fCmd.Metadata["parseAllSources"] == "true" {
					//IF parseAllSources
					fCmdConfigType := reflect.TypeOf(fCmd.Config)
					sCmdConfigType := reflect.TypeOf(s.command.Config)
					if fCmdConfigType != sCmdConfigType {
						return nil, fmt.Errorf("command %s : Config type doesn't match with root command config type. Expected %s got %s", fCmd.Name, sCmdConfigType.Name(), fCmdConfigType.Name())
					}
					s.command = fCmd
				} else {
					// ELSE (not parseAllSources)
					s.command, err = f.Parse(fCmd)
					return s.command.Config, err
				}
			}
		}
	}
	err := s.parseConfigAllSources(s.command)
	return s.command.Config, err
}

// Run calls the Run func of the command
// Warning, Run doesn't parse the config
func (s *Staert) Run() error {
	return s.command.Run()
}

//TomlSource impement Source
type TomlSource struct {
	filename     string
	dirNfullpath []string
	fullpath     string
}

// NewTomlSource creates and return a pointer on TomlSource.
// Parameter filename is the file name (without extension type, ".toml" will be added)
// dirNfullpath may contain directories or fullpath to the file.
func NewTomlSource(filename string, dirNfullpath []string) *TomlSource {
	return &TomlSource{filename, dirNfullpath, ""}
}

// ConfigFileUsed return config file used
func (ts *TomlSource) ConfigFileUsed() string {
	return ts.fullpath
}

func preprocessDir(dirIn string) (string, error) {
	dirOut := dirIn
	expanded := os.ExpandEnv(dirIn)
	dirOut, err := filepath.Abs(expanded)
	return dirOut, err
}

func findFile(filename string, dirNfile []string) string {
	for _, df := range dirNfile {
		if df != "" {
			fullPath, _ := preprocessDir(df)
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

// Parse calls toml.DecodeFile() func
func (ts *TomlSource) Parse(cmd *flaeg.Command) (*flaeg.Command, error) {
	ts.fullpath = findFile(ts.filename, ts.dirNfullpath)
	if len(ts.fullpath) < 2 {
		return cmd, nil
	}
	metadata, err := toml.DecodeFile(ts.fullpath, cmd.Config)
	if err != nil {
		return nil, err
	}
	boolFlags, err := flaeg.GetBoolFlags(cmd.Config)
	if err != nil {
		return nil, err
	}
	flaegArgs, hasUnderField, err := generateArgs(metadata, boolFlags)
	if err != nil {
		return nil, err
	}

	// fmt.Println(flaegArgs)
	err = flaeg.Load(cmd.Config, cmd.DefaultPointersConfig, flaegArgs)
	//if err!= missing parser err
	if err != nil && err != flaeg.ErrParserNotFound {
		return nil, err
	}
	if hasUnderField {
		_, err := toml.DecodeFile(ts.fullpath, cmd.Config)
		if err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

func generateArgs(metadata toml.MetaData, flags []string) ([]string, bool, error) {
	var flaegArgs []string
	keys := metadata.Keys()
	hasUnderField := false
	for i, key := range keys {
		// fmt.Println(key)
		if metadata.Type(key.String()) == "Hash" {
			// TOML hashes correspond to Go structs or maps.
			// fmt.Printf("%s could be a ptr on a struct, or a map\n", key)
			for j := i; j < len(keys); j++ {
				// fmt.Printf("%s =? %s\n", keys[j].String(), "."+key.String())
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
				flaegArgs = append(flaegArgs, "--"+strings.ToLower(key.String()))
			}
		}
	}
	return flaegArgs, hasUnderField, nil
}
