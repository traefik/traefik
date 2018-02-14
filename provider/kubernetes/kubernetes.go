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
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ provider.Provider = (*Provider)(nil)

const (
	ruleTypePathPrefix         = "PathPrefix"
	ruleTypeReplacePath        = "ReplacePath"
	traefikDefaultRealm        = "traefik"
	traefikDefaultIngressClass = "traefik"
)

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
	IngressClass           string     `description:"Value of kubernetes.io/ingress.class annotation to watch for" export:"true"`
	lastConfiguration      safe.Safe
}

func (p *Provider) newK8sClient() (Client, error) {
	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %v", p.Endpoint)
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		log.Infof("Creating in-cluster Provider client%s", withEndpoint)
		return NewInClusterClient(p.Endpoint)
	}

	log.Infof("Creating cluster-external Provider client%s", withEndpoint)
	return NewExternalClusterClient(p.Endpoint, p.Token, p.CertAuthFilePath)
}

// Provide allows the k8s provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	// Tell glog (used by client-go) to log into STDERR. Otherwise, we risk
	// certain kinds of API errors getting logged into a directory not
	// available in a `FROM scratch` Docker container, causing glog to abort
	// hard with an exit code > 0.
	err := flag.Set("logtostderr", "true")
	if err != nil {
		return err
	}

	// We require that IngressClasses start with `traefik` to reduce chances of
	// conflict with other Ingress Providers
	if len(p.IngressClass) > 0 && !strings.HasPrefix(p.IngressClass, traefikDefaultIngressClass) {
		return fmt.Errorf("value for IngressClass has to be empty or start with the prefix %q, instead found %q", traefikDefaultIngressClass, p.IngressClass)
	}

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
		annotationIngressClass := getAnnotationName(i.Annotations, annotationKubernetesIngressClass)
		ingressClass := i.Annotations[annotationIngressClass]

		if !p.shouldProcessIngress(ingressClass) {
			continue
		}

		tlsSection, err := getTLS(i, k8sClient)
		if err != nil {
			log.Errorf("Error configuring TLS for ingress %s/%s: %v", i.Namespace, i.Name, err)
			continue
		}
		templateObjects.TLS = append(templateObjects.TLS, tlsSection...)

		for _, r := range i.Spec.Rules {
			if r.HTTP == nil {
				log.Warn("Error in ingress: HTTP is nil")
				continue
			}

			for _, pa := range r.HTTP.Paths {
				baseName := r.Host + pa.Path
				if _, exists := templateObjects.Backends[baseName]; !exists {
					templateObjects.Backends[baseName] = &types.Backend{
						Servers: make(map[string]types.Server),
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					}
				}

				annotationAuthRealm := getAnnotationName(i.Annotations, annotationKubernetesAuthRealm)
				if realm := i.Annotations[annotationAuthRealm]; realm != "" && realm != traefikDefaultRealm {
					log.Errorf("Value for annotation %q on ingress %s/%s invalid: no realm customization supported", annotationAuthRealm, i.Namespace, i.Name)
					delete(templateObjects.Backends, baseName)
					continue
				}

				if _, exists := templateObjects.Frontends[baseName]; !exists {
					basicAuthCreds, err := handleBasicAuthConfig(i, k8sClient)
					if err != nil {
						log.Errorf("Failed to retrieve basic auth configuration for ingress %s/%s: %s", i.Namespace, i.Name, err)
						continue
					}

					passHostHeader := getBoolValue(i.Annotations, annotationKubernetesPreserveHost, !p.DisablePassHostHeaders)
					passTLSCert := getBoolValue(i.Annotations, annotationKubernetesPassTLSCert, p.EnablePassTLSCert)
					priority := getIntValue(i.Annotations, annotationKubernetesPriority, 0)
					entryPoints := getSliceStringValue(i.Annotations, annotationKubernetesFrontendEntryPoints)
					whitelistSourceRange := getSliceStringValue(i.Annotations, annotationKubernetesWhitelistSourceRange)

					templateObjects.Frontends[baseName] = &types.Frontend{
						Backend:              baseName,
						PassHostHeader:       passHostHeader,
						PassTLSCert:          passTLSCert,
						Routes:               make(map[string]types.Route),
						Priority:             priority,
						BasicAuth:            basicAuthCreds,
						WhitelistSourceRange: whitelistSourceRange,
						Redirect:             getFrontendRedirect(i),
						EntryPoints:          entryPoints,
						Headers:              getHeader(i),
						Errors:               getErrorPages(i),
						RateLimit:            getRateLimit(i),
					}
				}

				if len(r.Host) > 0 {
					if _, exists := templateObjects.Frontends[baseName].Routes[r.Host]; !exists {
						templateObjects.Frontends[baseName].Routes[r.Host] = types.Route{
							Rule: getRuleForHost(r.Host),
						}
					}
				}

				if rule := getRuleForPath(pa, i); rule != "" {
					templateObjects.Frontends[baseName].Routes[pa.Path] = types.Route{
						Rule: rule,
					}
				}

				service, exists, err := k8sClient.GetService(i.Namespace, pa.Backend.ServiceName)
				if err != nil {
					log.Errorf("Error while retrieving service information from k8s API %s/%s: %v", i.Namespace, pa.Backend.ServiceName, err)
					return nil, err
				}

				if !exists {
					log.Errorf("Service not found for %s/%s", i.Namespace, pa.Backend.ServiceName)
					delete(templateObjects.Frontends, baseName)
					continue
				}

				templateObjects.Backends[baseName].CircuitBreaker = getCircuitBreaker(service)
				templateObjects.Backends[baseName].LoadBalancer = getLoadBalancer(service)
				templateObjects.Backends[baseName].MaxConn = getMaxConn(service)
				templateObjects.Backends[baseName].Buffering = getBuffering(service)

				protocol := label.DefaultProtocol
				for _, port := range service.Spec.Ports {
					if equalPorts(port, pa.Backend.ServicePort) {
						if port.Port == 443 {
							protocol = "https"
						}

						if service.Spec.Type == "ExternalName" {
							url := protocol + "://" + service.Spec.ExternalName
							name := url

							templateObjects.Backends[baseName].Servers[name] = types.Server{
								URL:    url,
								Weight: 1,
							}
						} else {
							endpoints, exists, err := k8sClient.GetEndpoints(service.Namespace, service.Name)
							if err != nil {
								log.Errorf("Error retrieving endpoints %s/%s: %v", service.Namespace, service.Name, err)
								return nil, err
							}

							if !exists {
								log.Warnf("Endpoints not found for %s/%s", service.Namespace, service.Name)
								break
							}

							if len(endpoints.Subsets) == 0 {
								log.Warnf("Endpoints not available for %s/%s", service.Namespace, service.Name)
								break
							}

							for _, subset := range endpoints.Subsets {
								for _, address := range subset.Addresses {
									url := protocol + "://" + address.IP + ":" + strconv.Itoa(endpointPortNumber(port, subset.Ports))
									name := url
									if address.TargetRef != nil && address.TargetRef.Name != "" {
										name = address.TargetRef.Name
									}
									templateObjects.Backends[baseName].Servers[name] = types.Server{
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

func (p *Provider) loadConfig(templateObjects types.Configuration) *types.Configuration {
	var FuncMap = template.FuncMap{}
	configuration, err := p.GetConfiguration("templates/kubernetes.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func getRuleForPath(pa extensionsv1beta1.HTTPIngressPath, i *extensionsv1beta1.Ingress) string {
	if len(pa.Path) == 0 {
		return ""
	}

	ruleType := getStringValue(i.Annotations, annotationKubernetesRuleType, ruleTypePathPrefix)
	rules := []string{ruleType + ":" + pa.Path}

	if rewriteTarget := getStringValue(i.Annotations, annotationKubernetesRewriteTarget, ""); rewriteTarget != "" {
		rules = append(rules, ruleTypeReplacePath+":"+rewriteTarget)
	}

	return strings.Join(rules, ";")
}

func getRuleForHost(host string) string {
	if strings.Contains(host, "*") {
		return "HostRegexp:" + strings.Replace(host, "*", "{subdomain:[A-Za-z0-9-_]+}", 1)
	}
	return "Host:" + host
}

func handleBasicAuthConfig(i *extensionsv1beta1.Ingress, k8sClient Client) ([]string, error) {
	annotationAuthType := getAnnotationName(i.Annotations, annotationKubernetesAuthType)
	authType, exists := i.Annotations[annotationAuthType]
	if !exists {
		return nil, nil
	}

	if strings.ToLower(authType) != "basic" {
		return nil, fmt.Errorf("unsupported auth-type on annotation ingress.kubernetes.io/auth-type: %q", authType)
	}

	authSecret := getStringValue(i.Annotations, annotationKubernetesAuthSecret, "")
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

func getTLS(ingress *extensionsv1beta1.Ingress, k8sClient Client) ([]*tls.Configuration, error) {
	var tlsConfigs []*tls.Configuration

	for _, t := range ingress.Spec.TLS {
		tlsSecret, exists, err := k8sClient.GetSecret(ingress.Namespace, t.SecretName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch secret %s/%s: %v", ingress.Namespace, t.SecretName, err)
		}
		if !exists {
			return nil, fmt.Errorf("secret %s/%s does not exist", ingress.Namespace, t.SecretName)
		}

		tlsCrtData, tlsCrtExists := tlsSecret.Data["tls.crt"]
		tlsKeyData, tlsKeyExists := tlsSecret.Data["tls.key"]

		var missingEntries []string
		if !tlsCrtExists {
			missingEntries = append(missingEntries, "tls.crt")
		}
		if !tlsKeyExists {
			missingEntries = append(missingEntries, "tls.key")
		}
		if len(missingEntries) > 0 {
			return nil, fmt.Errorf("secret %s/%s is missing the following TLS data entries: %s",
				ingress.Namespace, t.SecretName, strings.Join(missingEntries, ", "))
		}

		entryPoints := getSliceStringValue(ingress.Annotations, annotationKubernetesFrontendEntryPoints)

		tlsConfig := &tls.Configuration{
			EntryPoints: entryPoints,
			Certificate: &tls.Certificate{
				CertFile: tls.FileOrContent(tlsCrtData),
				KeyFile:  tls.FileOrContent(tlsKeyData),
			},
		}

		tlsConfigs = append(tlsConfigs, tlsConfig)
	}

	return tlsConfigs, nil
}

func endpointPortNumber(servicePort corev1.ServicePort, endpointPorts []corev1.EndpointPort) int {
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

func equalPorts(servicePort corev1.ServicePort, ingressPort intstr.IntOrString) bool {
	if int(servicePort.Port) == ingressPort.IntValue() {
		return true
	}
	if servicePort.Name != "" && servicePort.Name == ingressPort.String() {
		return true
	}
	return false
}

func (p *Provider) shouldProcessIngress(annotationIngressClass string) bool {
	if len(p.IngressClass) == 0 {
		return len(annotationIngressClass) == 0 || annotationIngressClass == traefikDefaultIngressClass
	}
	return annotationIngressClass == p.IngressClass
}

func getFrontendRedirect(i *extensionsv1beta1.Ingress) *types.Redirect {
	permanent := getBoolValue(i.Annotations, annotationKubernetesRedirectPermanent, false)

	redirectEntryPoint := getStringValue(i.Annotations, annotationKubernetesRedirectEntryPoint, "")
	if len(redirectEntryPoint) > 0 {
		return &types.Redirect{
			EntryPoint: redirectEntryPoint,
			Permanent:  permanent,
		}
	}

	redirectRegex := getStringValue(i.Annotations, annotationKubernetesRedirectRegex, "")
	redirectReplacement := getStringValue(i.Annotations, annotationKubernetesRedirectReplacement, "")
	if len(redirectRegex) > 0 && len(redirectReplacement) > 0 {
		return &types.Redirect{
			Regex:       redirectRegex,
			Replacement: redirectReplacement,
			Permanent:   permanent,
		}
	}

	return nil
}

func getBuffering(service *corev1.Service) *types.Buffering {
	var buffering *types.Buffering

	bufferingRaw := getStringValue(service.Annotations, annotationKubernetesBuffering, "")

	if len(bufferingRaw) > 0 {
		buffering = &types.Buffering{}
		err := yaml.Unmarshal([]byte(bufferingRaw), buffering)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	return buffering
}

func getLoadBalancer(service *corev1.Service) *types.LoadBalancer {
	loadBalancer := &types.LoadBalancer{
		Method: "wrr",
	}

	if getStringValue(service.Annotations, annotationKubernetesLoadBalancerMethod, "") == "drr" {
		loadBalancer.Method = "drr"
	}

	if sticky := service.Annotations[label.TraefikBackendLoadBalancerSticky]; len(sticky) > 0 {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, annotationKubernetesAffinity)
		loadBalancer.Sticky = strings.EqualFold(strings.TrimSpace(sticky), "true")
	}

	if stickiness := getStickiness(service); stickiness != nil {
		loadBalancer.Stickiness = stickiness
	}

	return loadBalancer
}

func getStickiness(service *corev1.Service) *types.Stickiness {
	if getBoolValue(service.Annotations, annotationKubernetesAffinity, false) {
		stickiness := &types.Stickiness{}
		if cookieName := getStringValue(service.Annotations, annotationKubernetesSessionCookieName, ""); len(cookieName) > 0 {
			stickiness.CookieName = cookieName
		}
		return stickiness
	}
	return nil
}

func getHeader(i *extensionsv1beta1.Ingress) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    getMapValue(i.Annotations, annotationKubernetesCustomRequestHeaders),
		CustomResponseHeaders:   getMapValue(i.Annotations, annotationKubernetesCustomResponseHeaders),
		AllowedHosts:            getSliceStringValue(i.Annotations, annotationKubernetesAllowedHosts),
		HostsProxyHeaders:       getSliceStringValue(i.Annotations, annotationKubernetesProxyHeaders),
		SSLRedirect:             getBoolValue(i.Annotations, annotationKubernetesSSLRedirect, false),
		SSLTemporaryRedirect:    getBoolValue(i.Annotations, annotationKubernetesSSLTemporaryRedirect, false),
		SSLHost:                 getStringValue(i.Annotations, annotationKubernetesSSLHost, ""),
		SSLProxyHeaders:         getMapValue(i.Annotations, annotationKubernetesSSLProxyHeaders),
		STSSeconds:              getInt64Value(i.Annotations, annotationKubernetesHSTSMaxAge, 0),
		STSIncludeSubdomains:    getBoolValue(i.Annotations, annotationKubernetesHSTSIncludeSubdomains, false),
		STSPreload:              getBoolValue(i.Annotations, annotationKubernetesHSTSPreload, false),
		ForceSTSHeader:          getBoolValue(i.Annotations, annotationKubernetesForceHSTSHeader, false),
		FrameDeny:               getBoolValue(i.Annotations, annotationKubernetesFrameDeny, false),
		CustomFrameOptionsValue: getStringValue(i.Annotations, annotationKubernetesCustomFrameOptionsValue, ""),
		ContentTypeNosniff:      getBoolValue(i.Annotations, annotationKubernetesContentTypeNosniff, false),
		BrowserXSSFilter:        getBoolValue(i.Annotations, annotationKubernetesBrowserXSSFilter, false),
		ContentSecurityPolicy:   getStringValue(i.Annotations, annotationKubernetesContentSecurityPolicy, ""),
		PublicKey:               getStringValue(i.Annotations, annotationKubernetesPublicKey, ""),
		ReferrerPolicy:          getStringValue(i.Annotations, annotationKubernetesReferrerPolicy, ""),
		IsDevelopment:           getBoolValue(i.Annotations, annotationKubernetesIsDevelopment, false),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

func getMaxConn(service *corev1.Service) *types.MaxConn {
	amount := getInt64Value(service.Annotations, annotationKubernetesMaxConnAmount, -1)
	extractorFunc := getStringValue(service.Annotations, annotationKubernetesMaxConnExtractorFunc, "")
	if amount >= 0 && len(extractorFunc) > 0 {
		return &types.MaxConn{
			ExtractorFunc: extractorFunc,
			Amount:        amount,
		}
	}
	return nil
}

func getCircuitBreaker(service *corev1.Service) *types.CircuitBreaker {
	if expression := getStringValue(service.Annotations, annotationKubernetesCircuitBreakerExpression, ""); expression != "" {
		return &types.CircuitBreaker{
			Expression: expression,
		}
	}
	return nil
}

func getErrorPages(i *extensionsv1beta1.Ingress) map[string]*types.ErrorPage {
	var errorPages map[string]*types.ErrorPage

	pagesRaw := getStringValue(i.Annotations, annotationKubernetesErrorPages, "")
	if len(pagesRaw) > 0 {
		errorPages = make(map[string]*types.ErrorPage)
		err := yaml.Unmarshal([]byte(pagesRaw), errorPages)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	return errorPages
}

func getRateLimit(i *extensionsv1beta1.Ingress) *types.RateLimit {
	var rateLimit *types.RateLimit

	rateRaw := getStringValue(i.Annotations, annotationKubernetesRateLimit, "")
	if len(rateRaw) > 0 {
		rateLimit = &types.RateLimit{}
		err := yaml.Unmarshal([]byte(rateRaw), rateLimit)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	return rateLimit
}
