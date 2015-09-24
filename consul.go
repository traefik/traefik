package main

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/BurntSushi/ty/fun"
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

type Key struct {
	Value string
}

type ConsulProvider struct {
	Watch        bool
	Endpoint     string
	Prefix       string
	Filename     string
	consulClient *api.Client
}

var kvClient *api.KV

var ConsulFuncMap = template.FuncMap{
	"List": func(keys ...string) []string {
		joinedKeys := strings.Join(keys, "")
		keysPairs, _, err := kvClient.Keys(joinedKeys, "/", nil)
		if err != nil {
			log.Error("Error getting keys ", joinedKeys, err)
			return nil
		}
		keysPairs = fun.Filter(func(key string) bool {
			if key == joinedKeys {
				return false
			}
			return true
		}, keysPairs).([]string)
		return keysPairs
	},
	"Get": func(keys ...string) string {
		joinedKeys := strings.Join(keys, "")
		keyPair, _, err := kvClient.Get(joinedKeys, nil)
		if err != nil {
			log.Error("Error getting key ", joinedKeys, err)
			return ""
		}
		return string(keyPair.Value)
	},
	"Last": func(key string) string {
		splittedKey := strings.Split(key, "/")
		return splittedKey[len(splittedKey)-2]
	},
}

func NewConsulProvider() *ConsulProvider {
	consulProvider := new(ConsulProvider)
	// default values
	consulProvider.Watch = true
	consulProvider.Prefix = "traefik"

	return consulProvider
}

func (provider *ConsulProvider) Provide(configurationChan chan<- *Configuration) {
	config := &api.Config{
		Address:    provider.Endpoint,
		Scheme:     "http",
		HttpClient: http.DefaultClient,
	}
	consulClient, _ := api.NewClient(config)
	provider.consulClient = consulClient
	if provider.Watch {
		var waitIndex uint64
		keypairs, meta, err := consulClient.KV().Keys("", "", nil)
		if keypairs == nil && err == nil {
			log.Error("Key was not found.")
		}
		waitIndex = meta.LastIndex
		go func() {
			for {
				opts := api.QueryOptions{
					WaitIndex: waitIndex,
				}
				keypairs, meta, err := consulClient.KV().Keys("", "", &opts)
				if keypairs == nil && err == nil {
					log.Error("Key  was not found.")
				}
				waitIndex = meta.LastIndex
				configuration := provider.loadConsulConfig()
				if configuration != nil {
					configurationChan <- configuration
				}
			}
		}()
	}
	configuration := provider.loadConsulConfig()
	configurationChan <- configuration
}

func (provider *ConsulProvider) loadConsulConfig() *Configuration {
	configuration := new(Configuration)
	services := []*api.CatalogService{}
	kvClient = provider.consulClient.KV()

	servicesName, _, _ := provider.consulClient.Catalog().Services(nil)
	for serviceName := range servicesName {
		catalogServices, _, _ := provider.consulClient.Catalog().Service(serviceName, "", nil)
		for _, catalogService := range catalogServices {
			services = append(services, catalogService)
		}
	}

	templateObjects := struct {
		Services []*api.CatalogService
	}{
		services,
	}

	tmpl := template.New(provider.Filename).Funcs(ConsulFuncMap)
	if len(provider.Filename) > 0 {
		_, err := tmpl.ParseFiles(provider.Filename)
		if err != nil {
			log.Error("Error reading file", err)
			return nil
		}
	} else {
		buf, err := Asset("providerTemplates/consul.tmpl")
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
		log.Error("Error with consul template:", err)
		return nil
	}

	if _, err := toml.Decode(buffer.String(), configuration); err != nil {
		log.Error("Error creating consul configuration:", err)
		return nil
	}

	return configuration
}
