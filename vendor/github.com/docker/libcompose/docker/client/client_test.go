package client

import (
	"fmt"
	"os"
	"strings"
	"testing"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/go-connections/tlsconfig"
)

// TestCreateWithEnv creates client(s) using environment variables, using an empty Options.
func TestCreateWithEnv(t *testing.T) {
	cases := []struct {
		envs            map[string]string
		expectedError   string
		expectedVersion string
	}{
		{
			envs:            map[string]string{},
			expectedVersion: "v1.20",
		},
		{
			envs: map[string]string{
				"DOCKER_CERT_PATH": "invalid/path",
			},
			expectedError: "Could not load X509 key pair: open invalid/path/cert.pem: no such file or directory",
		},
		{
			envs: map[string]string{
				"DOCKER_HOST": "host",
			},
			expectedError: "unable to parse docker host",
		},
		{
			envs: map[string]string{
				"DOCKER_HOST": "invalid://url",
			},
			expectedVersion: "v1.20",
		},
		{
			envs: map[string]string{
				"DOCKER_API_VERSION": "anything",
			},
			expectedVersion: "anything",
		},
		{
			envs: map[string]string{
				"DOCKER_API_VERSION": "1.22",
			},
			expectedVersion: "1.22",
		},
	}
	for _, c := range cases {
		recoverEnvs := setupEnvs(t, c.envs)
		apiclient, err := Create(Options{})
		if c.expectedError != "" {
			if err == nil || !strings.Contains(err.Error(), c.expectedError) {
				t.Errorf("expected an error '%s', got '%s', for %v", c.expectedError, err.Error(), c)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			version := apiclient.ClientVersion()
			if version != c.expectedVersion {
				t.Errorf("expected %s, got %s, for %v", c.expectedVersion, version, c)
			}
		}
		recoverEnvs(t)
	}
}

func TestCreateWithOptions(t *testing.T) {
	cases := []struct {
		options         Options
		expectedError   string
		expectedVersion string
	}{
		{
			options: Options{
				Host: "host",
			},
			expectedError: "unable to parse docker host",
		},
		{
			options: Options{
				Host: "invalid://host",
			},
			expectedVersion: "v1.20",
		},
		{
			options: Options{
				Host:       "tcp://host",
				APIVersion: "anything",
			},
			expectedVersion: "anything",
		},
		{
			options: Options{
				Host:       "tcp://host",
				APIVersion: "v1.22",
			},
			expectedVersion: "v1.22",
		},
		{
			options: Options{
				Host:       "tcp://host",
				TLS:        true,
				APIVersion: "v1.22",
			},
			expectedError: fmt.Sprintf("Could not load X509 key pair: open %s/cert.pem: no such file or directory", cliconfig.Dir()),
		},
		{
			options: Options{
				Host: "tcp://host",
				TLS:  true,
				TLSOptions: tlsconfig.Options{
					CertFile: "invalid/cert/file",
					CAFile:   "invalid/ca/file",
					KeyFile:  "invalid/key/file",
				},
				TrustKey:   "invalid/trust/key",
				APIVersion: "v1.22",
			},
			expectedError: "Could not load X509 key pair: open invalid/cert/file: no such file or directory",
		},
		{
			options: Options{
				Host:      "host",
				TLSVerify: true,
				TLSOptions: tlsconfig.Options{
					CertFile: "fixtures/cert.pem",
					CAFile:   "fixtures/ca.pem",
					KeyFile:  "fixtures/key.pem",
				},
				APIVersion: "v1.22",
			},
			expectedError: "unable to parse docker host",
		},
		{
			options: Options{
				Host:      "tcp://host",
				TLSVerify: true,
				TLSOptions: tlsconfig.Options{
					CertFile: "fixtures/cert.pem",
					CAFile:   "fixtures/ca.pem",
					KeyFile:  "fixtures/key.pem",
				},
				APIVersion: "v1.22",
			},
			expectedVersion: "v1.22",
		},
	}
	for _, c := range cases {
		apiclient, err := Create(c.options)
		if c.expectedError != "" {
			if err == nil || !strings.Contains(err.Error(), c.expectedError) {
				t.Errorf("expected an error '%s', got '%s', for %v", c.expectedError, err.Error(), c)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			version := apiclient.ClientVersion()
			if version != c.expectedVersion {
				t.Errorf("expected %s, got %s, for %v", c.expectedVersion, version, c)
			}
		}
	}
}

func setupEnvs(t *testing.T, envs map[string]string) func(*testing.T) {
	oldEnvs := map[string]string{}
	for key, value := range envs {
		oldEnv := os.Getenv(key)
		oldEnvs[key] = oldEnv
		err := os.Setenv(key, value)
		if err != nil {
			t.Error(err)
		}
	}
	return func(t *testing.T) {
		for key, value := range oldEnvs {
			err := os.Setenv(key, value)
			if err != nil {
				t.Error(err)
			}
		}
	}
}
