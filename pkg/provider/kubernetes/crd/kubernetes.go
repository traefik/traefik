package crd

import (
	"context"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/job"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/tls"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	annotationKubernetesIngressClass = "kubernetes.io/ingress.class"
	traefikDefaultIngressClass       = "traefik"
)

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint               string   `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token                  string   `description:"Kubernetes bearer token (not needed for in-cluster client)." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty"`
	CertAuthFilePath       string   `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	DisablePassHostHeaders bool     `description:"Kubernetes disable PassHost Headers." json:"disablePassHostHeaders,omitempty" toml:"disablePassHostHeaders,omitempty" yaml:"disablePassHostHeaders,omitempty" export:"true"`
	Namespaces             []string `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector          string   `description:"Kubernetes label selector to use." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	IngressClass           string   `description:"Value of kubernetes.io/ingress.class annotation to watch for." json:"ingressClass,omitempty" toml:"ingressClass,omitempty" yaml:"ingressClass,omitempty" export:"true"`
	lastConfiguration      safe.Safe
}

func (p *Provider) newK8sClient(ctx context.Context, labelSelector string) (*clientWrapper, error) {
	labelSel, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector: %q", labelSelector)
	}
	log.FromContext(ctx).Infof("label selector is: %q", labelSel)

	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %v", p.Endpoint)
	}

	var client *clientWrapper
	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		log.FromContext(ctx).Infof("Creating in-cluster Provider client%s", withEndpoint)
		client, err = newInClusterClient(p.Endpoint)
	case os.Getenv("KUBECONFIG") != "":
		log.FromContext(ctx).Infof("Creating cluster-external Provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		client, err = newExternalClusterClientFromFile(os.Getenv("KUBECONFIG"))
	default:
		log.FromContext(ctx).Infof("Creating cluster-external Provider client%s", withEndpoint)
		client, err = newExternalClusterClient(p.Endpoint, p.Token, p.CertAuthFilePath)
	}

	if err == nil {
		client.labelSelector = labelSel
	}

	return client, err
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the k8s provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	ctxLog := log.With(context.Background(), log.Str(log.ProviderName, "kubernetescrd"))
	logger := log.FromContext(ctxLog)
	// Tell glog (used by client-go) to log into STDERR. Otherwise, we risk
	// certain kinds of API errors getting logged into a directory not
	// available in a `FROM scratch` Docker container, causing glog to abort
	// hard with an exit code > 0.
	err := flag.Set("logtostderr", "true")
	if err != nil {
		return err
	}

	logger.Debugf("Using label selector: %q", p.LabelSelector)
	k8sClient, err := p.newK8sClient(ctxLog, p.LabelSelector)
	if err != nil {
		return err
	}

	pool.Go(func(stop chan bool) {
		operation := func() error {
			stopWatch := make(chan struct{}, 1)
			defer close(stopWatch)
			eventsChan, err := k8sClient.WatchAll(p.Namespaces, stopWatch)
			if err != nil {
				logger.Errorf("Error watching kubernetes events: %v", err)
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
					conf := p.loadConfigurationFromCRD(ctxLog, k8sClient)

					if reflect.DeepEqual(p.lastConfiguration.Get(), conf) {
						logger.Debugf("Skipping Kubernetes event kind %T", event)
					} else {
						p.lastConfiguration.Set(conf)
						configurationChan <- dynamic.Message{
							ProviderName:  "kubernetescrd",
							Configuration: conf,
						}
					}
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Errorf("Provider connection error: %s; retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			logger.Errorf("Cannot connect to Provider: %s", err)
		}
	})

	return nil
}

func checkStringQuoteValidity(value string) error {
	_, err := strconv.Unquote(`"` + value + `"`)
	return err
}

func loadTCPServers(client Client, namespace string, svc v1alpha1.ServiceTCP) ([]dynamic.TCPServer, error) {
	service, exists, err := client.GetService(namespace, svc.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("service not found")
	}

	var portSpec *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		if svc.Port == p.Port {
			portSpec = &p
			break
		}
	}

	if portSpec == nil {
		return nil, errors.New("service port not found")
	}

	var servers []dynamic.TCPServer
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		servers = append(servers, dynamic.TCPServer{
			Address: fmt.Sprintf("%s:%d", service.Spec.ExternalName, portSpec.Port),
		})
	} else {
		endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, svc.Name)
		if endpointsErr != nil {
			return nil, endpointsErr
		}

		if !endpointsExists {
			return nil, errors.New("endpoints not found")
		}

		if len(endpoints.Subsets) == 0 {
			return nil, errors.New("subset not found")
		}

		var port int32
		for _, subset := range endpoints.Subsets {
			for _, p := range subset.Ports {
				if portSpec.Name == p.Name {
					port = p.Port
					break
				}
			}

			if port == 0 {
				return nil, errors.New("cannot define a port")
			}

			for _, addr := range subset.Addresses {
				servers = append(servers, dynamic.TCPServer{
					Address: fmt.Sprintf("%s:%d", addr.IP, port),
				})
			}
		}
	}

	return servers, nil
}

func loadServers(client Client, namespace string, svc v1alpha1.Service) ([]dynamic.Server, error) {
	strategy := svc.Strategy
	if strategy == "" {
		strategy = "RoundRobin"
	}
	if strategy != "RoundRobin" {
		return nil, fmt.Errorf("load balancing strategy %v is not supported", strategy)
	}

	service, exists, err := client.GetService(namespace, svc.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("service not found")
	}

	var portSpec *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		if svc.Port == p.Port {
			portSpec = &p
			break
		}
	}

	if portSpec == nil {
		return nil, errors.New("service port not found")
	}

	var servers []dynamic.Server
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		servers = append(servers, dynamic.Server{
			URL: fmt.Sprintf("http://%s:%d", service.Spec.ExternalName, portSpec.Port),
		})
	} else {
		endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, svc.Name)
		if endpointsErr != nil {
			return nil, endpointsErr
		}

		if !endpointsExists {
			return nil, errors.New("endpoints not found")
		}

		if len(endpoints.Subsets) == 0 {
			return nil, errors.New("subset not found")
		}

		var port int32
		for _, subset := range endpoints.Subsets {
			for _, p := range subset.Ports {
				if portSpec.Name == p.Name {
					port = p.Port
					break
				}
			}

			if port == 0 {
				return nil, errors.New("cannot define a port")
			}

			protocol := "http"
			switch svc.Scheme {
			case "http", "https", "h2c":
				protocol = svc.Scheme
			case "":
				if port == 443 || strings.HasPrefix(portSpec.Name, "https") {
					protocol = "https"
				}
			default:
				return nil, fmt.Errorf("invalid scheme %q specified", svc.Scheme)
			}

			for _, addr := range subset.Addresses {
				servers = append(servers, dynamic.Server{
					URL: fmt.Sprintf("%s://%s:%d", protocol, addr.IP, port),
				})
			}
		}
	}

	return servers, nil
}

func buildTLSOptions(ctx context.Context, client Client) map[string]tls.Options {
	tlsOptionsCRD := client.GetTLSOptions()
	var tlsOptions map[string]tls.Options

	if len(tlsOptionsCRD) == 0 {
		return tlsOptions
	}
	tlsOptions = make(map[string]tls.Options)

	for _, tlsOption := range tlsOptionsCRD {
		logger := log.FromContext(log.With(ctx, log.Str("tlsOption", tlsOption.Name), log.Str("namespace", tlsOption.Namespace)))
		var clientCAs []tls.FileOrContent

		for _, secretName := range tlsOption.Spec.ClientCA.SecretNames {
			secret, exists, err := client.GetSecret(tlsOption.Namespace, secretName)
			if err != nil {
				logger.Errorf("Failed to fetch secret %s/%s: %v", tlsOption.Namespace, secretName, err)
				continue
			}

			if !exists {
				logger.Warnf("Secret %s/%s does not exist", tlsOption.Namespace, secretName)
				continue
			}

			cert, err := getCABlocks(secret, tlsOption.Namespace, secretName)
			if err != nil {
				logger.Errorf("Failed to extract CA from secret %s/%s: %v", tlsOption.Namespace, secretName, err)
				continue
			}

			clientCAs = append(clientCAs, tls.FileOrContent(cert))
		}

		tlsOptions[makeID(tlsOption.Namespace, tlsOption.Name)] = tls.Options{
			MinVersion:   tlsOption.Spec.MinVersion,
			CipherSuites: tlsOption.Spec.CipherSuites,
			ClientCA: tls.ClientCA{
				Files:    clientCAs,
				Optional: tlsOption.Spec.ClientCA.Optional,
			},
			SniStrict: tlsOption.Spec.SniStrict,
		}
	}
	return tlsOptions
}

func (p *Provider) loadIngressRouteConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Middlewares: map[string]*dynamic.Middleware{},
		Services:    map[string]*dynamic.Service{},
	}

	for _, ingressRoute := range client.GetIngressRoutes() {
		logger := log.FromContext(log.With(ctx, log.Str("ingress", ingressRoute.Name), log.Str("namespace", ingressRoute.Namespace)))

		// TODO keep the name ingressClass?
		if !shouldProcessIngress(p.IngressClass, ingressRoute.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		err := getTLSHTTP(ctx, ingressRoute, client, tlsConfigs)
		if err != nil {
			logger.Errorf("Error configuring TLS: %v", err)
		}

		ingressName := ingressRoute.Name
		if len(ingressName) == 0 {
			ingressName = ingressRoute.GenerateName
		}

		for _, route := range ingressRoute.Spec.Routes {
			if route.Kind != "Rule" {
				logger.Errorf("Unsupported match kind: %s. Only \"Rule\" is supported for now.", route.Kind)
				continue
			}

			if len(route.Match) == 0 {
				logger.Errorf("Empty match rule")
				continue
			}

			if err := checkStringQuoteValidity(route.Match); err != nil {
				logger.Errorf("Invalid syntax for match rule: %s", route.Match)
				continue
			}

			var allServers []dynamic.Server
			for _, service := range route.Services {
				servers, err := loadServers(client, ingressRoute.Namespace, service)
				if err != nil {
					logger.
						WithField("serviceName", service.Name).
						WithField("servicePort", service.Port).
						Errorf("Cannot create service: %v", err)
					continue
				}

				allServers = append(allServers, servers...)
			}

			var mds []string
			for _, mi := range route.Middlewares {
				if strings.Contains(mi.Name, "@") {
					if len(mi.Namespace) > 0 {
						logger.
							WithField(log.MiddlewareName, mi.Name).
							Warnf("namespace %q is ignored in cross-provider context", mi.Namespace)
					}
					mds = append(mds, mi.Name)
					continue
				}

				ns := mi.Namespace
				if len(ns) == 0 {
					ns = ingressRoute.Namespace
				}
				mds = append(mds, makeID(ns, mi.Name))
			}

			key, err := makeServiceKey(route.Match, ingressName)
			if err != nil {
				logger.Error(err)
				continue
			}

			serviceName := makeID(ingressRoute.Namespace, key)

			conf.Routers[serviceName] = &dynamic.Router{
				Middlewares: mds,
				Priority:    route.Priority,
				EntryPoints: ingressRoute.Spec.EntryPoints,
				Rule:        route.Match,
				Service:     serviceName,
			}

			if ingressRoute.Spec.TLS != nil {
				tlsConf := &dynamic.RouterTLSConfig{}
				if ingressRoute.Spec.TLS.Options != nil && len(ingressRoute.Spec.TLS.Options.Name) > 0 {
					tlsOptionsName := ingressRoute.Spec.TLS.Options.Name
					// Is a Kubernetes CRD reference, (i.e. not a cross-provider reference)
					ns := ingressRoute.Spec.TLS.Options.Namespace
					if !strings.Contains(tlsOptionsName, "@") {
						if len(ns) == 0 {
							ns = ingressRoute.Namespace
						}
						tlsOptionsName = makeID(ns, tlsOptionsName)
					} else if len(ns) > 0 {
						logger.
							WithField("TLSoptions", ingressRoute.Spec.TLS.Options.Name).
							Warnf("namespace %q is ignored in cross-provider context", ns)
					}

					tlsConf.Options = tlsOptionsName
				}
				conf.Routers[serviceName].TLS = tlsConf
			}

			conf.Services[serviceName] = &dynamic.Service{
				LoadBalancer: &dynamic.LoadBalancerService{
					Servers: allServers,
					// TODO: support other strategies.
					PassHostHeader: true,
				},
			}
		}
	}

	return conf
}

func (p *Provider) loadIngressRouteTCPConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.TCPConfiguration {
	conf := &dynamic.TCPConfiguration{
		Routers:  map[string]*dynamic.TCPRouter{},
		Services: map[string]*dynamic.TCPService{},
	}

	for _, ingressRouteTCP := range client.GetIngressRouteTCPs() {
		logger := log.FromContext(log.With(ctx, log.Str("ingress", ingressRouteTCP.Name), log.Str("namespace", ingressRouteTCP.Namespace)))

		if !shouldProcessIngress(p.IngressClass, ingressRouteTCP.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		if ingressRouteTCP.Spec.TLS != nil && !ingressRouteTCP.Spec.TLS.Passthrough {
			err := getTLSTCP(ctx, ingressRouteTCP, client, tlsConfigs)
			if err != nil {
				logger.Errorf("Error configuring TLS: %v", err)
			}
		}

		ingressName := ingressRouteTCP.Name
		if len(ingressName) == 0 {
			ingressName = ingressRouteTCP.GenerateName
		}

		for _, route := range ingressRouteTCP.Spec.Routes {
			if len(route.Match) == 0 {
				logger.Errorf("Empty match rule")
				continue
			}

			if err := checkStringQuoteValidity(route.Match); err != nil {
				logger.Errorf("Invalid syntax for match rule: %s", route.Match)
				continue
			}

			var allServers []dynamic.TCPServer
			for _, service := range route.Services {
				servers, err := loadTCPServers(client, ingressRouteTCP.Namespace, service)
				if err != nil {
					logger.
						WithField("serviceName", service.Name).
						WithField("servicePort", service.Port).
						Errorf("Cannot create service: %v", err)
					continue
				}

				allServers = append(allServers, servers...)
			}

			key, e := makeServiceKey(route.Match, ingressName)
			if e != nil {
				logger.Error(e)
				continue
			}

			serviceName := makeID(ingressRouteTCP.Namespace, key)
			conf.Routers[serviceName] = &dynamic.TCPRouter{
				EntryPoints: ingressRouteTCP.Spec.EntryPoints,
				Rule:        route.Match,
				Service:     serviceName,
			}

			if ingressRouteTCP.Spec.TLS != nil {
				conf.Routers[serviceName].TLS = &dynamic.RouterTCPTLSConfig{
					Passthrough: ingressRouteTCP.Spec.TLS.Passthrough,
				}

				if ingressRouteTCP.Spec.TLS.Options != nil && len(ingressRouteTCP.Spec.TLS.Options.Name) > 0 {
					tlsOptionsName := ingressRouteTCP.Spec.TLS.Options.Name
					// Is a Kubernetes CRD reference (i.e. not a cross-provider reference)
					ns := ingressRouteTCP.Spec.TLS.Options.Namespace
					if !strings.Contains(tlsOptionsName, "@") {
						if len(ns) == 0 {
							ns = ingressRouteTCP.Namespace
						}
						tlsOptionsName = makeID(ns, tlsOptionsName)
					} else if len(ns) > 0 {
						logger.
							WithField("TLSoptions", ingressRouteTCP.Spec.TLS.Options.Name).
							Warnf("namespace %q is ignored in cross-provider context", ns)
					}

					conf.Routers[serviceName].TLS.Options = tlsOptionsName

				}
			}

			conf.Services[serviceName] = &dynamic.TCPService{
				LoadBalancer: &dynamic.TCPLoadBalancerService{
					Servers: allServers,
				},
			}
		}
	}

	return conf
}

func (p *Provider) loadConfigurationFromCRD(ctx context.Context, client Client) *dynamic.Configuration {
	tlsConfigs := make(map[string]*tls.CertAndStores)
	conf := &dynamic.Configuration{
		HTTP: p.loadIngressRouteConfiguration(ctx, client, tlsConfigs),
		TCP:  p.loadIngressRouteTCPConfiguration(ctx, client, tlsConfigs),
		TLS: &dynamic.TLSConfiguration{
			Certificates: getTLSConfig(tlsConfigs),
			Options:      buildTLSOptions(ctx, client),
		},
	}

	for _, middleware := range client.GetMiddlewares() {
		conf.HTTP.Middlewares[makeID(middleware.Namespace, middleware.Name)] = &middleware.Spec
	}

	return conf
}

func makeServiceKey(rule, ingressName string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(rule)); err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s-%.10x", ingressName, h.Sum(nil))

	return key, nil
}

func makeID(namespace, name string) string {
	if namespace == "" {
		return name
	}

	return namespace + "/" + name
}

func shouldProcessIngress(ingressClass string, ingressClassAnnotation string) bool {
	return ingressClass == ingressClassAnnotation ||
		(len(ingressClass) == 0 && ingressClassAnnotation == traefikDefaultIngressClass)
}

func getTLSHTTP(ctx context.Context, ingressRoute *v1alpha1.IngressRoute, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.FromContext(ctx).Debugf("Skipping TLS sub-section: No secret name provided")
		return nil
	}

	configKey := ingressRoute.Namespace + "/" + ingressRoute.Spec.TLS.SecretName
	if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
		tlsConf, err := getTLS(k8sClient, ingressRoute.Spec.TLS.SecretName, ingressRoute.Namespace)
		if err != nil {
			return err
		}

		tlsConfigs[configKey] = tlsConf
	}

	return nil
}

func getTLSTCP(ctx context.Context, ingressRoute *v1alpha1.IngressRouteTCP, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.FromContext(ctx).Debugf("Skipping TLS sub-section for TCP: No secret name provided")
		return nil
	}

	configKey := ingressRoute.Namespace + "/" + ingressRoute.Spec.TLS.SecretName
	if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
		tlsConf, err := getTLS(k8sClient, ingressRoute.Spec.TLS.SecretName, ingressRoute.Namespace)
		if err != nil {
			return err
		}

		tlsConfigs[configKey] = tlsConf
	}

	return nil
}

func getTLS(k8sClient Client, secretName, namespace string) (*tls.CertAndStores, error) {
	secret, exists, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret %s/%s: %v", namespace, secretName, err)
	}
	if !exists {
		return nil, fmt.Errorf("secret %s/%s does not exist", namespace, secretName)
	}

	cert, key, err := getCertificateBlocks(secret, namespace, secretName)
	if err != nil {
		return nil, err
	}

	return &tls.CertAndStores{
		Certificate: tls.Certificate{
			CertFile: tls.FileOrContent(cert),
			KeyFile:  tls.FileOrContent(key),
		},
	}, nil
}

func getTLSConfig(tlsConfigs map[string]*tls.CertAndStores) []*tls.CertAndStores {
	var secretNames []string
	for secretName := range tlsConfigs {
		secretNames = append(secretNames, secretName)
	}
	sort.Strings(secretNames)

	var configs []*tls.CertAndStores
	for _, secretName := range secretNames {
		configs = append(configs, tlsConfigs[secretName])
	}

	return configs
}

func getCertificateBlocks(secret *corev1.Secret, namespace, secretName string) (string, string, error) {
	var missingEntries []string

	tlsCrtData, tlsCrtExists := secret.Data["tls.crt"]
	if !tlsCrtExists {
		missingEntries = append(missingEntries, "tls.crt")
	}

	tlsKeyData, tlsKeyExists := secret.Data["tls.key"]
	if !tlsKeyExists {
		missingEntries = append(missingEntries, "tls.key")
	}

	if len(missingEntries) > 0 {
		return "", "", fmt.Errorf("secret %s/%s is missing the following TLS data entries: %s",
			namespace, secretName, strings.Join(missingEntries, ", "))
	}

	cert := string(tlsCrtData)
	if cert == "" {
		missingEntries = append(missingEntries, "tls.crt")
	}

	key := string(tlsKeyData)
	if key == "" {
		missingEntries = append(missingEntries, "tls.key")
	}

	if len(missingEntries) > 0 {
		return "", "", fmt.Errorf("secret %s/%s contains the following empty TLS data entries: %s",
			namespace, secretName, strings.Join(missingEntries, ", "))
	}

	return cert, key, nil
}

func getCABlocks(secret *corev1.Secret, namespace, secretName string) (string, error) {
	tlsCrtData, tlsCrtExists := secret.Data["tls.ca"]
	if !tlsCrtExists {
		return "", fmt.Errorf("the tls.ca entry is missing from secret %s/%s",
			namespace, secretName)
	}

	cert := string(tlsCrtData)
	if cert == "" {
		return "", fmt.Errorf("the tls.ca entry in secret %s/%s is empty",
			namespace, secretName)
	}

	return cert, nil
}
