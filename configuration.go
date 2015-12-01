package main

import (
	fmtlog "log"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/emilevauge/traefik/provider"
	"github.com/emilevauge/traefik/types"
)

// GlobalConfiguration holds global configuration (with providers, etc.).
// It's populated from the traefik configuration file passed as an argument to the binary.
type GlobalConfiguration struct {
	Port                      string
	GraceTimeOut              int64
	AccessLogsFile            string
	TraefikLogsFile           string
	Certificates              []Certificate
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

// LoadFileConfig returns a GlobalConfiguration from reading the specified file (a toml file).
func LoadFileConfig(file string) *GlobalConfiguration {
	configuration := NewGlobalConfiguration()
	if _, err := toml.DecodeFile(file, configuration); err != nil {
		fmtlog.Fatalf("Error reading file: %s", err)
	}
	return configuration
}

type configs map[string]*types.Configuration
