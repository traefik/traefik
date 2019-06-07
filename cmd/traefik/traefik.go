package main

import (
	"context"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containous/traefik/autogen/genstatic"
	"github.com/containous/traefik/cmd"
	"github.com/containous/traefik/cmd/healthcheck"
	cmdVersion "github.com/containous/traefik/cmd/version"
	"github.com/containous/traefik/pkg/cli"
	"github.com/containous/traefik/pkg/collector"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider/aggregator"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/server"
	"github.com/containous/traefik/pkg/server/router"
	traefiktls "github.com/containous/traefik/pkg/tls"
	"github.com/containous/traefik/pkg/version"
	"github.com/coreos/go-systemd/daemon"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/roundrobin"
)

func init() {
	goDebug := os.Getenv("GODEBUG")
	if len(goDebug) > 0 {
		goDebug += ","
	}
	os.Setenv("GODEBUG", goDebug+"tls13=1")
}

func main() {
	// traefik config inits
	tConfig := cmd.NewTraefikConfiguration()

	loaders := []cli.ResourceLoader{&cli.FileLoader{}, &cli.EnvLoader{}, &cli.FlagLoader{}}

	cmdTraefik := &cli.Command{
		Name: "traefik",
		Description: `Traefik is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
Complete documentation is available at https://traefik.io`,
		Configuration: tConfig,
		Resources:     loaders,
		Run: func(_ []string) error {
			return runCmd(&tConfig.Configuration, cli.GetConfigFile(loaders))
		},
	}

	err := cmdTraefik.AddCommand(healthcheck.NewCmd(&tConfig.Configuration, loaders))
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}

	err = cmdTraefik.AddCommand(cmdVersion.NewCmd())
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}

	err = cli.Execute(cmdTraefik)
	if err != nil {
		stdlog.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func runCmd(staticConfiguration *static.Configuration, configFile string) error {
	configureLogging(staticConfiguration)

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

	serverEntryPointsTCP := make(server.TCPEntryPoints)
	for entryPointName, config := range staticConfiguration.EntryPoints {
		ctx := log.With(context.Background(), log.Str(log.EntryPointName, entryPointName))
		serverEntryPointsTCP[entryPointName], err = server.NewTCPEntryPoint(ctx, config)
		if err != nil {
			return fmt.Errorf("error while building entryPoint %s: %v", entryPointName, err)
		}
		serverEntryPointsTCP[entryPointName].RouteAppenderFactory = router.NewRouteAppenderFactory(*staticConfiguration, entryPointName, acmeProvider)

	}

	tlsManager := traefiktls.NewManager()

	if acmeProvider != nil {
		acmeProvider.SetTLSManager(tlsManager)
		if acmeProvider.TLSChallenge != nil &&
			acmeProvider.HTTPChallenge == nil &&
			acmeProvider.DNSChallenge == nil {
			tlsManager.TLSAlpnGetter = acmeProvider.GetTLSALPNCertificate
		}
	}

	svr := server.NewServer(*staticConfiguration, providerAggregator, serverEntryPointsTCP, tlsManager)

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
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)

	// configure log level
	// an explicitly defined log level always has precedence. if none is
	// given and debug mode is disabled, the default is ERROR, and DEBUG
	// otherwise.
	var levelStr string
	if staticConfiguration.Log != nil {
		levelStr = strings.ToLower(staticConfiguration.Log.Level)
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
	if staticConfiguration.Global.SendAnonymousUsage == nil {
		log.WithoutContext().Error(`
You haven't specified the sendAnonymousUsage option, it will be enabled by default.
`)
		sendAnonymousUsage := true
		staticConfiguration.Global.SendAnonymousUsage = &sendAnonymousUsage
	}

	if *staticConfiguration.Global.SendAnonymousUsage {
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
