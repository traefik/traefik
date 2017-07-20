package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	// Blank import from pq package is needed for sql driver to initialize based on driver type passed as an
	//argument to sql open function
	_ "github.com/lib/pq"
	"strconv"
	"time"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configuration for provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	RefreshSeconds        int    `description:"Polling interval (in seconds)"`
	Endpoint              string `description:"The endpoint of a postgres RDS. Used for testing with a local postgres"`
	TableName             string `description:"The AWS postgres table that stores configuration for traefik"`
}

type postgresClient struct {
	db *sql.DB
}

// createClient configures aws credentials and createsrds a dynamoClient
func (p *Provider) createClient() (*postgresClient, error) {
	log.Info("Creating Provider client...")

	if p.Endpoint == "" {
		log.Errorf("Postgres endpoint is not configured")
		err := errors.New("Postgres endpoint is not configured")
		return nil, err
	}

	db, err := sql.Open("postgres", p.Endpoint)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &postgresClient{
		db,
	}, nil
}

// queryTable queries the given table and returns slice of all items in the table
func (p *Provider) queryTable(client *postgresClient) map[string][]string {
	log.Debugf("Scanning Provider table: %s ...", p.TableName)

	querystring := "select * from " + p.TableName

	rows, err := client.db.Query(querystring)
	if err != nil {
		log.Fatal(err)
		log.Errorf("Failed to query Provider table %s", p.TableName)
		return nil
	}
	log.Debugf("Successfully scanned Provider table %s", p.TableName)

	config := make(map[string][]string)

	var (
		id          int
		uri, ipport string
	)

	for rows.Next() {
		err := rows.Scan(&id, &uri, &ipport)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Retrieved following uri %s ipport %s info", uri, ipport)
		config[uri] = append(config[uri], ipport)
	}

	log.Debugf("Final map: %s ...", config)
	return config
}

// queryTable retrieves items from postgres and converts them into Backends and Frontends in a Configuration
func (p *Provider) loadPostgresConfig(client *postgresClient) (*types.Configuration, error) {
	config := p.queryTable(client)

	log.Debugf("Number of Items retrieved from Provider: %d", len(config))
	backends := make(map[string]*types.Backend)
	frontends := make(map[string]*types.Frontend)
	// unmarshal map into Backends and Frontends

	backendCounter := 1

	frontendCounter := 1
	for k, v := range config {
		log.Debugf("Provider Item: %d\n%v", k, v)

		servers := make(map[string]types.Server)
		backendName := "backend" + strconv.Itoa(backendCounter)
		log.Println("backendName %s ", backendName)

		for index, ipport := range v {
			backendServerName := "backends" + backendName + "servers.server" + strconv.Itoa(index)
			log.Println("backendServerName %s ", backendServerName)
			url := ipport
			wght := 1
			server := types.Server{URL: url, Weight: wght}
			servers[backendServerName] = server
		}

		tmpBackend := types.Backend{Servers: servers}
		backends[backendName] = &tmpBackend

		frontendName := "frontend" + strconv.Itoa(frontendCounter)
		log.Println("frontendName %s ", frontendName)

		frontendRouteName := "frontend." + frontendName + "routes"
		route := make(map[string]types.Route)
		rule := "Path: /" + k
		route[frontendRouteName] = types.Route{Rule: rule}

		tmpFrontend := types.Frontend{Backend: backendName, Routes: route}
		frontends[frontendName] = &tmpFrontend

		backendCounter++
		frontendCounter++
	}

	return &types.Configuration{
		Backends:  backends,
		Frontends: frontends,
	}, nil
}

// Provide provides the configuration to traefik via the configuration channel
// if watch is enabled it polls postgres
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	log.Debugf("Providing Provider...")
	p.Constraints = append(p.Constraints, constraints...)
	handleCanceled := func(ctx context.Context, err error) error {
		if ctx.Err() == context.Canceled || err == context.Canceled {
			return nil
		}
		return err
	}

	pool.Go(func(stop chan bool) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			select {
			case <-stop:
				cancel()
			}
		}()

		operation := func() error {
			aws, err := p.createClient()
			if err != nil {
				return handleCanceled(ctx, err)
			}

			configuration, err := p.loadPostgresConfig(aws)
			if err != nil {
				return handleCanceled(ctx, err)
			}

			configurationChan <- types.ConfigMessage{
				ProviderName:  "postgres",
				Configuration: configuration,
			}

			if p.Watch {
				reload := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
				defer reload.Stop()
				for {
					log.Debug("Watching Provider...")
					select {
					case <-reload.C:
						configuration, err := p.loadPostgresConfig(aws)
						if err != nil {
							return handleCanceled(ctx, err)
						}

						configurationChan <- types.ConfigMessage{
							ProviderName:  "postgres",
							Configuration: configuration,
						}
					case <-ctx.Done():
						return handleCanceled(ctx, ctx.Err())
					}
				}
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Provider error: %s time: %v", err.Error(), time)
		}

		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Failed to connect to Provider. %s", err.Error())
		}
	})
	return nil
}
