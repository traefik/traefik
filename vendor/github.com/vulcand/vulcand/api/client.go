package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/plugin"

	log "github.com/Sirupsen/logrus"
)

const CurrentVersion = "v2"

type Client struct {
	Addr     string
	Registry *plugin.Registry
}

func NewClient(addr string, registry *plugin.Registry) *Client {
	return &Client{Addr: addr, Registry: registry}
}

func (c *Client) GetStatus() error {
	_, err := c.Get(c.endpoint("status"), url.Values{})
	return err
}

func (c *Client) GetHosts() ([]engine.Host, error) {
	data, err := c.Get(c.endpoint("hosts"), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.HostsFromJSON(data)
}

func (c *Client) UpdateLogSeverity(s log.Level) error {
	return c.PutForm(c.endpoint("log", "severity"), url.Values{"severity": {s.String()}})
}

func (c *Client) GetLogSeverity() (log.Level, error) {
	data, err := c.Get(c.endpoint("log", "severity"), url.Values{})
	if err != nil {
		return 255, err
	}
	var sev *SeverityResponse
	if err := json.Unmarshal(data, &sev); err != nil {
		return 255, err
	}
	lvl, err := log.ParseLevel(strings.ToLower(sev.Severity))
	if err != nil {
		return 255, err
	}
	return lvl, nil
}

func (c *Client) GetHost(hk engine.HostKey) (*engine.Host, error) {
	response, err := c.Get(c.endpoint("hosts", hk.Name), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.HostFromJSON(response)
}

func (c *Client) UpsertHost(h engine.Host) error {
	_, err := c.Post(c.endpoint("hosts"), hostPack{Host: h})
	return err
}

func (c *Client) UpsertListener(l engine.Listener) error {
	_, err := c.Post(c.endpoint("listeners"), listenerPack{Listener: l})
	return err
}

func (c *Client) GetListener(lk engine.ListenerKey) (*engine.Listener, error) {
	data, err := c.Get(c.endpoint("listeners", lk.Id), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.ListenerFromJSON(data)
}

func (c *Client) GetListeners() ([]engine.Listener, error) {
	data, err := c.Get(c.endpoint("listeners"), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.ListenersFromJSON(data)
}

func (c *Client) DeleteListener(lk engine.ListenerKey) error {
	return c.Delete(c.endpoint("listeners", lk.Id))
}

func (c *Client) DeleteHost(hk engine.HostKey) error {
	return c.Delete(c.endpoint("hosts", hk.Name))
}

func (c *Client) UpsertFrontend(f engine.Frontend, ttl time.Duration) error {
	_, err := c.Post(c.endpoint("frontends"), frontendPack{Frontend: f, TTL: ttl.String()})
	return err
}

func (c *Client) GetFrontend(fk engine.FrontendKey) (*engine.Frontend, error) {
	response, err := c.Get(c.endpoint("frontends", fk.Id), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.FrontendFromJSON(c.Registry.GetRouter(), response)
}

func (c *Client) GetFrontends() ([]engine.Frontend, error) {
	data, err := c.Get(c.endpoint("frontends"), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.FrontendsFromJSON(c.Registry.GetRouter(), data)
}

func (c *Client) TopFrontends(bk *engine.BackendKey, limit int) ([]engine.Frontend, error) {
	values := url.Values{
		"limit": {fmt.Sprintf("%d", limit)},
	}
	if bk != nil {
		values["backendId"] = []string{bk.Id}
	}
	response, err := c.Get(c.endpoint("top", "frontends"), values)
	if err != nil {
		return nil, err
	}
	return engine.FrontendsFromJSON(c.Registry.GetRouter(), response)
}

func (c *Client) DeleteFrontend(fk engine.FrontendKey) error {
	return c.Delete(c.endpoint("frontends", fk.Id))
}

func (c *Client) UpsertBackend(b engine.Backend) error {
	if b.Id == "" {
		return fmt.Errorf("frontend id and middleware id can not be empty")
	}
	_, err := c.Post(c.endpoint("backends"), backendPack{Backend: b})
	return err
}

func (c *Client) DeleteBackend(bk engine.BackendKey) error {
	return c.Delete(c.endpoint("backends", bk.Id))
}

func (c *Client) GetBackend(bk engine.BackendKey) (*engine.Backend, error) {
	response, err := c.Get(c.endpoint("backends", bk.Id), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.BackendFromJSON(response)
}

func (c *Client) GetBackends() ([]engine.Backend, error) {
	data, err := c.Get(c.endpoint("backends"), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.BackendsFromJSON(data)
}

func (c *Client) UpsertServer(bk engine.BackendKey, srv engine.Server, ttl time.Duration) error {
	if bk.Id == "" || srv.Id == "" {
		return fmt.Errorf("backend id and server id can not be empty")
	}
	_, err := c.Post(c.endpoint("backends", bk.Id, "servers"), serverPack{Server: srv, TTL: ttl.String()})
	return err
}

func (c *Client) TopServers(bk *engine.BackendKey, limit int) ([]engine.Server, error) {
	values := url.Values{
		"limit": {fmt.Sprintf("%d", limit)},
	}
	if bk != nil {
		values["backendId"] = []string{bk.Id}
	}
	response, err := c.Get(c.endpoint("top", "servers"), values)
	if err != nil {
		return nil, err
	}
	var re *ServersResponse
	if err = json.Unmarshal(response, &re); err != nil {
		return nil, err
	}
	return re.Servers, nil
}

func (c *Client) GetServer(sk engine.ServerKey) (*engine.Server, error) {
	data, err := c.Get(c.endpoint("backends", sk.BackendKey.Id, "servers", sk.Id), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.ServerFromJSON(data)
}

func (c *Client) GetServers(bk engine.BackendKey) ([]engine.Server, error) {
	if bk.Id == "" {
		return nil, fmt.Errorf("backend id can not be empty")
	}
	data, err := c.Get(c.endpoint("backends", bk.Id, "servers"), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.ServersFromJSON(data)
}

func (c *Client) DeleteServer(sk engine.ServerKey) error {
	if sk.BackendKey.Id == "" {
		return fmt.Errorf("backend id can not be empty")
	}
	return c.Delete(c.endpoint("backends", sk.BackendKey.Id, "servers", sk.Id))
}

func (c *Client) UpsertMiddleware(fk engine.FrontendKey, m engine.Middleware, ttl time.Duration) error {
	if fk.Id == "" || m.Id == "" {
		return fmt.Errorf("frontend id and middleware id can not be empty")
	}
	_, err := c.Post(
		c.endpoint("frontends", fk.Id, "middlewares"), middlewarePack{Middleware: m, TTL: ttl.String()})
	return err
}

func (c *Client) GetMiddleware(mk engine.MiddlewareKey) (*engine.Middleware, error) {
	data, err := c.Get(c.endpoint("frontends", mk.FrontendKey.Id, "middlewares", mk.Id), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.MiddlewareFromJSON(data, c.Registry.GetSpec)
}

func (c *Client) GetMiddlewares(fk engine.FrontendKey) ([]engine.Middleware, error) {
	data, err := c.Get(c.endpoint("frontends", fk.Id, "middlewares"), url.Values{})
	if err != nil {
		return nil, err
	}
	return engine.MiddlewaresFromJSON(data, c.Registry.GetSpec)
}

func (c *Client) DeleteMiddleware(mk engine.MiddlewareKey) error {
	return c.Delete(c.endpoint("frontends", mk.FrontendKey.Id, "middlewares", mk.Id))
}

func (c *Client) PutForm(endpoint string, values url.Values) error {
	_, err := c.RoundTrip(func() (*http.Response, error) {
		req, err := http.NewRequest("PUT", endpoint, strings.NewReader(values.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return http.DefaultClient.Do(req)
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Post(endpoint string, in interface{}) ([]byte, error) {
	return c.RoundTrip(func() (*http.Response, error) {
		data, err := json.Marshal(in)
		if err != nil {
			return nil, err
		}
		return http.Post(endpoint, "application/json", bytes.NewBuffer(data))
	})
}

func (c *Client) Put(endpoint string, in interface{}) ([]byte, error) {
	return c.RoundTrip(func() (*http.Response, error) {
		data, err := json.Marshal(in)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		re, err := http.DefaultClient.Do(req)
		return re, err
	})
}

func (c *Client) Delete(endpoint string) error {
	data, err := c.RoundTrip(func() (*http.Response, error) {
		req, err := http.NewRequest("DELETE", endpoint, nil)
		if err != nil {
			return nil, err
		}
		return http.DefaultClient.Do(req)
	})
	if err != nil {
		return err
	}
	var re *StatusResponse
	err = json.Unmarshal(data, &re)
	return err
}

func (c *Client) Get(u string, params url.Values) ([]byte, error) {
	baseUrl, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	baseUrl.RawQuery = params.Encode()
	return c.RoundTrip(func() (*http.Response, error) {
		return http.Get(baseUrl.String())
	})
}

type RoundTripFn func() (*http.Response, error)

func (c *Client) RoundTrip(fn RoundTripFn) ([]byte, error) {
	response, err := fn()
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		var status *StatusResponse
		if err := json.Unmarshal(responseBody, &status); err != nil {
			return nil, fmt.Errorf("failed to decode response '%s', error: %v", responseBody, err)
		}
		if response.StatusCode == http.StatusNotFound {
			return nil, &engine.NotFoundError{Message: status.Message}
		}
		if response.StatusCode == http.StatusConflict {
			return nil, &engine.AlreadyExistsError{Message: status.Message}
		}
		return nil, status
	}
	return responseBody, nil
}

func (c *Client) endpoint(params ...string) string {
	return fmt.Sprintf("%s/%s/%s", c.Addr, CurrentVersion, strings.Join(params, "/"))
}

type BackendsResponse struct {
	Backends []engine.Backend
}

type ServersResponse struct {
	Servers []engine.Server
}

type StatusResponse struct {
	Message string
}

func (e *StatusResponse) Error() string {
	return e.Message
}

type ConnectionsResponse struct {
	Connections int
}

type SeverityResponse struct {
	Severity string
}
