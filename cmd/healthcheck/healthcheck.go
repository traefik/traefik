package healthcheck

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/cmd"
	"github.com/containous/traefik/configuration"
)

// NewCmd builds a new HealthCheck command
func NewCmd(traefikConfiguration *cmd.TraefikConfiguration, traefikPointersConfiguration *cmd.TraefikConfiguration) *flaeg.Command {
	return &flaeg.Command{
		Name:                  "healthcheck",
		Description:           `Calls traefik /ping to check health (web provider must be enabled)`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Run:                   runCmd(traefikConfiguration),
		Metadata: map[string]string{
			"parseAllSources": "true",
		},
	}
}

func runCmd(traefikConfiguration *cmd.TraefikConfiguration) func() error {
	return func() error {
		traefikConfiguration.GlobalConfiguration.SetEffectiveConfiguration(traefikConfiguration.ConfigFile)

		resp, errPing := Do(traefikConfiguration.GlobalConfiguration)
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
	}
}

// Do try to do a healthcheck
func Do(globalConfiguration configuration.GlobalConfiguration) (*http.Response, error) {
	if globalConfiguration.Ping == nil {
		return nil, errors.New("please enable `ping` to use health check")
	}
	pingEntryPoint, ok := globalConfiguration.EntryPoints[globalConfiguration.Ping.EntryPoint]
	if !ok {
		return nil, errors.New("missing `ping` entrypoint")
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
	path := "/"
	if globalConfiguration.Web != nil {
		path = globalConfiguration.Web.Path
	}
	return client.Head(protocol + "://" + pingEntryPoint.Address + path + "ping")
}
