package healthcheck

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v3/cmd"
	"github.com/traefik/traefik/v3/pkg/config/static"
)

// NewCmd builds a new HealthCheck command.
func NewCmd(traefikHealthCheckConfiguration *cmd.TraefikHealthCheckCmdConfiguration, loaders []cli.ResourceLoader) *cli.Command {
	return &cli.Command{
		Name:          "healthcheck",
		Description:   `Calls Traefik /ping endpoint (disabled by default) to check the health of Traefik.`,
		Configuration: traefikHealthCheckConfiguration,
		Run:           runCmd(traefikHealthCheckConfiguration),
		Resources:     loaders,
	}
}

func runCmd(traefikHealthCheckConfiguration *cmd.TraefikHealthCheckCmdConfiguration) func(args []string) error {
	return func(args []string) error {
		var resp *http.Response
		var errPing error
		if traefikHealthCheckConfiguration.URL != "" {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, errPing = client.Head(traefikHealthCheckConfiguration.URL)
		} else {
			traefikHealthCheckConfiguration.Configuration.SetEffectiveConfiguration()
			resp, errPing = Do(traefikHealthCheckConfiguration.Configuration)
		}

		if resp != nil {
			resp.Body.Close()
		}
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

// Do try to do a healthcheck.
func Do(staticConfiguration static.Configuration) (*http.Response, error) {
	if staticConfiguration.Ping == nil {
		return nil, errors.New("please enable `ping` to use health check")
	}

	ep := staticConfiguration.Ping.EntryPoint
	if ep == "" {
		ep = "traefik"
	}

	pingEntryPoint, ok := staticConfiguration.EntryPoints[ep]
	if !ok {
		return nil, fmt.Errorf("ping: missing %s entry point", ep)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	protocol := "http"

	// TODO Handle TLS on ping etc...
	// if pingEntryPoint.TLS != nil {
	// 	protocol = "https"
	// 	tr := &http.Transport{
	// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// 	}
	// 	client.Transport = tr
	// }

	path := "/"

	return client.Head(protocol + "://" + pingEntryPoint.GetAddress() + path + "ping")
}
