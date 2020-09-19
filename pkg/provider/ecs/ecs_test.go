package ecs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkIDs(t *testing.T) {
	provider := &Provider{}

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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var IDs []*string
			for v := 0; v < test.count; v++ {
				IDs = append(IDs, aws.String("a"))
			}

			var outCount []int
			for _, el := range provider.chunkIDs(IDs) {
				outCount = append(outCount, len(el))
			}

			assert.Equal(t, test.expected, outCount)
		})
	}
}

func TestProvider_loadECSInstanceFromTaskMetadata(t *testing.T) {
	tests := []struct {
		desc     string
		metadata *taskMetadata
		expData  []ecsInstance
		expErr   bool
	}{
		{
			desc:     "load ecs instance from v4 task metadata",
			metadata: &v4Metadata,
			expData: []ecsInstance{
				{
					Name: "query-metadata-query-metadata",
					ID:   "d4435207c04a",
					machine: &machine{
						state:        "RUNNING",
						privateIP:    "10.0.0.108",
						healthStatus: "RUNNING",
					},
					Labels: map[string]string{
						"com.amazonaws.ecs.cluster":                 "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:cluster/default",
						"com.amazonaws.ecs.container-name":          "query-metadata",
						"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:task/default/febee046097849aba589d4435207c04a",
						"com.amazonaws.ecs.task-definition-family":  "query-metadata",
						"com.amazonaws.ecs.task-definition-version": "7",
					},
				},
			},
		},
		{
			desc:     "load ecs instance from v3 task metadata",
			metadata: &v3Metadata,
			expData: []ecsInstance{
				{
					Name: "nginx-~internal~ecs~pause",
					ID:   "f63cb662a5d3",
					machine: &machine{
						state:        "RESOURCES_PROVISIONED",
						privateIP:    "10.0.2.106",
						healthStatus: "RESOURCES_PROVISIONED",
					},
					Labels: map[string]string{
						"com.amazonaws.ecs.cluster":                 "default",
						"com.amazonaws.ecs.container-name":          "~internal~ecs~pause",
						"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
						"com.amazonaws.ecs.task-definition-family":  "nginx",
						"com.amazonaws.ecs.task-definition-version": "5",
					},
				},
				{
					Name: "nginx-nginx-curl",
					ID:   "f63cb662a5d3",
					machine: &machine{
						state:        "RUNNING",
						privateIP:    "10.0.2.106",
						healthStatus: "RUNNING",
					},
					Labels: map[string]string{
						"com.amazonaws.ecs.cluster":                 "default",
						"com.amazonaws.ecs.container-name":          "nginx-curl",
						"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
						"com.amazonaws.ecs.task-definition-family":  "nginx",
						"com.amazonaws.ecs.task-definition-version": "5",
					},
				},
			},
		},
	}

	for _, test := range tests {
		p := Provider{}
		err := p.Init()
		require.NoError(t, err)

		t.Run(test.desc, func(t *testing.T) {
			instance, err := p.loadECSInstanceFromTaskMetadata(context.TODO(), test.metadata)
			if test.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, instance, test.expData)
		})
	}
}
