// Package provider holds the different provider implementation.
package provider

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"errors"
	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/emilevauge/backoff"
)

// Kv holds common configurations of key-value providers.
type Kv struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string     `description:"Comma sepparated server endpoints"`
	Prefix       string     `description:"Prefix used for KV store"`
	TLS          *ClientTLS `description:"Enable TLS support"`
	storeType    store.Backend
	kvclient     store.Store
}

func (provider *Kv) createStore() (store.Store, error) {
	storeConfig := &store.Config{
		ConnectionTimeout: 30 * time.Second,
		Bucket:            "traefik",
	}

	if provider.TLS != nil {
		var err error
		storeConfig.TLS, err = provider.TLS.CreateTLSConfig()
		if err != nil {
			return nil, err
		}
	}
	return libkv.NewStore(
		provider.storeType,
		strings.Split(provider.Endpoint, ","),
		storeConfig,
	)
}

func (provider *Kv) watchKv(configurationChan chan<- types.ConfigMessage, prefix string, stop chan bool) error {
	operation := func() error {
		events, err := provider.kvclient.WatchTree(provider.Prefix, make(chan struct{}))
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
				configuration := provider.loadConfig()
				if configuration != nil {
					configurationChan <- types.ConfigMessage{
						ProviderName:  string(provider.storeType),
						Configuration: configuration,
					}
				}
			}
		}
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("KV connection error: %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(operation, backoff.NewJobBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		return fmt.Errorf("Cannot connect to KV server: %v", err)
	}
	return nil
}

func (provider *Kv) provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints []types.Constraint) error {
	provider.Constraints = append(provider.Constraints, constraints...)
	storeConfig := &store.Config{
		ConnectionTimeout: 30 * time.Second,
		Bucket:            "traefik",
	}

	if provider.TLS != nil {
		caPool := x509.NewCertPool()

		if provider.TLS.CA != "" {
			ca, err := ioutil.ReadFile(provider.TLS.CA)

			if err != nil {
				return fmt.Errorf("Failed to read CA. %s", err)
			}

			caPool.AppendCertsFromPEM(ca)
		}

		cert, err := tls.LoadX509KeyPair(provider.TLS.Cert, provider.TLS.Key)

		if err != nil {
			return fmt.Errorf("Failed to load TLS keypair: %v", err)
		}

		storeConfig.TLS = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caPool,
			InsecureSkipVerify: provider.TLS.InsecureSkipVerify,
		}
	}

	operation := func() error {
		if _, err := provider.kvclient.Exists("qmslkjdfmqlskdjfmqlksjazçueznbvbwzlkajzebvkwjdcqmlsfj"); err != nil {
			return fmt.Errorf("Failed to test KV store connection: %v", err)
		}
		if provider.Watch {
			pool.Go(func(stop chan bool) {
				err := provider.watchKv(configurationChan, provider.Prefix, stop)
				if err != nil {
					log.Errorf("Cannot watch KV store: %v", err)
				}
			})
		}
		configuration := provider.loadConfig()
		configurationChan <- types.ConfigMessage{
			ProviderName:  string(provider.storeType),
			Configuration: configuration,
		}
		return nil
	}
	notify := func(err error, time time.Duration) {
		log.Errorf("KV connection error: %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(operation, backoff.NewJobBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		return fmt.Errorf("Cannot connect to KV server: %v", err)
	}
	return nil
}

func (provider *Kv) loadConfig() *types.Configuration {
	templateObjects := struct {
		Prefix string
	}{
		// Allow `/traefik/alias` to superesede `provider.Prefix`
		strings.TrimSuffix(provider.get(provider.Prefix, provider.Prefix+"/alias"), "/"),
	}

	var KvFuncMap = template.FuncMap{
		"List":             provider.list,
		"Get":              provider.get,
		"SplitGet":         provider.splitGet,
		"Last":             provider.last,
		"CheckConstraints": provider.checkConstraints,
	}

	configuration, err := provider.getConfiguration("templates/kv.tmpl", KvFuncMap, templateObjects)
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

func (provider *Kv) list(keys ...string) []string {
	joinedKeys := strings.Join(keys, "")
	keysPairs, err := provider.kvclient.List(joinedKeys)
	if err != nil {
		log.Errorf("Error getting keys %s %s ", joinedKeys, err)
		return nil
	}
	directoryKeys := make(map[string]string)
	for _, key := range keysPairs {
		directory := strings.Split(strings.TrimPrefix(key.Key, joinedKeys), "/")[0]
		directoryKeys[directory] = joinedKeys + directory
	}
	return fun.Values(directoryKeys).([]string)
}

func (provider *Kv) get(defaultValue string, keys ...string) string {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := provider.kvclient.Get(strings.TrimPrefix(joinedKeys, "/"))
	if err != nil {
		log.Warnf("Error getting key %s %s, setting default %s", joinedKeys, err, defaultValue)
		return defaultValue
	} else if keyPair == nil {
		log.Warnf("Error getting key %s, setting default %s", joinedKeys, defaultValue)
		return defaultValue
	}
	return string(keyPair.Value)
}

func (provider *Kv) splitGet(keys ...string) []string {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := provider.kvclient.Get(joinedKeys)
	if err != nil {
		log.Warnf("Error getting key %s %s, setting default empty", joinedKeys, err)
		return []string{}
	} else if keyPair == nil {
		log.Warnf("Error getting key %s, setting default %empty", joinedKeys)
		return []string{}
	}
	return strings.Split(string(keyPair.Value), ",")
}

func (provider *Kv) last(key string) string {
	splittedKey := strings.Split(key, "/")
	return splittedKey[len(splittedKey)-1]
}

func (provider *Kv) checkConstraints(keys ...string) string {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := provider.kvclient.Get(joinedKeys)

	value := ""
	if err == nil && keyPair != nil && keyPair.Value != nil {
		value = string(keyPair.Value)
	}

	constraintTags := strings.Split(value, ",")
	ok, failingConstraint := provider.MatchConstraints(constraintTags)
	if ok == false {
		if failingConstraint != nil {
			log.Debugf("Constraint %v not matching with following tags: %v", failingConstraint.String(), value)
		}
		return "false"
	}
	return "true"
}
