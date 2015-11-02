// Package provider holds the different provider implementation.
package provider

import (
	"bytes"
	"errors"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/emilevauge/traefik/autogen"
	"github.com/emilevauge/traefik/types"
)

// Kv holds common configurations of key-value providers.
type Kv struct {
	Watch     bool
	Endpoint  string
	Prefix    string
	Filename  string
	StoreType store.Backend
	kvclient  store.Store
}

// NewConsulProvider returns a Consul provider.
func NewConsulProvider(provider *Consul) *Kv {
	kvProvider := new(Kv)
	kvProvider.Watch = provider.Watch
	kvProvider.Endpoint = provider.Endpoint
	kvProvider.Prefix = provider.Prefix
	kvProvider.Filename = provider.Filename
	kvProvider.StoreType = store.CONSUL
	return kvProvider
}

// NewEtcdProvider returns a Etcd provider.
func NewEtcdProvider(provider *Etcd) *Kv {
	kvProvider := new(Kv)
	kvProvider.Watch = provider.Watch
	kvProvider.Endpoint = provider.Endpoint
	kvProvider.Prefix = provider.Prefix
	kvProvider.Filename = provider.Filename
	kvProvider.StoreType = store.ETCD
	return kvProvider
}

// NewZkProvider returns a Zookepper provider.
func NewZkProvider(provider *Zookepper) *Kv {
	kvProvider := new(Kv)
	kvProvider.Watch = provider.Watch
	kvProvider.Endpoint = provider.Endpoint
	kvProvider.Prefix = provider.Prefix
	kvProvider.Filename = provider.Filename
	kvProvider.StoreType = store.ZK
	return kvProvider
}

// NewBoltDbProvider returns a BoldDb provider.
func NewBoltDbProvider(provider *BoltDb) *Kv {
	kvProvider := new(Kv)
	kvProvider.Watch = provider.Watch
	kvProvider.Endpoint = provider.Endpoint
	kvProvider.Prefix = provider.Prefix
	kvProvider.Filename = provider.Filename
	kvProvider.StoreType = store.BOLTDB
	return kvProvider
}

func (provider *Kv) provide(configurationChan chan<- types.ConfigMessage) error {
	switch provider.StoreType {
	case store.CONSUL:
		consul.Register()
	case store.ETCD:
		etcd.Register()
	case store.ZK:
		zookeeper.Register()
	case store.BOLTDB:
		boltdb.Register()
	default:
		return errors.New("Invalid kv store: " + string(provider.StoreType))
	}
	kv, err := libkv.NewStore(
		provider.StoreType,
		[]string{provider.Endpoint},
		&store.Config{
			ConnectionTimeout: 30 * time.Second,
			Bucket:            "traefik",
		},
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
					configurationChan <- types.ConfigMessage{string(provider.StoreType), configuration}
				}
				defer close(stopCh)
			}
		}()
	}
	configuration := provider.loadConfig()
	configurationChan <- types.ConfigMessage{string(provider.StoreType), configuration}
	return nil
}

func (provider *Kv) loadConfig() *types.Configuration {
	configuration := new(types.Configuration)
	templateObjects := struct {
		Prefix string
	}{
		provider.Prefix,
	}
	var KvFuncMap = template.FuncMap{
		"List": func(keys ...string) []string {
			joinedKeys := strings.Join(keys, "")
			keysPairs, err := provider.kvclient.List(joinedKeys)
			if err != nil {
				log.Error("Error getting keys: ", joinedKeys, err)
				return nil
			}
			directoryKeys := make(map[string]string)
			for _, key := range keysPairs {
				directory := strings.Split(strings.TrimPrefix(key.Key, strings.TrimPrefix(joinedKeys, "/")), "/")[0]
				directoryKeys[directory] = joinedKeys + directory
			}
			return fun.Values(directoryKeys).([]string)
		},
		"Get": func(keys ...string) string {
			joinedKeys := strings.Join(keys, "")
			keyPair, err := provider.kvclient.Get(joinedKeys)
			if err != nil {
				log.Debug("Error getting key: ", joinedKeys, err)
				return ""
			} else if keyPair == nil {
				return ""
			}
			return string(keyPair.Value)
		},
		"Last": func(key string) string {
			splittedKey := strings.Split(key, "/")
			return splittedKey[len(splittedKey)-1]
		},
	}

	tmpl := template.New(provider.Filename).Funcs(KvFuncMap)
	if len(provider.Filename) > 0 {
		_, err := tmpl.ParseFiles(provider.Filename)
		if err != nil {
			log.Error("Error reading file", err)
			return nil
		}
	} else {
		buf, err := autogen.Asset("templates/kv.tmpl")
		if err != nil {
			log.Error("Error reading file", err)
		}
		_, err = tmpl.Parse(string(buf))
		if err != nil {
			log.Error("Error reading file", err)
			return nil
		}
	}

	var buffer bytes.Buffer

	err := tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		log.Error("Error with kv template:", err)
		return nil
	}

	if _, err := toml.Decode(buffer.String(), configuration); err != nil {
		log.Error("Error creating kv configuration:", err)
		log.Error(buffer.String())
		return nil
	}

	return configuration
}
