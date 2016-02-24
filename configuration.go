package main

import (
	"errors"
	"fmt"
	fmtlog "log"
	"regexp"
	"strings"
	"time"

	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// GlobalConfiguration holds global configuration (with providers, etc.).
// It's populated from the traefik configuration file passed as an argument to the binary.
type GlobalConfiguration struct {
	GraceTimeOut              int64
	AccessLogsFile            string
	TraefikLogsFile           string
	LogLevel                  string
	EntryPoints               EntryPoints
	DefaultEntryPoints        DefaultEntryPoints
	ProvidersThrottleDuration time.Duration
	MaxIdleConnsPerHost       int
	Docker                    *provider.Docker
	File                      *provider.File
	Web                       *WebProvider
	Marathon                  *provider.Marathon
	Consul                    *provider.Consul
	ConsulCatalog             *provider.ConsulCatalog
	Etcd                      *provider.Etcd
	Zookeeper                 *provider.Zookepper
	Boltdb                    *provider.BoltDb
}

// DefaultEntryPoints holds default entry points
type DefaultEntryPoints []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (dep *DefaultEntryPoints) String() string {
	return fmt.Sprintf("%#v", dep)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (dep *DefaultEntryPoints) Set(value string) error {
	entrypoints := strings.Split(value, ",")
	if len(entrypoints) == 0 {
		return errors.New("Bad DefaultEntryPoints format: " + value)
	}
	for _, entrypoint := range entrypoints {
		*dep = append(*dep, entrypoint)
	}
	return nil
}

// Type is type of the struct
func (dep *DefaultEntryPoints) Type() string {
	return fmt.Sprint("defaultentrypoints²")
}

// EntryPoints holds entry points configuration of the reverse proxy (ip, port, TLS...)
type EntryPoints map[string]*EntryPoint

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (ep *EntryPoints) String() string {
	return ""
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (ep *EntryPoints) Set(value string) error {
	regex := regexp.MustCompile("(?:Name:(?P<Name>\\S*))\\s*(?:Address:(?P<Address>\\S*))?\\s*(?:TLS:(?P<TLS>\\S*))?\\s*(?:Redirect.EntryPoint:(?P<RedirectEntryPoint>\\S*))?\\s*(?:Redirect.Regex:(?P<RedirectRegex>\\S*))?\\s*(?:Redirect.Replacement:(?P<RedirectReplacement>\\S*))?")
	match := regex.FindAllStringSubmatch(value, -1)
	if match == nil {
		return errors.New("Bad EntryPoints format: " + value)
	}
	matchResult := match[0]
	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			result[name] = matchResult[i]
		}
	}
	var tls *TLS
	if len(result["TLS"]) > 0 {
		certs := Certificates{}
		certs.Set(result["TLS"])
		tls = &TLS{
			Certificates: certs,
		}
	}
	var redirect *Redirect
	if len(result["RedirectEntryPoint"]) > 0 || len(result["RedirectRegex"]) > 0 || len(result["RedirectReplacement"]) > 0 {
		redirect = &Redirect{
			EntryPoint:  result["RedirectEntryPoint"],
			Regex:       result["RedirectRegex"],
			Replacement: result["RedirectReplacement"],
		}
	}

	(*ep)[result["Name"]] = &EntryPoint{
		Address:  result["Address"],
		TLS:      tls,
		Redirect: redirect,
	}

	return nil
}

// Type is type of the struct
func (ep *EntryPoints) Type() string {
	return fmt.Sprint("entrypoints²")
}

// EntryPoint holds an entry point configuration of the reverse proxy (ip, port, TLS...)
type EntryPoint struct {
	Network  string
	Address  string
	TLS      *TLS
	Redirect *Redirect
}

// Redirect configures a redirection of an entry point to another, or to an URL
type Redirect struct {
	EntryPoint  string
	Regex       string
	Replacement string
}

// TLS configures TLS for an entry point
type TLS struct {
	Certificates Certificates
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
	return new(GlobalConfiguration)
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
	viper.AddConfigPath("/etc/traefik/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/.traefik/") // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		fmtlog.Fatalf("Error reading file: %s", err)
	}

	if len(arguments.EntryPoints) > 0 {
		viper.Set("entryPoints", arguments.EntryPoints)
	}
	if len(arguments.DefaultEntryPoints) > 0 {
		viper.Set("defaultEntryPoints", arguments.DefaultEntryPoints)
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
	if arguments.consulCatalog {
		viper.Set("consulCatalog", arguments.ConsulCatalog)
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
	if err := unmarshal(&configuration); err != nil {
		fmtlog.Fatalf("Error reading file: %s", err)
	}

	if len(configuration.EntryPoints) == 0 {
		configuration.EntryPoints = make(map[string]*EntryPoint)
		configuration.EntryPoints["http"] = &EntryPoint{
			Address: ":80",
		}
		configuration.DefaultEntryPoints = []string{"http"}
	}

	if configuration.File != nil && len(configuration.File.Filename) == 0 {
		// no filename, setting to global config file
		configuration.File.Filename = viper.ConfigFileUsed()
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
