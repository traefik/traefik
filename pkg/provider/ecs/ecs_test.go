package ecs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkIDs(t *testing.T) {
	testCases := []struct {
		desc     string
		count    int
		expected []int
	}{
		{
			desc:     "0 element",
			count:    0,
			expected: []int(nil),
		},
		{
			desc:     "1 element",
			count:    1,
			expected: []int{1},
		},
		{
			desc:     "99 elements, 1 chunk",
			count:    99,
			expected: []int{99},
		},
		{
			desc:     "100 elements, 1 chunk",
			count:    100,
			expected: []int{100},
		},
		{
			desc:     "101 elements, 2 chunks",
			count:    101,
			expected: []int{100, 1},
		},
		{
			desc:     "199 elements, 2 chunks",
			count:    199,
			expected: []int{100, 99},
		},
		{
			desc:     "200 elements, 2 chunks",
			count:    200,
			expected: []int{100, 100},
		},
		{
			desc:     "201 elements, 3 chunks",
			count:    201,
			expected: []int{100, 100, 1},
		},
		{
			desc:     "555 elements, 5 chunks",
			count:    555,
			expected: []int{100, 100, 100, 100, 100, 55},
		},
		{
			desc:     "1001 elements, 11 chunks",
			count:    1001,
			expected: []int{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 1},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var IDs []string
			for range test.count {
				IDs = append(IDs, "a")
			}

			var outCount []int
			for el := range chunkIDs(IDs) {
				outCount = append(outCount, len(el))
			}

			assert.Equal(t, test.expected, outCount)
		})
	}
}

func TestLookupTaskDefinitionsSharesDefinitionAcrossTasks(t *testing.T) {
	const definitionARN = "arn:aws:ecs:us-east-1:123456789012:task-definition/web:1"
	definition := &ecstypes.TaskDefinition{TaskDefinitionArn: aws.String(definitionARN)}
	var calls []string
	client := newTaskDefinitionTestClient(t, &calls)

	definitions, err := (&Provider{}).lookupTaskDefinitions(t.Context(), client, map[string]ecstypes.Task{
		"task-a": {TaskDefinitionArn: aws.String(definitionARN)},
		"task-b": {TaskDefinitionArn: aws.String(definitionARN)},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{definitionARN}, calls)
	assert.Equal(t, map[string]*ecstypes.TaskDefinition{"task-a": definition, "task-b": definition}, definitions)
}

func TestLookupTaskDefinitionsReusesDefinitionAcrossRefresh(t *testing.T) {
	const definitionARN = "arn:aws:ecs:us-east-1:123456789012:task-definition/web:1"
	definition := &ecstypes.TaskDefinition{TaskDefinitionArn: aws.String(definitionARN)}
	var calls []string
	client := newTaskDefinitionTestClient(t, &calls)
	p := &Provider{}

	definitions, err := p.lookupTaskDefinitions(t.Context(), client, map[string]ecstypes.Task{
		"task-old": {TaskDefinitionArn: aws.String(definitionARN)},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{definitionARN}, calls)
	assert.Equal(t, map[string]*ecstypes.TaskDefinition{"task-old": definition}, definitions)

	definitions, err = p.lookupTaskDefinitions(t.Context(), client, map[string]ecstypes.Task{
		"task-replacement": {TaskDefinitionArn: aws.String(definitionARN)},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{definitionARN}, calls)
	assert.Equal(t, map[string]*ecstypes.TaskDefinition{"task-replacement": definition}, definitions)
}

func TestLookupTaskDefinitionsSeparatesDifferentDefinitions(t *testing.T) {
	const definitionARN = "arn:aws:ecs:us-east-1:123456789012:task-definition/web:1"
	const otherDefinitionARN = "arn:aws:ecs:us-east-1:123456789012:task-definition/web:2"
	var calls []string
	client := newTaskDefinitionTestClient(t, &calls)

	definitions, err := (&Provider{}).lookupTaskDefinitions(t.Context(), client, map[string]ecstypes.Task{
		"task-a": {TaskDefinitionArn: aws.String(definitionARN)},
		"task-b": {TaskDefinitionArn: aws.String(otherDefinitionARN)},
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{definitionARN, otherDefinitionARN}, calls)
	assert.Equal(t, map[string]*ecstypes.TaskDefinition{
		"task-a": {TaskDefinitionArn: aws.String(definitionARN)},
		"task-b": {TaskDefinitionArn: aws.String(otherDefinitionARN)},
	}, definitions)
}

func newTaskDefinitionTestClient(t *testing.T, calls *[]string) *awsClient {
	t.Helper()

	// Tests replace the global provider cache, so they must remain serial.
	previousCache := existingTaskDefCache
	existingTaskDefCache = cache.New(cache.NoExpiration, 0)
	t.Cleanup(func() { existingTaskDefCache = previousCache })

	return &awsClient{ecs: ecs.New(ecs.Options{
		Region:      "us-east-1",
		Credentials: aws.AnonymousCredentials{},
		HTTPClient: smithyhttp.ClientDoFunc(func(req *http.Request) (*http.Response, error) {
			var input struct {
				TaskDefinition string `json:"taskDefinition"`
			}
			require.NoError(t, json.NewDecoder(req.Body).Decode(&input))
			*calls = append(*calls, input.TaskDefinition)

			body, err := json.Marshal(map[string]any{
				"taskDefinition": map[string]string{"taskDefinitionArn": input.TaskDefinition},
			})
			require.NoError(t, err)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		}),
	})}
}
