package kv

import (
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
)

// Provider holds common configurations of key-value providers.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	Endpoint              string              `description:"Comma separated server endpoints"`
	Prefix                string              `description:"Prefix used for KV store"`
	TLS                   *provider.ClientTLS `description:"Enable TLS support"`
	StoreType             store.Backend
	Kvclient              store.Store
}

// CreateStore create the K/V store
func (p *Provider) CreateStore() (store.Store, error) {
	storeConfig := &store.Config{
		ConnectionTimeout: 30 * time.Second,
		Bucket:            "traefik",
	}

	if p.TLS != nil {
		var err error
		storeConfig.TLS, err = p.TLS.CreateTLSConfig()
		if err != nil {
			return nil, err
		}
	}
	return libkv.NewStore(
		p.StoreType,
		strings.Split(p.Endpoint, ","),
		storeConfig,
	)
}

func (p *Provider) watchKv(configurationChan chan<- types.ConfigMessage, prefix string, stop chan bool) error {
	operation := func() error {
		events, err := p.Kvclient.WatchTree(p.Prefix, make(chan struct{}))
		if err != nil {
			return fmt.Errorf("Failed to KV WatchTree: %v", err)
		}
		for {
			select {
			case <-stop:
				return nil
			case _, ok := <-events:
				if !ok {
					return errors.New("watchtree channel closed")
				}
				configuration := p.loadConfig()
				if configuration != nil {
					configurationChan <- types.ConfigMessage{
						ProviderName:  string(p.StoreType),
						Configuration: configuration,
					}
				}
			}
		}
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("KV connection error: %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		return fmt.Errorf("Cannot connect to KV server: %v", err)
	}
	return nil
}

// Provide provides the configuration to traefik via the configuration channel
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	p.Constraints = append(p.Constraints, constraints...)
	operation := func() error {
		if _, err := p.Kvclient.Exists("qmslkjdfmqlskdjfmqlksjazçueznbvbwzlkajzebvkwjdcqmlsfj"); err != nil {
			return fmt.Errorf("Failed to test KV store connection: %v", err)
		}
		if p.Watch {
			pool.Go(func(stop chan bool) {
				err := p.watchKv(configurationChan, p.Prefix, stop)
				if err != nil {
					log.Errorf("Cannot watch KV store: %v", err)
				}
			})
		}
		configuration := p.loadConfig()
		configurationChan <- types.ConfigMessage{
			ProviderName:  string(p.StoreType),
			Configuration: configuration,
		}
		return nil
	}
	notify := func(err error, time time.Duration) {
		log.Errorf("KV connection error: %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		return fmt.Errorf("Cannot connect to KV server: %v", err)
	}
	return nil
}

func (p *Provider) loadConfig() *types.Configuration {
	templateObjects := struct {
		Prefix string
	}{
		// Allow `/traefik/alias` to superesede `p.Prefix`
		strings.TrimSuffix(p.get(p.Prefix, p.Prefix+"/alias"), "/"),
	}

	var KvFuncMap = template.FuncMap{
		"List":        p.list,
		"ListServers": p.listServers,
		"Get":         p.get,
		"SplitGet":    p.splitGet,
		"Last":        p.last,
	}

	configuration, err := p.GetConfiguration("templates/kv.tmpl", KvFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	for key, frontend := range configuration.Frontends {
		if _, ok := configuration.Backends[frontend.Backend]; ok == false {
			delete(configuration.Frontends, key)
		}
	}

	return configuration
}

func (p *Provider) list(keys ...string) []string {
	joinedKeys := strings.Join(keys, "")
	keysPairs, err := p.Kvclient.List(joinedKeys)
	if err != nil {
		log.Debugf("Cannot get keys %s %s ", joinedKeys, err)
		return nil
	}
	directoryKeys := make(map[string]string)
	for _, key := range keysPairs {
		directory := strings.Split(strings.TrimPrefix(key.Key, joinedKeys), "/")[0]
		directoryKeys[directory] = joinedKeys + directory
	}
	return fun.Values(directoryKeys).([]string)
}

func (p *Provider) listServers(backend string) []string {
	serverNames := p.list(backend, "/servers/")
	return fun.Filter(func(serverName string) bool {
		key := fmt.Sprint(serverName, "/url")
		if _, err := p.Kvclient.Get(key); err != nil {
			if err != store.ErrKeyNotFound {
				log.Errorf("Failed to retrieve value for key %s: %s", key, err)
			}
			return false
		}
		return p.checkConstraints(serverName, "/tags")
	}, serverNames).([]string)
}

func (p *Provider) get(defaultValue string, keys ...string) string {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := p.Kvclient.Get(strings.TrimPrefix(joinedKeys, "/"))
	if err != nil {
		log.Debugf("Cannot get key %s %s, setting default %s", joinedKeys, err, defaultValue)
		return defaultValue
	} else if keyPair == nil {
		log.Debugf("Cannot get key %s, setting default %s", joinedKeys, defaultValue)
		return defaultValue
	}
	return string(keyPair.Value)
}

func (p *Provider) splitGet(keys ...string) []string {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := p.Kvclient.Get(joinedKeys)
	if err != nil {
		log.Debugf("Cannot get key %s %s, setting default empty", joinedKeys, err)
		return []string{}
	} else if keyPair == nil {
		log.Debugf("Cannot get key %s, setting default %empty", joinedKeys)
		return []string{}
	}
	return strings.Split(string(keyPair.Value), ",")
}

func (p *Provider) last(key string) string {
	splittedKey := strings.Split(key, "/")
	return splittedKey[len(splittedKey)-1]
}

func (p *Provider) checkConstraints(keys ...string) bool {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := p.Kvclient.Get(joinedKeys)

	value := ""
	if err == nil && keyPair != nil && keyPair.Value != nil {
		value = string(keyPair.Value)
	}

	constraintTags := strings.Split(value, ",")
	ok, failingConstraint := p.MatchConstraints(constraintTags)
	if ok == false {
		if failingConstraint != nil {
			log.Debugf("Constraint %v not matching with following tags: %v", failingConstraint.String(), value)
		}
		return false
	}
	return true
}
