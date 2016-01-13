package main

import (
	fmtlog "log"
	"time"

	"fmt"
	"github.com/emilevauge/traefik/provider"
	"github.com/emilevauge/traefik/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/wendal/errors"
	"strings"
)

// GlobalConfiguration holds global configuration (with providers, etc.).
// It's populated from the traefik configuration file passed as an argument to the binary.
type GlobalConfiguration struct {
	Port                      string
	GraceTimeOut              int64
	AccessLogsFile            string
	TraefikLogsFile           string
	Certificates              Certificates
	LogLevel                  string
	ProvidersThrottleDuration time.Duration
	Docker                    *provider.Docker
	File                      *provider.File
	Web                       *WebProvider
	Marathon                  *provider.Marathon
	Consul                    *provider.Consul
	Etcd                      *provider.Etcd
	Zookeeper                 *provider.Zookepper
	Boltdb                    *provider.BoltDb
}

// Certificates defines traefik certificates type
type Certificates []Certificate

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (certs *Certificates) String() string {
	if len(*certs) == 0 {
		return ""
	}
	return (*certs)[0].CertFile + "," + (*certs)[0].KeyFile
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (certs *Certificates) Set(value string) error {
	files := strings.Split(value, ",")
	if len(files) != 2 {
		return errors.New("Bad certificates format: " + value)
	}
	*certs = append(*certs, Certificate{
		CertFile: files[0],
		KeyFile:  files[1],
	})
	return nil
}

// Type is type of the struct
func (certs *Certificates) Type() string {
	return fmt.Sprint("certificates")
}

// Certificate holds a SSL cert/key pair
type Certificate struct {
	CertFile string
	KeyFile  string
}

// NewGlobalConfiguration returns a GlobalConfiguration with default values.
func NewGlobalConfiguration() *GlobalConfiguration {
	globalConfiguration := new(GlobalConfiguration)
	// default values
	globalConfiguration.Port = ":80"
	globalConfiguration.GraceTimeOut = 10
	globalConfiguration.LogLevel = "ERROR"
	globalConfiguration.ProvidersThrottleDuration = time.Duration(2 * time.Second)

	return globalConfiguration
}

// LoadConfiguration returns a GlobalConfiguration.
func LoadConfiguration() *GlobalConfiguration {
	configuration := NewGlobalConfiguration()
	viper.SetEnvPrefix("traefik")
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	if len(viper.GetString("configFile")) > 0 {
		viper.SetConfigFile(viper.GetString("configFile"))
	} else {
		viper.SetConfigName("traefik") // name of config file (without extension)
	}
	viper.AddConfigPath("/etc/traefik/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.traefik") // call multiple times to add many search paths
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		fmtlog.Fatalf("Error reading file: %s", err)
	}
	if len(arguments.Certificates) > 0 {
		viper.Set("certificates", arguments.Certificates)
	}
	if arguments.web {
		viper.Set("web", arguments.Web)
	}
	if arguments.file {
		viper.Set("file", arguments.File)
	}
	if !arguments.dockerTLS {
		arguments.Docker.TLS = nil
	}
	if arguments.docker {
		viper.Set("docker", arguments.Docker)
	}
	if arguments.marathon {
		viper.Set("marathon", arguments.Marathon)
	}
	if arguments.consul {
		viper.Set("consul", arguments.Consul)
	}
	if arguments.zookeeper {
		viper.Set("zookeeper", arguments.Zookeeper)
	}
	if arguments.etcd {
		viper.Set("etcd", arguments.Etcd)
	}
	if arguments.boltdb {
		viper.Set("boltdb", arguments.Boltdb)
	}
	err = unmarshal(&configuration)
	if err != nil {
		fmtlog.Fatalf("Error reading file: %s", err)
	}

	return configuration
}

func unmarshal(rawVal interface{}) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
		Metadata:         nil,
		Result:           rawVal,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	err = decoder.Decode(viper.AllSettings())
	if err != nil {
		return err
	}
	return nil
}

type configs map[string]*types.Configuration
