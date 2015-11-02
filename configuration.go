package main

import (
	"time"

	"github.com/emilevauge/traefik/provider"
	"github.com/emilevauge/traefik/types"
)

type GlobalConfiguration struct {
	Port                      string
	GraceTimeOut              int64
	AccessLogsFile            string
	TraefikLogsFile           string
	CertFile, KeyFile         string
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

func NewGlobalConfiguration() *GlobalConfiguration {
	globalConfiguration := new(GlobalConfiguration)
	// default values
	globalConfiguration.Port = ":80"
	globalConfiguration.GraceTimeOut = 10
	globalConfiguration.LogLevel = "ERROR"
	globalConfiguration.ProvidersThrottleDuration = time.Duration(2 * time.Second)

	return globalConfiguration
}

type configs map[string]*types.Configuration
