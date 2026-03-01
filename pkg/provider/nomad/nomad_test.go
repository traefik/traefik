package nomad

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
)

var responses = map[string][]byte{}

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	m.Run()
}

func Test_globalConfig(t *testing.T) {
	cases := []struct {
		Name             string
		Prefix           string
		Tags             []string
		ExposedByDefault bool
		exp              configuration
	}{
		{
			Name:             "expose_by_default_no_tags",
			Prefix:           "traefik",
			Tags:             nil,
			ExposedByDefault: true,
			exp:              configuration{Enable: true},
		},
		{
			Name:             "not_expose_by_default_no_tags",
			Prefix:           "traefik",
			Tags:             nil,
			ExposedByDefault: false,
			exp:              configuration{Enable: false},
		},
		{
			Name:             "expose_by_default_tags_enable",
			Prefix:           "traefik",
			Tags:             []string{"traefik.enable=true"},
			ExposedByDefault: true,
			exp:              configuration{Enable: true},
		},
		{
			Name:             "expose_by_default_tags_disable",
			Prefix:           "traefik",
			Tags:             []string{"traefik.enable=false"},
			ExposedByDefault: true,
			exp:              configuration{Enable: false},
		},
		{
			Name:             "expose_by_default_tags_enable_custom_prefix",
			Prefix:           "custom",
			Tags:             []string{"custom.enable=true"},
			ExposedByDefault: true,
			exp:              configuration{Enable: true},
		},
		{
			Name:             "expose_by_default_tags_disable_custom_prefix",
			Prefix:           "custom",
			Tags:             []string{"custom.enable=false"},
			ExposedByDefault: true,
			exp:              configuration{Enable: false},
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			p := Provider{
				Configuration: Configuration{
					ExposedByDefault: test.ExposedByDefault,
					Prefix:           test.Prefix,
				},
			}
			result := p.getExtraConf(test.Tags)
			require.Equal(t, test.exp, result)
		})
	}
}

func TestProvider_SetDefaults_Endpoint(t *testing.T) {
	testCases := []struct {
		desc     string
		envs     map[string]string
		expected *EndpointConfig
	}{
		{
			desc: "without env vars",
			envs: map[string]string{},
			expected: &EndpointConfig{
				Address: "http://127.0.0.1:4646",
			},
		},
		{
			desc: "with env vars",
			envs: map[string]string{
				"NOMAD_ADDR":        "https://nomad.example.com",
				"NOMAD_REGION":      "us-west",
				"NOMAD_TOKEN":       "almighty_token",
				"NOMAD_CACERT":      "/etc/ssl/private/nomad-agent-ca.pem",
				"NOMAD_CLIENT_CERT": "/etc/ssl/private/global-client-nomad.pem",
				"NOMAD_CLIENT_KEY":  "/etc/ssl/private/global-client-nomad-key.pem",
				"NOMAD_SKIP_VERIFY": "true",
			},
			expected: &EndpointConfig{
				Address: "https://nomad.example.com",
				Region:  "us-west",
				Token:   "almighty_token",
				TLS: &types.ClientTLS{
					CA:                 "/etc/ssl/private/nomad-agent-ca.pem",
					Cert:               "/etc/ssl/private/global-client-nomad.pem",
					Key:                "/etc/ssl/private/global-client-nomad-key.pem",
					InsecureSkipVerify: true,
				},
				EndpointWaitTime: 0,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			for k, v := range test.envs {
				t.Setenv(k, v)
			}

			p := &Provider{}
			p.SetDefaults()

			assert.Equal(t, test.expected, p.Endpoint)
		})
	}
}

func Test_getNomadServiceDataWithEmptyServices_GroupService_Scaling1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job1"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job1"):
			_, _ = w.Write(responses["job_job1_WithGroupService_Scaling1"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job1"):
			_, _ = w.Write(responses["service_job1"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func Test_getNomadServiceDataWithEmptyServices_GroupService_Scaling0(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job2"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job2"):
			_, _ = w.Write(responses["job_job2_WithGroupService_Scaling0"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job2"):
			_, _ = w.Write(responses["service_job2"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func Test_getNomadServiceDataWithEmptyServices_GroupService_ScalingDisabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job3"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job3"):
			_, _ = w.Write(responses["job_job3_WithGroupService_ScalingDisabled"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job3"):
			_, _ = w.Write(responses["service_job3"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func Test_getNomadServiceDataWithEmptyServices_GroupService_ScalingDisabled_Stopped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job4"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job4"):
			_, _ = w.Write(responses["job_job4_WithGroupService_ScalingDisabled_Stopped"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job4"):
			_, _ = w.Write(responses["service_job4"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)

	// Should not be listed as job is stopped
	require.Empty(t, items)
}

func Test_getNomadServiceDataWithEmptyServices_GroupTaskService_Scaling1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job5"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job5"):
			_, _ = w.Write(responses["job_job5_WithGroupTaskService_Scaling1"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job5task1"):
			_, _ = w.Write(responses["service_job5task1"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job5task2"):
			_, _ = w.Write(responses["service_job5task2"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 2)
}

func Test_getNomadServiceDataWithEmptyServices_GroupTaskService_Scaling0(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job6"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job6"):
			_, _ = w.Write(responses["job_job6_WithGroupTaskService_Scaling0"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job6task1"):
			_, _ = w.Write(responses["service_job6task1"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job6task2"):
			_, _ = w.Write(responses["service_job6task2"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 2)
}

func Test_getNomadServiceDataWithEmptyServices_TCP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job7"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job7"):
			_, _ = w.Write(responses["job_job7_TCP"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job7"):
			_, _ = w.Write(responses["service_job7"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func Test_getNomadServiceDataWithEmptyServices_UDP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job8"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job8"):
			_, _ = w.Write(responses["job_job8_UDP"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job8"):
			_, _ = w.Write(responses["service_job8"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func Test_getNomadServiceDataWithEmptyServices_ScalingEnabled_Stopped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/jobs"):
			_, _ = w.Write(responses["jobs_job9"])
		case strings.HasSuffix(r.RequestURI, "/v1/job/job9"):
			_, _ = w.Write(responses["job_job9_ScalingEnabled_Stopped"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/job9"):
			_, _ = w.Write(responses["service_job9"])
		}
	}))

	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceDataWithEmptyServices(t.Context())
	require.NoError(t, err)

	// Should not be listed as job is stopped
	require.Empty(t, items)
}

func setup() error {
	responsesDir := "./fixtures"
	files, err := os.ReadDir(responsesDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			content, err := os.ReadFile(filepath.Join(responsesDir, file.Name()))
			if err != nil {
				return err
			}
			responses[strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name()))] = content
		}
	}
	return nil
}

func Test_getNomadServiceData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/services"):
			_, _ = w.Write(responses["services"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/redis"):
			_, _ = w.Write(responses["service_redis"])
		case strings.HasSuffix(r.RequestURI, "/v1/service/hello-nomad"):
			_, _ = w.Write(responses["service_hello"])
		}
	}))
	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceData(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 2)
}
