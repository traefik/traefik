package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/kvtools/redis"
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
	DB       int              `description:"Database to be selected after connecting to the server." json:"db,omitempty" toml:"db,omitempty" yaml:"db,omitempty"`
	Sentinel *Sentinel        `description:"Enable Sentinel support." json:"sentinel,omitempty" toml:"sentinel,omitempty" yaml:"sentinel,omitempty" export:"true"`
}

// Sentinel holds the Redis Sentinel configuration.
type Sentinel struct {
	MasterName string `description:"Name of the master." json:"masterName,omitempty" toml:"masterName,omitempty" yaml:"masterName,omitempty" export:"true"`
	Username   string `description:"Username for Sentinel authentication." json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty" loggable:"false"`
	Password   string `description:"Password for Sentinel authentication." json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" loggable:"false"`

	LatencyStrategy bool `description:"Defines whether to route commands to the closest master or replica nodes (mutually exclusive with RandomStrategy and ReplicaStrategy)." json:"latencyStrategy,omitempty" toml:"latencyStrategy,omitempty" yaml:"latencyStrategy,omitempty" export:"true"`
	RandomStrategy  bool `description:"Defines whether to route commands randomly to master or replica nodes (mutually exclusive with LatencyStrategy and ReplicaStrategy)." json:"randomStrategy,omitempty" toml:"randomStrategy,omitempty" yaml:"randomStrategy,omitempty" export:"true"`
	ReplicaStrategy bool `description:"Defines whether to route all commands to replica nodes (mutually exclusive with LatencyStrategy and RandomStrategy)." json:"replicaStrategy,omitempty" toml:"replicaStrategy,omitempty" yaml:"replicaStrategy,omitempty" export:"true"`

	UseDisconnectedReplicas bool `description:"Use replicas disconnected with master when cannot get connected replicas." json:"useDisconnectedReplicas,omitempty" toml:"useDisconnectedReplicas,omitempty" yaml:"useDisconnectedReplicas,omitempty" export:"true"`
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

	if p.Sentinel != nil {
		switch {
		case p.Sentinel.LatencyStrategy && !(p.Sentinel.RandomStrategy || p.Sentinel.ReplicaStrategy):
		case p.Sentinel.RandomStrategy && !(p.Sentinel.LatencyStrategy || p.Sentinel.ReplicaStrategy):
		case p.Sentinel.ReplicaStrategy && !(p.Sentinel.RandomStrategy || p.Sentinel.LatencyStrategy):
			return errors.New("latencyStrategy, randomStrategy and replicaStrategy options are mutually exclusive, please use only one of those options")
		}

		clusterClient := p.Sentinel.LatencyStrategy || p.Sentinel.RandomStrategy
		config.Sentinel = &redis.Sentinel{
			MasterName:              p.Sentinel.MasterName,
			Username:                p.Sentinel.Username,
			Password:                p.Sentinel.Password,
			ClusterClient:           clusterClient,
			RouteByLatency:          p.Sentinel.LatencyStrategy,
			RouteRandomly:           p.Sentinel.RandomStrategy,
			ReplicaOnly:             p.Sentinel.ReplicaStrategy,
			UseDisconnectedReplicas: p.Sentinel.UseDisconnectedReplicas,
		}
	}

	return p.Provider.Init(redis.StoreName, "redis", config)
}
