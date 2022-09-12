package zk

import (
	"time"

	"github.com/kvtools/zookeeper"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/kv"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider `yaml:",inline" export:"true"`

	Username string `description:"Username for authentication." json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" loggable:"false"`
	Password string `description:"Password for authentication." json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" loggable:"false"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Provider.SetDefaults()
	p.Endpoints = []string{"127.0.0.1:2181"}
}

// Init the provider.
func (p *Provider) Init() error {
	config := &zookeeper.Config{
		ConnectionTimeout: 3 * time.Second,
		Username:          p.Username,
		Password:          p.Password,
	}

	return p.Provider.Init(zookeeper.StoreName, "zookeeper", config)
}
