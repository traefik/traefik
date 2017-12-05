package dynamodb

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/containous/traefik/types"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	testWithError bool
}

var backend = &types.Backend{
	HealthCheck: &types.HealthCheck{
		Path: "/build",
	},
	Servers: map[string]types.Server{
		"server1": {
			URL: "http://test.traefik.io",
		},
	},
}

var frontend = &types.Frontend{
	EntryPoints: []string{"http"},
	Backend:     "test.traefik.io",
	Routes: map[string]types.Route{
		"route1": {
			Rule: "Host:test.traefik.io",
		},
	},
}

// ScanPages simulates a call to ScanPages (see https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#DynamoDB.ScanPages)
// by running the fn function twice and returning an item each time.
func (m *mockDynamoDBClient) ScanPages(input *dynamodb.ScanInput, fn func(*dynamodb.ScanOutput, bool) bool) error {
	if m.testWithError {
		return errors.New("fake error")
	}
	attributeBackend, err := dynamodbattribute.Marshal(backend)
	if err != nil {
		return err
	}

	attributeFrontend, err := dynamodbattribute.Marshal(frontend)
	if err != nil {
		return err
	}

	fn(&dynamodb.ScanOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				"id": &dynamodb.AttributeValue{
					S: aws.String("test.traefik.io_backend"),
				},
				"name": &dynamodb.AttributeValue{
					S: aws.String("test.traefik.io"),
				},
				"backend": attributeBackend,
			},
		},
	}, false)

	fn(&dynamodb.ScanOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				"id": &dynamodb.AttributeValue{
					S: aws.String("test.traefik.io_frontend"),
				},
				"name": &dynamodb.AttributeValue{
					S: aws.String("test.traefik.io"),
				},
				"frontend": attributeFrontend,
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
	loadedConfig, err := provider.buildConfiguration(dbiface)
	if err != nil {
		t.Fatal(err)
	}
	expectedConfig := &types.Configuration{
		Backends: map[string]*types.Backend{
			"test.traefik.io": backend,
		},
		Frontends: map[string]*types.Frontend{
			"test.traefik.io": frontend,
		},
	}
	if !reflect.DeepEqual(loadedConfig, expectedConfig) {
		t.Fatalf("Configurations did not match: %v %v", loadedConfig, expectedConfig)
	}
}

func TestBuildConfigurationFailure(t *testing.T) {
	dbiface := &dynamoClient{
		db: &mockDynamoDBClient{
			testWithError: true,
		},
	}
	provider := Provider{}
	_, err := provider.buildConfiguration(dbiface)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestCreateClientSuccessful(t *testing.T) {
	provider := Provider{
		Region: "us-east-1",
	}
	_, err := provider.createClient()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateClientFailure(t *testing.T) {
	provider := Provider{}
	_, err := provider.createClient()
	if err == nil {
		t.Fatal("Expected error")
	}
}
