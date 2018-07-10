package dynamodb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configuration for provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	AccessKeyID           string `description:"The AWS credentials access key to use for making requests"`
	RefreshSeconds        int    `description:"Polling interval (in seconds)" export:"true"`
	Region                string `description:"The AWS region to use for requests" export:"true"`
	SecretAccessKey       string `description:"The AWS credentials secret key to use for making requests"`
	TableName             string `description:"The AWS dynamodb table that stores configuration for traefik" export:"true"`
	Endpoint              string `description:"The endpoint of a dynamodb. Used for testing with a local dynamodb"`
}

type dynamoClient struct {
	db dynamodbiface.DynamoDBAPI
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	return p.BaseProvider.Init(constraints)
}

// createClient configures aws credentials and creates a dynamoClient
func (p *Provider) createClient() (*dynamoClient, error) {
	log.Info("Creating Provider client...")
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	if p.Region == "" {
		return nil, errors.New("no Region provided for Provider")
	}
	cfg := &aws.Config{
		Region: &p.Region,
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     p.AccessKeyID,
						SecretAccessKey: p.SecretAccessKey,
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				defaults.RemoteCredProvider(*(defaults.Config()), defaults.Handlers()),
			}),
	}

	if p.Trace {
		cfg.WithLogger(aws.LoggerFunc(func(args ...interface{}) {
			log.Debug(args...)
		}))
	}

	if p.Endpoint != "" {
		cfg.Endpoint = aws.String(p.Endpoint)
	}

	return &dynamoClient{
		dynamodb.New(sess, cfg),
	}, nil
}

// scanTable scans the given table and returns slice of all items in the table
func (p *Provider) scanTable(client *dynamoClient) ([]map[string]*dynamodb.AttributeValue, error) {
	log.Debugf("Scanning Provider table: %s ...", p.TableName)
	params := &dynamodb.ScanInput{
		TableName: aws.String(p.TableName),
	}
	items := make([]map[string]*dynamodb.AttributeValue, 0)
	err := client.db.ScanPages(params,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			items = append(items, page.Items...)
			return !lastPage
		})
	if err != nil {
		log.Errorf("Failed to scan Provider table %s", p.TableName)
		return nil, err
	}
	log.Debugf("Successfully scanned Provider table %s", p.TableName)
	return items, nil
}

// buildConfiguration retrieves items from dynamodb and converts them into Backends and Frontends in a Configuration
func (p *Provider) buildConfiguration(client *dynamoClient) (*types.Configuration, error) {
	items, err := p.scanTable(client)
	if err != nil {
		return nil, err
	}
	log.Debugf("Number of Items retrieved from Provider: %d", len(items))
	backends := make(map[string]*types.Backend)
	frontends := make(map[string]*types.Frontend)
	// unmarshal dynamoAttributes into Backends and Frontends
	for i, item := range items {
		log.Debugf("Provider Item: %d\n%v", i, item)
		// verify the type of each item by checking to see if it has
		// the corresponding type, backend or frontend map
		if backend, exists := item["backend"]; exists {
			log.Debug("Unmarshaling backend from Provider...")
			tmpBackend := &types.Backend{}
			err = dynamodbattribute.Unmarshal(backend, tmpBackend)
			if err != nil {
				log.Errorf(err.Error())
			} else {
				backends[*item["name"].S] = tmpBackend
				log.Debug("Backend from Provider unmarshalled successfully")
			}
		} else if frontend, exists := item["frontend"]; exists {
			log.Debugf("Unmarshaling frontend from Provider...")
			tmpFrontend := &types.Frontend{}
			err = dynamodbattribute.Unmarshal(frontend, tmpFrontend)
			if err != nil {
				log.Errorf(err.Error())
			} else {
				frontends[*item["name"].S] = tmpFrontend
				log.Debug("Frontend from Provider unmarshalled successfully")
			}
		} else {
			log.Warnf("Error in format of Provider Item: %v", item)
		}
	}

	return &types.Configuration{
		Backends:  backends,
		Frontends: frontends,
	}, nil
}

// Provide provides the configuration to traefik via the configuration channel
// if watch is enabled it polls dynamodb
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	handleCanceled := func(ctx context.Context, err error) error {
		if ctx.Err() == context.Canceled || err == context.Canceled {
			return nil
		}
		return err
	}

	pool.Go(func(stop chan bool) {
		ctx, cancel := context.WithCancel(context.Background())
		safe.Go(func() {
			<-stop
			cancel()
		})

		operation := func() error {
			awsClient, err := p.createClient()
			if err != nil {
				return handleCanceled(ctx, err)
			}

			configuration, err := p.buildConfiguration(awsClient)
			if err != nil {
				return handleCanceled(ctx, err)
			}

			configurationChan <- types.ConfigMessage{
				ProviderName:  "dynamodb",
				Configuration: configuration,
			}

			if p.Watch {
				reload := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
				defer reload.Stop()
				for {
					log.Debug("Watching Provider...")
					select {
					case <-reload.C:
						configuration, err := p.buildConfiguration(awsClient)
						if err != nil {
							return handleCanceled(ctx, err)
						}

						configurationChan <- types.ConfigMessage{
							ProviderName:  "dynamodb",
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
