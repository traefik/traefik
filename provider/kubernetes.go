package provider

import (
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/util/intstr"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/k8s"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
)

var _ Provider = (*Kubernetes)(nil)

// Kubernetes holds configurations of the Kubernetes provider.
type Kubernetes struct {
	BaseProvider           `mapstructure:",squash"`
	Endpoint               string         `description:"Kubernetes server endpoint"`
	DisablePassHostHeaders bool           `description:"Kubernetes disable PassHost Headers"`
	Namespaces             k8s.Namespaces `description:"Kubernetes namespaces"`
	LabelSelector          string         `description:"Kubernetes api label selector to use"`
	lastConfiguration      safe.Safe
}

func (provider *Kubernetes) newK8sClient() (k8s.Client, error) {
	if provider.Endpoint != "" {
		log.Infof("Creating in cluster Kubernetes client with endpoint %", provider.Endpoint)
		return k8s.NewInClusterClientWithEndpoint(provider.Endpoint)
	}
	log.Info("Creating in cluster Kubernetes client")
	return k8s.NewInClusterClient()
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Kubernetes) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	k8sClient, err := provider.newK8sClient()
	if err != nil {
		return err
	}
	provider.Constraints = append(provider.Constraints, constraints...)

	pool.Go(func(stop chan bool) {
		operation := func() error {
			for {
				stopWatch := make(chan bool, 5)
				defer close(stopWatch)
				log.Debugf("Using label selector: '%s'", provider.LabelSelector)
				eventsChan, err := k8sClient.WatchAll(provider.LabelSelector, stopWatch)
				if err != nil {
					log.Errorf("Error watching kubernetes events: %v", err)
					timer := time.NewTimer(1 * time.Second)
					select {
					case <-timer.C:
						return err
					case <-stop:
						return nil
					}
				}
				for {
					select {
					case <-stop:
						stopWatch <- true
						return nil
					case event := <-eventsChan:
						log.Debugf("Received event from kubernetes %+v", event)
						templateObjects, err := provider.loadIngresses(k8sClient)
						if err != nil {
							return err
						}
						if reflect.DeepEqual(provider.lastConfiguration.Get(), templateObjects) {
							log.Debugf("Skipping event from kubernetes %+v", event)
						} else {
							provider.lastConfiguration.Set(templateObjects)
							configurationChan <- types.ConfigMessage{
								ProviderName:  "kubernetes",
								Configuration: provider.loadConfig(*templateObjects),
							}
						}
					}
				}
			}
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Kubernetes connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Kubernetes server %+v", err)
		}
	})

	return nil
}

func (provider *Kubernetes) loadIngresses(k8sClient k8s.Client) (*types.Configuration, error) {
	ingresses := k8sClient.GetIngresses(provider.Namespaces)

	templateObjects := types.Configuration{
		map[string]*types.Backend{},
		map[string]*types.Frontend{},
	}
	PassHostHeader := provider.getPassHostHeader()
	for _, i := range ingresses {
		for _, r := range i.Spec.Rules {
			for _, pa := range r.HTTP.Paths {
				if _, exists := templateObjects.Backends[r.Host+pa.Path]; !exists {
					templateObjects.Backends[r.Host+pa.Path] = &types.Backend{
						Servers: make(map[string]types.Server),
					}
				}
				if _, exists := templateObjects.Frontends[r.Host+pa.Path]; !exists {
					templateObjects.Frontends[r.Host+pa.Path] = &types.Frontend{
						Backend:        r.Host + pa.Path,
						PassHostHeader: PassHostHeader,
						Routes:         make(map[string]types.Route),
						Priority:       len(pa.Path),
					}
				}
				if len(r.Host) > 0 {
					if _, exists := templateObjects.Frontends[r.Host+pa.Path].Routes[r.Host]; !exists {
						templateObjects.Frontends[r.Host+pa.Path].Routes[r.Host] = types.Route{
							Rule: "Host:" + r.Host,
						}
					}
				}
				if len(pa.Path) > 0 {
					ruleType := i.Annotations["traefik.frontend.rule.type"]

					switch strings.ToLower(ruleType) {
					case "pathprefixstrip":
						ruleType = "PathPrefixStrip"
					case "pathstrip":
						ruleType = "PathStrip"
					case "path":
						ruleType = "Path"
					case "pathprefix":
						ruleType = "PathPrefix"
					case "":
						ruleType = "PathPrefix"
					default:
						log.Warnf("Unknown RuleType %s for %s/%s, falling back to PathPrefix", ruleType, i.ObjectMeta.Namespace, i.ObjectMeta.Name)
						ruleType = "PathPrefix"
					}

					templateObjects.Frontends[r.Host+pa.Path].Routes[pa.Path] = types.Route{
						Rule: ruleType + ":" + pa.Path,
					}
				}
				service, exists, err := k8sClient.GetService(i.ObjectMeta.Namespace, pa.Backend.ServiceName)
				if err != nil || !exists {
					log.Warnf("Error retrieving service %s/%s: %v", i.ObjectMeta.Namespace, pa.Backend.ServiceName, err)
					delete(templateObjects.Frontends, r.Host+pa.Path)
					continue
				}

				protocol := "http"
				for _, port := range service.Spec.Ports {
					if equalPorts(port, pa.Backend.ServicePort) {
						if port.Port == 443 {
							protocol = "https"
						}
						endpoints, exists, err := k8sClient.GetEndpoints(service.ObjectMeta.Namespace, service.ObjectMeta.Name)
						if err != nil || !exists {
							log.Errorf("Error retrieving endpoints %s/%s: %v", service.ObjectMeta.Namespace, service.ObjectMeta.Name, err)
							continue
						}
						if len(endpoints.Subsets) == 0 {
							log.Warnf("Endpoints not found for %s/%s, falling back to Service ClusterIP", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
							templateObjects.Backends[r.Host+pa.Path].Servers[string(service.UID)] = types.Server{
								URL:    protocol + "://" + service.Spec.ClusterIP + ":" + strconv.Itoa(int(port.Port)),
								Weight: 1,
							}
						} else {
							for _, subset := range endpoints.Subsets {
								for _, address := range subset.Addresses {
									url := protocol + "://" + address.IP + ":" + strconv.Itoa(endpointPortNumber(port, subset.Ports))
									name := url
									if address.TargetRef != nil && address.TargetRef.Name != "" {
										name = address.TargetRef.Name
									}
									templateObjects.Backends[r.Host+pa.Path].Servers[name] = types.Server{
										URL:    url,
										Weight: 1,
									}
								}
							}
						}
						break
					}
				}
			}
		}
	}
	return &templateObjects, nil
}

func endpointPortNumber(servicePort v1.ServicePort, endpointPorts []v1.EndpointPort) int {
	if len(endpointPorts) > 0 {
		//name is optional if there is only one port
		port := endpointPorts[0]
		for _, endpointPort := range endpointPorts {
			if servicePort.Name == endpointPort.Name {
				port = endpointPort
			}
		}
		return int(port.Port)
	}
	return int(servicePort.Port)
}

func equalPorts(servicePort v1.ServicePort, ingressPort intstr.IntOrString) bool {
	if int(servicePort.Port) == ingressPort.IntValue() {
		return true
	}
	if servicePort.Name != "" && servicePort.Name == ingressPort.String() {
		return true
	}
	return false
}

func (provider *Kubernetes) getPassHostHeader() bool {
	if provider.DisablePassHostHeaders {
		return false
	}
	return true
}

func (provider *Kubernetes) loadConfig(templateObjects types.Configuration) *types.Configuration {
	var FuncMap = template.FuncMap{}
	configuration, err := provider.getConfiguration("templates/kubernetes.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}
