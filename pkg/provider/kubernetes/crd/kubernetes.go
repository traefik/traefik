package crd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/job"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/tls"
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
		id := makeID(middleware.Namespace, middleware.Name)
		ctxMid := log.With(ctx, log.Str(log.MiddlewareName, id))

		basicAuth, err := createBasicAuthMiddleware(client, middleware.Namespace, middleware.Spec.BasicAuth)
		if err != nil {
			log.FromContext(ctxMid).Errorf("Error while reading basic auth middleware: %v", err)
			continue
		}

		digestAuth, err := createDigestAuthMiddleware(client, middleware.Namespace, middleware.Spec.DigestAuth)
		if err != nil {
			log.FromContext(ctxMid).Errorf("Error while reading digest auth middleware: %v", err)
			continue
		}

		forwardAuth, err := createForwardAuthMiddleware(client, middleware.Namespace, middleware.Spec.ForwardAuth)
		if err != nil {
			log.FromContext(ctxMid).Errorf("Error while reading digest auth middleware: %v", err)
			continue
		}

		conf.HTTP.Middlewares[id] = &dynamic.Middleware{
			AddPrefix:         middleware.Spec.AddPrefix,
			StripPrefix:       middleware.Spec.StripPrefix,
			StripPrefixRegex:  middleware.Spec.StripPrefixRegex,
			ReplacePath:       middleware.Spec.ReplacePath,
			ReplacePathRegex:  middleware.Spec.ReplacePathRegex,
			Chain:             createChainMiddleware(ctxMid, middleware.Namespace, middleware.Spec.Chain),
			IPWhiteList:       middleware.Spec.IPWhiteList,
			Headers:           middleware.Spec.Headers,
			Errors:            middleware.Spec.Errors,
			RateLimit:         middleware.Spec.RateLimit,
			RedirectRegex:     middleware.Spec.RedirectRegex,
			RedirectScheme:    middleware.Spec.RedirectScheme,
			BasicAuth:         basicAuth,
			DigestAuth:        digestAuth,
			ForwardAuth:       forwardAuth,
			InFlightReq:       middleware.Spec.InFlightReq,
			Buffering:         middleware.Spec.Buffering,
			CircuitBreaker:    middleware.Spec.CircuitBreaker,
			Compress:          middleware.Spec.Compress,
			PassTLSClientCert: middleware.Spec.PassTLSClientCert,
			Retry:             middleware.Spec.Retry,
		}

	}

	return conf
}

func createForwardAuthMiddleware(k8sClient Client, namespace string, auth *v1alpha1.ForwardAuth) (*dynamic.ForwardAuth, error) {
	if auth == nil {
		return nil, nil
	}
	if len(auth.Address) == 0 {
		return nil, fmt.Errorf("forward authentication requires an address")
	}

	forwardAuth := &dynamic.ForwardAuth{
		Address:             auth.Address,
		TrustForwardHeader:  auth.TrustForwardHeader,
		AuthResponseHeaders: auth.AuthResponseHeaders,
	}

	if auth.TLS != nil {
		forwardAuth.TLS = &dynamic.ClientTLS{
			CAOptional:         auth.TLS.CAOptional,
			InsecureSkipVerify: auth.TLS.InsecureSkipVerify,
		}

		if len(auth.TLS.CASecret) > 0 {
			caSecret, err := loadCASecret(namespace, auth.TLS.CASecret, k8sClient)
			if err != nil {
				return nil, err
			}
			forwardAuth.TLS.CA = caSecret
		}

		if len(auth.TLS.CertSecret) > 0 {
			authSecretCert, authSecretKey, err := loadAuthTLSSecret(namespace, auth.TLS.CertSecret, k8sClient)
			if err != nil {
				return nil, fmt.Errorf("failed to load auth secret: %s", err)
			}
			forwardAuth.TLS.Cert = authSecretCert
			forwardAuth.TLS.Key = authSecretKey
		}
	}

	return forwardAuth, nil
}

func loadCASecret(namespace, secretName string, k8sClient Client) (string, error) {
	secret, ok, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to fetch secret %q/%q: %s", namespace, secretName, err)
	}
	if !ok {
		return "", fmt.Errorf("secret %q/%q not found", namespace, secretName)
	}
	if secret == nil {
		return "", fmt.Errorf("data for secret %q/%q must not be nil", namespace, secretName)
	}
	if len(secret.Data) != 1 {
		return "", fmt.Errorf("found %d elements for secret %q/%q, must be single element exactly", len(secret.Data), namespace, secretName)
	}

	for _, v := range secret.Data {
		return string(v), nil
	}
	return "", nil
}

func loadAuthTLSSecret(namespace, secretName string, k8sClient Client) (string, string, error) {
	secret, exists, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch secret %q/%q: %s", namespace, secretName, err)
	}
	if !exists {
		return "", "", fmt.Errorf("secret %q/%q does not exist", namespace, secretName)
	}
	if secret == nil {
		return "", "", fmt.Errorf("data for secret %q/%q must not be nil", namespace, secretName)
	}
	if len(secret.Data) != 2 {
		return "", "", fmt.Errorf("found %d elements for secret %q/%q, must be two elements exactly", len(secret.Data), namespace, secretName)
	}

	return getCertificateBlocks(secret, namespace, secretName)
}

func createBasicAuthMiddleware(client Client, namespace string, basicAuth *v1alpha1.BasicAuth) (*dynamic.BasicAuth, error) {
	if basicAuth == nil {
		return nil, nil
	}

	credentials, err := getAuthCredentials(client, basicAuth.Secret, namespace)
	if err != nil {
		return nil, err
	}

	return &dynamic.BasicAuth{
		Users:        credentials,
		Realm:        basicAuth.Realm,
		RemoveHeader: basicAuth.RemoveHeader,
		HeaderField:  basicAuth.HeaderField,
	}, nil
}

func createDigestAuthMiddleware(client Client, namespace string, digestAuth *v1alpha1.DigestAuth) (*dynamic.DigestAuth, error) {
	if digestAuth == nil {
		return nil, nil
	}

	credentials, err := getAuthCredentials(client, digestAuth.Secret, namespace)
	if err != nil {
		return nil, err
	}

	return &dynamic.DigestAuth{
		Users:        credentials,
		Realm:        digestAuth.Realm,
		RemoveHeader: digestAuth.RemoveHeader,
		HeaderField:  digestAuth.HeaderField,
	}, nil
}

func getAuthCredentials(k8sClient Client, authSecret, namespace string) ([]string, error) {
	if authSecret == "" {
		return nil, fmt.Errorf("auth secret must be set")
	}

	auth, err := loadAuthCredentials(namespace, authSecret, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load auth credentials: %s", err)
	}

	return auth, nil
}

func loadAuthCredentials(namespace, secretName string, k8sClient Client) ([]string, error) {
	secret, ok, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret %q/%q: %s", namespace, secretName, err)
	}
	if !ok {
		return nil, fmt.Errorf("secret %q/%q not found", namespace, secretName)
	}
	if secret == nil {
		return nil, fmt.Errorf("data for secret %q/%q must not be nil", namespace, secretName)
	}
	if len(secret.Data) != 1 {
		return nil, fmt.Errorf("found %d elements for secret %q/%q, must be single element exactly", len(secret.Data), namespace, secretName)
	}

	var firstSecret []byte
	for _, v := range secret.Data {
		firstSecret = v
		break
	}

	var credentials []string
	scanner := bufio.NewScanner(bytes.NewReader(firstSecret))
	for scanner.Scan() {
		if cred := scanner.Text(); len(cred) > 0 {
			credentials = append(credentials, cred)
		}
	}

	if len(credentials) == 0 {
		return nil, fmt.Errorf("secret %q/%q does not contain any credentials", namespace, secretName)
	}

	return credentials, nil
}

func createChainMiddleware(ctx context.Context, namespace string, chain *v1alpha1.Chain) *dynamic.Chain {
	if chain == nil {
		return nil
	}

	var mds []string
	for _, mi := range chain.Middlewares {
		if strings.Contains(mi.Name, "@") {
			if len(mi.Namespace) > 0 {
				log.FromContext(ctx).
					Warnf("namespace %q is ignored in cross-provider context", mi.Namespace)
			}
			mds = append(mds, mi.Name)
			continue
		}

		ns := mi.Namespace
		if len(ns) == 0 {
			ns = namespace
		}
		mds = append(mds, makeID(ns, mi.Name))
	}
	return &dynamic.Chain{Middlewares: mds}
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

		for _, secretName := range tlsOption.Spec.ClientAuth.SecretNames {
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
			ClientAuth: tls.ClientAuth{
				CAFiles:        clientCAs,
				ClientAuthType: tlsOption.Spec.ClientAuth.ClientAuthType,
			},
			SniStrict: tlsOption.Spec.SniStrict,
		}
	}
	return tlsOptions
}

func checkStringQuoteValidity(value string) error {
	_, err := strconv.Unquote(`"` + value + `"`)
	return err
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
