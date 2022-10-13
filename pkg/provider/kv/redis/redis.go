package redis

import (
	"context"
	"fmt"

	"github.com/kvtools/redis"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/kv"
	"github.com/traefik/traefik/v2/pkg/types"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	kv.Provider `yaml:",inline" export:"true"`

	TLS      *types.ClientTLS `description:"Enable TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	Username string           `description:"Username for authentication." json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" loggable:"false"`
	Password string           `description:"Password for authentication." json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" loggable:"false"`
	DB       int              `description:"Database to be selected after connecting to the server." json:"db,omitempty" toml:"db,omitempty" yaml:"db,omitempty"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Provider.SetDefaults()
	p.Endpoints = []string{"127.0.0.1:6379"}
}

// Init the provider.
func (p *Provider) Init() error {
	config := &redis.Config{
		Username: p.Username,
		Password: p.Password,
		DB:       p.DB,
	}

	if p.TLS != nil {
		var err error
		config.TLS, err = p.TLS.CreateTLSConfig(context.Background())
		if err != nil {
			return fmt.Errorf("unable to create client TLS configuration: %w", err)
		}
	}

	return p.Provider.Init(redis.StoreName, "redis", config)
}
