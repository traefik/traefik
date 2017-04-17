package provider

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	rancher "github.com/rancher/go-rancher/client"
)

const (
	// RancherDefaultWatchTime is the duration of the interval when polling rancher
	RancherDefaultWatchTime = 15 * time.Second
)

var _ Provider = (*Rancher)(nil)

// Rancher holds configurations of the Rancher provider.
type Rancher struct {
	BaseProvider     `mapstructure:",squash"`
	Endpoint         string `description:"Rancher server HTTP(S) endpoint."`
	AccessKey        string `description:"Rancher server access key."`
	SecretKey        string `description:"Rancher server Secret Key."`
	ExposedByDefault bool   `description:"Expose Services by default"`
	Domain           string `description:"Default domain used"`
}

type rancherData struct {
	Name       string
	Labels     map[string]string // List of labels set to container or service
	Containers []string
	Health     string
}

func (r rancherData) String() string {
	return fmt.Sprintf("{name:%s, labels:%v, containers: %v, health: %s}", r.Name, r.Labels, r.Containers, r.Health)
}

// Frontend Labels
func (provider *Rancher) getPassHostHeader(service rancherData) string {
	if passHostHeader, err := getServiceLabel(service, "traefik.frontend.passHostHeader"); err == nil {
		return passHostHeader
	}
	return "true"
}

func (provider *Rancher) getPriority(service rancherData) string {
	if priority, err := getServiceLabel(service, "traefik.frontend.priority"); err == nil {
		return priority
	}
	return "0"
}

func (provider *Rancher) getEntryPoints(service rancherData) []string {
	if entryPoints, err := getServiceLabel(service, "traefik.frontend.entryPoints"); err == nil {
		return strings.Split(entryPoints, ",")
	}
	return []string{}
}

func (provider *Rancher) getFrontendRule(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.frontend.rule"); err == nil {
		return label
	}
	return "Host:" + strings.ToLower(strings.Replace(service.Name, "/", ".", -1)) + "." + provider.Domain
}

func (provider *Rancher) getFrontendName(service rancherData) string {
	// Replace '.' with '-' in quoted keys because of this issue https://github.com/BurntSushi/toml/issues/78
	return Normalize(provider.getFrontendRule(service))
}

// Backend Labels
func (provider *Rancher) getLoadBalancerMethod(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.backend.loadbalancer.method"); err == nil {
		return label
	}
	return "wrr"
}

func (provider *Rancher) hasLoadBalancerLabel(service rancherData) bool {
	_, errMethod := getServiceLabel(service, "traefik.backend.loadbalancer.method")
	_, errSticky := getServiceLabel(service, "traefik.backend.loadbalancer.sticky")
	if errMethod != nil && errSticky != nil {
		return false
	}
	return true
}

func (provider *Rancher) hasCircuitBreakerLabel(service rancherData) bool {
	if _, err := getServiceLabel(service, "traefik.backend.circuitbreaker.expression"); err != nil {
		return false
	}
	return true
}

func (provider *Rancher) getCircuitBreakerExpression(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.backend.circuitbreaker.expression"); err == nil {
		return label
	}
	return "NetworkErrorRatio() > 1"
}

func (provider *Rancher) getSticky(service rancherData) string {
	if _, err := getServiceLabel(service, "traefik.backend.loadbalancer.sticky"); err == nil {
		return "true"
	}
	return "false"
}

func (provider *Rancher) getBackend(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.backend"); err == nil {
		return Normalize(label)
	}
	return Normalize(service.Name)
}

// Generall Application Stuff
func (provider *Rancher) getPort(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.port"); err == nil {
		return label
	}
	return ""
}

func (provider *Rancher) getProtocol(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.protocol"); err == nil {
		return label
	}
	return "http"
}

func (provider *Rancher) getWeight(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.weight"); err == nil {
		return label
	}
	return "0"
}

func (provider *Rancher) getDomain(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.domain"); err == nil {
		return label
	}
	return provider.Domain
}

func (provider *Rancher) hasMaxConnLabels(service rancherData) bool {
	if _, err := getServiceLabel(service, "traefik.backend.maxconn.amount"); err != nil {
		return false
	}
	if _, err := getServiceLabel(service, "traefik.backend.maxconn.extractorfunc"); err != nil {
		return false
	}
	return true
}

func (provider *Rancher) getMaxConnAmount(service rancherData) int64 {
	if label, err := getServiceLabel(service, "traefik.backend.maxconn.amount"); err == nil {
		i, errConv := strconv.ParseInt(label, 10, 64)
		if errConv != nil {
			log.Errorf("Unable to parse traefik.backend.maxconn.amount %s", label)
			return math.MaxInt64
		}
		return i
	}
	return math.MaxInt64
}

func (provider *Rancher) getMaxConnExtractorFunc(service rancherData) string {
	if label, err := getServiceLabel(service, "traefik.backend.maxconn.extractorfunc"); err == nil {
		return label
	}
	return "request.host"
}

func getServiceLabel(service rancherData, label string) (string, error) {
	for key, value := range service.Labels {
		if key == label {
			return value, nil
		}
	}
	return "", errors.New("Label not found:" + label)
}

func (provider *Rancher) createClient() (*rancher.RancherClient, error) {

	rancherURL := getenv("CATTLE_URL", provider.Endpoint)
	accessKey := getenv("CATTLE_ACCESS_KEY", provider.AccessKey)
	secretKey := getenv("CATTLE_SECRET_KEY", provider.SecretKey)

	return rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       rancherURL,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Rancher) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {

	safe.Go(func() {
		operation := func() error {
			rancherClient, err := provider.createClient()

			if err != nil {
				log.Errorf("Failed to create a client for rancher, error: %s", err)
				return err
			}

			ctx := context.Background()
			var environments = listRancherEnvironments(rancherClient)
			var services = listRancherServices(rancherClient)
			var container = listRancherContainer(rancherClient)

			var rancherData = parseRancherData(environments, services, container)

			configuration := provider.loadRancherConfig(rancherData)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "rancher",
				Configuration: configuration,
			}

			if provider.Watch {
				_, cancel := context.WithCancel(ctx)
				ticker := time.NewTicker(RancherDefaultWatchTime)
				pool.Go(func(stop chan bool) {
					for {
						select {
						case <-ticker.C:

							log.Debugf("Refreshing new Data from Rancher API")
							var environments = listRancherEnvironments(rancherClient)
							var services = listRancherServices(rancherClient)
							var container = listRancherContainer(rancherClient)

							rancherData := parseRancherData(environments, services, container)

							configuration := provider.loadRancherConfig(rancherData)
							if configuration != nil {
								configurationChan <- types.ConfigMessage{
									ProviderName:  "rancher",
									Configuration: configuration,
								}
							}
						case <-stop:
							ticker.Stop()
							cancel()
							return
						}
					}
				})
			}

			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Rancher connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Rancher Endpoint %+v", err)
		}
	})

	return nil
}

func listRancherEnvironments(client *rancher.RancherClient) []*rancher.Environment {

	var environmentList = []*rancher.Environment{}

	environments, err := client.Environment.List(nil)

	if err != nil {
		log.Errorf("Cannot get Rancher Environments %+v", err)
	}

	for k := range environments.Data {
		environmentList = append(environmentList, &environments.Data[k])
	}

	return environmentList
}

func listRancherServices(client *rancher.RancherClient) []*rancher.Service {

	var servicesList = []*rancher.Service{}

	services, err := client.Service.List(nil)

	if err != nil {
		log.Errorf("Cannot get Rancher Services %+v", err)
	}

	for k := range services.Data {
		servicesList = append(servicesList, &services.Data[k])
	}

	return servicesList
}

func listRancherContainer(client *rancher.RancherClient) []*rancher.Container {

	containerList := []*rancher.Container{}

	container, err := client.Container.List(nil)

	log.Debugf("first container len: %i", len(container.Data))

	if err != nil {
		log.Errorf("Cannot get Rancher Services %+v", err)
	}

	valid := true

	for valid {
		for k := range container.Data {
			containerList = append(containerList, &container.Data[k])
		}

		container, err = container.Next()

		if err != nil {
			break
		}

		if container == nil || len(container.Data) == 0 {
			valid = false
		}
	}

	return containerList
}

func parseRancherData(environments []*rancher.Environment, services []*rancher.Service, containers []*rancher.Container) []rancherData {
	var rancherDataList []rancherData

	for _, environment := range environments {

		for _, service := range services {
			if service.EnvironmentId != environment.Id {
				continue
			}

			rancherData := rancherData{
				Name:       environment.Name + "/" + service.Name,
				Health:     service.HealthState,
				Labels:     make(map[string]string),
				Containers: []string{},
			}

			for key, value := range service.LaunchConfig.Labels {
				rancherData.Labels[key] = value.(string)
			}

			for _, container := range containers {
				for key, value := range container.Labels {

					if key == "io.rancher.stack_service.name" && value == rancherData.Name {
						rancherData.Containers = append(rancherData.Containers, container.PrimaryIpAddress)
					}
				}
			}
			rancherDataList = append(rancherDataList, rancherData)
		}
	}

	return rancherDataList
}

func (provider *Rancher) loadRancherConfig(services []rancherData) *types.Configuration {

	var RancherFuncMap = template.FuncMap{
		"getPort":                     provider.getPort,
		"getBackend":                  provider.getBackend,
		"getWeight":                   provider.getWeight,
		"getDomain":                   provider.getDomain,
		"getProtocol":                 provider.getProtocol,
		"getPassHostHeader":           provider.getPassHostHeader,
		"getPriority":                 provider.getPriority,
		"getEntryPoints":              provider.getEntryPoints,
		"getFrontendRule":             provider.getFrontendRule,
		"hasCircuitBreakerLabel":      provider.hasCircuitBreakerLabel,
		"getCircuitBreakerExpression": provider.getCircuitBreakerExpression,
		"hasLoadBalancerLabel":        provider.hasLoadBalancerLabel,
		"getLoadBalancerMethod":       provider.getLoadBalancerMethod,
		"hasMaxConnLabels":            provider.hasMaxConnLabels,
		"getMaxConnAmount":            provider.getMaxConnAmount,
		"getMaxConnExtractorFunc":     provider.getMaxConnExtractorFunc,
		"getSticky":                   provider.getSticky,
	}

	// filter services
	filteredServices := fun.Filter(func(service rancherData) bool {
		return provider.serviceFilter(service)
	}, services).([]rancherData)

	frontends := map[string]rancherData{}
	backends := map[string]rancherData{}

	for _, service := range filteredServices {
		frontendName := provider.getFrontendName(service)
		frontends[frontendName] = service
		backendName := provider.getBackend(service)
		backends[backendName] = service
	}

	templateObjects := struct {
		Frontends map[string]rancherData
		Backends  map[string]rancherData
		Domain    string
	}{
		frontends,
		backends,
		provider.Domain,
	}

	configuration, err := provider.GetConfiguration("templates/rancher.tmpl", RancherFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration

}

func (provider *Rancher) serviceFilter(service rancherData) bool {

	if service.Labels["traefik.port"] == "" {
		log.Debugf("Filtering service %s without traefik.port label", service.Name)
		return false
	}

	if !isServiceEnabled(service, provider.ExposedByDefault) {
		log.Debugf("Filtering disabled service %s", service.Name)
		return false
	}

	if service.Health != "" && service.Health != "healthy" {
		log.Debugf("Filtering unhealthy or starting service %s", service.Name)
		return false
	}

	return true
}

func isServiceEnabled(service rancherData, exposedByDefault bool) bool {

	if service.Labels["traefik.enable"] != "" {
		var v = service.Labels["traefik.enable"]
		return exposedByDefault && v != "false" || v == "true"
	}
	return exposedByDefault
}
