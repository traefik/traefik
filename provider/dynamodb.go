package provider

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
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

var _ Provider = (*DynamoDB)(nil)

// DynamoDB holds configuration for DynamoDB provider.
type DynamoDB struct {
	BaseProvider `mapstructure:",squash"`

	AccessKeyID     string `description:"The AWS credentials access key to use for making requests"`
	RefreshSeconds  int    `description:"Polling interval (in seconds)"`
	Region          string `description:"The AWS region to use for requests"`
	SecretAccessKey string `description:"The AWS credentals secret key to use for making requests"`
	TableName       string `description:"The AWS dynamodb table that stores configuration for traefik"`
	Endpoint        string `description:"The endpoint of a dynamodb. Used for testing with a local dynamodb"`
}

type dynamoClient struct {
	db dynamodbiface.DynamoDBAPI
}

// createClient configures aws credentials and creates a dynamoClient
func (provider *DynamoDB) createClient() (*dynamoClient, error) {
	log.Infof("Creating DynamoDB client...")
	sess := session.New()
	if provider.Region == "" {
		return nil, errors.New("no Region provided for DynamoDB")
	}
	cfg := &aws.Config{
		Region: &provider.Region,
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     provider.AccessKeyID,
						SecretAccessKey: provider.SecretAccessKey,
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				defaults.RemoteCredProvider(*(defaults.Config()), defaults.Handlers()),
			}),
	}

	if provider.Endpoint != "" {
		cfg.Endpoint = aws.String(provider.Endpoint)
	}

	return &dynamoClient{
		dynamodb.New(sess, cfg),
	}, nil
}

// scanTable scans the given table and returns slice of all items in the table
func (provider *DynamoDB) scanTable(client *dynamoClient) ([]map[string]*dynamodb.AttributeValue, error) {
	log.Debugf("Scanning DynamoDB table: %s ...", provider.TableName)
	params := &dynamodb.ScanInput{
		TableName: aws.String(provider.TableName),
	}
	items := make([]map[string]*dynamodb.AttributeValue, 0)
	err := client.db.ScanPages(params,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			items = append(items, page.Items...)
			return !lastPage
		})
	if err != nil {
		log.Errorf("Failed to scan DynamoDB table %s", provider.TableName)
		return nil, err
	}
	log.Debugf("Successfully scanned DynamoDB table %s", provider.TableName)
	return items, nil
}

// loadDynamoConfig retrieves items from dynamodb and converts them into Backends and Frontends in a Configuration
func (provider *DynamoDB) loadDynamoConfig(client *dynamoClient) (*types.Configuration, error) {
	items, err := provider.scanTable(client)
	if err != nil {
		return nil, err
	}
	log.Debugf("Number of Items retrieved from DynamoDB: %d", len(items))
	backends := make(map[string]*types.Backend)
	frontends := make(map[string]*types.Frontend)
	// unmarshal dynamoAttributes into Backends and Frontends
	for i, item := range items {
		log.Debugf("DynamoDB Item: %d\n%v", i, item)
		// verify the type of each item by checking to see if it has
		// the corresponding type, backend or frontend map
		if backend, exists := item["backend"]; exists {
			log.Debugf("Unmarshaling backend from DynamoDB...")
			tmpBackend := &types.Backend{}
			err = dynamodbattribute.Unmarshal(backend, tmpBackend)
			if err != nil {
				log.Errorf(err.Error())
			} else {
				backends[*item["name"].S] = tmpBackend
				log.Debugf("Backend from DynamoDB unmarshalled successfully")
			}
		} else if frontend, exists := item["frontend"]; exists {
			log.Debugf("Unmarshaling frontend from DynamoDB...")
			tmpFrontend := &types.Frontend{}
			err = dynamodbattribute.Unmarshal(frontend, tmpFrontend)
			if err != nil {
				log.Errorf(err.Error())
			} else {
				frontends[*item["name"].S] = tmpFrontend
				log.Debugf("Frontend from DynamoDB unmarshalled successfully")
			}
		} else {
			log.Warnf("Error in format of DynamoDB Item: %v", item)
		}
	}

	return &types.Configuration{
		Backends:  backends,
		Frontends: frontends,
	}, nil
}

// Provide provides the configuration to traefik via the configuration channel
// if watch is enabled it polls dynamodb
func (provider *DynamoDB) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	log.Debugf("Providing DynamoDB...")
	provider.Constraints = append(provider.Constraints, constraints...)
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
			aws, err := provider.createClient()
			if err != nil {
				return handleCanceled(ctx, err)
			}

			configuration, err := provider.loadDynamoConfig(aws)
			if err != nil {
				return handleCanceled(ctx, err)
			}

			configurationChan <- types.ConfigMessage{
				ProviderName:  "dynamodb",
				Configuration: configuration,
			}

			if provider.Watch {
				reload := time.NewTicker(time.Second * time.Duration(provider.RefreshSeconds))
				defer reload.Stop()
				for {
					log.Debugf("Watching DynamoDB...")
					select {
					case <-reload.C:
						configuration, err := provider.loadDynamoConfig(aws)
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
			log.Errorf("DynamoDB error: %s time: %v", err.Error(), time)
		}

		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Failed to connect to DynamoDB. %s", err.Error())
		}
	})
	return nil
}
