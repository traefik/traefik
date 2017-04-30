package kubernetes

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

var _ provider.Provider = (*Provider)(nil)

const (
	annotationFrontendRuleType = "traefik.frontend.rule.type"
	ruleTypePathPrefixStrip    = "PathPrefixStrip"
	ruleTypePathStrip          = "PathStrip"
	ruleTypePath               = "Path"
	ruleTypePathPrefix         = "PathPrefix"

	annotationKubernetesWhitelistSourceRange = "ingress.kubernetes.io/whitelist-source-range"
)

const traefikDefaultRealm = "traefik"

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider  `mapstructure:",squash"`
	Endpoint               string     `description:"Kubernetes server endpoint (required for external cluster client)"`
	Token                  string     `description:"Kubernetes bearer token (not needed for in-cluster client)"`
	CertAuthFilePath       string     `description:"Kubernetes certificate authority file path (not needed for in-cluster client)"`
	DisablePassHostHeaders bool       `description:"Kubernetes disable PassHost Headers"`
	Namespaces             Namespaces `description:"Kubernetes namespaces"`
	LabelSelector          string     `description:"Kubernetes api label selector to use"`
	lastConfiguration      safe.Safe
}

func (p *Provider) newK8sClient() (Client, error) {
	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %v", p.Endpoint)
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		log.Infof("Creating in-cluster Provider client%s\n", withEndpoint)
		return NewInClusterClient(p.Endpoint)
	}

	log.Infof("Creating cluster-external Provider client%s\n", withEndpoint)
	return NewExternalClusterClient(p.Endpoint, p.Token, p.CertAuthFilePath)
}

// Provide allows the k8s provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	k8sClient, err := p.newK8sClient()
	if err != nil {
		return err
	}
	p.Constraints = append(p.Constraints, constraints...)

	pool.Go(func(stop chan bool) {
		operation := func() error {
			for {
				stopWatch := make(chan struct{}, 1)
				defer close(stopWatch)
				log.Debugf("Using label selector: '%s'", p.LabelSelector)
				eventsChan, err := k8sClient.WatchAll(p.LabelSelector, stopWatch)
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
						return nil
					case event := <-eventsChan:
						log.Debugf("Received event from kubernetes %+v", event)
						templateObjects, err := p.loadIngresses(k8sClient)
						if err != nil {
							return err
						}
						if reflect.DeepEqual(p.lastConfiguration.Get(), templateObjects) {
							log.Debugf("Skipping event from kubernetes %+v", event)
						} else {
							p.lastConfiguration.Set(templateObjects)
							configurationChan <- types.ConfigMessage{
								ProviderName:  "kubernetes",
								Configuration: p.loadConfig(*templateObjects),
							}
						}
					}
				}
			}
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Provider server %+v", err)
		}
	})

	return nil
}

func (p *Provider) loadIngresses(k8sClient Client) (*types.Configuration, error) {
	ingresses := k8sClient.GetIngresses(p.Namespaces)

	templateObjects := types.Configuration{
		map[string]*types.Backend{},
		map[string]*types.Frontend{},
	}
	for _, i := range ingresses {
		ingressClass := i.Annotations["kubernetes.io/ingress.class"]

		if !shouldProcessIngress(ingressClass) {
			continue
		}

		for _, r := range i.Spec.Rules {
			if r.HTTP == nil {
				log.Warnf("Error in ingress: HTTP is nil")
				continue
			}
			for _, pa := range r.HTTP.Paths {
				if _, exists := templateObjects.Backends[r.Host+pa.Path]; !exists {
					templateObjects.Backends[r.Host+pa.Path] = &types.Backend{
						Servers: make(map[string]types.Server),
						LoadBalancer: &types.LoadBalancer{
							Sticky: false,
							Method: "wrr",
						},
					}
				}

				PassHostHeader := p.getPassHostHeader()

				passHostHeaderAnnotation, ok := i.Annotations["traefik.frontend.passHostHeader"]
				switch {
				case !ok:
					// No op.
				case passHostHeaderAnnotation == "false":
					PassHostHeader = false
				case passHostHeaderAnnotation == "true":
					PassHostHeader = true
				default:
					log.Warnf("Unknown value '%s' for traefik.frontend.passHostHeader, falling back to %s", passHostHeaderAnnotation, PassHostHeader)
				}
				if realm := i.Annotations["ingress.kubernetes.io/auth-realm"]; realm != "" && realm != traefikDefaultRealm {
					return nil, errors.New("no realm customization supported")
				}

				witelistSourceRangeAnnotation := i.Annotations[annotationKubernetesWhitelistSourceRange]
				whitelistSourceRange := provider.SplitAndTrimString(witelistSourceRangeAnnotation)

				if _, exists := templateObjects.Frontends[r.Host+pa.Path]; !exists {
					basicAuthCreds, err := handleBasicAuthConfig(i, k8sClient)
					if err != nil {
						return nil, err
					}
					templateObjects.Frontends[r.Host+pa.Path] = &types.Frontend{
						Backend:              r.Host + pa.Path,
						PassHostHeader:       PassHostHeader,
						Routes:               make(map[string]types.Route),
						Priority:             len(pa.Path),
						BasicAuth:            basicAuthCreds,
						WhitelistSourceRange: whitelistSourceRange,
					}
				}
				if len(r.Host) > 0 {
					rule := "Host:" + r.Host

					if strings.Contains(r.Host, "*") {
						rule = "HostRegexp:" + strings.Replace(r.Host, "*", "{subdomain:[A-Za-z0-9-_]+}", 1)
					}

					if _, exists := templateObjects.Frontends[r.Host+pa.Path].Routes[r.Host]; !exists {
						templateObjects.Frontends[r.Host+pa.Path].Routes[r.Host] = types.Route{
							Rule: rule,
						}
					}
				}

				if len(pa.Path) > 0 {
					ruleType, unknown := getRuleTypeFromAnnotation(i.Annotations)
					switch {
					case unknown:
						log.Warnf("Unknown RuleType '%s' for Ingress %s/%s, falling back to PathPrefix", ruleType, i.ObjectMeta.Namespace, i.ObjectMeta.Name)
						fallthrough
					case ruleType == "":
						ruleType = ruleTypePathPrefix
					}

					templateObjects.Frontends[r.Host+pa.Path].Routes[pa.Path] = types.Route{
						Rule: ruleType + ":" + pa.Path,
					}
				}

				service, exists, err := k8sClient.GetService(i.ObjectMeta.Namespace, pa.Backend.ServiceName)
				if err != nil {
					log.Errorf("Error while retrieving service information from k8s API %s/%s: %v", i.ObjectMeta.Namespace, pa.Backend.ServiceName, err)
					return nil, err
				}

				if !exists {
					log.Errorf("Service not found for %s/%s", i.ObjectMeta.Namespace, pa.Backend.ServiceName)
					delete(templateObjects.Frontends, r.Host+pa.Path)
					continue
				}

				if expression := service.Annotations["traefik.backend.circuitbreaker"]; expression != "" {
					templateObjects.Backends[r.Host+pa.Path].CircuitBreaker = &types.CircuitBreaker{
						Expression: expression,
					}
				}
				if service.Annotations["traefik.backend.loadbalancer.method"] == "drr" {
					templateObjects.Backends[r.Host+pa.Path].LoadBalancer.Method = "drr"
				}
				if service.Annotations["traefik.backend.loadbalancer.sticky"] == "true" {
					templateObjects.Backends[r.Host+pa.Path].LoadBalancer.Sticky = true
				}

				protocol := "http"
				for _, port := range service.Spec.Ports {
					if equalPorts(port, pa.Backend.ServicePort) {
						if port.Port == 443 {
							protocol = "https"
						}
						if service.Spec.Type == "ExternalName" {
							url := protocol + "://" + service.Spec.ExternalName
							name := url

							templateObjects.Backends[r.Host+pa.Path].Servers[name] = types.Server{
								URL:    url,
								Weight: 1,
							}
						} else {
							endpoints, exists, err := k8sClient.GetEndpoints(service.ObjectMeta.Namespace, service.ObjectMeta.Name)
							if err != nil {
								log.Errorf("Error retrieving endpoints %s/%s: %v", service.ObjectMeta.Namespace, service.ObjectMeta.Name, err)
								return nil, err
							}

							if !exists {
								log.Errorf("Endpoints not found for %s/%s", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
								continue
							}

							if len(endpoints.Subsets) == 0 {
								log.Warnf("Service endpoints not found for %s/%s, falling back to Service ClusterIP", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
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
						}
						break
					}
				}
			}
		}
	}
	return &templateObjects, nil
}

func handleBasicAuthConfig(i *v1beta1.Ingress, k8sClient Client) ([]string, error) {
	authType, exists := i.Annotations["ingress.kubernetes.io/auth-type"]
	if !exists {
		return nil, nil
	}
	if strings.ToLower(authType) != "basic" {
		return nil, fmt.Errorf("unsupported auth-type: %q", authType)
	}
	authSecret := i.Annotations["ingress.kubernetes.io/auth-secret"]
	if authSecret == "" {
		return nil, errors.New("auth-secret annotation must be set")
	}
	basicAuthCreds, err := loadAuthCredentials(i.Namespace, authSecret, k8sClient)
	if err != nil {
		return nil, err
	}
	if len(basicAuthCreds) == 0 {
		return nil, errors.New("secret file without credentials")
	}
	return basicAuthCreds, nil
}

func loadAuthCredentials(namespace, secretName string, k8sClient Client) ([]string, error) {
	secret, ok, err := k8sClient.GetSecret(namespace, secretName)
	switch { // keep order of case conditions
	case err != nil:
		return nil, fmt.Errorf("failed to fetch secret %q/%q: %s", namespace, secretName, err)
	case !ok:
		return nil, fmt.Errorf("secret %q/%q not found", namespace, secretName)
	case secret == nil:
		return nil, errors.New("secret data must not be nil")
	case len(secret.Data) != 1:
		return nil, errors.New("secret must contain single element only")
	default:
	}
	var firstSecret []byte
	for _, v := range secret.Data {
		firstSecret = v
		break
	}
	creds := make([]string, 0)
	scanner := bufio.NewScanner(bytes.NewReader(firstSecret))
	for scanner.Scan() {
		if cred := scanner.Text(); cred != "" {
			creds = append(creds, cred)
		}
	}
	return creds, nil
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

func shouldProcessIngress(ingressClass string) bool {
	switch ingressClass {
	case "", "traefik":
		return true
	default:
		return false
	}
}

func (p *Provider) getPassHostHeader() bool {
	if p.DisablePassHostHeaders {
		return false
	}
	return true
}

func (p *Provider) loadConfig(templateObjects types.Configuration) *types.Configuration {
	var FuncMap = template.FuncMap{}
	configuration, err := p.GetConfiguration("templates/kubernetes.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func getRuleTypeFromAnnotation(annotations map[string]string) (ruleType string, unknown bool) {
	ruleType = annotations[annotationFrontendRuleType]
	for _, knownRuleType := range []string{
		ruleTypePathPrefixStrip,
		ruleTypePathStrip,
		ruleTypePath,
		ruleTypePathPrefix,
	} {
		if strings.ToLower(ruleType) == strings.ToLower(knownRuleType) {
			return knownRuleType, false
		}
	}

	if ruleType != "" {
		// Annotation is set but does not match anything we know.
		unknown = true
	}

	return ruleType, unknown
}
