package pilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/version"
)

const baseURL = "https://instance-info.pilot.traefik.io/public"

const tokenHeader = "X-Token"

const (
	pilotTimer     = 5 * time.Minute
	maxElapsedTime = 4 * time.Minute
)

// RunTimeRepresentation is the configuration information exposed by the API handler.
type RunTimeRepresentation struct {
	Routers     map[string]*runtime.RouterInfo        `json:"routers,omitempty"`
	Middlewares map[string]*runtime.MiddlewareInfo    `json:"middlewares,omitempty"`
	Services    map[string]*serviceInfoRepresentation `json:"services,omitempty"`
	TCPRouters  map[string]*runtime.TCPRouterInfo     `json:"tcpRouters,omitempty"`
	TCPServices map[string]*runtime.TCPServiceInfo    `json:"tcpServices,omitempty"`
	UDPRouters  map[string]*runtime.UDPRouterInfo     `json:"udpRouters,omitempty"`
	UDPServices map[string]*runtime.UDPServiceInfo    `json:"udpServices,omitempty"`
}

type serviceInfoRepresentation struct {
	*runtime.ServiceInfo
	ServerStatus map[string]string `json:"serverStatus,omitempty"`
}

type instanceInfo struct {
	ID            string                `json:"id,omitempty"`
	Configuration RunTimeRepresentation `json:"configuration,omitempty"`
	Metrics       []metrics.PilotMetric `json:"metrics,omitempty"`
}

// New creates a new Pilot.
func New(token string, metricsRegistry *metrics.PilotRegistry, pool *safe.Pool) *Pilot {
	return &Pilot{
		rtConfChan: make(chan *runtime.Configuration),
		client: &client{
			token:      token,
			httpClient: &http.Client{Timeout: 5 * time.Second},
			baseURL:    baseURL,
		},
		routinesPool:    pool,
		metricsRegistry: metricsRegistry,
	}
}

// Pilot connector with Pilot.
type Pilot struct {
	routinesPool *safe.Pool
	client       *client

	rtConf          *runtime.Configuration
	rtConfChan      chan *runtime.Configuration
	metricsRegistry *metrics.PilotRegistry
}

// SetRuntimeConfiguration stores the runtime configuration.
func (p *Pilot) SetRuntimeConfiguration(rtConf *runtime.Configuration) {
	p.rtConfChan <- rtConf
}

func (p *Pilot) getRepresentation() RunTimeRepresentation {
	if p.rtConf == nil {
		return RunTimeRepresentation{}
	}

	siRepr := make(map[string]*serviceInfoRepresentation, len(p.rtConf.Services))
	for k, v := range p.rtConf.Services {
		siRepr[k] = &serviceInfoRepresentation{
			ServiceInfo:  v,
			ServerStatus: v.GetAllStatus(),
		}
	}

	result := RunTimeRepresentation{
		Routers:     p.rtConf.Routers,
		Middlewares: p.rtConf.Middlewares,
		Services:    siRepr,
		TCPRouters:  p.rtConf.TCPRouters,
		TCPServices: p.rtConf.TCPServices,
		UDPRouters:  p.rtConf.UDPRouters,
		UDPServices: p.rtConf.UDPServices,
	}

	return result
}

func (p *Pilot) sendData(ctx context.Context, conf RunTimeRepresentation, pilotMetrics []metrics.PilotMetric) {
	err := p.client.SendData(ctx, conf, pilotMetrics)
	if err != nil {
		log.WithoutContext().Error(err)
	}
}

// Tick sends data periodically.
func (p *Pilot) Tick(ctx context.Context) {
	select {
	case rtConf := <-p.rtConfChan:
		p.rtConf = rtConf
		break
	case <-ctx.Done():
		return
	}

	conf := p.getRepresentation()
	pilotMetrics := p.metricsRegistry.Data()

	p.routinesPool.GoCtx(func(ctxRt context.Context) {
		p.sendData(ctxRt, conf, pilotMetrics)
	})

	ticker := time.NewTicker(pilotTimer)
	for {
		select {
		case tick := <-ticker.C:
			log.WithoutContext().Debugf("Send to pilot: %s", tick)

			conf := p.getRepresentation()
			pilotMetrics := p.metricsRegistry.Data()

			p.routinesPool.GoCtx(func(ctxRt context.Context) {
				p.sendData(ctxRt, conf, pilotMetrics)
			})
		case rtConf := <-p.rtConfChan:
			p.rtConf = rtConf
		case <-ctx.Done():
			return
		}
	}
}

type client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	uuid       string
}

func (c *client) createUUID() (string, error) {
	data := []byte(`{"version":"` + version.Version + `","codeName":"` + version.Codename + `"}`)
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(tokenHeader, c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed call Pilot: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed read response body: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("wrong status code while sending configuration: %d: %s", resp.StatusCode, body)
	}

	created := instanceInfo{}
	err = json.Unmarshal(body, &created)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return created.ID, nil
}

// SendData sends data to Pilot.
func (c *client) SendData(ctx context.Context, rtConf RunTimeRepresentation, pilotMetrics []metrics.PilotMetric) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = maxElapsedTime

	return backoff.RetryNotify(
		func() error {
			return c.sendData(rtConf, pilotMetrics)
		},
		backoff.WithContext(exponentialBackOff, ctx),
		func(err error, duration time.Duration) {
			log.WithoutContext().Errorf("retry in %s due to: %v ", duration, err)
		})
}

func (c *client) sendData(_ RunTimeRepresentation, pilotMetrics []metrics.PilotMetric) error {
	if len(c.uuid) == 0 {
		var err error
		c.uuid, err = c.createUUID()
		if err != nil {
			return fmt.Errorf("failed to create UUID: %w", err)
		}

		version.UUID = c.uuid
	}

	info := instanceInfo{
		ID:      c.uuid,
		Metrics: pilotMetrics,
	}

	b, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshall request body: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, c.baseURL+"/command", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(tokenHeader, c.token)

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to call Pilot: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wrong status code while sending configuration: %d: %s", resp.StatusCode, body)
	}

	return nil
}
