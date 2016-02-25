// Package provider holds the different provider implementation.
package provider

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
)

// Kv holds common configurations of key-value providers.
type Kv struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string
	Prefix       string
	TLS          *KvTLS
	storeType    store.Backend
	kvclient     store.Store
}

// KvTLS holds TLS specific configurations
type KvTLS struct {
	CA                 string
	Cert               string
	Key                string
	InsecureSkipVerify bool
}

func (provider *Kv) provide(configurationChan chan<- types.ConfigMessage) error {
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
			return fmt.Errorf("Failed to load keypair. %s", err)
		}

		storeConfig.TLS = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caPool,
			InsecureSkipVerify: provider.TLS.InsecureSkipVerify,
		}
	}

	kv, err := libkv.NewStore(
		provider.storeType,
		[]string{provider.Endpoint},
		storeConfig,
	)
	if err != nil {
		return err
	}
	if _, err := kv.List(""); err != nil {
		return err
	}
	provider.kvclient = kv
	if provider.Watch {
		stopCh := make(chan struct{})
		chanKeys, err := kv.WatchTree(provider.Prefix, stopCh)
		if err != nil {
			return err
		}
		go func() {
			for {
				<-chanKeys
				configuration := provider.loadConfig()
				if configuration != nil {
					configurationChan <- types.ConfigMessage{
						ProviderName:  string(provider.storeType),
						Configuration: configuration,
					}
				}
				defer close(stopCh)
			}
		}()
	}
	configuration := provider.loadConfig()
	configurationChan <- types.ConfigMessage{
		ProviderName:  string(provider.storeType),
		Configuration: configuration,
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
		"List":     provider.list,
		"Get":      provider.get,
		"SplitGet": provider.splitGet,
		"Last":     provider.last,
	}

	configuration, err := provider.getConfiguration("templates/kv.tmpl", KvFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
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
		directory := strings.Split(strings.TrimPrefix(key.Key, strings.TrimPrefix(joinedKeys, "/")), "/")[0]
		directoryKeys[directory] = joinedKeys + directory
	}
	return fun.Values(directoryKeys).([]string)
}

func (provider *Kv) get(defaultValue string, keys ...string) string {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := provider.kvclient.Get(joinedKeys)
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
