package crd

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/job"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider"
	"github.com/containous/traefik/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"github.com/containous/traefik/pkg/provider/kubernetes/k8s"
	"github.com/containous/traefik/pkg/safe"
	"github.com/containous/traefik/pkg/tls"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	annotationKubernetesIngressClass = "kubernetes.io/ingress.class"
	traefikDefaultIngressClass       = "traefik"
)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider  `mapstructure:",squash" export:"true"`
	Endpoint               string         `description:"Kubernetes server endpoint (required for external cluster client)"`
	Token                  string         `description:"Kubernetes bearer token (not needed for in-cluster client)"`
	CertAuthFilePath       string         `description:"Kubernetes certificate authority file path (not needed for in-cluster client)"`
	DisablePassHostHeaders bool           `description:"Kubernetes disable PassHost Headers" export:"true"`
	Namespaces             k8s.Namespaces `description:"Kubernetes namespaces" export:"true"`
	LabelSelector          string         `description:"Kubernetes label selector to use" export:"true"`
	IngressClass           string         `description:"Value of kubernetes.io/ingress.class annotation to watch for" export:"true"`
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
	return p.BaseProvider.Init()
}

// Provide allows the k8s provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- config.Message, pool *safe.Pool) error {
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

	if p.LabelSelector != "" {
		logger.Debugf("Using label selector: %q", p.LabelSelector)
	}
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
					conf := p.loadConfigurationFromIngresses(ctxLog, k8sClient)

					if reflect.DeepEqual(p.lastConfiguration.Get(), conf) {
						logger.Debugf("Skipping Kubernetes event kind %T", event)
					} else {
						p.lastConfiguration.Set(conf)
						configurationChan <- config.Message{
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

func loadServers(client Client, namespace string, svc v1alpha1.Service) ([]config.Server, error) {
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

	var portSpec corev1.ServicePort
	var match bool
	// TODO: support name ports? do we actually care?
	for _, p := range service.Spec.Ports {
		if svc.Port == p.Port {
			portSpec = p
			match = true
			break
		}
	}

	if !match {
		return nil, errors.New("service port not found")
	}

	var servers []config.Server
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		servers = append(servers, config.Server{
			URL:    fmt.Sprintf("http://%s:%d", service.Spec.ExternalName, portSpec.Port),
			Weight: 1,
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
			if port == 443 || strings.HasPrefix(portSpec.Name, "https") {
				protocol = "https"
			}

			for _, addr := range subset.Addresses {
				servers = append(servers, config.Server{
					URL:    fmt.Sprintf("%s://%s:%d", protocol, addr.IP, port),
					Weight: 1,
				})
			}
		}
	}

	return servers, nil
}

func (p *Provider) loadConfigurationFromIngresses(ctx context.Context, client Client) *config.Configuration {

	conf := &config.Configuration{
		HTTP: &config.HTTPConfiguration{
			Routers:     map[string]*config.Router{},
			Middlewares: map[string]*config.Middleware{},
			Services:    map[string]*config.Service{},
		},
		TCP: &config.TCPConfiguration{},
	}
	tlsConfigs := make(map[string]*tls.Configuration)

	for _, ingressRoute := range client.GetIngressRoutes() {
		logger := log.FromContext(log.With(ctx, log.Str("ingress", ingressRoute.Name), log.Str("namespace", ingressRoute.Namespace)))

		// TODO keep the name ingressClass?
		if !shouldProcessIngress(p.IngressClass, ingressRoute.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		err := getTLS(ctx, ingressRoute, client, tlsConfigs)
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

			var allServers []config.Server
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

			// TODO: support middlewares from other providers.
			// Mechanism: in the spec, prefix the name with the provider name,
			// with dot as the separator. In which case. we ignore the
			// namespace.

			var mds []string
			for _, mi := range route.Middlewares {
				ns := mi.Namespace
				if len(ns) == 0 {
					ns = ingressRoute.Namespace
				}
				mds = append(mds, makeID(ns, mi.Name))
			}

			h := sha256.New()
			_, err = h.Write([]byte(route.Match))
			if err != nil {
				logger.Error(err)
				continue
			}
			key := fmt.Sprintf("%s-%.10x", ingressName, h.Sum(nil))

			serviceName := makeID(ingressRoute.Namespace, key)

			conf.HTTP.Routers[serviceName] = &config.Router{
				Middlewares: mds,
				Priority:    route.Priority,
				EntryPoints: ingressRoute.Spec.EntryPoints,
				Rule:        route.Match,
				Service:     serviceName,
			}
			if ingressRoute.Spec.TLS != nil {
				conf.HTTP.Routers[serviceName].TLS = &config.RouterTLSConfig{}
			}
			conf.HTTP.Services[serviceName] = &config.Service{
				LoadBalancer: &config.LoadBalancerService{
					Servers: allServers,
					// TODO: support other strategies.
					Method:         "wrr",
					PassHostHeader: true,
				},
			}
		}
	}

	conf.TLS = getTLSConfig(tlsConfigs)

	for _, middleware := range client.GetMiddlewares() {
		conf.HTTP.Middlewares[makeID(middleware.Namespace, middleware.Name)] = &middleware.Spec
	}

	return conf
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

func getTLS(ctx context.Context, ingressRoute *v1alpha1.IngressRoute, k8sClient Client, tlsConfigs map[string]*tls.Configuration) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.FromContext(ctx).Debugf("Skipping TLS sub-section: No secret name provided")
		return nil
	}

	configKey := ingressRoute.Namespace + "/" + ingressRoute.Spec.TLS.SecretName
	if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
		secret, exists, err := k8sClient.GetSecret(ingressRoute.Namespace, ingressRoute.Spec.TLS.SecretName)
		if err != nil {
			return fmt.Errorf("failed to fetch secret %s/%s: %v", ingressRoute.Namespace, ingressRoute.Spec.TLS.SecretName, err)
		}
		if !exists {
			return fmt.Errorf("secret %s/%s does not exist", ingressRoute.Namespace, ingressRoute.Spec.TLS.SecretName)
		}

		cert, key, err := getCertificateBlocks(secret, ingressRoute.Namespace, ingressRoute.Spec.TLS.SecretName)
		if err != nil {
			return err
		}

		tlsConfigs[configKey] = &tls.Configuration{
			Certificate: &tls.Certificate{
				CertFile: tls.FileOrContent(cert),
				KeyFile:  tls.FileOrContent(key),
			},
		}
	}

	return nil
}

func getTLSConfig(tlsConfigs map[string]*tls.Configuration) []*tls.Configuration {
	var secretNames []string
	for secretName := range tlsConfigs {
		secretNames = append(secretNames, secretName)
	}
	sort.Strings(secretNames)

	var configs []*tls.Configuration
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
