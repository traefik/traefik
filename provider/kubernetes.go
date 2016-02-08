package provider

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/provider/k8s"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"io/ioutil"
	"os"
	"text/template"
	"time"
)

const (
	serviceAccountToken  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	serviceAccountCACert = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

// Kubernetes holds configurations of the Kubernetes provider.
type Kubernetes struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Kubernetes) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	var token string
	tokenBytes, err := ioutil.ReadFile(serviceAccountToken)
	if err == nil {
		token = string(tokenBytes)
		log.Debugf("Kubernetes token: %s", token)
	} else {
		log.Debugf("Kubernetes load token error: %s", err)
	}
	caCert, err := ioutil.ReadFile(serviceAccountCACert)
	if err == nil {
		log.Debugf("Kubernetes CA cert: %s", serviceAccountCACert)
	} else {
		log.Debugf("Kubernetes load token error: %s", err)
	}
	kubernetesHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	kubernetesPort := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	if len(kubernetesPort) > 0 && len(kubernetesHost) > 0 {
		provider.Endpoint = "https://" + kubernetesHost + ":" + kubernetesPort
	}
	log.Debugf("Kubernetes endpoint: %s", provider.Endpoint)
	k8sClient, err := k8s.NewClient(provider.Endpoint, caCert, token)
	if err != nil {
		return err
	}

	pool.Go(func(stop chan bool) {
		stopWatch := make(chan bool)
		operation := func() error {
			select {
			case <-stop:
				return nil
			default:
			}
			ingressesChan, errChan, err := k8sClient.WatchIngresses(func(ingress k8s.Ingress) bool {
				return true
			}, stopWatch)
			if err != nil {
				log.Errorf("Error retrieving ingresses: %v", err)
				return err
			}
			for {
				templateObjects := types.Configuration{
					map[string]*types.Backend{},
					map[string]*types.Frontend{},
				}
				select {
				case <-stop:
					stopWatch <- true
					return nil
				case err := <-errChan:
					return err
				case event := <-ingressesChan:
					log.Debugf("Received event from kubenetes %+v", event)
					ingresses, err := k8sClient.GetIngresses(func(ingress k8s.Ingress) bool {
						return true
					})
					if err != nil {
						log.Errorf("Error retrieving ingresses: %+v", err)
						continue
					}
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
										Backend: r.Host + pa.Path,
										Routes:  make(map[string]types.Route),
									}
								}
								if _, exists := templateObjects.Frontends[r.Host+pa.Path].Routes[r.Host]; !exists {
									templateObjects.Frontends[r.Host+pa.Path].Routes[r.Host] = types.Route{
										Rule: "Host:" + r.Host,
									}
								}
								if len(pa.Path) > 0 {
									templateObjects.Frontends[r.Host+pa.Path].Routes[pa.Path] = types.Route{
										Rule: "Path:" + pa.Path,
									}
								}
								services, err := k8sClient.GetServices(func(service k8s.Service) bool {
									return service.Name == pa.Backend.ServiceName
								})
								if err != nil {
									log.Errorf("Error retrieving services: %v", err)
									continue
								}
								for _, service := range services {
									var protocol string
									for _, port := range service.Spec.Ports {
										if port.Port == pa.Backend.ServicePort.IntValue() {
											protocol = port.Name
											break
										}
									}
									templateObjects.Backends[r.Host+pa.Path].Servers[string(service.UID)] = types.Server{
										URL:    protocol + "://" + service.Spec.ClusterIP + ":" + pa.Backend.ServicePort.String(),
										Weight: 1,
									}
								}
							}
						}
					}

					configurationChan <- types.ConfigMessage{
						ProviderName:  "kubernetes",
						Configuration: provider.loadConfig(templateObjects),
					}
				}
			}
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Kubernetes connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
		if err != nil {
			log.Fatalf("Cannot connect to Kubernetes server %+v", err)
		}
	})

	return nil
}

func (provider *Kubernetes) loadConfig(templateObjects types.Configuration) *types.Configuration {
	var FuncMap = template.FuncMap{}
	configuration, err := provider.getConfiguration("templates/kubernetes.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}
