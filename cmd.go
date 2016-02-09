/*
Copyright
*/
package main

import (
	fmtlog "log"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/emilevauge/traefik/middlewares"
	"github.com/emilevauge/traefik/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

var traefikCmd = &cobra.Command{
	Use:   "traefik",
	Short: "traefik, a modern reverse proxy",
	Long: `traefik is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
Complete documentation is available at http://traefik.io`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long:  `Print version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmtlog.Println(Version + " built on the " + BuildDate)
		os.Exit(0)
	},
}

var arguments = struct {
	GlobalConfiguration
	web       bool
	file      bool
	docker    bool
	dockerTLS bool
	marathon  bool
	consul    bool
	zookeeper bool
	etcd      bool
	boltdb    bool
}{
	GlobalConfiguration{
		Docker: &provider.Docker{
			TLS: &provider.DockerTLS{},
		},
		File:      &provider.File{},
		Web:       &WebProvider{},
		Marathon:  &provider.Marathon{},
		Consul:    &provider.Consul{},
		Zookeeper: &provider.Zookepper{},
		Etcd:      &provider.Etcd{},
		Boltdb:    &provider.BoltDb{},
	},
	false,
	false,
	false,
	false,
	false,
	false,
	false,
	false,
	false,
}

func init() {
	traefikCmd.AddCommand(versionCmd)
	traefikCmd.PersistentFlags().StringP("configFile", "c", "", "Configuration file to use (TOML, JSON, YAML, HCL).")
	traefikCmd.PersistentFlags().StringP("port", "p", ":80", "Reverse proxy port")
	traefikCmd.PersistentFlags().StringP("graceTimeOut", "g", "10", "Timeout in seconds. Duration to give active requests a chance to finish during hot-reloads")
	traefikCmd.PersistentFlags().String("accessLogsFile", "log/access.log", "Access logs file")
	traefikCmd.PersistentFlags().String("traefikLogsFile", "log/traefik.log", "Traefik logs file")
	traefikCmd.PersistentFlags().Var(&arguments.Certificates, "certificates", "SSL certificates and keys pair, ie 'tests/traefik.crt,tests/traefik.key'. You may add several certificate/key pairs to terminate HTTPS for multiple domain names using TLS SNI")
	traefikCmd.PersistentFlags().StringP("logLevel", "l", "ERROR", "Log level")
	traefikCmd.PersistentFlags().DurationVar(&arguments.ProvidersThrottleDuration, "providersThrottleDuration", time.Duration(2*time.Second), "Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time.")
	traefikCmd.PersistentFlags().Int("maxIdleConnsPerHost", 0, "If non-zero, controls the maximum idle (keep-alive) to keep per-host.  If zero, DefaultMaxIdleConnsPerHost is used")

	traefikCmd.PersistentFlags().BoolVar(&arguments.web, "web", false, "Enable Web backend")
	traefikCmd.PersistentFlags().StringVar(&arguments.Web.Address, "web.address", ":8080", "Web administration port")
	traefikCmd.PersistentFlags().StringVar(&arguments.Web.CertFile, "web.cerFile", "", "SSL certificate")
	traefikCmd.PersistentFlags().StringVar(&arguments.Web.KeyFile, "web.keyFile", "", "SSL certificate")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Web.ReadOnly, "web.readOnly", false, "Enable read only API")

	traefikCmd.PersistentFlags().BoolVar(&arguments.file, "file", false, "Enable File backend")
	traefikCmd.PersistentFlags().BoolVar(&arguments.File.Watch, "file.watch", true, "Watch provider")
	traefikCmd.PersistentFlags().StringVar(&arguments.File.Filename, "file.filename", "", "Override default configuration template. For advanced users :)")

	traefikCmd.PersistentFlags().BoolVar(&arguments.docker, "docker", false, "Enable Docker backend")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Docker.Watch, "docker.watch", true, "Watch provider")
	traefikCmd.PersistentFlags().StringVar(&arguments.Docker.Filename, "docker.filename", "", "Override default configuration template. For advanced users :)")
	traefikCmd.PersistentFlags().StringVar(&arguments.Docker.Endpoint, "docker.endpoint", "unix:///var/run/docker.sock", "Docker server endpoint. Can be a tcp or a unix socket endpoint")
	traefikCmd.PersistentFlags().StringVar(&arguments.Docker.Domain, "docker.domain", "", "Default domain used")
	traefikCmd.PersistentFlags().BoolVar(&arguments.dockerTLS, "docker.tls", false, "Enable Docker TLS support")
	traefikCmd.PersistentFlags().StringVar(&arguments.Docker.TLS.CA, "docker.tls.ca", "", "TLS CA")
	traefikCmd.PersistentFlags().StringVar(&arguments.Docker.TLS.Cert, "docker.tls.cert", "", "TLS cert")
	traefikCmd.PersistentFlags().StringVar(&arguments.Docker.TLS.Key, "docker.tls.key", "", "TLS key")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Docker.TLS.InsecureSkipVerify, "docker.tls.insecureSkipVerify", false, "TLS insecure skip verify")

	traefikCmd.PersistentFlags().BoolVar(&arguments.marathon, "marathon", false, "Enable Marathon backend")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Marathon.Watch, "marathon.watch", true, "Watch provider")
	traefikCmd.PersistentFlags().StringVar(&arguments.Marathon.Filename, "marathon.filename", "", "Override default configuration template. For advanced users :)")
	traefikCmd.PersistentFlags().StringVar(&arguments.Marathon.Endpoint, "marathon.endpoint", "http://127.0.0.1:8080", "Marathon server endpoint. You can also specify multiple endpoint for Marathon")
	traefikCmd.PersistentFlags().StringVar(&arguments.Marathon.Domain, "marathon.domain", "", "Default domain used")
	traefikCmd.PersistentFlags().StringVar(&arguments.Marathon.NetworkInterface, "marathon.networkInterface", "eth0", "Network interface used to call Marathon web services. Needed in case of multiple network interfaces")

	traefikCmd.PersistentFlags().BoolVar(&arguments.consul, "consul", false, "Enable Consul backend")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Consul.Watch, "consul.watch", true, "Watch provider")
	traefikCmd.PersistentFlags().StringVar(&arguments.Consul.Filename, "consul.filename", "", "Override default configuration template. For advanced users :)")
	traefikCmd.PersistentFlags().StringVar(&arguments.Consul.Endpoint, "consul.endpoint", "127.0.0.1:8500", "Consul server endpoint")
	traefikCmd.PersistentFlags().StringVar(&arguments.Consul.Prefix, "consul.prefix", "/traefik", "Prefix used for KV store")

	traefikCmd.PersistentFlags().BoolVar(&arguments.zookeeper, "zookeeper", false, "Enable Zookeeper backend")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Zookeeper.Watch, "zookeeper.watch", true, "Watch provider")
	traefikCmd.PersistentFlags().StringVar(&arguments.Zookeeper.Filename, "zookeeper.filename", "", "Override default configuration template. For advanced users :)")
	traefikCmd.PersistentFlags().StringVar(&arguments.Zookeeper.Endpoint, "zookeeper.endpoint", "127.0.0.1:2181", "Zookeeper server endpoint")
	traefikCmd.PersistentFlags().StringVar(&arguments.Zookeeper.Prefix, "zookeeper.prefix", "/traefik", "Prefix used for KV store")

	traefikCmd.PersistentFlags().BoolVar(&arguments.etcd, "etcd", false, "Enable Etcd backend")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Etcd.Watch, "etcd.watch", true, "Watch provider")
	traefikCmd.PersistentFlags().StringVar(&arguments.Etcd.Filename, "etcd.filename", "", "Override default configuration template. For advanced users :)")
	traefikCmd.PersistentFlags().StringVar(&arguments.Etcd.Endpoint, "etcd.endpoint", "127.0.0.1:4001", "Etcd server endpoint")
	traefikCmd.PersistentFlags().StringVar(&arguments.Etcd.Prefix, "etcd.prefix", "/traefik", "Prefix used for KV store")

	traefikCmd.PersistentFlags().BoolVar(&arguments.boltdb, "boltdb", false, "Enable Boltdb backend")
	traefikCmd.PersistentFlags().BoolVar(&arguments.Boltdb.Watch, "boltdb.watch", true, "Watch provider")
	traefikCmd.PersistentFlags().StringVar(&arguments.Boltdb.Filename, "boltdb.filename", "", "Override default configuration template. For advanced users :)")
	traefikCmd.PersistentFlags().StringVar(&arguments.Boltdb.Endpoint, "boltdb.endpoint", "127.0.0.1:4001", "Boltdb server endpoint")
	traefikCmd.PersistentFlags().StringVar(&arguments.Boltdb.Prefix, "boltdb.prefix", "/traefik", "Prefix used for KV store")

	viper.BindPFlag("configFile", traefikCmd.PersistentFlags().Lookup("configFile"))
	viper.BindPFlag("port", traefikCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("graceTimeOut", traefikCmd.PersistentFlags().Lookup("graceTimeOut"))
	//	viper.BindPFlag("certificates", TraefikCmd.PersistentFlags().Lookup("certificates"))
	viper.BindPFlag("logLevel", traefikCmd.PersistentFlags().Lookup("logLevel"))
	// TODO: wait for this issue to be corrected: https://github.com/spf13/viper/issues/105
	viper.BindPFlag("providersThrottleDuration", traefikCmd.PersistentFlags().Lookup("providersThrottleDuration"))
	viper.BindPFlag("maxIdleConnsPerHost", traefikCmd.PersistentFlags().Lookup("maxIdleConnsPerHost"))
	viper.SetDefault("certificates", &Certificates{})
	viper.SetDefault("providersThrottleDuration", time.Duration(2*time.Second))
}

func run() {
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)

	// load global configuration
	globalConfiguration := LoadConfiguration()

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
	defer server.Close()
	log.Info("Shutting down")
}
