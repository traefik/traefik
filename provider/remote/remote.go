package remote

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/go-cleanhttp"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	URL                   string           `description:"External URL to call using HTTP GET to retrieve the content of a .toml configuration" export:"true"`
	TLS                   *types.ClientTLS `description:"Enable TLS support" export:"true"`
	LongPollDuration      string           `description:"In case 'watch' is set to true, this indicates how long to wait for a socket long poll before retrying.  This value will be passed as additional 'wait' query parameter in seconds" export:"true"`
	RepeatInterval        string           `description:"In case 'watch' is set to true, but no 'longPollDuration' is configured, this indicates how long to wait till repeating HTTP call again.  If no value is set, 10s will be used by default." export:"true"`
	httpClient            *http.Client
	httpTransport         *http.Transport
	pollDuration          time.Duration
	refreshDuration       time.Duration
}

// Provide allows the remote http provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	p.httpTransport = cleanhttp.DefaultPooledTransport()
	if p.TLS != nil {
		tlsConfig, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return err
		}
		p.httpTransport.TLSClientConfig = tlsConfig
	}

	if len(p.LongPollDuration) > 0 {
		longPollDuration, err := time.ParseDuration(p.LongPollDuration)
		if err != nil {
			return err
		}
		p.pollDuration = longPollDuration
	}
	if len(p.RepeatInterval) > 0 {
		interval, err := time.ParseDuration(p.RepeatInterval)
		if err != nil {
			return err
		}
		p.refreshDuration = interval
	}
	p.httpClient = &http.Client{
		Transport: p.httpTransport,
		Timeout:   p.pollDuration,
	}
	configuration, err := p.loadConfig()

	if err != nil {
		return err
	}

	if p.Watch {
		if err := p.addWatcher(pool, configurationChan, p.watcherCallback); err != nil {
			return err
		}
	}

	sendConfigToChannel(configurationChan, configuration)
	return nil
}

func (p *Provider) addWatcher(pool *safe.Pool, configurationChan chan<- types.ConfigMessage, callback func(chan<- types.ConfigMessage)) error {
	// Process events
	pool.Go(func(stop chan bool) {
		defer p.httpTransport.CloseIdleConnections()
		for {
			select {
			case <-stop:
				return
			default:
				if p.Watch && p.refreshDuration != 0 {
					time.Sleep(p.refreshDuration)
				}
				// Guard against spamming the HTTP server, if no RepeatInterval and LongPollDuration are configured, but Wait is selected.
				if p.Watch && p.refreshDuration == 0 && p.pollDuration == 0 {
					time.Sleep(10 * time.Second)
				}
				callback(configurationChan)
			}
		}
	})
	return nil
}

func sendConfigToChannel(configurationChan chan<- types.ConfigMessage, configuration *types.Configuration) {
	configurationChan <- types.ConfigMessage{
		ProviderName:  "remote",
		Configuration: configuration,
	}
}

func parseConfigFromContent(data io.Reader) (*types.Configuration, error) {
	configuration := new(types.Configuration)
	if _, err := toml.DecodeReader(data, configuration); err != nil {
		return nil, fmt.Errorf("error reading configuration content: %s", err)
	}
	return configuration, nil
}

func (p *Provider) watcherCallback(configurationChan chan<- types.ConfigMessage) {
	configuration, err := p.loadConfig()

	if err != nil {
		log.Errorf("Error occurred during watcher callback: %s", err)
		return
	}

	sendConfigToChannel(configurationChan, configuration)
}

func (p *Provider) createHTTPRequest() (*http.Request, error) {
	u, err := url.Parse(p.URL)
	if err != nil {
		return nil, err
	}

	if p.Watch && p.pollDuration != 0 {
		query := u.Query()
		query.Add("wait", fmt.Sprintf("%ds", int(p.pollDuration.Seconds())))
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	req.Header.Add("User-Agent", "Traefik")

	// TODO: Add other relevant/useful headers, etc.
	return req, err
}

func (p *Provider) loadConfig() (*types.Configuration, error) {
	req, err := p.createHTTPRequest()

	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)

	if err != nil || resp == nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to retrieve remote config via calling %s: received %d-%s", p.URL, resp.StatusCode, resp.Status)
	}

	return parseConfigFromContent(resp.Body)
}
