package kubernetes

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
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
	ruleTypePathPrefix  = "PathPrefix"
	ruleTypeReplacePath = "ReplacePath"

	annotationKubernetesIngressClass            = "kubernetes.io/ingress.class"
	annotationKubernetesAuthRealm               = "ingress.kubernetes.io/auth-realm"
	annotationKubernetesAuthType                = "ingress.kubernetes.io/auth-type"
	annotationKubernetesAuthSecret              = "ingress.kubernetes.io/auth-secret"
	annotationKubernetesRewriteTarget           = "ingress.kubernetes.io/rewrite-target"
	annotationKubernetesWhitelistSourceRange    = "ingress.kubernetes.io/whitelist-source-range"
	annotationKubernetesSSLRedirect             = "ingress.kubernetes.io/ssl-redirect"
	annotationKubernetesHSTSMaxAge              = "ingress.kubernetes.io/hsts-max-age"
	annotationKubernetesHSTSIncludeSubdomains   = "ingress.kubernetes.io/hsts-include-subdomains"
	annotationKubernetesCustomRequestHeaders    = "ingress.kubernetes.io/custom-request-headers"
	annotationKubernetesCustomResponseHeaders   = "ingress.kubernetes.io/custom-response-headers"
	annotationKubernetesAllowedHosts            = "ingress.kubernetes.io/allowed-hosts"
	annotationKubernetesProxyHeaders            = "ingress.kubernetes.io/proxy-headers"
	annotationKubernetesSSLTemporaryRedirect    = "ingress.kubernetes.io/ssl-temporary-redirect"
	annotationKubernetesSSLHost                 = "ingress.kubernetes.io/ssl-host"
	annotationKubernetesSSLProxyHeaders         = "ingress.kubernetes.io/ssl-proxy-headers"
	annotationKubernetesHSTSPreload             = "ingress.kubernetes.io/hsts-preload"
	annotationKubernetesForceHSTSHeader         = "ingress.kubernetes.io/force-hsts"
	annotationKubernetesFrameDeny               = "ingress.kubernetes.io/frame-deny"
	annotationKubernetesCustomFrameOptionsValue = "ingress.kubernetes.io/custom-frame-options-value"
	annotationKubernetesContentTypeNosniff      = "ingress.kubernetes.io/content-type-nosniff"
	annotationKubernetesBrowserXSSFilter        = "ingress.kubernetes.io/browser-xss-filter"
	annotationKubernetesContentSecurityPolicy   = "ingress.kubernetes.io/content-security-policy"
	annotationKubernetesPublicKey               = "ingress.kubernetes.io/public-key"
	annotationKubernetesReferrerPolicy          = "ingress.kubernetes.io/referrer-policy"
	annotationKubernetesIsDevelopment           = "ingress.kubernetes.io/is-development"
)

const traefikDefaultRealm = "traefik"

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider  `mapstructure:",squash" export:"true"`
	Endpoint               string     `description:"Kubernetes server endpoint (required for external cluster client)"`
	Token                  string     `description:"Kubernetes bearer token (not needed for in-cluster client)"`
	CertAuthFilePath       string     `description:"Kubernetes certificate authority file path (not needed for in-cluster client)"`
	DisablePassHostHeaders bool       `description:"Kubernetes disable PassHost Headers" export:"true"`
	EnablePassTLSCert      bool       `description:"Kubernetes enable Pass TLS Client Certs" export:"true"`
	Namespaces             Namespaces `description:"Kubernetes namespaces" export:"true"`
	LabelSelector          string     `description:"Kubernetes api label selector to use" export:"true"`
	lastConfiguration      safe.Safe
}

func (p Provider) newK8sClient() (Client, error) {
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
	// Tell glog (used by client-go) to log into STDERR. Otherwise, we risk
	// certain kinds of API errors getting logged into a directory not
	// available in a `FROM scratch` Docker container, causing glog to abort
	// hard with an exit code > 0.
	flag.Set("logtostderr", "true")

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
				eventsChan, err := k8sClient.WatchAll(p.Namespaces, p.LabelSelector, stopWatch)
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
						log.Debugf("Received Kubernetes event kind %T", event)
						templateObjects, err := p.loadIngresses(k8sClient)
						if err != nil {
							return err
						}
						if reflect.DeepEqual(p.lastConfiguration.Get(), templateObjects) {
							log.Debugf("Skipping Kubernetes event kind %T", event)
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
			log.Errorf("Provider connection error: %s; retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Provider: %s", err)
		}
	})

	return nil
}

func (p *Provider) loadIngresses(k8sClient Client) (*types.Configuration, error) {
	ingresses := k8sClient.GetIngresses()

	templateObjects := types.Configuration{
		Backends:  map[string]*types.Backend{},
		Frontends: map[string]*types.Frontend{},
	}
	for _, i := range ingresses {
		ingressClass := i.Annotations[annotationKubernetesIngressClass]

		if !shouldProcessIngress(ingressClass) {
			continue
		}

		for _, r := range i.Spec.Rules {
			if r.HTTP == nil {
				log.Warn("Error in ingress: HTTP is nil")
				continue
			}

			for _, pa := range r.HTTP.Paths {
				if _, exists := templateObjects.Backends[r.Host+pa.Path]; !exists {
					templateObjects.Backends[r.Host+pa.Path] = &types.Backend{
						Servers: make(map[string]types.Server),
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					}
				}

				passHostHeader := getBoolAnnotation(i, types.LabelFrontendPassHostHeader, !p.DisablePassHostHeaders)
				passTLSCert := getBoolAnnotation(i, types.LabelFrontendPassTLSCert, p.EnablePassTLSCert)

				if realm := i.Annotations[annotationKubernetesAuthRealm]; realm != "" && realm != traefikDefaultRealm {
					log.Errorf("Value for annotation %q on ingress %s/%s invalid: no realm customization supported", annotationKubernetesAuthRealm, i.ObjectMeta.Namespace, i.ObjectMeta.Name)
					delete(templateObjects.Backends, r.Host+pa.Path)
					continue
				}

				entryPoints := getSliceAnnotation(i, types.LabelFrontendEntryPoints)

				whitelistSourceRange := getSliceAnnotation(i, annotationKubernetesWhitelistSourceRange)

				entryPointRedirect, _ := i.Annotations[types.LabelFrontendRedirect]

				if _, exists := templateObjects.Frontends[r.Host+pa.Path]; !exists {
					basicAuthCreds, err := handleBasicAuthConfig(i, k8sClient)
					if err != nil {
						log.Errorf("Failed to retrieve basic auth configuration for ingress %s/%s: %s", i.ObjectMeta.Namespace, i.ObjectMeta.Name, err)
						continue
					}

					priority := getPriority(i)

					headers := types.Headers{
						CustomRequestHeaders:    getMapAnnotation(i, annotationKubernetesCustomRequestHeaders),
						CustomResponseHeaders:   getMapAnnotation(i, annotationKubernetesCustomResponseHeaders),
						AllowedHosts:            getSliceAnnotation(i, annotationKubernetesAllowedHosts),
						HostsProxyHeaders:       getSliceAnnotation(i, annotationKubernetesProxyHeaders),
						SSLRedirect:             getBoolAnnotation(i, annotationKubernetesSSLRedirect, false),
						SSLTemporaryRedirect:    getBoolAnnotation(i, annotationKubernetesSSLTemporaryRedirect, false),
						SSLHost:                 getStringAnnotation(i, annotationKubernetesSSLHost),
						SSLProxyHeaders:         getMapAnnotation(i, annotationKubernetesSSLProxyHeaders),
						STSSeconds:              getSTSSeconds(i),
						STSIncludeSubdomains:    getBoolAnnotation(i, annotationKubernetesHSTSIncludeSubdomains, false),
						STSPreload:              getBoolAnnotation(i, annotationKubernetesHSTSPreload, false),
						ForceSTSHeader:          getBoolAnnotation(i, annotationKubernetesForceHSTSHeader, false),
						FrameDeny:               getBoolAnnotation(i, annotationKubernetesFrameDeny, false),
						CustomFrameOptionsValue: getStringAnnotation(i, annotationKubernetesCustomFrameOptionsValue),
						ContentTypeNosniff:      getBoolAnnotation(i, annotationKubernetesContentTypeNosniff, false),
						BrowserXSSFilter:        getBoolAnnotation(i, annotationKubernetesBrowserXSSFilter, false),
						ContentSecurityPolicy:   getStringAnnotation(i, annotationKubernetesContentSecurityPolicy),
						PublicKey:               getStringAnnotation(i, annotationKubernetesPublicKey),
						ReferrerPolicy:          getStringAnnotation(i, annotationKubernetesReferrerPolicy),
						IsDevelopment:           getBoolAnnotation(i, annotationKubernetesIsDevelopment, false),
					}

					templateObjects.Frontends[r.Host+pa.Path] = &types.Frontend{
						Backend:              r.Host + pa.Path,
						PassHostHeader:       passHostHeader,
						PassTLSCert:          passTLSCert,
						Routes:               make(map[string]types.Route),
						Priority:             priority,
						BasicAuth:            basicAuthCreds,
						WhitelistSourceRange: whitelistSourceRange,
						Redirect:             entryPointRedirect,
						EntryPoints:          entryPoints,
						Headers:              headers,
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

				rule := getRuleForPath(pa, i)
				if rule != "" {
					templateObjects.Frontends[r.Host+pa.Path].Routes[pa.Path] = types.Route{
						Rule: rule,
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

				if expression := service.Annotations[types.LabelTraefikBackendCircuitbreaker]; expression != "" {
					templateObjects.Backends[r.Host+pa.Path].CircuitBreaker = &types.CircuitBreaker{
						Expression: expression,
					}
				}

				if service.Annotations[types.LabelBackendLoadbalancerMethod] == "drr" {
					templateObjects.Backends[r.Host+pa.Path].LoadBalancer.Method = "drr"
				}

				if sticky := service.Annotations[types.LabelBackendLoadbalancerSticky]; len(sticky) > 0 {
					log.Warnf("Deprecated configuration found: %s. Please use %s.", types.LabelBackendLoadbalancerSticky, types.LabelBackendLoadbalancerStickiness)
					templateObjects.Backends[r.Host+pa.Path].LoadBalancer.Sticky = strings.EqualFold(strings.TrimSpace(sticky), "true")
				}

				if service.Annotations[types.LabelBackendLoadbalancerStickiness] == "true" {
					templateObjects.Backends[r.Host+pa.Path].LoadBalancer.Stickiness = &types.Stickiness{}
					if cookieName := service.Annotations[types.LabelBackendLoadbalancerStickinessCookieName]; len(cookieName) > 0 {
						templateObjects.Backends[r.Host+pa.Path].LoadBalancer.Stickiness.CookieName = cookieName
					}
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
								log.Warnf("Endpoints not found for %s/%s", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
								break
							}

							if len(endpoints.Subsets) == 0 {
								log.Warnf("Endpoints not available for %s/%s", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
								break
							}

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

func (p Provider) loadConfig(templateObjects types.Configuration) *types.Configuration {
	var FuncMap = template.FuncMap{}
	configuration, err := p.GetConfiguration("templates/kubernetes.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func getSTSSeconds(i *v1beta1.Ingress) int64 {
	value, err := strconv.ParseInt(i.ObjectMeta.Annotations[annotationKubernetesHSTSMaxAge], 10, 64)
	if err == nil && value > 0 {
		return value
	}
	return 0
}

func getRuleForPath(pa v1beta1.HTTPIngressPath, i *v1beta1.Ingress) string {
	if len(pa.Path) == 0 {
		return ""
	}

	ruleType := i.Annotations[types.LabelFrontendRuleType]
	if ruleType == "" {
		ruleType = ruleTypePathPrefix
	}

	rules := []string{ruleType + ":" + pa.Path}

	if rewriteTarget := i.Annotations[annotationKubernetesRewriteTarget]; rewriteTarget != "" {
		rules = append(rules, ruleTypeReplacePath+":"+rewriteTarget)
	}

	return strings.Join(rules, ";")
}

func getPriority(i *v1beta1.Ingress) int {
	priority := 0

	priorityRaw, ok := i.Annotations[types.LabelFrontendPriority]
	if ok {
		priorityParsed, err := strconv.Atoi(priorityRaw)

		if err == nil {
			priority = priorityParsed
		} else {
			log.Errorf("Error in ingress: failed to parse %q value %q.", types.LabelFrontendPriority, priorityRaw)
		}
	}

	return priority
}

func handleBasicAuthConfig(i *v1beta1.Ingress, k8sClient Client) ([]string, error) {
	authType, exists := i.Annotations[annotationKubernetesAuthType]
	if !exists {
		return nil, nil
	}
	if strings.ToLower(authType) != "basic" {
		return nil, fmt.Errorf("unsupported auth-type on annotation ingress.kubernetes.io/auth-type: %q", authType)
	}
	authSecret := i.Annotations[annotationKubernetesAuthSecret]
	if authSecret == "" {
		return nil, errors.New("auth-secret annotation ingress.kubernetes.io/auth-secret must be set")
	}
	basicAuthCreds, err := loadAuthCredentials(i.Namespace, authSecret, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load auth credentials: %s", err)
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
		return nil, fmt.Errorf("data for secret %q/%q must not be nil", namespace, secretName)
	case len(secret.Data) != 1:
		return nil, fmt.Errorf("found %d elements for secret %q/%q, must be single element exactly", len(secret.Data), namespace, secretName)
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
	if len(creds) == 0 {
		return nil, fmt.Errorf("secret %q/%q does not contain any credentials", namespace, secretName)
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
