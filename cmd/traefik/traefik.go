package main

import (
	"context"
	"encoding/json"
	"fmt"
	fmtlog "log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/containous/traefik/autogen/genstatic"
	"github.com/containous/traefik/cmd"
	"github.com/containous/traefik/cmd/bug"
	"github.com/containous/traefik/cmd/healthcheck"
	"github.com/containous/traefik/cmd/storeconfig"
	cmdVersion "github.com/containous/traefik/cmd/version"
	"github.com/containous/traefik/collector"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/old/provider/ecs"
	"github.com/containous/traefik/old/provider/kubernetes"
	oldtypes "github.com/containous/traefik/old/types"
	"github.com/containous/traefik/provider/aggregator"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/server"
	"github.com/containous/traefik/server/router"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	"github.com/coreos/go-systemd/daemon"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/ogier/pflag"
	"github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/roundrobin"
)

// sliceOfStrings is the parser for []string
type sliceOfStrings []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (s *sliceOfStrings) String() string {
	return strings.Join(*s, ",")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (s *sliceOfStrings) Set(value string) error {
	strings := strings.Split(value, ",")
	if len(strings) == 0 {
		return fmt.Errorf("bad []string format: %s", value)
	}
	for _, entrypoint := range strings {
		*s = append(*s, entrypoint)
	}
	return nil
}

// Get return the []string
func (s *sliceOfStrings) Get() interface{} {
	return *s
}

// SetValue sets the []string with val
func (s *sliceOfStrings) SetValue(val interface{}) {
	*s = val.([]string)
}

// Type is type of the struct
func (s *sliceOfStrings) Type() string {
	return "sliceOfStrings"
}

func main() {
	// traefik config inits
	traefikConfiguration := cmd.NewTraefikConfiguration()
	traefikPointersConfiguration := cmd.NewTraefikDefaultPointersConfiguration()

	// traefik Command init
	traefikCmd := &flaeg.Command{
		Name: "traefik",
		Description: `traefik is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
Complete documentation is available at https://traefik.io`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Run: func() error {
			return runCmd(&traefikConfiguration.Configuration, traefikConfiguration.ConfigFile)
		},
	}

	// storeconfig Command init
	storeConfigCmd := storeconfig.NewCmd(traefikConfiguration, traefikPointersConfiguration)

	// init flaeg source
	f := flaeg.New(traefikCmd, os.Args[1:])
	// add custom parsers
	f.AddParser(reflect.TypeOf(static.EntryPoints{}), &static.EntryPoints{})

	f.AddParser(reflect.SliceOf(reflect.TypeOf("")), &sliceOfStrings{})
	f.AddParser(reflect.TypeOf(traefiktls.FilesOrContents{}), &traefiktls.FilesOrContents{})
	f.AddParser(reflect.TypeOf(types.Constraints{}), &types.Constraints{})
	f.AddParser(reflect.TypeOf(kubernetes.Namespaces{}), &kubernetes.Namespaces{})
	f.AddParser(reflect.TypeOf(ecs.Clusters{}), &ecs.Clusters{})
	f.AddParser(reflect.TypeOf([]types.Domain{}), &types.Domains{})
	f.AddParser(reflect.TypeOf(types.DNSResolvers{}), &types.DNSResolvers{})
	f.AddParser(reflect.TypeOf(types.Buckets{}), &types.Buckets{})

	f.AddParser(reflect.TypeOf(types.StatusCodes{}), &types.StatusCodes{})
	f.AddParser(reflect.TypeOf(types.FieldNames{}), &types.FieldNames{})
	f.AddParser(reflect.TypeOf(types.FieldHeaderNames{}), &types.FieldHeaderNames{})

	// FIXME Remove with ACME
	f.AddParser(reflect.TypeOf([]oldtypes.Domain{}), &oldtypes.Domains{})
	// FIXME Remove with old providers
	f.AddParser(reflect.TypeOf(oldtypes.Constraints{}), &oldtypes.Constraints{})

	// add commands
	f.AddCommand(cmdVersion.NewCmd())
	f.AddCommand(bug.NewCmd(traefikConfiguration, traefikPointersConfiguration))
	f.AddCommand(storeConfigCmd)
	f.AddCommand(healthcheck.NewCmd(traefikConfiguration, traefikPointersConfiguration))

	usedCmd, err := f.GetCommand()
	if err != nil {
		fmtlog.Println(err)
		os.Exit(1)
	}

	if _, err := f.Parse(usedCmd); err != nil {
		if err == pflag.ErrHelp {
			os.Exit(0)
		}
		fmtlog.Printf("Error parsing command: %s\n", err)
		os.Exit(1)
	}

	// staert init
	s := staert.NewStaert(traefikCmd)
	// init TOML source
	toml := staert.NewTomlSource("traefik", []string{traefikConfiguration.ConfigFile, "/etc/traefik/", "$HOME/.traefik/", "."})

	// add sources to staert
	s.AddSource(toml)
	s.AddSource(f)
	if _, err := s.LoadConfig(); err != nil {
		fmtlog.Printf("Error reading TOML config file %s : %s\n", toml.ConfigFileUsed(), err)
		os.Exit(1)
	}

	traefikConfiguration.ConfigFile = toml.ConfigFileUsed()

	kv, err := storeconfig.CreateKvSource(traefikConfiguration)
	if err != nil {
		fmtlog.Printf("Error creating kv store: %s\n", err)
		os.Exit(1)
	}
	storeConfigCmd.Run = storeconfig.Run(kv, traefikConfiguration)

	// if a KV Store is enable and no sub-command called in args
	if kv != nil && usedCmd == traefikCmd {
		s.AddSource(kv)
		operation := func() error {
			_, err := s.LoadConfig()
			return err
		}
		notify := func(err error, time time.Duration) {
			log.WithoutContext().Errorf("Load config error: %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			fmtlog.Printf("Error loading configuration: %s\n", err)
			os.Exit(1)
		}
	}

	if err := s.Run(); err != nil {
		fmtlog.Printf("Error running traefik: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func runCmd(staticConfiguration *static.Configuration, configFile string) error {
	configureLogging(staticConfiguration)

	if len(configFile) > 0 {
		log.WithoutContext().Infof("Using TOML configuration file %s", configFile)
	}

	http.DefaultTransport.(*http.Transport).Proxy = http.ProxyFromEnvironment

	if err := roundrobin.SetDefaultWeight(0); err != nil {
		log.WithoutContext().Errorf("Could not set roundrobin default weight: %v", err)
	}

	staticConfiguration.SetEffectiveConfiguration(configFile)
	staticConfiguration.ValidateConfiguration()

	log.WithoutContext().Infof("Traefik version %s built on %s", version.Version, version.BuildDate)

	jsonConf, err := json.Marshal(staticConfiguration)
	if err != nil {
		log.WithoutContext().Errorf("Could not marshal static configuration: %v", err)
		log.WithoutContext().Debugf("Static configuration loaded [struct] %#v", staticConfiguration)
	} else {
		log.WithoutContext().Debugf("Static configuration loaded %s", string(jsonConf))
	}

	if staticConfiguration.API != nil && staticConfiguration.API.Dashboard {
		staticConfiguration.API.DashboardAssets = &assetfs.AssetFS{Asset: genstatic.Asset, AssetInfo: genstatic.AssetInfo, AssetDir: genstatic.AssetDir, Prefix: "static"}
	}

	if staticConfiguration.Global.CheckNewVersion {
		checkNewVersion()
	}

	stats(staticConfiguration)

	providerAggregator := aggregator.NewProviderAggregator(*staticConfiguration.Providers)

	acmeProvider, err := staticConfiguration.InitACMEProvider()
	if err != nil {
		log.WithoutContext().Errorf("Unable to initialize ACME provider: %v", err)
	} else if acmeProvider != nil {
		if err := providerAggregator.AddProvider(acmeProvider); err != nil {
			log.WithoutContext().Errorf("Unable to add ACME provider to the providers list: %v", err)
			acmeProvider = nil
		}
	}

	serverEntryPoints := make(server.EntryPoints)
	for entryPointName, config := range staticConfiguration.EntryPoints {
		ctx := log.With(context.Background(), log.Str(log.EntryPointName, entryPointName))
		logger := log.FromContext(ctx)

		serverEntryPoint, err := server.NewEntryPoint(ctx, config)
		if err != nil {
			return fmt.Errorf("error while building entryPoint %s: %v", entryPointName, err)
		}

		serverEntryPoint.RouteAppenderFactory = router.NewRouteAppenderFactory(*staticConfiguration, entryPointName, acmeProvider)

		if acmeProvider != nil && entryPointName == acmeProvider.EntryPoint {
			logger.Debugf("Setting Acme Certificate store from Entrypoint")
			acmeProvider.SetCertificateStore(serverEntryPoint.Certs)

			if acmeProvider.OnDemand {
				serverEntryPoint.OnDemandListener = acmeProvider.ListenRequest
			}

			// TLS ALPN 01
			if acmeProvider.TLSChallenge != nil && acmeProvider.HTTPChallenge == nil && acmeProvider.DNSChallenge == nil {
				serverEntryPoint.TLSALPNGetter = acmeProvider.GetTLSALPNCertificate
			}
		}

		serverEntryPoints[entryPointName] = serverEntryPoint
	}

	svr := server.NewServer(*staticConfiguration, providerAggregator, serverEntryPoints)

	if acmeProvider != nil && acmeProvider.OnHostRule {
		acmeProvider.SetConfigListenerChan(make(chan config.Configuration))
		svr.AddListener(acmeProvider.ListenConfiguration)
	}
	ctx := cmd.ContextWithSignal(context.Background())

	if staticConfiguration.Ping != nil {
		staticConfiguration.Ping.WithContext(ctx)
	}

	svr.Start(ctx)
	defer svr.Close()

	sent, err := daemon.SdNotify(false, "READY=1")
	if !sent && err != nil {
		log.WithoutContext().Errorf("Failed to notify: %v", err)
	}

	t, err := daemon.SdWatchdogEnabled(false)
	if err != nil {
		log.WithoutContext().Errorf("Could not enable Watchdog: %v", err)
	} else if t != 0 {
		// Send a ping each half time given
		t /= 2
		log.WithoutContext().Infof("Watchdog activated with timer duration %s", t)
		safe.Go(func() {
			tick := time.Tick(t)
			for range tick {
				_, errHealthCheck := healthcheck.Do(*staticConfiguration)
				if staticConfiguration.Ping == nil || errHealthCheck == nil {
					if ok, _ := daemon.SdNotify(false, "WATCHDOG=1"); !ok {
						log.WithoutContext().Error("Fail to tick watchdog")
					}
				} else {
					log.WithoutContext().Error(errHealthCheck)
				}
			}
		})
	}

	svr.Wait()
	log.WithoutContext().Info("Shutting down")
	logrus.Exit(0)
	return nil
}

func configureLogging(staticConfiguration *static.Configuration) {
	// configure default log flags
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)

	// configure log level
	// an explicitly defined log level always has precedence. if none is
	// given and debug mode is disabled, the default is ERROR, and DEBUG
	// otherwise.
	var levelStr string
	if staticConfiguration.Log != nil {
		levelStr = strings.ToLower(staticConfiguration.Log.LogLevel)
	}
	if levelStr == "" {
		levelStr = "error"
		if staticConfiguration.Global.Debug {
			levelStr = "debug"
		}
	}
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		log.WithoutContext().Errorf("Error getting level: %v", err)
	}
	log.SetLevel(level)

	var logFile string
	if staticConfiguration.Log != nil && len(staticConfiguration.Log.FilePath) > 0 {
		logFile = staticConfiguration.Log.FilePath
	}

	// configure log format
	var formatter logrus.Formatter
	if staticConfiguration.Log != nil && staticConfiguration.Log.Format == "json" {
		formatter = &logrus.JSONFormatter{}
	} else {
		disableColors := len(logFile) > 0
		formatter = &logrus.TextFormatter{DisableColors: disableColors, FullTimestamp: true, DisableSorting: true}
	}
	log.SetFormatter(formatter)

	if len(logFile) > 0 {
		dir := filepath.Dir(logFile)

		if err := os.MkdirAll(dir, 0755); err != nil {
			log.WithoutContext().Errorf("Failed to create log path %s: %s", dir, err)
		}

		err = log.OpenFile(logFile)
		logrus.RegisterExitHandler(func() {
			if err := log.CloseFile(); err != nil {
				log.WithoutContext().Errorf("Error while closing log: %v", err)
			}
		})
		if err != nil {
			log.WithoutContext().Errorf("Error while opening log file %s: %v", logFile, err)
		}
	}
}

func checkNewVersion() {
	ticker := time.Tick(24 * time.Hour)
	safe.Go(func() {
		for time.Sleep(10 * time.Minute); ; <-ticker {
			version.CheckNewVersion()
		}
	})
}

func stats(staticConfiguration *static.Configuration) {
	if staticConfiguration.Global.SendAnonymousUsage {
		log.WithoutContext().Info(`
Stats collection is enabled.
Many thanks for contributing to Traefik's improvement by allowing us to receive anonymous information from your configuration.
Help us improve Traefik by leaving this feature on :)
More details on: https://docs.traefik.io/basics/#collected-data
`)
		collect(staticConfiguration)
	} else {
		log.WithoutContext().Info(`
Stats collection is disabled.
Help us improve Traefik by turning this feature on :)
More details on: https://docs.traefik.io/basics/#collected-data
`)
	}
}

func collect(staticConfiguration *static.Configuration) {
	ticker := time.Tick(24 * time.Hour)
	safe.Go(func() {
		for time.Sleep(10 * time.Minute); ; <-ticker {
			if err := collector.Collect(staticConfiguration); err != nil {
				log.WithoutContext().Debug(err)
			}
		}
	})
}
