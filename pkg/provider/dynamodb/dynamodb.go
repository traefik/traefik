package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/cenkalti/backoff/v4"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/job"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/safe"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configuration for provider.
type Provider struct {
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
func (p *Provider) Init() error {
	return nil
}

// createClient configures aws credentials and creates a dynamoClient
func (p *Provider) createClient(logger log.Logger) (*dynamoClient, error) {
	logger.Infoln("Creating Provider client...")
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

	if p.Endpoint != "" {
		cfg.Endpoint = aws.String(p.Endpoint)
	}

	cfg.WithLogger(aws.LoggerFunc(func(args ...interface{}) {
		logger.Debug(args...)
	}))

	return &dynamoClient{
		dynamodb.New(sess, cfg),
	}, nil
}

// scanTable scans the given table and returns slice of all items in the table
func (p *Provider) scanTable(client *dynamoClient, logger log.Logger) ([]map[string]*dynamodb.AttributeValue, error) {
	logger.Debugf("Scanning Provider table: %s ...", p.TableName)
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
		logger.Errorf("Failed to scan Provider table %s", p.TableName)
		return nil, err
	}
	logger.Debugf("Successfully scanned Provider table %s", p.TableName)
	return items, nil
}

// Provide provides the configuration to traefik via the configuration channel
// if watch is enabled it polls dynamodb
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, "ecs"))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			awsClient, err := p.createClient(logger)
			if err != nil {
				return fmt.Errorf("unable to create AWS client: %w", err)
			}

			configuration, err := p.buildConfiguration(ctxLog, awsClient)
			if err != nil {
				return fmt.Errorf("failed to get DynamoDB configuration: %w", err)
			}

			configurationChan <- dynamic.Message{
				ProviderName:  "dynamodb",
				Configuration: configuration,
			}

			ticker := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					err = p.loadConfiguration(ctxLog, awsClient, configurationChan)
					if err != nil {
						return fmt.Errorf("failed to refresh ECS configuration: %w", err)
					}

				case <-routineCtx.Done():
					return nil
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Errorf("Provider error: %s time: %v", err.Error(), time)
		}

		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), routineCtx), notify)
		if err != nil {
			logger.Errorf("Failed to connect to Provider. %s", err.Error())
		}
	})
	return nil
}

func (p *Provider) loadConfiguration(ctx context.Context, client *dynamoClient, configurationChan chan<- dynamic.Message) error {
	conf, err := p.buildConfiguration(ctx, client)
	if err != nil {
		return err
	}

	configurationChan <- dynamic.Message{
		ProviderName:  "dynamodb",
		Configuration: conf,
	}

	return nil
}
