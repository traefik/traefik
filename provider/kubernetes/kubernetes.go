package kubernetes

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ provider.Provider = (*Provider)(nil)

const (
	ruleTypePathPrefix         = "PathPrefix"
	ruleTypeReplacePath        = "ReplacePath"
	traefikDefaultRealm        = "traefik"
	traefikDefaultIngressClass = "traefik"
)

// IngressEndpoint holds the endpoint information for the Kubernetes provider
type IngressEndpoint struct {
	IP               string `description:"IP used for Kubernetes Ingress endpoints"`
	Hostname         string `description:"Hostname used for Kubernetes Ingress endpoints"`
	PublishedService string `description:"Published Kubernetes Service to copy status from"`
}

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider  `mapstructure:",squash" export:"true"`
	Endpoint               string           `description:"Kubernetes server endpoint (required for external cluster client)"`
	Token                  string           `description:"Kubernetes bearer token (not needed for in-cluster client)"`
	CertAuthFilePath       string           `description:"Kubernetes certificate authority file path (not needed for in-cluster client)"`
	DisablePassHostHeaders bool             `description:"Kubernetes disable PassHost Headers" export:"true"`
	EnablePassTLSCert      bool             `description:"Kubernetes enable Pass TLS Client Certs" export:"true"`
	Namespaces             Namespaces       `description:"Kubernetes namespaces" export:"true"`
	LabelSelector          string           `description:"Kubernetes Ingress label selector to use" export:"true"`
	IngressClass           string           `description:"Value of kubernetes.io/ingress.class annotation to watch for" export:"true"`
	IngressEndpoint        *IngressEndpoint `description:"Kubernetes Ingress Endpoint"`
	lastConfiguration      safe.Safe
}

func (p *Provider) newK8sClient(ingressLabelSelector string) (Client, error) {
	ingLabelSel, err := labels.Parse(ingressLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid ingress label selector: %q", ingressLabelSelector)
	}
	log.Infof("ingress label selector is: %q", ingLabelSel)

	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %v", p.Endpoint)
	}

	var cl *clientImpl
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		log.Infof("Creating in-cluster Provider client%s", withEndpoint)
		cl, err = newInClusterClient(p.Endpoint)
	} else {
		log.Infof("Creating cluster-external Provider client%s", withEndpoint)
		cl, err = newExternalClusterClient(p.Endpoint, p.Token, p.CertAuthFilePath)
	}

	if err == nil {
		cl.ingressLabelSelector = ingLabelSel
	}

	return cl, err
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

	log.Debugf("Using Ingress label selector: %q", p.LabelSelector)
	k8sClient, err := p.newK8sClient(p.LabelSelector)
	if err != nil {
		return err
	}
	p.Constraints = append(p.Constraints, constraints...)

	pool.Go(func(stop chan bool) {
		operation := func() error {
			for {
				stopWatch := make(chan struct{}, 1)
				defer close(stopWatch)
				eventsChan, err := k8sClient.WatchAll(p.Namespaces, stopWatch)
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

		backendPercentageWeightMap, err := getPercentageWeightMap(i)
		if err != nil {
			log.Errorf("Invalid yaml format for backend weight annotation of ingress %s/%s: %v", i.Namespace, i.Name, err)
			continue
		}

		for _, r := range i.Spec.Rules {
			if r.HTTP == nil {
				log.Warn("Error in ingress: HTTP is nil")
				continue
			}

			var leftFractionPercentage float64 = 1
			leftFractionInstanceCount := 0
			for _, pa := range r.HTTP.Paths {
				percentageWeight, found := backendPercentageWeightMap[pa.Backend.ServiceName]
				if found {
					leftFractionPercentage -= percentageWeight.Float64()
					continue
				}
				endpoints, exist, err := k8sClient.GetEndpoints(i.Namespace, pa.Backend.ServiceName)
				if err == nil && exist {
					for _, subset := range endpoints.Subsets {
						leftFractionInstanceCount += len(subset.Addresses)
					}
				}
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

					templateObjects.Frontends[baseName] = &types.Frontend{
						Backend:        baseName,
						PassHostHeader: passHostHeader,
						PassTLSCert:    passTLSCert,
						Routes:         make(map[string]types.Route),
						Priority:       priority,
						BasicAuth:      basicAuthCreds,
						WhiteList:      getWhiteList(i),
						Redirect:       getFrontendRedirect(i),
						EntryPoints:    entryPoints,
						Headers:        getHeader(i),
						Errors:         getErrorPages(i),
						RateLimit:      getRateLimit(i),
					}
				}

				if len(r.Host) > 0 {
					if _, exists := templateObjects.Frontends[baseName].Routes[r.Host]; !exists {
						templateObjects.Frontends[baseName].Routes[r.Host] = types.Route{
							Rule: getRuleForHost(r.Host),
						}
					}
				}

				rule, err := getRuleForPath(pa, i)
				if err != nil {
					log.Errorf("Failed to get rule for ingress %s/%s: %s", i.Namespace, i.Name, err)
					delete(templateObjects.Frontends, baseName)
					continue
				}
				if rule != "" {
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
				serverWeight := label.DefaultWeight

				for _, port := range service.Spec.Ports {
					if equalPorts(port, pa.Backend.ServicePort) {
						if port.Port == 443 || strings.HasPrefix(port.Name, "https") {
							protocol = "https"
						}

						if service.Spec.Type == "ExternalName" {
							url := protocol + "://" + service.Spec.ExternalName
							if port.Port != 443 && port.Port != 80 {
								url = fmt.Sprintf("%s:%d", url, port.Port)
							}

							templateObjects.Backends[baseName].Servers[url] = types.Server{
								URL:    url,
								Weight: serverWeight,
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

							if percentageWeight, found := backendPercentageWeightMap[pa.Backend.ServiceName]; found {
								instanceCount := 0
								for _, subset := range endpoints.Subsets {
									instanceCount = instanceCount + len(subset.Addresses)
								}
								serverWeight = int(percentageWeight.RawValue()) / instanceCount
							} else if leftFractionPercentage < 1 {
								serverWeight = int(PercentageValueFromFloat64(leftFractionPercentage).RawValue() / int64(leftFractionInstanceCount))
							}

							for _, subset := range endpoints.Subsets {
								endpointPort := endpointPortNumber(port, subset.Ports)
								if endpointPort == 0 {
									// endpoint port does not match service.
									continue
								}
								for _, address := range subset.Addresses {
									url := protocol + "://" + net.JoinHostPort(address.IP, strconv.FormatInt(int64(endpointPort), 10))
									name := url
									if address.TargetRef != nil && address.TargetRef.Name != "" {
										name = address.TargetRef.Name
									}
									templateObjects.Backends[baseName].Servers[name] = types.Server{
										URL:    url,
										Weight: serverWeight,
									}
								}
							}
						}
						break
					}
				}
			}
		}

		err = p.updateIngressStatus(i, k8sClient)
		if err != nil {
			log.Errorf("Cannot update Ingress %s/%s due to error: %v", i.Namespace, i.Name, err)
		}
	}
	return &templateObjects, nil
}

func (p *Provider) updateIngressStatus(i *extensionsv1beta1.Ingress, k8sClient Client) error {
	// Only process if an IngressEndpoint has been configured
	if p.IngressEndpoint == nil {
		return nil
	}

	if len(p.IngressEndpoint.PublishedService) == 0 {
		if len(p.IngressEndpoint.IP) == 0 && len(p.IngressEndpoint.Hostname) == 0 {
			return errors.New("publishedService or ip or hostname must be defined")
		}

		return k8sClient.UpdateIngressStatus(i.Namespace, i.Name, p.IngressEndpoint.IP, p.IngressEndpoint.Hostname)
	}

	serviceInfo := strings.Split(p.IngressEndpoint.PublishedService, "/")
	if len(serviceInfo) != 2 {
		return fmt.Errorf("invalid publishedService format (expected 'namespace/service' format): %s", p.IngressEndpoint.PublishedService)
	}
	serviceNamespace, serviceName := serviceInfo[0], serviceInfo[1]

	service, exists, err := k8sClient.GetService(serviceNamespace, serviceName)
	if err != nil {
		return fmt.Errorf("cannot get service %s, received error: %s", p.IngressEndpoint.PublishedService, err)
	}

	if exists && service.Status.LoadBalancer.Ingress == nil {
		// service exists, but has no Load Balancer status
		log.Debugf("Skipping updating Ingress %s/%s due to service %s having no status set", i.Namespace, i.Name, p.IngressEndpoint.PublishedService)
		return nil
	}

	if !exists {
		return fmt.Errorf("missing service: %s", p.IngressEndpoint.PublishedService)
	}

	return k8sClient.UpdateIngressStatus(i.Namespace, i.Name, service.Status.LoadBalancer.Ingress[0].IP, service.Status.LoadBalancer.Ingress[0].Hostname)
}

func (p *Provider) loadConfig(templateObjects types.Configuration) *types.Configuration {
	var FuncMap = template.FuncMap{}
	configuration, err := p.GetConfiguration("templates/kubernetes.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration
}

func getRuleForPath(pa extensionsv1beta1.HTTPIngressPath, i *extensionsv1beta1.Ingress) (string, error) {
	if len(pa.Path) == 0 {
		return "", nil
	}

	ruleType := getStringValue(i.Annotations, annotationKubernetesRuleType, ruleTypePathPrefix)
	rules := []string{ruleType + ":" + pa.Path}

	var pathReplaceAnnotation string
	if ruleType == ruleTypeReplacePath {
		pathReplaceAnnotation = annotationKubernetesRuleType
	}

	if rewriteTarget := getStringValue(i.Annotations, annotationKubernetesRewriteTarget, ""); rewriteTarget != "" {
		if pathReplaceAnnotation != "" {
			return "", fmt.Errorf("rewrite-target must not be used together with annotation %q", pathReplaceAnnotation)
		}
		rules = append(rules, ruleTypeReplacePath+":"+rewriteTarget)
		pathReplaceAnnotation = annotationKubernetesRewriteTarget
	}

	if rootPath := getStringValue(i.Annotations, annotationKubernetesAppRoot, ""); rootPath != "" && pa.Path == "/" {
		if pathReplaceAnnotation != "" {
			return "", fmt.Errorf("app-root must not be used together with annotation %q", pathReplaceAnnotation)
		}
		rules = append(rules, ruleTypeReplacePath+":"+rootPath)
	}
	return strings.Join(rules, ";"), nil
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

// endpointPortNumber returns the port to be used for this endpoint. It is zero
// if the endpoint does not match the given service port.
func endpointPortNumber(servicePort corev1.ServicePort, endpointPorts []corev1.EndpointPort) int32 {
	// Is this reasonable to assume?
	if len(endpointPorts) == 0 {
		return servicePort.Port
	}

	for _, endpointPort := range endpointPorts {
		// For matching endpoints, the port names must correspond, either by
		// being empty or non-empty. Multi-port services mandate non-empty
		// names and allow us to filter for the right addresses.
		if servicePort.Name == endpointPort.Name {
			return endpointPort.Port
		}
	}
	return 0
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

func getWhiteList(i *extensionsv1beta1.Ingress) *types.WhiteList {
	ranges := getSliceStringValue(i.Annotations, annotationKubernetesWhiteListSourceRange)
	if len(ranges) <= 0 {
		return nil
	}

	return &types.WhiteList{
		SourceRange:      ranges,
		UseXForwardedFor: getBoolValue(i.Annotations, annotationKubernetesWhiteListUseXForwardedFor, false),
	}
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
		SSLForceHost:            getBoolValue(i.Annotations, annotationKubernetesSSLForceHost, false),
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
		CustomBrowserXSSValue:   getStringValue(i.Annotations, annotationKubernetesCustomBrowserXSSValue, ""),
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

func getPercentageWeightMap(i *extensionsv1beta1.Ingress) (map[string]*PercentageValue, error) {
	backendPercentageAnnotationValue := getStringValue(i.Annotations, annotationKubernetesBackendPercentageWeights, "")
	backendPercentageAnnotationMap := make(map[string]string)
	if err := yaml.Unmarshal([]byte(backendPercentageAnnotationValue), &backendPercentageAnnotationMap); err != nil {
		return nil, err
	}
	backendPercentageWeightMap := make(map[string]*PercentageValue)
	for serviceName, percentageStr := range backendPercentageAnnotationMap {
		percentageValue, err := PercentageValueFromString(percentageStr)
		if err != nil {
			log.Errorf("Invalid percentage value %q in ingress %s/%s: %s", percentageStr, i.Name, err)
			backendPercentageWeightMap = make(map[string]*PercentageValue)
			break
		}
		backendPercentageWeightMap[serviceName] = percentageValue
	}
	return backendPercentageWeightMap, nil
}
