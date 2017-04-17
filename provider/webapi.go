package provider

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// WebAPI holds configurations of the WebAPI provider.
type WebAPI struct {
	Endpoint      string `description:"Comma sepparated server endpoints"`
	Cluster       string `description:"Web cluster"`
	Watch         bool   `description:"Watch provider"`
	CheckInterval int    `description:"Check interval with seconds for config"`
	version       int
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *WebAPI) Provide(configurationChan chan<- types.ConfigMessage, _ *safe.Pool, _ types.Constraints) error {
	if provider.CheckInterval == 0 {
		provider.CheckInterval = 30
	}

	log.Infof("webapi > {Endpoint: %s, Watch: %v, Cluster: %s, Checkinterval: %s}",
		provider.Endpoint, provider.Watch, provider.Cluster, provider.CheckInterval)

	if provider.Watch {
		go func() {
			for {
				provider.watch(configurationChan)
			}
		}()
	}

	version := provider.loadVersion()
	cfg, err := provider.loadConfig()
	if err == nil {
		provider.version = version
		configurationChan <- cfg
	}
	return err
}

func (provider *WebAPI) loadVersion() (verson int) {
	data, err := provider.request("/traefik/version")
	if err != nil {
		log.Errorf("webapi > load version failed: %v", err)
		return
	}

	v := struct {
		Version int `json:"version" bson:"-"`
	}{}
	err = json.Unmarshal(data, &v)
	if err != nil {
		log.Errorf("webapi > unmarshal verson failed: %v", err)
		return
	}

	return v.Version
}

func (provider *WebAPI) loadConfig() (cfg types.ConfigMessage, err error) {
	var data []byte
	data, err = provider.request("/traefik/config")
	if err != nil {
		return
	}

	cfg = types.ConfigMessage{
		ProviderName:  "webapi",
		Configuration: new(types.Configuration),
	}
	err = json.Unmarshal(data, &cfg.Configuration)
	if err != nil {
		log.Errorf("unmarshal config failed: %v", err)
	}
	return
}

func (provider *WebAPI) watch(configurationChan chan<- types.ConfigMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("webapi > refresh panic: %v", r)
		}
	}()

	time.Sleep(time.Duration(provider.CheckInterval) * time.Second)

	if version := provider.loadVersion(); version != provider.version {
		log.Infof("webapi > refresh ok: %v", version)
		cfg, err := provider.loadConfig()
		if err == nil {
			provider.version = version
			configurationChan <- cfg
		} else {
			log.Errorf("webapi > refresh failed: %v", err)
		}
	}
}

func (provider *WebAPI) request(path string) (data []byte, err error) {
	servers := strings.Split(provider.Endpoint, ",")
	if len(servers) == 0 {
		err = errors.New("webapi > endpoint must be configured")
		return
	}

	var resp *http.Response
	for _, server := range servers {
		u := server + path
		if provider.Cluster != "" {
			u = u + "?cluster=" + provider.Cluster
		}

		client := http.Client{Timeout: time.Second * 5}
		resp, err = client.Get(u)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			err = errors.New(string(data))
			continue
		}

		return data, nil
	}
	return
}
