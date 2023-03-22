package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/kvtools/etcdv3"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/provider/kv"
	"github.com/traefik/traefik/v3/pkg/types"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider `yaml:",inline" export:"true"`

	TLS      *types.ClientTLS `description:"Enable TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	Username string           `description:"Username for authentication." json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" loggable:"false"`
	Password string           `description:"Password for authentication." json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" loggable:"false"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Provider.SetDefaults()
	p.Endpoints = []string{"127.0.0.1:2379"}
}

// Init the provider.
func (p *Provider) Init() error {
	config := &etcdv3.Config{
		ConnectionTimeout: 3 * time.Second,
		Username:          p.Username,
		Password:          p.Password,
	}

	if p.TLS != nil {
		var err error
		config.TLS, err = p.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return fmt.Errorf("unable to create client TLS configuration: %w", err)
		}
	}

	return p.Provider.Init(etcdv3.StoreName, "etcd", config)
}
