package provider

import (
	"fmt"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/k8s"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
)

const (
	serviceAccountToken  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	serviceAccountCACert = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	defaultKubeEndpoint  = "http://127.0.0.1:8080"
)

// Namespaces holds kubernetes namespaces
type Namespaces []string

//Set adds strings elem into the the parser
//it splits str on , and ;
func (ns *Namespaces) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*ns = append(*ns, slice...)
	return nil
}

//Get []string
func (ns *Namespaces) Get() interface{} { return Namespaces(*ns) }

//String return slice in a string
func (ns *Namespaces) String() string { return fmt.Sprintf("%v", *ns) }

//SetValue sets []string into the parser
func (ns *Namespaces) SetValue(val interface{}) {
	*ns = Namespaces(val.(Namespaces))
}

var _ Provider = (*Kubernetes)(nil)

// Kubernetes holds configurations of the Kubernetes provider.
type Kubernetes struct {
	BaseProvider           `mapstructure:",squash"`
	Endpoint               string     `description:"Kubernetes server endpoint"`
	DisablePassHostHeaders bool       `description:"Kubernetes disable PassHost Headers"`
	Namespaces             Namespaces `description:"Kubernetes namespaces"`
	LabelSelector          string     `description:"Kubernetes api label selector to use"`
	lastConfiguration      safe.Safe
}

func (provider *Kubernetes) createClient() (k8s.Client, error) {
	var token string
	tokenBytes, err := ioutil.ReadFile(serviceAccountToken)
	if err == nil {
		token = string(tokenBytes)
		log.Debugf("Kubernetes token: %s", token)
	} else {
		log.Errorf("Kubernetes load token error: %s", err)
	}
	caCert, err := ioutil.ReadFile(serviceAccountCACert)
	if err == nil {
		log.Debugf("Kubernetes CA cert: %s", serviceAccountCACert)
	} else {
		log.Errorf("Kubernetes load token error: %s", err)
	}
	kubernetesHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	kubernetesPort := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	// Prioritize user provided kubernetes endpoint since kube container runtime will almost always have it
	if provider.Endpoint == "" && len(kubernetesPort) > 0 && len(kubernetesHost) > 0 {
		log.Debugf("Using environment provided kubernetes endpoint")
		provider.Endpoint = "https://" + kubernetesHost + ":" + kubernetesPort
	}
	if provider.Endpoint == "" {
		log.Debugf("Using default kubernetes api endpoint")
		provider.Endpoint = defaultKubeEndpoint
	}
	log.Debugf("Kubernetes endpoint: %s", provider.Endpoint)
	return k8s.NewClient(provider.Endpoint, caCert, token)
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Kubernetes) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	k8sClient, err := provider.createClient()
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
				eventsChan, errEventsChan, err := k8sClient.WatchAll(provider.LabelSelector, stopWatch)
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
					case err, _ := <-errEventsChan:
						stopWatch <- true
						return err
					case event := <-eventsChan:
						log.Debugf("Received event from kubernetes %+v", event)
						templateObjects, err := provider.loadIngresses(k8sClient)
						if err != nil {
							stopWatch <- true
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
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Kubernetes server %+v", err)
		}
	})

	templateObjects, err := provider.loadIngresses(k8sClient)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(provider.lastConfiguration.Get(), templateObjects) {
		log.Debugf("Skipping configuration from kubernetes %+v", templateObjects)
	} else {
		provider.lastConfiguration.Set(templateObjects)
		configurationChan <- types.ConfigMessage{
			ProviderName:  "kubernetes",
			Configuration: provider.loadConfig(*templateObjects),
		}
	}

	return nil
}

func (provider *Kubernetes) loadIngresses(k8sClient k8s.Client) (*types.Configuration, error) {
	ingresses, err := k8sClient.GetIngresses(provider.LabelSelector, func(ingress k8s.Ingress) bool {
		if len(provider.Namespaces) == 0 {
			return true
		}
		for _, n := range provider.Namespaces {
			if ingress.ObjectMeta.Namespace == n {
				return true
			}
		}
		return false
	})
	if err != nil {
		log.Errorf("Error retrieving ingresses: %+v", err)
		return nil, err
	}
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
				service, err := k8sClient.GetService(pa.Backend.ServiceName, i.ObjectMeta.Namespace)
				if err != nil {
					log.Warnf("Error retrieving services: %v", err)
					delete(templateObjects.Frontends, r.Host+pa.Path)
					log.Warnf("Error retrieving services %s", pa.Backend.ServiceName)
					continue
				}
				protocol := "http"
				for _, port := range service.Spec.Ports {
					if equalPorts(port, pa.Backend.ServicePort) {
						if port.Port == 443 {
							protocol = "https"
						}
						endpoints, err := k8sClient.GetEndpoints(service.ObjectMeta.Name, service.ObjectMeta.Namespace)
						if err != nil {
							log.Errorf("Error retrieving endpoints: %v", err)
							continue
						}
						if len(endpoints.Subsets) == 0 {
							log.Warnf("Endpoints not found for %s/%s, falling back to Service ClusterIP", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
							templateObjects.Backends[r.Host+pa.Path].Servers[string(service.UID)] = types.Server{
								URL:    protocol + "://" + service.Spec.ClusterIP + ":" + strconv.Itoa(port.Port),
								Weight: 0,
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
										Weight: 0,
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

func endpointPortNumber(servicePort k8s.ServicePort, endpointPorts []k8s.EndpointPort) int {
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
	return servicePort.Port
}

func equalPorts(servicePort k8s.ServicePort, ingressPort k8s.IntOrString) bool {
	if servicePort.Port == ingressPort.IntValue() {
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
