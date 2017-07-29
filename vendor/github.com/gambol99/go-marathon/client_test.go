/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := Config{
		URL: "http://marathon",
	}
	cl, err := NewClient(config)

	if !assert.Nil(t, err) {
		return
	}

	conf := cl.(*marathonClient).config

	assert.Equal(t, conf.HTTPClient, http.DefaultClient)
	assert.Equal(t, conf.PollingWaitTime, defaultPollingWaitTime)
}

func TestInvalidConfig(t *testing.T) {
	config := Config{
		URL: "",
	}
	_, err := NewClient(config)
	assert.Error(t, err)
}

func TestPing(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	pong, err := endpoint.Client.Ping()
	assert.NoError(t, err)
	assert.True(t, pong)
}

func TestGetMarathonURL(t *testing.T) {
	endpoint := newFakeMarathonEndpoint(t, nil)
	defer endpoint.Close()

	assert.Equal(t, endpoint.Client.GetMarathonURL(), endpoint.URL)
}

func TestAPIRequest(t *testing.T) {
	cases := []struct {
		Username       string
		Password       string
		ServerUsername string
		ServerPassword string
		Ok             bool
	}{
		{
			Username:       "should_pass",
			Password:       "",
			ServerUsername: "",
			ServerPassword: "",
			Ok:             true,
		},
		{
			Username:       "bad_username",
			Password:       "",
			ServerUsername: "test",
			ServerPassword: "password",
			Ok:             false,
		},
		{
			Username:       "test",
			Password:       "bad_password",
			ServerUsername: "test",
			ServerPassword: "password",
			Ok:             false,
		},
		{
			Username:       "",
			Password:       "",
			ServerUsername: "test",
			ServerPassword: "password",
			Ok:             false,
		},
		{
			Username:       "test",
			Password:       "password",
			ServerUsername: "test",
			ServerPassword: "password",
			Ok:             true,
		},
	}
	for i, x := range cases {
		var endpoint *endpoint

		config := NewDefaultConfig()
		config.HTTPBasicAuthUser = x.Username
		config.HTTPBasicPassword = x.Password

		endpoint = newFakeMarathonEndpoint(t, &configContainer{
			client: &config,
			server: &serverConfig{
				username: x.ServerUsername,
				password: x.ServerPassword,
			},
		})

		_, err := endpoint.Client.Applications(nil)

		if x.Ok && err != nil {
			t.Errorf("case %d, did not expect an error: %s", i, err)
		}
		if !x.Ok && err == nil {
			t.Errorf("case %d, expected to received an error", i)
		}

		endpoint.Close()
	}
}

func TestBuildApiRequestFailure(t *testing.T) {
	tests := []struct {
		name              string
		expectedError     error
		expectedErrorType interface{}
		path              string
		clusterDown       bool
	}{
		{
			name:          "cluster down",
			expectedError: ErrMarathonDown,
			clusterDown:   true,
		},
		{
			name:              "invalid request parameter",
			expectedErrorType: newRequestError{},
			path:              "%zzzzz",
		},
	}

	for _, test := range tests {
		if test.expectedError == nil && test.expectedErrorType == nil {
			panic("Testcase requires at least one of 'expectedError' or 'expectedErrorType'")
		}

		clientCfg := NewDefaultConfig()
		config := configContainer{client: &clientCfg}
		endpoint := newFakeMarathonEndpoint(t, &config)

		client := endpoint.Client.(*marathonClient)

		if test.clusterDown {
			for _, member := range client.hosts.members {
				member.status = memberStatusDown
			}
		}

		_, _, err := client.buildAPIRequest("GET", test.path, nil)

		if test.expectedError != nil {
			assert.Equal(t, test.expectedError, err)
		}
		if test.expectedErrorType != nil {
			assert.IsType(t, test.expectedErrorType, err)
		}

		endpoint.Close()
	}
}

func TestOneLogLine(t *testing.T) {
	in := `
	a
	b    c
	d\n
	  efgh
	i\r\n
	j\t
	{"json":  "works",
		"f o o": "ba    r"
	}
	`
	assert.Equal(t, `a\n b    c\n d\n\n efgh\n i\r\n\n j\t\n {"json":  "works",\n "f o o": "ba    r"\n }\n `, string(oneLogLine([]byte(in))))
}

func TestAPIRequestDCOS(t *testing.T) {
	cases := []struct {
		DCOSToken       string
		ServerDCOSToken string
		ServerUsername  string
		ServerPassword  string
		Ok              bool
	}{
		{
			DCOSToken:       "should_pass",
			ServerDCOSToken: "should_pass",
			ServerUsername:  "",
			ServerPassword:  "",
			Ok:              true,
		},
		{
			DCOSToken:       "should_pass",
			ServerDCOSToken: "",
			ServerUsername:  "",
			ServerPassword:  "",
			Ok:              true,
		},
		{
			DCOSToken:       "should_not_pass",
			ServerDCOSToken: "different_token",
			ServerUsername:  "",
			ServerPassword:  "",
			Ok:              false,
		},
	}
	for i, x := range cases {
		var endpoint *endpoint

		config := NewDefaultConfig()
		config.DCOSToken = x.DCOSToken

		endpoint = newFakeMarathonEndpoint(t, &configContainer{
			client: &config,
			server: &serverConfig{
				dcosToken: x.ServerDCOSToken,
				username:  x.ServerUsername,
				password:  x.ServerPassword,
			},
		})

		_, err := endpoint.Client.Applications(nil)

		if x.Ok && err != nil {
			t.Errorf("case %d, did not expect an error: %s", i, err)
		}
		if !x.Ok && err == nil {
			t.Errorf("case %d, expected to received an error", i)
		}

		endpoint.Close()
	}
}
