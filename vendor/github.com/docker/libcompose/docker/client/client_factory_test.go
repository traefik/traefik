package client

import (
	"strings"
	"testing"
)

func TestFactoryWithEnv(t *testing.T) {
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
			expectedError:   "Could not load X509 key pair: open invalid/path/cert.pem: no such file or directory",
			expectedVersion: "v1.20",
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
		factory, err := NewDefaultFactory(Options{})
		if c.expectedError != "" {
			if err == nil || !strings.Contains(err.Error(), c.expectedError) {
				t.Errorf("expected an error %s, got %s, for %v", c.expectedError, err.Error(), c)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			apiclient := factory.Create(nil)
			version := apiclient.ClientVersion()
			if version != c.expectedVersion {
				t.Errorf("expected %s, got %s, for %v", c.expectedVersion, version, c)
			}
		}
		recoverEnvs(t)
	}
}

func TestFactoryWithOptions(t *testing.T) {
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
				APIVersion: "v1.22",
			},
			expectedVersion: "v1.22",
		},
	}
	for _, c := range cases {
		factory, err := NewDefaultFactory(c.options)
		if c.expectedError != "" {
			if err == nil || !strings.Contains(err.Error(), c.expectedError) {
				t.Errorf("expected an error %s, got %s, for %v", c.expectedError, err.Error(), c)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			apiclient := factory.Create(nil)
			version := apiclient.ClientVersion()
			if version != c.expectedVersion {
				t.Errorf("expected %s, got %s, for %v", c.expectedVersion, version, c)
			}
		}
	}
}
