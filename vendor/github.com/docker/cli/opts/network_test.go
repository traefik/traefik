package opts

import (
	"testing"

	"github.com/docker/docker/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNetworkOptLegacySyntax(t *testing.T) {
	testCases := []struct {
		value    string
		expected []NetworkAttachmentOpts
	}{
		{
			value: "docknet1",
			expected: []NetworkAttachmentOpts{
				{
					Target: "docknet1",
				},
			},
		},
	}
	for _, tc := range testCases {
		var network NetworkOpt
		assert.NoError(t, network.Set(tc.value))
		assert.Equal(t, tc.expected, network.Value())
	}
}

func TestNetworkOptCompleteSyntax(t *testing.T) {
	testCases := []struct {
		value    string
		expected []NetworkAttachmentOpts
	}{
		{
			value: "name=docknet1,alias=web,driver-opt=field1=value1",
			expected: []NetworkAttachmentOpts{
				{
					Target:  "docknet1",
					Aliases: []string{"web"},
					DriverOpts: map[string]string{
						"field1": "value1",
					},
				},
			},
		},
		{
			value: "name=docknet1,alias=web1,alias=web2,driver-opt=field1=value1,driver-opt=field2=value2",
			expected: []NetworkAttachmentOpts{
				{
					Target:  "docknet1",
					Aliases: []string{"web1", "web2"},
					DriverOpts: map[string]string{
						"field1": "value1",
						"field2": "value2",
					},
				},
			},
		},
		{
			value: "name=docknet1",
			expected: []NetworkAttachmentOpts{
				{
					Target:  "docknet1",
					Aliases: []string{},
				},
			},
		},
	}
	for _, tc := range testCases {
		var network NetworkOpt
		assert.NoError(t, network.Set(tc.value))
		assert.Equal(t, tc.expected, network.Value())
	}
}

func TestNetworkOptInvalidSyntax(t *testing.T) {
	testCases := []struct {
		value         string
		expectedError string
	}{
		{
			value:         "invalidField=docknet1",
			expectedError: "invalid field",
		},
		{
			value:         "network=docknet1,invalid=web",
			expectedError: "invalid field",
		},
		{
			value:         "driver-opt=field1=value1,driver-opt=field2=value2",
			expectedError: "network name/id is not specified",
		},
	}
	for _, tc := range testCases {
		var network NetworkOpt
		testutil.ErrorContains(t, network.Set(tc.value), tc.expectedError)
	}
}
