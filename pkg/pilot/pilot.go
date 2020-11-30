package pilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/traefik/traefik/v2/pkg/anonymize"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/version"
)

const (
	baseInstanceInfoURL = "https://instance-info.pilot.traefik.io/public"
	baseTelemetryURL    = "https://gateway.pilot.traefik.io"
	tokenHeader         = "X-Token"
	tokenHashHeader     = "X-Token-Hash"

	pilotInstanceInfoTimer = 5 * time.Minute
	pilotTelemetryTimer    = 12 * time.Hour
	maxElapsedTime         = 4 * time.Minute
)

type instanceInfo struct {
	ID      string                `json:"id,omitempty"`
	Metrics []metrics.PilotMetric `json:"metrics,omitempty"`
}

// New creates a new Pilot.
func New(token string, metricsRegistry *metrics.PilotRegistry, pool *safe.Pool) (*Pilot, error) {
	tokenHash := fnv.New64a()

	if _, err := tokenHash.Write([]byte(token)); err != nil {
		return nil, fmt.Errorf("unable to hash token: %w", err)
	}

	return &Pilot{
		dynamicConfigCh: make(chan dynamic.Configuration),
		client: &client{
			token:               token,
			tokenHash:           fmt.Sprintf("%x", tokenHash.Sum64()),
			httpClient:          &http.Client{Timeout: 5 * time.Second},
			baseInstanceInfoURL: baseInstanceInfoURL,
			baseTelemetryURL:    baseTelemetryURL,
		},
		routinesPool:    pool,
		metricsRegistry: metricsRegistry,
	}, nil
}

// Pilot connector with Pilot.
type Pilot struct {
	routinesPool *safe.Pool
	client       *client

	dynamicConfig   dynamic.Configuration
	dynamicConfigCh chan dynamic.Configuration
	metricsRegistry *metrics.PilotRegistry
}

// SetDynamicConfiguration stores the dynamic configuration.
func (p *Pilot) SetDynamicConfiguration(dynamicConfig dynamic.Configuration) {
	p.dynamicConfigCh <- dynamicConfig
}

func (p *Pilot) sendTelemetry(ctx context.Context, config dynamic.Configuration) {
	err := p.client.SendTelemetry(ctx, config)
	if err != nil {
		log.WithoutContext().Error(err)
	}
}

func (p *Pilot) sendInstanceInfo(ctx context.Context, pilotMetrics []metrics.PilotMetric) {
	err := p.client.SendInstanceInfo(ctx, pilotMetrics)
	if err != nil {
		log.WithoutContext().Error(err)
	}
}

// Tick sends data periodically.
func (p *Pilot) Tick(ctx context.Context) {
	pilotMetrics := p.metricsRegistry.Data()

	p.routinesPool.GoCtx(func(ctxRt context.Context) {
		p.sendInstanceInfo(ctxRt, pilotMetrics)
	})

	instanceInfoTicker := time.NewTicker(pilotInstanceInfoTimer)
	telemetryTicker := time.NewTicker(pilotTelemetryTimer)

	for {
		select {
		case tick := <-instanceInfoTicker.C:
			log.WithoutContext().Debugf("Send instance info to pilot: %s", tick)

			pilotMetrics := p.metricsRegistry.Data()

			p.routinesPool.GoCtx(func(ctxRt context.Context) {
				p.sendInstanceInfo(ctxRt, pilotMetrics)
			})
		case tick := <-telemetryTicker.C:
			log.WithoutContext().Debugf("Send telemetry to pilot: %s", tick)

			p.routinesPool.GoCtx(func(ctxRt context.Context) {
				p.sendTelemetry(ctxRt, p.dynamicConfig)
			})
		case dynamicConfig := <-p.dynamicConfigCh:
			p.dynamicConfig = dynamicConfig
		case <-ctx.Done():
			return
		}
	}
}

type client struct {
	httpClient          *http.Client
	baseInstanceInfoURL string
	baseTelemetryURL    string
	token               string
	tokenHash           string
	uuid                string
}

func (c *client) createUUID() (string, error) {
	data := []byte(`{"version":"` + version.Version + `","codeName":"` + version.Codename + `"}`)
	req, err := http.NewRequest(http.MethodPost, c.baseInstanceInfoURL+"/", bytes.NewBuffer(data))
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

// SendTelemetry sends telemetry to Pilot.
func (c *client) SendTelemetry(ctx context.Context, config dynamic.Configuration) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = maxElapsedTime

	anonConfig, err := anonymize.Do(&config, false)
	if err != nil {
		return fmt.Errorf("unable to anonymize dynamic configuration: %w", err)
	}

	return backoff.RetryNotify(
		func() error {
			req, err := http.NewRequest(http.MethodPost, c.baseTelemetryURL+"/collect", bytes.NewReader([]byte(anonConfig)))
			if err != nil {
				return fmt.Errorf("failed to create telemetry request: %w", err)
			}

			return c.sendData(req)
		},
		backoff.WithContext(exponentialBackOff, ctx),
		func(err error, duration time.Duration) {
			log.WithoutContext().Errorf("retry in %s due to: %v ", duration, err)
		})
}

// SendInstanceInfo sends instance information to Pilot.
func (c *client) SendInstanceInfo(ctx context.Context, pilotMetrics []metrics.PilotMetric) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = maxElapsedTime

	return backoff.RetryNotify(
		func() error {
			return c.sendInstanceInfo(pilotMetrics)
		},
		backoff.WithContext(exponentialBackOff, ctx),
		func(err error, duration time.Duration) {
			log.WithoutContext().Errorf("retry in %s due to: %v ", duration, err)
		})
}

func (c *client) sendInstanceInfo(pilotMetrics []metrics.PilotMetric) error {
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

	req, err := http.NewRequest(http.MethodPost, c.baseInstanceInfoURL+"/command", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create instance info request: %w", err)
	}

	return c.sendData(req)
}

func (c *client) sendData(req *http.Request) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(tokenHeader, c.token)
	req.Header.Set(tokenHashHeader, c.tokenHash)

	resp, err := c.httpClient.Do(req)
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
