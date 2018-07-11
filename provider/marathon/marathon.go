package marathon

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
	"github.com/sirupsen/logrus"
)

const (
	traceMaxScanTokenSize = 1024 * 1024
	marathonEventIDs      = marathon.EventIDApplications |
		marathon.EventIDAddHealthCheck |
		marathon.EventIDDeploymentSuccess |
		marathon.EventIDDeploymentFailed |
		marathon.EventIDDeploymentInfo |
		marathon.EventIDDeploymentStepSuccess |
		marathon.EventIDDeploymentStepFailed
)

// TaskState denotes the Mesos state a task can have.
type TaskState string

const (
	taskStateRunning TaskState = "TASK_RUNNING"
	taskStateStaging TaskState = "TASK_STAGING"
)

const (
	labelIPAddressIdx         = "traefik.ipAddressIdx"
	labelLbCompatibilityGroup = "HAPROXY_GROUP"
	labelLbCompatibility      = "HAPROXY_0_VHOST"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configuration of the provider.
type Provider struct {
	provider.BaseProvider
	Endpoint                  string           `description:"Marathon server endpoint. You can also specify multiple endpoint for Marathon" export:"true"`
	Domain                    string           `description:"Default domain used" export:"true"`
	ExposedByDefault          bool             `description:"Expose Marathon apps by default" export:"true"`
	GroupsAsSubDomains        bool             `description:"Convert Marathon groups to subdomains" export:"true"`
	DCOSToken                 string           `description:"DCOSToken for DCOS environment, This will override the Authorization header" export:"true"`
	MarathonLBCompatibility   bool             `description:"Add compatibility with marathon-lb labels" export:"true"`
	FilterMarathonConstraints bool             `description:"Enable use of Marathon constraints in constraint filtering" export:"true"`
	TLS                       *types.ClientTLS `description:"Enable TLS support" export:"true"`
	DialerTimeout             flaeg.Duration   `description:"Set a dialer timeout for Marathon" export:"true"`
	ResponseHeaderTimeout     flaeg.Duration   `description:"Set a response header timeout for Marathon" export:"true"`
	TLSHandshakeTimeout       flaeg.Duration   `description:"Set a TLS handhsake timeout for Marathon" export:"true"`
	KeepAlive                 flaeg.Duration   `description:"Set a TCP Keep Alive time in seconds" export:"true"`
	ForceTaskHostname         bool             `description:"Force to use the task's hostname." export:"true"`
	Basic                     *Basic           `description:"Enable basic authentication" export:"true"`
	RespectReadinessChecks    bool             `description:"Filter out tasks with non-successful readiness checks during deployments" export:"true"`
	readyChecker              *readinessChecker
	marathonClient            marathon.Marathon
}

// Basic holds basic authentication specific configurations
type Basic struct {
	HTTPBasicAuthUser string `description:"Basic authentication User"`
	HTTPBasicPassword string `description:"Basic authentication Password"`
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	return p.BaseProvider.Init(constraints)
}

// Provide allows the marathon provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	operation := func() error {
		config := marathon.NewDefaultConfig()
		config.URL = p.Endpoint
		config.EventsTransport = marathon.EventsTransportSSE
		if p.Trace {
			config.LogOutput = log.CustomWriterLevel(logrus.DebugLevel, traceMaxScanTokenSize)
		}
		if p.Basic != nil {
			config.HTTPBasicAuthUser = p.Basic.HTTPBasicAuthUser
			config.HTTPBasicPassword = p.Basic.HTTPBasicPassword
		}
		var rc *readinessChecker
		if p.RespectReadinessChecks {
			log.Debug("Enabling Marathon readiness checker")
			rc = defaultReadinessChecker(p.Trace)
		}
		p.readyChecker = rc

		if len(p.DCOSToken) > 0 {
			config.DCOSToken = p.DCOSToken
		}
		TLSConfig, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return err
		}
		config.HTTPClient = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					KeepAlive: time.Duration(p.KeepAlive),
					Timeout:   time.Duration(p.DialerTimeout),
				}).DialContext,
				ResponseHeaderTimeout: time.Duration(p.ResponseHeaderTimeout),
				TLSHandshakeTimeout:   time.Duration(p.TLSHandshakeTimeout),
				TLSClientConfig:       TLSConfig,
			},
		}
		client, err := marathon.NewClient(config)
		if err != nil {
			log.Errorf("Failed to create a client for marathon, error: %s", err)
			return err
		}
		p.marathonClient = client

		if p.Watch {
			update, err := client.AddEventsListener(marathonEventIDs)
			if err != nil {
				log.Errorf("Failed to register for events, %s", err)
				return err
			}
			pool.Go(func(stop chan bool) {
				defer close(update)
				for {
					select {
					case <-stop:
						return
					case event := <-update:
						log.Debugf("Received provider event %s", event)

						configuration := p.getConfiguration()
						if configuration != nil {
							configurationChan <- types.ConfigMessage{
								ProviderName:  "marathon",
								Configuration: configuration,
							}
						}
					}
				}
			})
		}

		configuration := p.getConfiguration()
		configurationChan <- types.ConfigMessage{
			ProviderName:  "marathon",
			Configuration: configuration,
		}
		return nil
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("Provider connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		log.Errorf("Cannot connect to Provider server %+v", err)
	}
	return nil
}

func (p *Provider) getConfiguration() *types.Configuration {
	applications, err := p.getApplications()
	if err != nil {
		log.Errorf("Failed to retrieve Marathon applications: %v", err)
		return nil
	}

	return p.buildConfiguration(applications)
}

func (p *Provider) getApplications() (*marathon.Applications, error) {
	v := url.Values{}
	v.Add("embed", "apps.tasks")
	v.Add("embed", "apps.deployments")
	v.Add("embed", "apps.readiness")

	return p.marathonClient.Applications(v)
}
