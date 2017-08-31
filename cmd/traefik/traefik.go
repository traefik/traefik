package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	fmtlog "log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/ecs"
	"github.com/containous/traefik/provider/kubernetes"
	"github.com/containous/traefik/provider/rancher"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/server"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	"github.com/coreos/go-systemd/daemon"
	"github.com/docker/libkv/store"
	"github.com/satori/go.uuid"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//traefik config inits
	traefikConfiguration := NewTraefikConfiguration()
	traefikPointersConfiguration := NewTraefikDefaultPointersConfiguration()
	//traefik Command init
	traefikCmd := &flaeg.Command{
		Name: "traefik",
		Description: `traefik is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
Complete documentation is available at https://traefik.io`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Run: func() error {
			globalConfiguration := traefikConfiguration.GlobalConfiguration
			if globalConfiguration.File != nil && len(globalConfiguration.File.Filename) == 0 {
				// no filename, setting to global config file
				if len(traefikConfiguration.ConfigFile) != 0 {
					globalConfiguration.File.Filename = traefikConfiguration.ConfigFile
				} else {
					log.Errorln("Error using file configuration backend, no filename defined")
				}
			}
			if len(traefikConfiguration.ConfigFile) != 0 {
				log.Infof("Using TOML configuration file %s", traefikConfiguration.ConfigFile)
			}
			run(&globalConfiguration)
			return nil
		},
	}

	//storeconfig Command init
	var kv *staert.KvSource
	var err error

	storeConfigCmd := &flaeg.Command{
		Name:                  "storeconfig",
		Description:           `Store the static traefik configuration into a Key-value stores. Traefik will not start.`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Run: func() error {
			if kv == nil {
				return fmt.Errorf("Error using command storeconfig, no Key-value store defined")
			}
			jsonConf, err := json.Marshal(traefikConfiguration.GlobalConfiguration)
			if err != nil {
				return err
			}
			fmtlog.Printf("Storing configuration: %s\n", jsonConf)
			err = kv.StoreConfig(traefikConfiguration.GlobalConfiguration)
			if err != nil {
				return err
			}
			if traefikConfiguration.GlobalConfiguration.ACME != nil && len(traefikConfiguration.GlobalConfiguration.ACME.StorageFile) > 0 {
				// convert ACME json file to KV store
				localStore := acme.NewLocalStore(traefikConfiguration.GlobalConfiguration.ACME.StorageFile)
				object, err := localStore.Load()
				if err != nil {
					return err
				}
				meta := cluster.NewMetadata(object)
				err = meta.Marshall()
				if err != nil {
					return err
				}
				source := staert.KvSource{
					Store:  kv,
					Prefix: traefikConfiguration.GlobalConfiguration.ACME.Storage,
				}
				err = source.StoreConfig(meta)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Metadata: map[string]string{
			"parseAllSources": "true",
		},
	}

	healthCheckCmd := &flaeg.Command{
		Name:                  "healthcheck",
		Description:           `Calls traefik /ping to check health (web provider must be enabled)`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Run: func() error {
			if traefikConfiguration.Web == nil {
				fmt.Println("Please enable the web provider to use healtcheck.")
				os.Exit(1)
			}
			client := &http.Client{Timeout: 5 * time.Second}
			protocol := "http"
			if len(traefikConfiguration.Web.CertFile) > 0 {
				protocol = "https"
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client.Transport = tr
			}
			resp, err := client.Head(protocol + "://" + traefikConfiguration.Web.Address + "/ping")
			if err != nil {
				fmt.Printf("Error calling healthcheck: %s\n", err)
				os.Exit(1)
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Bad healthcheck status: %s\n", resp.Status)
				os.Exit(1)
			}
			fmt.Printf("OK: %s\n", resp.Request.URL)
			os.Exit(0)
			return nil
		},
		Metadata: map[string]string{
			"parseAllSources": "true",
		},
	}

	//init flaeg source
	f := flaeg.New(traefikCmd, os.Args[1:])
	//add custom parsers
	f.AddParser(reflect.TypeOf(configuration.EntryPoints{}), &configuration.EntryPoints{})
	f.AddParser(reflect.TypeOf(configuration.DefaultEntryPoints{}), &configuration.DefaultEntryPoints{})
	f.AddParser(reflect.TypeOf(configuration.RootCAs{}), &configuration.RootCAs{})
	f.AddParser(reflect.TypeOf(types.Constraints{}), &types.Constraints{})
	f.AddParser(reflect.TypeOf(kubernetes.Namespaces{}), &kubernetes.Namespaces{})
	f.AddParser(reflect.TypeOf(ecs.Clusters{}), &ecs.Clusters{})
	f.AddParser(reflect.TypeOf([]acme.Domain{}), &acme.Domains{})
	f.AddParser(reflect.TypeOf(types.Buckets{}), &types.Buckets{})

	//add commands
	f.AddCommand(newVersionCmd())
	f.AddCommand(newBugCmd(traefikConfiguration, traefikPointersConfiguration))
	f.AddCommand(storeConfigCmd)
	f.AddCommand(healthCheckCmd)

	usedCmd, err := f.GetCommand()
	if err != nil {
		fmtlog.Println(err)
		os.Exit(-1)
	}

	if _, err := f.Parse(usedCmd); err != nil {
		fmtlog.Printf("Error parsing command: %s\n", err)
		os.Exit(-1)
	}

	//staert init
	s := staert.NewStaert(traefikCmd)
	//init toml source
	toml := staert.NewTomlSource("traefik", []string{traefikConfiguration.ConfigFile, "/etc/traefik/", "$HOME/.traefik/", "."})

	//add sources to staert
	s.AddSource(toml)
	s.AddSource(f)
	if _, err := s.LoadConfig(); err != nil {
		fmtlog.Println(fmt.Errorf("Error reading TOML config file %s : %s", toml.ConfigFileUsed(), err))
		os.Exit(-1)
	}

	traefikConfiguration.ConfigFile = toml.ConfigFileUsed()

	kv, err = CreateKvSource(traefikConfiguration)
	if err != nil {
		fmtlog.Printf("Error creating kv store: %s\n", err)
		os.Exit(-1)
	}

	// IF a KV Store is enable and no sub-command called in args
	if kv != nil && usedCmd == traefikCmd {
		if traefikConfiguration.Cluster == nil {
			traefikConfiguration.Cluster = &types.Cluster{Node: uuid.NewV4().String()}
		}
		if traefikConfiguration.Cluster.Store == nil {
			traefikConfiguration.Cluster.Store = &types.Store{Prefix: kv.Prefix, Store: kv.Store}
		}
		s.AddSource(kv)
		if _, err := s.LoadConfig(); err != nil {
			fmtlog.Printf("Error loading configuration: %s\n", err)
			os.Exit(-1)
		}
	}

	if err := s.Run(); err != nil {
		fmtlog.Printf("Error running traefik: %s\n", err)
		os.Exit(-1)
	}

	os.Exit(0)
}

func run(globalConfiguration *configuration.GlobalConfiguration) {
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)

	http.DefaultTransport.(*http.Transport).Proxy = http.ProxyFromEnvironment

	if len(globalConfiguration.EntryPoints) == 0 {
		globalConfiguration.EntryPoints = map[string]*configuration.EntryPoint{"http": {Address: ":80"}}
		globalConfiguration.DefaultEntryPoints = []string{"http"}
	}

	if globalConfiguration.Rancher != nil {
		// Ensure backwards compatibility for now
		if len(globalConfiguration.Rancher.AccessKey) > 0 ||
			len(globalConfiguration.Rancher.Endpoint) > 0 ||
			len(globalConfiguration.Rancher.SecretKey) > 0 {

			if globalConfiguration.Rancher.API == nil {
				globalConfiguration.Rancher.API = &rancher.APIConfiguration{
					AccessKey: globalConfiguration.Rancher.AccessKey,
					SecretKey: globalConfiguration.Rancher.SecretKey,
					Endpoint:  globalConfiguration.Rancher.Endpoint,
				}
			}
			log.Warn("Deprecated configuration found: rancher.[accesskey|secretkey|endpoint]. " +
				"Please use rancher.api.[accesskey|secretkey|endpoint] instead.")
		}

		if globalConfiguration.Rancher.Metadata != nil && len(globalConfiguration.Rancher.Metadata.Prefix) == 0 {
			globalConfiguration.Rancher.Metadata.Prefix = "latest"
		}
	}

	if globalConfiguration.Debug {
		globalConfiguration.LogLevel = "DEBUG"
	}

	// logging
	level, err := logrus.ParseLevel(strings.ToLower(globalConfiguration.LogLevel))
	if err != nil {
		log.Error("Error getting level", err)
	}
	log.SetLevel(level)
	if len(globalConfiguration.TraefikLogsFile) > 0 {
		dir := filepath.Dir(globalConfiguration.TraefikLogsFile)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Errorf("Failed to create log path %s: %s", dir, err)
		}

		err = log.OpenFile(globalConfiguration.TraefikLogsFile)
		defer func() {
			if err := log.CloseFile(); err != nil {
				log.Error("Error closing log", err)
			}
		}()
		if err != nil {
			log.Error("Error opening file", err)
		} else {
			log.SetFormatter(&logrus.TextFormatter{DisableColors: true, FullTimestamp: true, DisableSorting: true})
		}
	} else {
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, DisableSorting: true})
	}
	jsonConf, _ := json.Marshal(globalConfiguration)
	log.Infof("Traefik version %s built on %s", version.Version, version.BuildDate)

	if globalConfiguration.CheckNewVersion {
		ticker := time.NewTicker(24 * time.Hour)
		safe.Go(func() {
			version.CheckNewVersion()
			for {
				select {
				case <-ticker.C:
					version.CheckNewVersion()
				}
			}
		})
	}

	log.Debugf("Global configuration loaded %s", string(jsonConf))
	svr := server.NewServer(*globalConfiguration)
	svr.Start()
	defer svr.Close()
	sent, err := daemon.SdNotify(false, "READY=1")
	if !sent && err != nil {
		log.Error("Fail to notify", err)
	}
	t, err := daemon.SdWatchdogEnabled(false)
	if err != nil {
		log.Error("Problem with watchdog", err)
	} else if t != 0 {
		// Send a ping each half time given
		t = t / 2
		log.Info("Watchdog activated with timer each ", t)
		safe.Go(func() {
			tick := time.Tick(t)
			for range tick {
				if ok, _ := daemon.SdNotify(false, "WATCHDOG=1"); !ok {
					log.Error("Fail to tick watchdog")
				}
			}
		})
	}
	svr.Wait()
	log.Info("Shutting down")
}

// CreateKvSource creates KvSource
// TLS support is enable for Consul and Etcd backends
func CreateKvSource(traefikConfiguration *TraefikConfiguration) (*staert.KvSource, error) {
	var kv *staert.KvSource
	var kvStore store.Store
	var err error

	switch {
	case traefikConfiguration.Consul != nil:
		kvStore, err = traefikConfiguration.Consul.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Consul.Prefix,
		}
	case traefikConfiguration.Etcd != nil:
		kvStore, err = traefikConfiguration.Etcd.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Etcd.Prefix,
		}
	case traefikConfiguration.Zookeeper != nil:
		kvStore, err = traefikConfiguration.Zookeeper.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Zookeeper.Prefix,
		}
	case traefikConfiguration.Boltdb != nil:
		kvStore, err = traefikConfiguration.Boltdb.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Boltdb.Prefix,
		}
	}
	return kv, err
}
