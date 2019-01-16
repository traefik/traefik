package marathon

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/old/types"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/gambol99/go-marathon"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultTemplateRule The default template for the default rule.
	DefaultTemplateRule   = "Host:{{ normalize .Name }}"
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

var _ provider.Provider = (*Provider)(nil)

// Provider holds configuration of the provider.
type Provider struct {
	provider.BaseProvider
	Endpoint                  string           `description:"Marathon server endpoint. You can also specify multiple endpoint for Marathon" export:"true"`
	DefaultRule               string           `description:"Default rule"`
	ExposedByDefault          bool             `description:"Expose Marathon apps by default" export:"true"`
	DCOSToken                 string           `description:"DCOSToken for DCOS environment, This will override the Authorization header" export:"true"`
	FilterMarathonConstraints bool             `description:"Enable use of Marathon constraints in constraint filtering" export:"true"`
	TLS                       *types.ClientTLS `description:"Enable TLS support" export:"true"`
	DialerTimeout             parse.Duration   `description:"Set a dialer timeout for Marathon" export:"true"`
	ResponseHeaderTimeout     parse.Duration   `description:"Set a response header timeout for Marathon" export:"true"`
	TLSHandshakeTimeout       parse.Duration   `description:"Set a TLS handhsake timeout for Marathon" export:"true"`
	KeepAlive                 parse.Duration   `description:"Set a TCP Keep Alive time in seconds" export:"true"`
	ForceTaskHostname         bool             `description:"Force to use the task's hostname." export:"true"`
	Basic                     *Basic           `description:"Enable basic authentication" export:"true"`
	RespectReadinessChecks    bool             `description:"Filter out tasks with non-successful readiness checks during deployments" export:"true"`
	readyChecker              *readinessChecker
	marathonClient            marathon.Marathon
	defaultRuleTpl            *template.Template
}

// Basic holds basic authentication specific configurations
type Basic struct {
	HTTPBasicAuthUser string `description:"Basic authentication User"`
	HTTPBasicPassword string `description:"Basic authentication Password"`
}

// Init the provider
func (p *Provider) Init() error {
	fm := template.FuncMap{
		"strsToItfs": func(values []string) []interface{} {
			var r []interface{}
			for _, v := range values {
				r = append(r, v)
			}
			return r
		},
	}

	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, fm)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %v", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	return p.BaseProvider.Init()
}

// Provide allows the marathon provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- config.Message, pool *safe.Pool) error {
	ctx := log.With(context.Background(), log.Str(log.ProviderName, "marathon"))
	logger := log.FromContext(ctx)

	operation := func() error {

		confg := marathon.NewDefaultConfig()
		confg.URL = p.Endpoint
		confg.EventsTransport = marathon.EventsTransportSSE
		if p.Trace {
			confg.LogOutput = log.CustomWriterLevel(logrus.DebugLevel, traceMaxScanTokenSize)
		}
		if p.Basic != nil {
			confg.HTTPBasicAuthUser = p.Basic.HTTPBasicAuthUser
			confg.HTTPBasicPassword = p.Basic.HTTPBasicPassword
		}
		var rc *readinessChecker
		if p.RespectReadinessChecks {
			logger.Debug("Enabling Marathon readiness checker")
			rc = defaultReadinessChecker(p.Trace)
		}
		p.readyChecker = rc

		if len(p.DCOSToken) > 0 {
			confg.DCOSToken = p.DCOSToken
		}
		TLSConfig, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return err
		}
		confg.HTTPClient = &http.Client{
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
		client, err := marathon.NewClient(confg)
		if err != nil {
			logger.Errorf("Failed to create a client for marathon, error: %s", err)
			return err
		}
		p.marathonClient = client

		if p.Watch {
			update, err := client.AddEventsListener(marathonEventIDs)
			if err != nil {
				logger.Errorf("Failed to register for events, %s", err)
				return err
			}
			pool.Go(func(stop chan bool) {
				defer close(update)
				for {
					select {
					case <-stop:
						return
					case event := <-update:
						logger.Debugf("Received provider event %s", event)

						conf := p.getConfigurations(ctx)
						if conf != nil {
							configurationChan <- config.Message{
								ProviderName:  "marathon",
								Configuration: conf,
							}
						}
					}
				}
			})
		}

		configuration := p.getConfigurations(ctx)
		configurationChan <- config.Message{
			ProviderName:  "marathon",
			Configuration: configuration,
		}
		return nil
	}

	notify := func(err error, time time.Duration) {
		logger.Errorf("Provider connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		logger.Errorf("Cannot connect to Provider server: %+v", err)
	}
	return nil
}

func (p *Provider) getConfigurations(ctx context.Context) *config.Configuration {
	applications, err := p.getApplications()
	if err != nil {
		log.FromContext(ctx).Errorf("Failed to retrieve Marathon applications: %v", err)
		return nil
	}

	return p.buildConfiguration(ctx, applications)
}

func (p *Provider) getApplications() (*marathon.Applications, error) {
	v := url.Values{}
	v.Add("embed", "apps.tasks")
	v.Add("embed", "apps.deployments")
	v.Add("embed", "apps.readiness")

	return p.marathonClient.Applications(v)
}
