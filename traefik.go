package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/provider"
	fmtlog "log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//traefik config inits
	traefikConfiguration := NewTraefikConfiguration()
	traefikPointersConfiguration := NewTraefikPointersConfiguration()
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
			fmtlog.Println(Version + " built on the " + BuildDate)
			return nil
		},
	}

	//init flaeg source
	f := flaeg.New(traefikCmd, os.Args[1:])
	//add custom parsers
	f.AddParser(reflect.TypeOf(EntryPoints{}), &EntryPoints{})
	f.AddParser(reflect.TypeOf(DefaultEntryPoints{}), &DefaultEntryPoints{})
	f.AddParser(reflect.TypeOf(provider.Namespaces{}), &provider.Namespaces{})
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
	if _, err := s.GetConfig(); err != nil {
		fmtlog.Println(err)
	}
	if traefikConfiguration.File != nil && len(traefikConfiguration.File.Filename) == 0 {
		// no filename, setting to global config file
		log.Debugf("ConfigFileUsed %s", toml.ConfigFileUsed())
		traefikConfiguration.File.Filename = toml.ConfigFileUsed()
	}
	if len(traefikConfiguration.EntryPoints) == 0 {
		traefikConfiguration.EntryPoints = map[string]*EntryPoint{"http": &EntryPoint{Address: ":80"}}
		traefikConfiguration.DefaultEntryPoints = []string{"http"}
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
	log.Debugf("Global configuration loaded %s", string(jsonConf))
	server := NewServer(globalConfiguration)
	server.Start()
	defer server.Close()
	log.Info("Shutting down")
}
