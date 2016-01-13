package main

import (
	"errors"
	fmtlog "log"
	"os"
	"reflect"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/emilevauge/traefik/middlewares"
	"github.com/thoas/stats"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	globalConfigFile = kingpin.Arg("conf", "Main configration file.").Default("traefik.toml").String()
	version          = kingpin.Flag("version", "Get Version.").Short('v').Bool()
	metrics          = stats.New()
	oxyLogger        = &OxyLogger{}
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	kingpin.Version(Version + " built on the " + BuildDate)
	kingpin.Parse()
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)

	// load global configuration
	globalConfiguration := LoadFileConfig(*globalConfigFile)

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
		defer fi.Close()
		if err != nil {
			log.Fatal("Error opening file", err)
		} else {
			log.SetOutput(fi)
			log.SetFormatter(&log.TextFormatter{DisableColors: true, FullTimestamp: true, DisableSorting: true})
		}
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableSorting: true})
	}
	log.Debugf("Global configuration loaded %+v", globalConfiguration)
	server := NewServer(*globalConfiguration)
	server.Start()
	server.Close()
	log.Info("Shutting down")
}

// Invoke calls the specified method with the specified arguments on the specified interface.
// It uses the go(lang) reflect package.
func invoke(any interface{}, name string, args ...interface{}) ([]reflect.Value, error) {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	method := reflect.ValueOf(any).MethodByName(name)
	if method.IsValid() {
		return method.Call(inputs), nil
	}
	return nil, errors.New("Method not found: " + name)
}
