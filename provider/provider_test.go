package provider

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type myProvider struct {
	BaseProvider
	TLS *types.ClientTLS
}

func (p *myProvider) Foo() string {
	return "bar"
}

func TestConfigurationErrors(t *testing.T) {
	templateErrorFile, err := ioutil.TempFile("", "provider-configuration-error")
	require.NoError(t, err)

	defer os.RemoveAll(templateErrorFile.Name())

	data := []byte("Not a valid template {{ Bar }}")

	err = ioutil.WriteFile(templateErrorFile.Name(), data, 0700)
	require.NoError(t, err)

	templateInvalidTOMLFile, err := ioutil.TempFile("", "provider-configuration-error")
	require.NoError(t, err)

	defer os.RemoveAll(templateInvalidTOMLFile.Name())

	data = []byte(`Hello {{ .Name }}
{{ Foo }}`)

	err = ioutil.WriteFile(templateInvalidTOMLFile.Name(), data, 0700)
	require.NoError(t, err)

	invalids := []struct {
		provider        *myProvider
		defaultTemplate string
		expectedError   string
		funcMap         template.FuncMap
		templateObjects interface{}
	}{
		{
			provider: &myProvider{
				BaseProvider: BaseProvider{
					Filename: "/non/existent/template.tmpl",
				},
			},
			expectedError: "open /non/existent/template.tmpl: no such file or directory",
		},
		{
			provider:        &myProvider{},
			defaultTemplate: "non/existent/template.tmpl",
			expectedError:   "Asset non/existent/template.tmpl not found",
		},
		{
			provider: &myProvider{
				BaseProvider: BaseProvider{
					Filename: templateErrorFile.Name(),
				},
			},
			expectedError: `function "Bar" not defined`,
		},
		{
			provider: &myProvider{
				BaseProvider: BaseProvider{
					Filename: templateInvalidTOMLFile.Name(),
				},
			},
			expectedError: "Near line 1 (last key parsed 'Hello'): expected key separator '=', but got '<' instead",
			funcMap: template.FuncMap{
				"Foo": func() string {
					return "bar"
				},
			},
			templateObjects: struct{ Name string }{Name: "bar"},
		},
	}

	for _, invalid := range invalids {
		configuration, err := invalid.provider.GetConfiguration(invalid.defaultTemplate, invalid.funcMap, nil)
		if err == nil || !strings.Contains(err.Error(), invalid.expectedError) {
			t.Fatalf("should have generate an error with %q, got %v", invalid.expectedError, err)
		}

		assert.Nil(t, configuration)
	}
}

func TestGetConfiguration(t *testing.T) {
	templateFile, err := ioutil.TempFile("", "provider-configuration")
	require.NoError(t, err)

	defer os.RemoveAll(templateFile.Name())

	data := []byte(`[backends]
  [backends.backend1]
    [backends.backend1.circuitbreaker]
      expression = "NetworkErrorRatio() > 0.5"
    [backends.backend1.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1

[frontends]
  [frontends.frontend1]
  backend = "backend1"
  passHostHeader = true
    [frontends.frontend11.routes.test_2]
    rule = "Path"
    value = "/test"`)

	err = ioutil.WriteFile(templateFile.Name(), data, 0700)
	require.NoError(t, err)

	provider := &myProvider{
		BaseProvider: BaseProvider{
			Filename: templateFile.Name(),
		},
	}

	configuration, err := provider.GetConfiguration(templateFile.Name(), nil, nil)
	require.NoError(t, err)

	assert.NotNil(t, configuration)
}

func TestGetConfigurationReturnsCorrectMaxConnConfiguration(t *testing.T) {
	templateFile, err := ioutil.TempFile("", "provider-configuration")
	require.NoError(t, err)

	defer os.RemoveAll(templateFile.Name())

	data := []byte(`[backends]
  [backends.backend1]
    [backends.backend1.maxconn]
      amount = 10
      extractorFunc = "request.host"`)

	err = ioutil.WriteFile(templateFile.Name(), data, 0700)
	require.NoError(t, err)

	provider := &myProvider{
		BaseProvider: BaseProvider{
			Filename: templateFile.Name(),
		},
	}

	configuration, err := provider.GetConfiguration(templateFile.Name(), nil, nil)
	require.NoError(t, err)

	require.NotNil(t, configuration)
	require.Contains(t, configuration.Backends, "backend1")
	assert.EqualValues(t, 10, configuration.Backends["backend1"].MaxConn.Amount)
	assert.Equal(t, "request.host", configuration.Backends["backend1"].MaxConn.ExtractorFunc)
}

func TestNilClientTLS(t *testing.T) {
	p := &myProvider{
		BaseProvider: BaseProvider{
			Filename: "",
		},
	}

	_, err := p.TLS.CreateTLSConfig()
	require.NoError(t, err, "CreateTLSConfig should assume that consumer does not want a TLS configuration if input is nil")
}

func TestInsecureSkipVerifyClientTLS(t *testing.T) {
	p := &myProvider{
		BaseProvider: BaseProvider{
			Filename: "",
		},
		TLS: &types.ClientTLS{
			InsecureSkipVerify: true,
		},
	}

	config, err := p.TLS.CreateTLSConfig()
	require.NoError(t, err, "CreateTLSConfig should assume that consumer does not want a TLS configuration if input is nil")

	assert.True(t, config.InsecureSkipVerify, "CreateTLSConfig should support setting only InsecureSkipVerify property")
}

func TestInsecureSkipVerifyFalseClientTLS(t *testing.T) {
	p := &myProvider{
		BaseProvider: BaseProvider{
			Filename: "",
		},
		TLS: &types.ClientTLS{
			InsecureSkipVerify: false,
		},
	}

	_, err := p.TLS.CreateTLSConfig()
	assert.Errorf(t, err, "CreateTLSConfig should error if consumer does not set a TLS cert or key configuration and not chooses InsecureSkipVerify to be true")
}

func TestMatchingConstraints(t *testing.T) {
	testCases := []struct {
		desc        string
		constraints types.Constraints
		tags        []string
		expected    bool
	}{
		// simple test: must match
		{
			desc: "tag==us-east-1 with us-east-1",
			constraints: types.Constraints{
				{
					Key:       "tag",
					MustMatch: true,
					Regex:     "us-east-1",
				},
			},
			tags: []string{
				"us-east-1",
			},
			expected: true,
		},
		// simple test: must match but does not match
		{
			desc: "tag==us-east-1 with us-east-2",
			constraints: types.Constraints{
				{
					Key:       "tag",
					MustMatch: true,
					Regex:     "us-east-1",
				},
			},
			tags: []string{
				"us-east-2",
			},
			expected: false,
		},
		// simple test: must not match
		{
			desc: "tag!=us-east-1 with us-east-1",
			constraints: types.Constraints{
				{
					Key:       "tag",
					MustMatch: false,
					Regex:     "us-east-1",
				},
			},
			tags: []string{
				"us-east-1",
			},
			expected: false,
		},
		// complex test: globbing
		{
			desc: "tag!=us-east-* with us-east-1",
			constraints: types.Constraints{
				{
					Key:       "tag",
					MustMatch: true,
					Regex:     "us-east-*",
				},
			},
			tags: []string{
				"us-east-1",
			},
			expected: true,
		},
		// complex test: multiple constraints
		{
			desc: "tag==us-east-* & tag!=api with us-east-1 & api",
			constraints: types.Constraints{
				{
					Key:       "tag",
					MustMatch: true,
					Regex:     "us-east-*",
				},
				{
					Key:       "tag",
					MustMatch: false,
					Regex:     "api",
				},
			},
			tags: []string{
				"api",
				"us-east-1",
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		p := myProvider{
			BaseProvider: BaseProvider{
				Constraints: test.constraints,
			},
		}

		actual, _ := p.MatchConstraints(test.tags)
		assert.Equal(t, test.expected, actual)
	}
}

func TestDefaultFuncMap(t *testing.T) {
	templateFile, err := ioutil.TempFile("", "provider-configuration")
	require.NoError(t, err)
	defer os.RemoveAll(templateFile.Name())

	data := []byte(`
  [backends]
  [backends.{{ "backend-1" | replace  "-" "" }}]
    [backends.{{ "BACKEND1" | tolower }}.circuitbreaker]
      expression = "NetworkErrorRatio() > 0.5"
    [backends.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1

[frontends]
  [frontends.{{normalize "frontend/1"}}]
  {{ $backend := "backend1/test/value" | split  "/" }}
  {{ $backendid := index $backend 1 }}
  {{ if "backend1" | contains "backend" }}
  backend = "backend1"
  {{end}}
  passHostHeader = true
    [frontends.frontend-1.routes.test_2]
    rule = "Path"
    value = "/test"`)

	err = ioutil.WriteFile(templateFile.Name(), data, 0700)
	require.NoError(t, err)

	provider := &myProvider{
		BaseProvider: BaseProvider{
			Filename: templateFile.Name(),
		},
	}

	configuration, err := provider.GetConfiguration(templateFile.Name(), nil, nil)
	require.NoError(t, err)

	require.NotNil(t, configuration)
	assert.Contains(t, configuration.Backends, "backend1")
	assert.Contains(t, configuration.Frontends, "frontend-1")
}

func TestSprigFunctions(t *testing.T) {
	templateFile, err := ioutil.TempFile("", "provider-configuration")
	require.NoError(t, err)

	defer os.RemoveAll(templateFile.Name())

	data := []byte(`
  {{$backend_name := trimAll "-" uuidv4}}
  [backends]
  [backends.{{$backend_name}}]
    [backends.{{$backend_name}}.circuitbreaker]
    [backends.{{$backend_name}}.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1

[frontends]
  [frontends.{{normalize "frontend/1"}}]
  backend = "{{$backend_name}}"
  passHostHeader = true
    [frontends.frontend-1.routes.test_2]
    rule = "Path"
    value = "/test"`)

	err = ioutil.WriteFile(templateFile.Name(), data, 0700)
	require.NoError(t, err)

	provider := &myProvider{
		BaseProvider: BaseProvider{
			Filename: templateFile.Name(),
		},
	}

	configuration, err := provider.GetConfiguration(templateFile.Name(), nil, nil)
	require.NoError(t, err)

	require.NotNil(t, configuration)
	assert.Len(t, configuration.Backends, 1)
	assert.Contains(t, configuration.Frontends, "frontend-1")
}

func TestBaseProvider_GetConfiguration(t *testing.T) {
	baseProvider := BaseProvider{}

	testCases := []struct {
		name                string
		defaultTemplateFile string
		expectedContent     string
	}{
		{
			defaultTemplateFile: "templates/docker.tmpl",
			expectedContent:     readTemplateFile(t, "./../templates/docker.tmpl"),
		},
		{
			defaultTemplateFile: `template content`,
			expectedContent:     `template content`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {

			content, err := baseProvider.getTemplateContent(test.defaultTemplateFile)
			require.NoError(t, err)

			assert.Equal(t, test.expectedContent, content)
		})
	}
}

func TestNormalize(t *testing.T) {
	testCases := []struct {
		desc     string
		name     string
		expected string
	}{
		{
			desc:     "without special chars",
			name:     "foobar",
			expected: "foobar",
		},
		{
			desc:     "with special chars",
			name:     "foo.foo.foo;foo:foo!foo/foo\\foo)foo_123-ç_àéè",
			expected: "foo-foo-foo-foo-foo-foo-foo-foo-foo-123-ç-àéè",
		},
		{
			desc:     "starts with special chars",
			name:     ".foo.foo",
			expected: "foo-foo",
		},
		{
			desc:     "ends with special chars",
			name:     "foo.foo.",
			expected: "foo-foo",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := Normalize(test.name)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func readTemplateFile(t *testing.T, path string) string {
	t.Helper()
	expectedContent, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(expectedContent)
}
