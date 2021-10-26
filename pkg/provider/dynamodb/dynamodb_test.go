package dynamodb

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"

	"github.com/stretchr/testify/assert"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	testWithError bool
}

var service = &dynamic.Service{
	LoadBalancer: &dynamic.ServersLoadBalancer{
		HealthCheck: &dynamic.ServerHealthCheck{
			Path: "/build",
		},
		Servers: []dynamic.Server{
			{
				URL: "http://test.traefik.io",
			},
		},
	},
}

var router = &dynamic.Router{
	EntryPoints: []string{"http"},
	Service:     "test.traefik.io",
	Rule:        "Host(`test.traefik.io`)",
}

// ScanPages simulates a call to ScanPages (see https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#DynamoDB.ScanPages)
// by running the fn function twice and returning an item each time.
func (m *mockDynamoDBClient) ScanPages(input *dynamodb.ScanInput, fn func(*dynamodb.ScanOutput, bool) bool) error {
	if m.testWithError {
		return errors.New("fake error")
	}
	attributeService, err := dynamodbattribute.Marshal(service)
	if err != nil {
		return err
	}

	attributeRouter, err := dynamodbattribute.Marshal(router)
	if err != nil {
		return err
	}

	fn(&dynamodb.ScanOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				"id": &dynamodb.AttributeValue{
					S: aws.String("test.traefik.io_service"),
				},
				"name": &dynamodb.AttributeValue{
					S: aws.String("test.traefik.io"),
				},
				"service": attributeService,
				"router":  attributeRouter,
			},
		},
	}, true)

	return nil
}

func TestBuildConfigurationSuccessful(t *testing.T) {
	dbiface := &dynamoClient{
		db: &mockDynamoDBClient{
			testWithError: false,
		},
	}
	provider := Provider{}
	loadedConfig, err := provider.buildConfiguration(context.Background(), dbiface)
	if err != nil {
		t.Fatal(err)
	}
	expectedConfig := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Services: map[string]*dynamic.Service{
				"test.traefik.io": service,
			},
			Routers: map[string]*dynamic.Router{
				"test.traefik.io": router,
			},
			Middlewares:       map[string]*dynamic.Middleware{},
			ServersTransports: map[string]*dynamic.ServersTransport{},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:     map[string]*dynamic.TCPRouter{},
			Services:    map[string]*dynamic.TCPService{},
			Middlewares: map[string]*dynamic.TCPMiddleware{},
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  map[string]*dynamic.UDPRouter{},
			Services: map[string]*dynamic.UDPService{},
		},
	}

	assert.Equal(t, expectedConfig, loadedConfig)
}

func TestBuildConfigurationFailure(t *testing.T) {
	dbiface := &dynamoClient{
		db: &mockDynamoDBClient{
			testWithError: true,
		},
	}
	provider := Provider{}
	_, err := provider.buildConfiguration(context.Background(), dbiface)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestCreateClientSuccessful(t *testing.T) {
	logger := log.FromContext(context.Background())
	provider := Provider{
		Region: "us-east-1",
	}
	_, err := provider.createClient(logger)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateClientFailure(t *testing.T) {
	logger := log.FromContext(context.Background())
	provider := Provider{}
	_, err := provider.createClient(logger)
	if err == nil {
		t.Fatal("Expected error")
	}
}
