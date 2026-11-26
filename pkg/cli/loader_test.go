package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	paersercli "github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v2/cmd"
	"github.com/traefik/traefik/v2/pkg/cli"
)

func TestFileLoader_CrossProviderNamespaces(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		content  string
		expected *[]string
	}{
		{
			desc:     "YAML unset",
			filename: "traefik.yml",
			content: `
providers:
  kubernetesCRD: {}
  kubernetesIngress: {}
  kubernetesGateway: {}
`,
			expected: nil,
		},
		{
			desc:     "YAML empty list",
			filename: "traefik.yml",
			content: `
providers:
  kubernetesCRD:
    crossProviderNamespaces: []
  kubernetesIngress:
    crossProviderNamespaces: []
  kubernetesGateway:
    crossProviderNamespaces: []
`,
			expected: &[]string{},
		},
		{
			desc:     "YAML list",
			filename: "traefik.yml",
			content: `
providers:
  kubernetesCRD:
    crossProviderNamespaces: [foo, bar]
  kubernetesIngress:
    crossProviderNamespaces: [foo, bar]
  kubernetesGateway:
    crossProviderNamespaces: [foo, bar]
`,
			expected: &[]string{"foo", "bar"},
		},
		{
			desc:     "TOML unset",
			filename: "traefik.toml",
			content: `
[providers.kubernetesCRD]
[providers.kubernetesIngress]
[providers.kubernetesGateway]
`,
			expected: nil,
		},
		{
			desc:     "TOML empty list",
			filename: "traefik.toml",
			content: `
[providers.kubernetesCRD]
  crossProviderNamespaces = []
[providers.kubernetesIngress]
  crossProviderNamespaces = []
[providers.kubernetesGateway]
  crossProviderNamespaces = []
`,
			expected: &[]string{},
		},
		{
			desc:     "TOML list",
			filename: "traefik.toml",
			content: `
[providers.kubernetesCRD]
  crossProviderNamespaces = ["foo", "bar"]
[providers.kubernetesIngress]
  crossProviderNamespaces = ["foo", "bar"]
[providers.kubernetesGateway]
  crossProviderNamespaces = ["foo", "bar"]
`,
			expected: &[]string{"foo", "bar"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), test.filename)
			require.NoError(t, os.WriteFile(path, []byte(test.content), 0o600))

			conf := cmd.NewTraefikConfiguration()
			ok, err := (&cli.FileLoader{}).Load([]string{"--configFile=" + path}, &paersercli.Command{Configuration: conf})
			require.NoError(t, err)
			require.True(t, ok)

			assertCrossProviderNamespaces(t, test.expected, conf)
		})
	}
}

func TestFlagLoader_CrossProviderNamespaces(t *testing.T) {
	testCases := []struct {
		desc     string
		args     []string
		expected *[]string
	}{
		{
			desc: "unset",
			args: []string{
				"--providers.kubernetescrd",
				"--providers.kubernetesingress",
				"--providers.kubernetesgateway",
			},
			expected: nil,
		},
		{
			desc: "empty list",
			args: []string{
				"--providers.kubernetescrd.crossprovidernamespaces=",
				"--providers.kubernetesingress.crossprovidernamespaces=",
				"--providers.kubernetesgateway.crossprovidernamespaces=",
			},
			expected: &[]string{},
		},
		{
			desc: "list",
			args: []string{
				"--providers.kubernetescrd.crossprovidernamespaces=foo,bar",
				"--providers.kubernetesingress.crossprovidernamespaces=foo,bar",
				"--providers.kubernetesgateway.crossprovidernamespaces=foo,bar",
			},
			expected: &[]string{"foo", "bar"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			conf := cmd.NewTraefikConfiguration()
			ok, err := (&cli.FlagLoader{}).Load(test.args, &paersercli.Command{Configuration: conf})
			require.NoError(t, err)
			require.True(t, ok)

			assertCrossProviderNamespaces(t, test.expected, conf)
		})
	}
}

func TestEnvLoader_CrossProviderNamespaces(t *testing.T) {
	testCases := []struct {
		desc     string
		env      map[string]string
		expected *[]string
	}{
		{
			desc: "unset",
			env: map[string]string{
				"TRAEFIK_PROVIDERS_KUBERNETESCRD":     "true",
				"TRAEFIK_PROVIDERS_KUBERNETESINGRESS": "true",
				"TRAEFIK_PROVIDERS_KUBERNETESGATEWAY": "true",
			},
			expected: nil,
		},
		{
			desc: "empty list",
			env: map[string]string{
				"TRAEFIK_PROVIDERS_KUBERNETESCRD_CROSSPROVIDERNAMESPACES":     "",
				"TRAEFIK_PROVIDERS_KUBERNETESINGRESS_CROSSPROVIDERNAMESPACES": "",
				"TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_CROSSPROVIDERNAMESPACES": "",
			},
			expected: &[]string{},
		},
		{
			desc: "list",
			env: map[string]string{
				"TRAEFIK_PROVIDERS_KUBERNETESCRD_CROSSPROVIDERNAMESPACES":     "foo,bar",
				"TRAEFIK_PROVIDERS_KUBERNETESINGRESS_CROSSPROVIDERNAMESPACES": "foo,bar",
				"TRAEFIK_PROVIDERS_KUBERNETESGATEWAY_CROSSPROVIDERNAMESPACES": "foo,bar",
			},
			expected: &[]string{"foo", "bar"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			for key, value := range test.env {
				t.Setenv(key, value)
			}

			conf := cmd.NewTraefikConfiguration()
			ok, err := (&cli.EnvLoader{}).Load(nil, &paersercli.Command{Configuration: conf})
			require.NoError(t, err)
			require.True(t, ok)

			assertCrossProviderNamespaces(t, test.expected, conf)
		})
	}
}

func assertCrossProviderNamespaces(t *testing.T, expected *[]string, conf *cmd.TraefikCmdConfiguration) {
	t.Helper()

	require.NotNil(t, conf.Providers.KubernetesCRD)
	require.NotNil(t, conf.Providers.KubernetesIngress)
	require.NotNil(t, conf.Providers.KubernetesGateway)

	for _, actual := range []*[]string{
		conf.Providers.KubernetesCRD.CrossProviderNamespaces,
		conf.Providers.KubernetesIngress.CrossProviderNamespaces,
		conf.Providers.KubernetesGateway.CrossProviderNamespaces,
	} {
		if expected == nil {
			assert.Nil(t, actual)
			continue
		}

		require.NotNil(t, actual)
		assert.ElementsMatch(t, *expected, *actual)
	}
}
