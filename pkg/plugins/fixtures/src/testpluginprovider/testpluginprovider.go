package testpluginprovider

import (
	"context"
	"encoding/json"
	"errors"
)

type Config struct {
	RouterName string
}

func CreateConfig() *Config {
	return &Config{}
}

type Provider struct {
	routerName string
	cancel     func()
}

func New(ctx context.Context, config *Config, name string) (*Provider, error) {
	return &Provider{routerName: config.RouterName}, nil
}

func (p *Provider) Init() error {
	if p.routerName == "" {
		return errors.New("missing router name")
	}

	return nil
}

func (p *Provider) Provide(cfgChan chan<- json.Marshaler) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go func() {
		select {
		case cfgChan <- configuration{routerName: p.routerName}:
		case <-ctx.Done():
		}
	}()

	return nil
}

func (p *Provider) Stop() error {
	p.cancel()
	return nil
}

type configuration struct {
	routerName string
}

func (c configuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"http": map[string]any{
			"routers": map[string]any{
				c.routerName: map[string]any{"rule": "Host(`example.com`)", "service": "noop@internal"},
			},
		},
	})
}
