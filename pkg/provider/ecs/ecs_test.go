package ecs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

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

func TestLookupTaskDefinitions(t *testing.T) {
	const definitionARN = "arn:aws:ecs:us-east-1:123456789012:task-definition/web:1"
	const otherDefinitionARN = "arn:aws:ecs:us-east-1:123456789012:task-definition/web:2"

	testCases := []struct {
		desc          string
		batches       []map[string]*string
		cachedARN     *string
		expired       bool
		apiError      bool
		expectedError bool
		expectedCalls map[string]int
	}{
		{
			desc:          "tasks sharing a definition",
			batches:       []map[string]*string{{"task-a": aws.String(definitionARN), "task-b": aws.String(definitionARN), "task-c": aws.String(definitionARN)}},
			expectedCalls: map[string]int{definitionARN: 1},
		},
		{
			desc:          "different definition revisions",
			batches:       []map[string]*string{{"task-a": aws.String(definitionARN), "task-b": aws.String(otherDefinitionARN)}},
			expectedCalls: map[string]int{definitionARN: 1, otherDefinitionARN: 1},
		},
		{
			desc:          "cached definition for a new task",
			batches:       []map[string]*string{{"task-new": aws.String(definitionARN)}},
			cachedARN:     aws.String(definitionARN),
			expectedCalls: map[string]int{},
		},
		{
			desc:          "replacement task on next refresh",
			batches:       []map[string]*string{{"task-old": aws.String(definitionARN)}, {"task-new": aws.String(definitionARN)}},
			expectedCalls: map[string]int{definitionARN: 1},
		},
		{
			desc:          "expired definition",
			batches:       []map[string]*string{{"task-a": aws.String(definitionARN)}},
			cachedARN:     aws.String(definitionARN),
			expired:       true,
			expectedCalls: map[string]int{definitionARN: 1},
		},
		{
			desc:          "errors are not cached",
			batches:       []map[string]*string{{"task-a": aws.String(definitionARN)}, {"task-a": aws.String(definitionARN)}},
			apiError:      true,
			expectedError: true,
			expectedCalls: map[string]int{definitionARN: 2},
		},
		{
			desc:          "missing definition ARN ignores empty cache key",
			batches:       []map[string]*string{{"task-a": nil}},
			cachedARN:     aws.String(""),
			expectedError: true,
			expectedCalls: map[string]int{},
		},
		{
			desc:          "empty definition ARN ignores empty cache key",
			batches:       []map[string]*string{{"task-a": aws.String("")}},
			cachedARN:     aws.String(""),
			apiError:      true,
			expectedError: true,
			expectedCalls: map[string]int{"": 1},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// The provider cache is global, so these subtests must remain serial.
			previousCache := existingTaskDefCache
			existingTaskDefCache = cache.New(30*time.Minute, 0)
			t.Cleanup(func() { existingTaskDefCache = previousCache })

			cachedDefinition := &ecstypes.TaskDefinition{TaskDefinitionArn: test.cachedARN}
			if test.cachedARN != nil {
				if test.expired {
					existingTaskDefCache = cache.NewFrom(30*time.Minute, 0, map[string]cache.Item{
						*test.cachedARN: {Object: cachedDefinition, Expiration: 1},
					})
				} else {
					existingTaskDefCache.Set(*test.cachedARN, cachedDefinition, cache.DefaultExpiration)
				}
			}
			initialCache := existingTaskDefCache.Items()

			calls := make(map[string]int)
			client := &awsClient{ecs: ecs.New(ecs.Options{
				Region:           "us-east-1",
				Credentials:      aws.AnonymousCredentials{},
				RetryMaxAttempts: 1,
				HTTPClient: smithyhttp.ClientDoFunc(func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "AmazonEC2ContainerServiceV20141113.DescribeTaskDefinition", req.Header.Get("X-Amz-Target"))
					var input struct {
						TaskDefinition string `json:"taskDefinition"`
					}
					require.NoError(t, json.NewDecoder(req.Body).Decode(&input))
					calls[input.TaskDefinition]++

					status := http.StatusOK
					body, err := json.Marshal(map[string]any{
						"taskDefinition": map[string]string{"taskDefinitionArn": input.TaskDefinition},
					})
					require.NoError(t, err)
					if test.apiError {
						status = http.StatusBadRequest
						body = []byte(`{"__type":"ClientException","message":"Invalid task definition"}`)
					}

					return &http.Response{
						StatusCode: status,
						Header:     http.Header{"Content-Type": []string{"application/x-amz-json-1.1"}},
						Body:       io.NopCloser(bytes.NewReader(body)),
					}, nil
				}),
			})}

			for _, batch := range test.batches {
				tasks := make(map[string]ecstypes.Task)
				for taskARN, defARN := range batch {
					tasks[taskARN] = ecstypes.Task{TaskArn: aws.String(taskARN), TaskDefinitionArn: defARN}
				}

				definitions, err := (&Provider{}).lookupTaskDefinitions(t.Context(), client, tasks)
				if test.expectedError {
					require.ErrorContains(t, err, "describing task definition")
					assert.Nil(t, definitions)
					assert.Equal(t, initialCache, existingTaskDefCache.Items())
					continue
				}
				require.NoError(t, err)
				require.Len(t, definitions, len(tasks))
				for taskARN, defARN := range batch {
					require.Contains(t, definitions, taskARN)
					require.NotNil(t, definitions[taskARN])
					assert.Equal(t, *defARN, aws.ToString(definitions[taskARN].TaskDefinitionArn))
				}
			}

			assert.Equal(t, test.expectedCalls, calls)
			if test.cachedARN != nil && !test.expired {
				assert.Equal(t, initialCache[*test.cachedARN], existingTaskDefCache.Items()[*test.cachedARN])
			}
		})
	}
}
