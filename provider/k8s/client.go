package k8s

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/containous/traefik/safe"
	"github.com/parnurzeal/gorequest"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// APIEndpoint defines the base path for kubernetes API resources.
	APIEndpoint        = "/api/v1"
	extentionsEndpoint = "/apis/extensions/v1beta1"
	defaultIngress     = "/ingresses"
)

// Client is a client for the Kubernetes master.
type Client interface {
	GetIngresses(predicate func(Ingress) bool) ([]Ingress, error)
	GetServices(predicate func(Service) bool) ([]Service, error)
	WatchAll(stopCh <-chan bool) (chan interface{}, chan error, error)
}

type clientImpl struct {
	endpointURL string
	tls         *tls.Config
	token       string
	caCert      []byte
}

// NewClient returns a new Kubernetes client.
// The provided host is an url (scheme://hostname[:port]) of a
// Kubernetes master without any path.
// The provided client is an authorized http.Client used to perform requests to the Kubernetes API master.
func NewClient(baseURL string, caCert []byte, token string) (Client, error) {
	validURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %v", baseURL, err)
	}
	return &clientImpl{
		endpointURL: strings.TrimSuffix(validURL.String(), "/"),
		token:       token,
		caCert:      caCert,
	}, nil
}

// GetIngresses returns all services in the cluster
func (c *clientImpl) GetIngresses(predicate func(Ingress) bool) ([]Ingress, error) {
	getURL := c.endpointURL + extentionsEndpoint + defaultIngress

	body, err := c.do(c.request(getURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", getURL, err)
	}

	var ingressList IngressList
	if err := json.Unmarshal(body, &ingressList); err != nil {
		return nil, fmt.Errorf("failed to decode list of ingress resources: %v", err)
	}
	ingresses := ingressList.Items[:0]
	for _, ingress := range ingressList.Items {
		if predicate(ingress) {
			ingresses = append(ingresses, ingress)
		}
	}
	return ingresses, nil
}

// WatchIngresses returns all ingresses in the cluster
func (c *clientImpl) WatchIngresses(stopCh <-chan bool) (chan interface{}, chan error, error) {
	getURL := c.endpointURL + extentionsEndpoint + defaultIngress
	return c.watch(getURL, stopCh)
}

// GetServices returns all services in the cluster
func (c *clientImpl) GetServices(predicate func(Service) bool) ([]Service, error) {
	getURL := c.endpointURL + APIEndpoint + "/services"

	body, err := c.do(c.request(getURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", getURL, err)
	}

	var serviceList ServiceList
	if err := json.Unmarshal(body, &serviceList); err != nil {
		return nil, fmt.Errorf("failed to decode list of services resources: %v", err)
	}
	services := serviceList.Items[:0]
	for _, service := range serviceList.Items {
		if predicate(service) {
			services = append(services, service)
		}
	}
	return services, nil
}

// WatchServices returns all services in the cluster
func (c *clientImpl) WatchServices(stopCh <-chan bool) (chan interface{}, chan error, error) {
	getURL := c.endpointURL + APIEndpoint + "/services"
	return c.watch(getURL, stopCh)
}

// WatchEvents returns events in the cluster
func (c *clientImpl) WatchEvents(stopCh <-chan bool) (chan interface{}, chan error, error) {
	getURL := c.endpointURL + APIEndpoint + "/events"
	return c.watch(getURL, stopCh)
}

// WatchPods returns pods in the cluster
func (c *clientImpl) WatchPods(stopCh <-chan bool) (chan interface{}, chan error, error) {
	getURL := c.endpointURL + APIEndpoint + "/pods"
	return c.watch(getURL, stopCh)
}

// WatchReplicationControllers returns ReplicationControllers in the cluster
func (c *clientImpl) WatchReplicationControllers(stopCh <-chan bool) (chan interface{}, chan error, error) {
	getURL := c.endpointURL + APIEndpoint + "/replicationcontrollers"
	return c.watch(getURL, stopCh)
}

// WatchAll returns events in the cluster
func (c *clientImpl) WatchAll(stopCh <-chan bool) (chan interface{}, chan error, error) {
	watchCh := make(chan interface{})
	errCh := make(chan error)

	stopIngresses := make(chan bool)
	chanIngresses, chanIngressesErr, err := c.WatchIngresses(stopIngresses)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create watch %v", err)
	}
	stopServices := make(chan bool)
	chanServices, chanServicesErr, err := c.WatchServices(stopServices)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create watch %v", err)
	}
	stopPods := make(chan bool)
	chanPods, chanPodsErr, err := c.WatchPods(stopPods)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create watch %v", err)
	}
	stopReplicationControllers := make(chan bool)
	chanReplicationControllers, chanReplicationControllersErr, err := c.WatchReplicationControllers(stopReplicationControllers)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create watch %v", err)
	}
	go func() {
		defer close(watchCh)
		defer close(errCh)
		defer close(stopIngresses)
		defer close(stopServices)
		defer close(stopPods)
		defer close(stopReplicationControllers)

		for {
			select {
			case <-stopCh:
				stopIngresses <- true
				stopServices <- true
				stopPods <- true
				stopReplicationControllers <- true
				break
			case err := <-chanIngressesErr:
				errCh <- err
			case err := <-chanServicesErr:
				errCh <- err
			case err := <-chanPodsErr:
				errCh <- err
			case err := <-chanReplicationControllersErr:
				errCh <- err
			case event := <-chanIngresses:
				watchCh <- event
			case event := <-chanServices:
				watchCh <- event
			case event := <-chanPods:
				watchCh <- event
			case event := <-chanReplicationControllers:
				watchCh <- event
			}
		}
	}()

	return watchCh, errCh, nil
}

func (c *clientImpl) do(request *gorequest.SuperAgent) ([]byte, error) {
	res, body, errs := request.EndBytes()
	if errs != nil {
		return nil, fmt.Errorf("failed to create request: GET %q : %v", request.Url, errs)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d GET %q: %q", res.StatusCode, request.Url, string(body))
	}
	return body, nil
}

func (c *clientImpl) request(url string) *gorequest.SuperAgent {
	// Make request to Kubernetes API
	request := gorequest.New().Get(url)
	if len(c.token) > 0 {
		request.Header["Authorization"] = "Bearer " + c.token
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(c.caCert)
		c.tls = &tls.Config{RootCAs: pool}
	}
	return request.TLSClientConfig(c.tls)
}

// GenericObject generic object
type GenericObject struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata,omitempty"`
}

func (c *clientImpl) watch(url string, stopCh <-chan bool) (chan interface{}, chan error, error) {
	watchCh := make(chan interface{})
	errCh := make(chan error)

	// get version
	body, err := c.do(c.request(url))
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create request: GET %q : %v", url, err)
	}

	var generic GenericObject
	if err := json.Unmarshal(body, &generic); err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create request: GET %q : %v", url, err)
	}
	resourceVersion := generic.ResourceVersion

	url = url + "?watch&resourceVersion=" + resourceVersion
	// Make request to Kubernetes API
	request := c.request(url)
	request.Transport.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := net.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		// No timeout for long-polling request
		conn.SetDeadline(time.Now())
		return conn, nil
	}
	req, err := request.TLSClientConfig(c.tls).MakeRequest()
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create request: GET %q : %v", url, err)
	}
	res, err := request.Client.Do(req)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to make request: GET %q: %v", url, err)
	}

	shouldStop := safe.New(false)

	go func() {
		select {
		case <-stopCh:
			shouldStop.Set(true)
			res.Body.Close()
			return
		}
	}()

	go func() {
		defer close(watchCh)
		defer close(errCh)
		for {
			var eventList interface{}
			if err := json.NewDecoder(res.Body).Decode(&eventList); err != nil {
				if !shouldStop.Get().(bool) {
					errCh <- fmt.Errorf("failed to decode watch event: %v", err)
				}
				return
			}
			watchCh <- eventList
		}
	}()
	return watchCh, errCh, nil
}
