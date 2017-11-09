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
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/ecs"
	"github.com/containous/traefik/provider/kubernetes"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/server"
	"github.com/containous/traefik/server/uuid"
	traefikTls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	"github.com/coreos/go-systemd/daemon"
	"github.com/docker/libkv/store"
)

func main() {
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
			run(&traefikConfiguration.GlobalConfiguration, traefikConfiguration.ConfigFile)
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
		Description:           `Calls traefik /ping to check health (ping must be enabled)`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Run: func() error {
			traefikConfiguration.GlobalConfiguration.SetEffectiveConfiguration(traefikConfiguration.ConfigFile)

			if traefikConfiguration.Ping == nil {
				fmt.Println("Please enable `ping` to use healtcheck.")
				os.Exit(1)
			}

			pingEntryPoint, ok := traefikConfiguration.EntryPoints[traefikConfiguration.Ping.EntryPoint]
			if !ok {
				pingEntryPoint = &configuration.EntryPoint{Address: ":8080"}
			}

			client := &http.Client{Timeout: 5 * time.Second}
			protocol := "http"
			if pingEntryPoint.TLS != nil {
				protocol = "https"
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client.Transport = tr
			}

			resp, errPing := client.Head(protocol + "://" + pingEntryPoint.Address + traefikConfiguration.Web.Path + "ping")
			if errPing != nil {
				fmt.Printf("Error calling healthcheck: %s\n", errPing)
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
	f.AddParser(reflect.TypeOf(traefikTls.RootCAs{}), &traefikTls.RootCAs{})
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
		fmtlog.Printf("Error reading TOML config file %s : %s\n", toml.ConfigFileUsed(), err)
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
			traefikConfiguration.Cluster = &types.Cluster{Node: uuid.Get()}
		}
		if traefikConfiguration.Cluster.Store == nil {
			traefikConfiguration.Cluster.Store = &types.Store{Prefix: kv.Prefix, Store: kv.Store}
		}
		s.AddSource(kv)
		operation := func() error {
			_, err := s.LoadConfig()
			return err
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Load config error: %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
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

func run(globalConfiguration *configuration.GlobalConfiguration, configFile string) {
	configureLogging(globalConfiguration)

	if len(configFile) > 0 {
		log.Infof("Using TOML configuration file %s", configFile)
	}

	http.DefaultTransport.(*http.Transport).Proxy = http.ProxyFromEnvironment

	globalConfiguration.SetEffectiveConfiguration(configFile)

	jsonConf, _ := json.Marshal(globalConfiguration)
	log.Infof("Traefik version %s built on %s", version.Version, version.BuildDate)

	if globalConfiguration.CheckNewVersion {
		checkNewVersion()
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
	logrus.Exit(0)
}

func configureLogging(globalConfiguration *configuration.GlobalConfiguration) {
	// configure default log flags
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)

	if globalConfiguration.Debug {
		globalConfiguration.LogLevel = "DEBUG"
	}

	// configure log level
	level, err := logrus.ParseLevel(strings.ToLower(globalConfiguration.LogLevel))
	if err != nil {
		log.Error("Error getting level", err)
	}
	log.SetLevel(level)

	// configure log output file
	logFile := globalConfiguration.TraefikLogsFile
	if len(logFile) > 0 {
		log.Warn("top-level traefikLogsFile has been deprecated -- please use traefiklog.filepath")
	}
	if globalConfiguration.TraefikLog != nil && len(globalConfiguration.TraefikLog.FilePath) > 0 {
		logFile = globalConfiguration.TraefikLog.FilePath
	}

	// configure log format
	var formatter logrus.Formatter
	if globalConfiguration.TraefikLog != nil && globalConfiguration.TraefikLog.Format == "json" {
		formatter = &logrus.JSONFormatter{}
	} else {
		disableColors := false
		if len(logFile) > 0 {
			disableColors = true
		}
		formatter = &logrus.TextFormatter{DisableColors: disableColors, FullTimestamp: true, DisableSorting: true}
	}
	log.SetFormatter(formatter)

	if len(logFile) > 0 {
		dir := filepath.Dir(logFile)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Errorf("Failed to create log path %s: %s", dir, err)
		}

		err = log.OpenFile(logFile)
		logrus.RegisterExitHandler(func() {
			if err := log.CloseFile(); err != nil {
				log.Error("Error closing log", err)
			}
		})
		if err != nil {
			log.Error("Error opening file", err)
		}
	}
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

func checkNewVersion() {
	ticker := time.NewTicker(24 * time.Hour)
	safe.Go(func() {
		time.Sleep(10 * time.Minute)
		version.CheckNewVersion()
		for {
			select {
			case <-ticker.C:
				version.CheckNewVersion()
			}
		}
	})
}
