package k8s

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"net/url"
	"strings"
)

const (
	// APIEndpoint defines the base path for kubernetes API resources.
	APIEndpoint        = "/api/v1"
	extentionsEndpoint = "/apis/extensions/v1beta1"
	defaultIngress     = "/ingresses"
	namespaces         = "/namespaces/"
)

// Client is a client for the Kubernetes master.
type Client interface {
	GetIngresses(predicate func(Ingress) bool) ([]Ingress, error)
	GetService(name, namespace string) (Service, error)
	GetEndpoints(name, namespace string) (Endpoints, error)
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

// GetIngresses returns all ingresses in the cluster
func (c *clientImpl) GetIngresses(predicate func(Ingress) bool) ([]Ingress, error) {
	getURL := c.endpointURL + extentionsEndpoint + defaultIngress

	body, err := c.do(c.request(getURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create ingresses request: GET %q : %v", getURL, err)
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

// GetService returns the named service from the named namespace
func (c *clientImpl) GetService(name, namespace string) (Service, error) {
	getURL := c.endpointURL + APIEndpoint + namespaces + namespace + "/services/" + name

	body, err := c.do(c.request(getURL))
	if err != nil {
		return Service{}, fmt.Errorf("failed to create services request: GET %q : %v", getURL, err)
	}

	var service Service
	if err := json.Unmarshal(body, &service); err != nil {
		return Service{}, fmt.Errorf("failed to decode service resource: %v", err)
	}
	return service, nil
}

// WatchServices returns all services in the cluster
func (c *clientImpl) WatchServices(stopCh <-chan bool) (chan interface{}, chan error, error) {
	getURL := c.endpointURL + APIEndpoint + "/services"
	return c.watch(getURL, stopCh)
}

// GetEndpoints returns the named Endpoints
// Endpoints have the same name as the coresponding service
func (c *clientImpl) GetEndpoints(name, namespace string) (Endpoints, error) {
	getURL := c.endpointURL + APIEndpoint + namespaces + namespace + "/endpoints/" + name

	body, err := c.do(c.request(getURL))
	if err != nil {
		return Endpoints{}, fmt.Errorf("failed to create endpoints request: GET %q : %v", getURL, err)
	}

	var endpoints Endpoints
	if err := json.Unmarshal(body, &endpoints); err != nil {
		return Endpoints{}, fmt.Errorf("failed to decode endpoints resources: %v", err)
	}
	return endpoints, nil
}

// WatchEndpoints returns endpoints in the cluster
func (c *clientImpl) WatchEndpoints(stopCh <-chan bool) (chan interface{}, chan error, error) {
	getURL := c.endpointURL + APIEndpoint + "/endpoints"
	return c.watch(getURL, stopCh)
}

// WatchAll returns events in the cluster
func (c *clientImpl) WatchAll(stopCh <-chan bool) (chan interface{}, chan error, error) {
	watchCh := make(chan interface{}, 10)
	errCh := make(chan error, 10)

	stopIngresses := make(chan bool)
	chanIngresses, chanIngressesErr, err := c.WatchIngresses(stopIngresses)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create watch: %v", err)
	}
	stopServices := make(chan bool)
	chanServices, chanServicesErr, err := c.WatchServices(stopServices)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create watch: %v", err)
	}
	stopEndpoints := make(chan bool)
	chanEndpoints, chanEndpointsErr, err := c.WatchEndpoints(stopEndpoints)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to create watch: %v", err)
	}
	go func() {
		defer close(watchCh)
		defer close(errCh)
		defer close(stopIngresses)
		defer close(stopServices)
		defer close(stopEndpoints)

		for {
			select {
			case <-stopCh:
				stopIngresses <- true
				stopServices <- true
				stopEndpoints <- true
				return
			case err := <-chanIngressesErr:
				errCh <- err
			case err := <-chanServicesErr:
				errCh <- err
			case err := <-chanEndpointsErr:
				errCh <- err
			case event := <-chanIngresses:
				watchCh <- event
			case event := <-chanServices:
				watchCh <- event
			case event := <-chanEndpoints:
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
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error %d GET %q: %q", res.StatusCode, request.Url, string(body))
	}
	return body, nil
}

func (c *clientImpl) request(url string) *gorequest.SuperAgent {
	// Make request to Kubernetes API
	request := gorequest.New().Get(url)
	request.Transport.DisableKeepAlives = true

	if strings.HasPrefix(url, "http://") {
		return request
	}

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
	watchCh := make(chan interface{}, 10)
	errCh := make(chan error, 10)

	// get version
	body, err := c.do(c.request(url))
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to do version request: GET %q : %v", url, err)
	}

	var generic GenericObject
	if err := json.Unmarshal(body, &generic); err != nil {
		return watchCh, errCh, fmt.Errorf("failed to decode version %v", err)
	}
	resourceVersion := generic.ResourceVersion

	url = url + "?watch&resourceVersion=" + resourceVersion
	// Make request to Kubernetes API
	request := c.request(url)
	req, err := request.MakeRequest()
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to make watch request: GET %q : %v", url, err)
	}
	request.Client.Transport = request.Transport

	res, err := request.Client.Do(req)
	if err != nil {
		return watchCh, errCh, fmt.Errorf("failed to do watch request: GET %q: %v", url, err)
	}

	go func() {
		finishCh := make(chan bool)
		defer close(finishCh)
		defer close(watchCh)
		defer close(errCh)
		go func() {
			defer res.Body.Close()
			for {
				var eventList interface{}
				if err := json.NewDecoder(res.Body).Decode(&eventList); err != nil {
					if !strings.Contains(err.Error(), "net/http: request canceled") {
						errCh <- fmt.Errorf("failed to decode watch event: GET %q : %v", url, err)
					}
					finishCh <- true
					return
				}
				watchCh <- eventList
			}
		}()
		select {
		case <-stopCh:
			go func() {
				request.Transport.CancelRequest(req)
			}()
			<-finishCh
			return
		}
	}()
	return watchCh, errCh, nil
}
