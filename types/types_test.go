package types

import (
	"testing"
)

func TestServer_UnmarshalTOML(t *testing.T) {
	var testCases = []struct {
		name             string
		data             interface{}
		expectedServer   Server
		expectedErrorMsg string
	}{
		{
			name: "toml server without value (default)",
			data: map[string]interface{}{},
			expectedServer: Server{
				Weight: 1,
			},
		},
		{
			name: "toml server with URL only",
			data: map[string]interface{}{
				"url": "http://example.com",
			},
			expectedServer: Server{
				URL:    "http://example.com",
				Weight: 1,
			},
		},
		{
			name: "toml server with weight 0",
			data: map[string]interface{}{
				"weight": int64(0),
			},
			expectedServer: Server{
				Weight: 0,
			},
		},
		{
			name: "toml server with URL and weight 0",
			data: map[string]interface{}{
				"url":    "http://example.com",
				"weight": int64(0),
			},
			expectedServer: Server{
				URL:    "http://example.com",
				Weight: 0,
			},
		},
		{
			name:             "toml server with invalid type",
			data:             "xxx",
			expectedErrorMsg: "Server UnmarshalTOML want (map[string]interface{})",
		},
		{
			name: "toml server with invalid URL type",
			data: map[string]interface{}{
				"url": []string{"xxx"},
			},
			expectedErrorMsg: "toml: server url must be string",
		},
		{
			name: "toml server with invalid weight type",
			data: map[string]interface{}{
				"weight": "xxx",
			},
			expectedErrorMsg: "toml: server weight must be int64",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server := Server{}
			err := server.UnmarshalTOML(test.data)

			if err != nil && err.Error() != test.expectedErrorMsg {
				t.Fatalf("UnmarshalTOML(%#v): unexepected error message, got %q, want %s.", test.data, test.expectedErrorMsg, err)
			}

			if server != test.expectedServer {
				t.Fatalf("UnmarshalTOML(%#v): invalid server, got %+v, want %+v", test.data, server, test.expectedServer)
			}
		})
	}
}
