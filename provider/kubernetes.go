package provider

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/provider/k8s"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	serviceAccountToken  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	serviceAccountCACert = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

// Kubernetes holds configurations of the Kubernetes provider.
type Kubernetes struct {
	BaseProvider           `mapstructure:",squash"`
	Endpoint               string
	disablePassHostHeaders bool
	Namespaces             []string
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
	if len(kubernetesPort) > 0 && len(kubernetesHost) > 0 {
		provider.Endpoint = "https://" + kubernetesHost + ":" + kubernetesPort
	}
	log.Debugf("Kubernetes endpoint: %s", provider.Endpoint)
	return k8s.NewClient(provider.Endpoint, caCert, token)
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Kubernetes) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	k8sClient, err := provider.createClient()
	if err != nil {
		return err
	}
	backOff := backoff.NewExponentialBackOff()

	pool.Go(func(stop chan bool) {
		operation := func() error {
			for {
				stopWatch := make(chan bool, 5)
				defer close(stopWatch)
				eventsChan, errEventsChan, err := k8sClient.WatchAll(stopWatch)
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
			Watch:
				for {
					select {
					case <-stop:
						stopWatch <- true
						return nil
					case err, ok := <-errEventsChan:
						stopWatch <- true
						if ok && strings.Contains(err.Error(), io.EOF.Error()) {
							// edge case, kubernetes long-polling disconnection
							break Watch
						}
						return err
					case event := <-eventsChan:
						log.Debugf("Received event from kubernetes %+v", event)
						templateObjects, err := provider.loadIngresses(k8sClient)
						if err != nil {
							return err
						}
						configurationChan <- types.ConfigMessage{
							ProviderName:  "kubernetes",
							Configuration: provider.loadConfig(*templateObjects),
						}
					}
				}
			}
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Kubernetes connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, backOff, notify)
		if err != nil {
			log.Fatalf("Cannot connect to Kubernetes server %+v", err)
		}
	})

	templateObjects, err := provider.loadIngresses(k8sClient)
	if err != nil {
		return err
	}
	configurationChan <- types.ConfigMessage{
		ProviderName:  "kubernetes",
		Configuration: provider.loadConfig(*templateObjects),
	}

	return nil
}

func (provider *Kubernetes) loadIngresses(k8sClient k8s.Client) (*types.Configuration, error) {
	ingresses, err := k8sClient.GetIngresses(func(ingress k8s.Ingress) bool {
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
					default:
						log.Warnf("Unknown RuleType `%s`, falling back to `PathPrefix", ruleType)
						ruleType = "PathPrefix"
					}

					templateObjects.Frontends[r.Host+pa.Path].Routes[pa.Path] = types.Route{
						Rule: ruleType + ":" + pa.Path,
					}
				}
				services, err := k8sClient.GetServices(func(service k8s.Service) bool {
					return service.ObjectMeta.Namespace == i.ObjectMeta.Namespace && service.Name == pa.Backend.ServiceName
				})
				if err != nil {
					log.Warnf("Error retrieving services: %v", err)
					continue
				}
				if len(services) == 0 {
					// no backends found, delete frontend...
					delete(templateObjects.Frontends, r.Host+pa.Path)
					log.Warnf("Error retrieving services %s", pa.Backend.ServiceName)
				}
				for _, service := range services {
					protocol := "http"
					for _, port := range service.Spec.Ports {
						if equalPorts(port, pa.Backend.ServicePort) {
							if port.Port == 443 {
								protocol = "https"
							}
							templateObjects.Backends[r.Host+pa.Path].Servers[string(service.UID)] = types.Server{
								URL:    protocol + "://" + service.Spec.ClusterIP + ":" + strconv.Itoa(port.Port),
								Weight: 1,
							}
							break
						}
					}
				}
			}
		}
	}
	return &templateObjects, nil
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
	if provider.disablePassHostHeaders {
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
