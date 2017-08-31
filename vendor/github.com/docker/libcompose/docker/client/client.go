package client

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/homedir"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/libcompose/version"
)

const (
	// DefaultAPIVersion is the default docker API version set by libcompose
	DefaultAPIVersion   = "v1.20"
	defaultTrustKeyFile = "key.json"
	defaultCaFile       = "ca.pem"
	defaultKeyFile      = "key.pem"
	defaultCertFile     = "cert.pem"
)

var (
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
)

func init() {
	if dockerCertPath == "" {
		dockerCertPath = cliconfig.Dir()
	}
}

// Options holds docker client options (host, tls, ..)
type Options struct {
	TLS        bool
	TLSVerify  bool
	TLSOptions tlsconfig.Options
	TrustKey   string
	Host       string
	APIVersion string
}

// Create creates a docker client based on the specified options.
func Create(c Options) (client.APIClient, error) {
	if c.Host == "" {
		if os.Getenv("DOCKER_API_VERSION") == "" {
			os.Setenv("DOCKER_API_VERSION", DefaultAPIVersion)
		}
		client, err := client.NewEnvClient()
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	apiVersion := c.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	if c.TLSOptions.CAFile == "" {
		c.TLSOptions.CAFile = filepath.Join(dockerCertPath, defaultCaFile)
	}
	if c.TLSOptions.CertFile == "" {
		c.TLSOptions.CertFile = filepath.Join(dockerCertPath, defaultCertFile)
	}
	if c.TLSOptions.KeyFile == "" {
		c.TLSOptions.KeyFile = filepath.Join(dockerCertPath, defaultKeyFile)
	}
	if c.TrustKey == "" {
		c.TrustKey = filepath.Join(homedir.Get(), ".docker", defaultTrustKeyFile)
	}
	if c.TLSVerify {
		c.TLS = true
	}
	if c.TLS {
		c.TLSOptions.InsecureSkipVerify = !c.TLSVerify
	}

	var httpClient *http.Client
	if c.TLS {
		config, err := tlsconfig.Client(c.TLSOptions)
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: config,
		}
		proto, addr, _, err := client.ParseHost(c.Host)
		if err != nil {
			return nil, err
		}

		if err := sockets.ConfigureTransport(tr, proto, addr); err != nil {
			return nil, err
		}

		httpClient = &http.Client{
			Transport: tr,
		}
	}

	customHeaders := map[string]string{}
	customHeaders["User-Agent"] = fmt.Sprintf("Libcompose-Client/%s (%s)", version.VERSION, runtime.GOOS)

	client, err := client.NewClient(c.Host, apiVersion, httpClient, customHeaders)
	if err != nil {
		return nil, err
	}
	return client, nil
}
