package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v3/cmd"
)

func TestRoutingLoader_Load(t *testing.T) {
	testCases := []struct {
		desc          string
		args          []string
		configContent string
		expectWarning bool
		env           map[string]string
	}{
		{
			desc: "should warn when http config found in install config",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"

[http]
  [http.routers]
    [http.routers.test]
      rule = "Host(` + "`test.example.com`" + `)"
`,
			expectWarning: true,
		},
		{
			desc: "should warn when tcp config found in install config",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"

[tcp]
  [tcp.routers]
    [tcp.routers.test]
      rule = "HostSNI(` + "`test.example.com`" + `)"
`,
			expectWarning: true,
		},
		{
			desc: "should warn when udp config found in install config",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"

[udp]
  [udp.routers]
    [udp.routers.test]
      entrypoints = ["web"]
`,
			expectWarning: true,
		},
		{
			desc: "should warn when tls config found in install config",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"

[tls]
  [[tls.certificates]]
    certFile = "path/to/cert.crt"
    keyFile = "path/to/cert.key"
`,
			expectWarning: true,
		},
		{
			desc: "should warn when multiple routing configs found in install config",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"

[http]
  [http.routers]
    [http.routers.test]
      rule = "Host(` + "`test.example.com`" + `)"

[tcp]
  [tcp.routers]
    [tcp.routers.test]
      rule = "HostSNI(` + "`test.example.com`" + `)"
`,
			expectWarning: true,
		},
		{
			desc: "should not warn when only install config is present",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"

[providers]
  [providers.file]
    filename = "routing.toml"

[api]
  dashboard = true
`,
			expectWarning: false,
		},
		{
			desc: "should not warn for self-reference scenario",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"

[providers]
  [providers.file]
    filename = "traefik.toml"

[http]
  [http.routers]
    [http.routers.test]
      rule = "Host(` + "`test.example.com`" + `)"
`,
			expectWarning: false,
		},
		{
			desc: "should not warn with empty configuration",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"
`,
			expectWarning: false,
		},
		{
			desc: "should handle environment variable configuration",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"
`,
			env: map[string]string{
				"TRAEFIK_HTTP_ROUTERS_TEST_RULE": "Host(`test.example.com`)",
			},
			expectWarning: true,
		},
		{
			desc: "should not warn for env config with self-reference",
			args: []string{"traefik"},
			configContent: `
[entrypoints]
  [entrypoints.web]
    address = ":80"
`,
			env: map[string]string{
				"TRAEFIK_PROVIDERS_FILE_FILENAME": "traefik.toml",
				"TRAEFIK_HTTP_ROUTERS_TEST_RULE":  "Host(`test.example.com`)",
			},
			expectWarning: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Set up environment variables
			for name, val := range test.env {
				t.Setenv(name, val)
			}

			// Create temporary config file
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "traefik.toml")

			// Use the test config content directly
			configContent := test.configContent

			if configContent != "" {
				err := os.WriteFile(configFile, []byte(configContent), 0o644)
				require.NoError(t, err)
			}

			// Set up command configuration
			tconfig := cmd.NewTraefikConfiguration()
			c := &cli.Command{Configuration: tconfig}
			loader := RoutingConfigLoader{}

			// Add config file flag if content was provided
			args := test.args
			if configContent != "" {
				args = append(args, "--configfile="+configFile)
			}

			// Execute the loader
			loaded, err := loader.Load(args, c)
			require.NoError(t, err)
			assert.False(t, loaded, "RoutingConfigLoader should never block startup")
		})
	}
}

func TestRoutingConfiguration_CheckRoutingElements(t *testing.T) {
	testCases := []struct {
		desc          string
		config        *routingConfiguration
		expectLogging bool
	}{
		{
			desc:          "should not log for nil config",
			config:        nil,
			expectLogging: false,
		},
		{
			desc:          "should not log for empty config",
			config:        &routingConfiguration{},
			expectLogging: false,
		},
		{
			desc: "should log when HTTP config is present",
			config: &routingConfiguration{
				HTTP: map[string]interface{}{"routers": map[string]interface{}{"test": "value"}},
			},
			expectLogging: true,
		},
		{
			desc: "should log when TCP config is present",
			config: &routingConfiguration{
				TCP: map[string]interface{}{"routers": map[string]interface{}{"test": "value"}},
			},
			expectLogging: true,
		},
		{
			desc: "should log when UDP config is present",
			config: &routingConfiguration{
				UDP: map[string]interface{}{"routers": map[string]interface{}{"test": "value"}},
			},
			expectLogging: true,
		},
		{
			desc: "should log when TLS config is present",
			config: &routingConfiguration{
				TLS: map[string]interface{}{"certificates": []interface{}{"cert1"}},
			},
			expectLogging: true,
		},
		{
			desc: "should log when multiple configs are present",
			config: &routingConfiguration{
				HTTP: map[string]interface{}{"routers": map[string]interface{}{"test": "value"}},
				TCP:  map[string]interface{}{"routers": map[string]interface{}{"test": "value"}},
			},
			expectLogging: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := zerolog.New(&buf)

			if test.config != nil {
				test.config.checkRoutingElements(logger)
			}

			logOutput := buf.String()
			if test.expectLogging {
				assert.NotEmpty(t, logOutput)
			} else {
				assert.Empty(t, logOutput)
			}
		})
	}
}

func TestIsSelfReference(t *testing.T) {
	testCases := []struct {
		desc       string
		config     *routingConfiguration
		configFile string
		expected   bool
	}{
		{
			desc: "should return true when explicit config file matches providers filename",
			config: &routingConfiguration{
				Providers: &providersConfig{
					File: &fileProviderConfig{
						Filename: "traefik.toml",
					},
				},
			},
			configFile: "traefik.toml",
			expected:   true,
		},
		{
			desc: "should return false when explicit config file differs from providers filename",
			config: &routingConfiguration{
				Providers: &providersConfig{
					File: &fileProviderConfig{
						Filename: "routing.toml",
					},
				},
			},
			configFile: "traefik.toml",
			expected:   false,
		},
		{
			desc: "should return false when providers.file.filename is empty",
			config: &routingConfiguration{
				Providers: &providersConfig{
					File: &fileProviderConfig{
						Filename: "",
					},
				},
			},
			configFile: "traefik.toml",
			expected:   false,
		},
		{
			desc: "should return false when file provider config is nil",
			config: &routingConfiguration{
				Providers: &providersConfig{
					File: nil,
				},
			},
			configFile: "traefik.toml",
			expected:   false,
		},
		{
			desc: "should return false when providers config is nil",
			config: &routingConfiguration{
				Providers: nil,
			},
			configFile: "traefik.toml",
			expected:   false,
		},
		{
			desc:       "should return false when config is nil",
			config:     nil,
			configFile: "traefik.toml",
			expected:   false,
		},
		{
			desc:       "should return false when config is empty",
			config:     &routingConfiguration{},
			configFile: "traefik.toml",
			expected:   false,
		},
		{
			desc: "should handle relative path comparison correctly",
			config: &routingConfiguration{
				Providers: &providersConfig{
					File: &fileProviderConfig{
						Filename: "./traefik.toml",
					},
				},
			},
			configFile: "traefik.toml",
			expected:   true,
		},
		{
			desc: "should return false when no config file specified",
			config: &routingConfiguration{
				Providers: &providersConfig{
					File: &fileProviderConfig{
						Filename: "some-file.toml",
					},
				},
			},
			configFile: "",
			expected:   false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, isSelfReference(test.config, test.configFile))
		})
	}
}
