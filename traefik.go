package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/types"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	fmtlog "log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"text/template"
	"time"
)

var versionTemplate = `Version:      {{.Version}}
Codename:     {{.Codename}}
Go version:   {{.GoVersion}}
Built:        {{.BuildTime}}
OS/Arch:      {{.Os}}/{{.Arch}}`

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
			run(traefikConfiguration)
			return nil
		},
	}

	//version Command init
	versionCmd := &flaeg.Command{
		Name:                  "version",
		Description:           `Print version`,
		Config:                struct{}{},
		DefaultPointersConfig: struct{}{},
		Run: func() error {
			tmpl, err := template.New("").Parse(versionTemplate)
			if err != nil {
				return err
			}

			v := struct {
				Version   string
				Codename  string
				GoVersion string
				BuildTime string
				Os        string
				Arch      string
			}{
				Version:   Version,
				Codename:  Codename,
				GoVersion: runtime.Version(),
				BuildTime: BuildDate,
				Os:        runtime.GOOS,
				Arch:      runtime.GOARCH,
			}

			if err := tmpl.Execute(os.Stdout, v); err != nil {
				return err
			}
			fmt.Printf("\n")
			return nil

		},
	}

	//init flaeg source
	f := flaeg.New(traefikCmd, os.Args[1:])
	//add custom parsers
	f.AddParser(reflect.TypeOf(EntryPoints{}), &EntryPoints{})
	f.AddParser(reflect.TypeOf(DefaultEntryPoints{}), &DefaultEntryPoints{})
	f.AddParser(reflect.TypeOf(types.Constraints{}), &types.Constraints{})
	f.AddParser(reflect.TypeOf(provider.Namespaces{}), &provider.Namespaces{})
	f.AddParser(reflect.TypeOf([]acme.Domain{}), &acme.Domains{})

	//add version command
	f.AddCommand(versionCmd)
	if _, err := f.Parse(traefikCmd); err != nil {
		fmtlog.Println(err)
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
	}

	traefikConfiguration.ConfigFile = toml.ConfigFileUsed()

	var kv *staert.KvSource
	var err error
	storeConfig := &store.Config{
		ConnectionTimeout: 30 * time.Second,
		Bucket:            "traefik",
	}
	if traefikConfiguration.Consul != nil {
		//init KvSource
		if traefikConfiguration.Consul.TLS != nil {
			storeConfig.TLS, err = traefikConfiguration.Consul.TLS.CreateTLSConfig()
			if err != nil {
				fmtlog.Println(err)
				os.Exit(-1)
			}
		}
		consul.Register()
		kv, err = staert.NewKvSource(
			store.CONSUL,
			strings.Split(traefikConfiguration.Consul.Endpoint, ","),
			storeConfig,
			strings.TrimPrefix(traefikConfiguration.Consul.Prefix, "/"),
		)
	} else if traefikConfiguration.Etcd != nil {
		//init KvSource
		if traefikConfiguration.Etcd.TLS != nil {
			storeConfig.TLS, err = traefikConfiguration.Etcd.TLS.CreateTLSConfig()
			if err != nil {
				fmtlog.Println(err)
				os.Exit(-1)
			}
		}
		etcd.Register()
		kv, err = staert.NewKvSource(
			store.ETCD,
			strings.Split(traefikConfiguration.Etcd.Endpoint, ","),
			storeConfig,
			traefikConfiguration.Etcd.Prefix,
		)
	} else if traefikConfiguration.Zookeeper != nil {
		//init KvSource
		zookeeper.Register()
		kv, err = staert.NewKvSource(
			store.ZK,
			strings.Split(traefikConfiguration.Zookeeper.Endpoint, ","),
			nil,
			traefikConfiguration.Zookeeper.Prefix,
		)
	} else if traefikConfiguration.Boltdb != nil {
		//init KvSource
		boltdb.Register()
		kv, err = staert.NewKvSource(
			store.BOLTDB,
			strings.Split(traefikConfiguration.Boltdb.Endpoint, ","),
			nil,
			traefikConfiguration.Boltdb.Prefix,
		)
	}
	if err != nil {
		fmtlog.Println(err)
		os.Exit(-1)
	}

	if kv != nil {
		s.AddSource(kv)
		if _, err := s.LoadConfig(); err != nil {
			fmtlog.Println(err)
		}
	}

	if err := s.Run(); err != nil {
		fmtlog.Println(err)
		os.Exit(-1)
	}

	os.Exit(0)
}

func run(traefikConfiguration *TraefikConfiguration) {
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)

	// load global configuration
	globalConfiguration := traefikConfiguration.GlobalConfiguration

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = globalConfiguration.MaxIdleConnsPerHost
	loggerMiddleware := middlewares.NewLogger(globalConfiguration.AccessLogsFile)
	defer loggerMiddleware.Close()

	if globalConfiguration.File != nil && len(globalConfiguration.File.Filename) == 0 {
		// no filename, setting to global config file
		if len(traefikConfiguration.ConfigFile) != 0 {
			globalConfiguration.File.Filename = traefikConfiguration.ConfigFile
		} else {
			log.Errorln("Error using file configuration backend, no filename defined")
		}
	}

	if len(globalConfiguration.EntryPoints) == 0 {
		globalConfiguration.EntryPoints = map[string]*EntryPoint{"http": {Address: ":80"}}
		globalConfiguration.DefaultEntryPoints = []string{"http"}
	}

	if globalConfiguration.Debug {
		globalConfiguration.LogLevel = "DEBUG"
	}

	// logging
	level, err := log.ParseLevel(strings.ToLower(globalConfiguration.LogLevel))
	if err != nil {
		log.Fatal("Error getting level", err)
	}
	log.SetLevel(level)
	if len(globalConfiguration.TraefikLogsFile) > 0 {
		fi, err := os.OpenFile(globalConfiguration.TraefikLogsFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer func() {
			if err := fi.Close(); err != nil {
				log.Error("Error closinf file", err)
			}
		}()
		if err != nil {
			log.Fatal("Error opening file", err)
		} else {
			log.SetOutput(fi)
			log.SetFormatter(&log.TextFormatter{DisableColors: true, FullTimestamp: true, DisableSorting: true})
		}
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableSorting: true})
	}
	jsonConf, _ := json.Marshal(globalConfiguration)
	log.Infof("Traefik version %s built on %s", Version, BuildDate)
	if len(traefikConfiguration.ConfigFile) != 0 {
		log.Infof("Using TOML configuration file %s", traefikConfiguration.ConfigFile)
	}
	log.Debugf("Global configuration loaded %s", string(jsonConf))
	server := NewServer(globalConfiguration)
	server.Start()
	defer server.Close()
	log.Info("Shutting down")
}
