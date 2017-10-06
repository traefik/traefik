package integration

import (
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/containous/traefik/integration/try"
	"github.com/containous/traefik/types"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

type DynamoDBSuite struct {
	BaseSuite
}

type DynamoDBItem struct {
	ID   string `dynamodbav:"id"`
	Name string `dynamodbav:"name"`
}

type DynamoDBBackendItem struct {
	DynamoDBItem
	Backend types.Backend `dynamodbav:"backend"`
}

type DynamoDBFrontendItem struct {
	DynamoDBItem
	Frontend types.Frontend `dynamodbav:"frontend"`
}

func (s *DynamoDBSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "dynamodb")
	s.composeProject.Start(c)
	dynamoURL := "http://" + s.composeProject.Container(c, "dynamo").NetworkSettings.IPAddress + ":8000"
	config := &aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials("id", "secret", ""),
		Endpoint:    aws.String(dynamoURL),
	}
	var sess *session.Session
	err := try.Do(60*time.Second, func() error {
		_, err := session.NewSession(config)
		if err != nil {
			return err
		}
		sess = session.New(config)
		return nil
	})
	c.Assert(err, checker.IsNil)
	svc := dynamodb.New(sess)

	// create dynamodb table
	params := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: aws.String("traefik"),
	}
	_, err = svc.CreateTable(params)
	if err != nil {
		c.Error(err)
		return
	}

	// load config into dynamodb
	whoami1 := "http://" + s.composeProject.Container(c, "whoami1").NetworkSettings.IPAddress + ":80"
	whoami2 := "http://" + s.composeProject.Container(c, "whoami2").NetworkSettings.IPAddress + ":80"
	whoami3 := "http://" + s.composeProject.Container(c, "whoami3").NetworkSettings.IPAddress + ":80"

	backend := DynamoDBBackendItem{
		Backend: types.Backend{
			Servers: map[string]types.Server{
				"whoami1": {
					URL: whoami1,
				},
				"whoami2": {
					URL: whoami2,
				},
				"whoami3": {
					URL: whoami3,
				},
			},
		},
		DynamoDBItem: DynamoDBItem{
			ID:   "whoami_backend",
			Name: "whoami",
		},
	}

	frontend := DynamoDBFrontendItem{
		Frontend: types.Frontend{
			EntryPoints: []string{
				"http",
			},
			Backend: "whoami",
			Routes: map[string]types.Route{
				"hostRule": {
					Rule: "Host:test.traefik.io",
				},
			},
		},
		DynamoDBItem: DynamoDBItem{
			ID:   "whoami_frontend",
			Name: "whoami",
		},
	}
	backendAttributeValue, err := dynamodbattribute.MarshalMap(backend)
	c.Assert(err, checker.IsNil)
	frontendAttributeValue, err := dynamodbattribute.MarshalMap(frontend)
	c.Assert(err, checker.IsNil)
	putParams := &dynamodb.PutItemInput{
		Item:      backendAttributeValue,
		TableName: aws.String("traefik"),
	}
	_, err = svc.PutItem(putParams)
	c.Assert(err, checker.IsNil)

	putParams = &dynamodb.PutItemInput{
		Item:      frontendAttributeValue,
		TableName: aws.String("traefik"),
	}
	_, err = svc.PutItem(putParams)
	c.Assert(err, checker.IsNil)
}

func (s *DynamoDBSuite) TestSimpleConfiguration(c *check.C) {
	dynamoURL := "http://" + s.composeProject.Container(c, "dynamo").NetworkSettings.IPAddress + ":8000"
	file := s.adaptFile(c, "fixtures/dynamodb/simple.toml", struct{ DynamoURL string }{dynamoURL})
	defer os.Remove(file)
	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	err = try.GetRequest("http://127.0.0.1:8081/api/providers", 120*time.Second, try.BodyContains("Host:test.traefik.io"))
	c.Assert(err, checker.IsNil)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	c.Assert(err, checker.IsNil)
	req.Host = "test.traefik.io"

	err = try.Request(req, 200*time.Millisecond, try.StatusCodeIs(http.StatusOK))
	c.Assert(err, checker.IsNil)
}

func (s *DynamoDBSuite) TearDownSuite(c *check.C) {
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}
