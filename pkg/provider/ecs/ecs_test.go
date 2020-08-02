package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
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
