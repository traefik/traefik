package nomad

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
			p := Provider{ExposedByDefault: test.ExposedByDefault, Prefix: test.Prefix}
			result := p.globalConfig(test.Tags)
			require.Equal(t, test.exp, result)
		})
	}
}

func Test_getNomadServiceData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.RequestURI, "/v1/services"):
			_, _ = w.Write([]byte(services))
		case strings.HasSuffix(r.RequestURI, "/v1/service/redis"):
			_, _ = w.Write([]byte(redis))
		case strings.HasSuffix(r.RequestURI, "/v1/service/hello-nomad"):
			_, _ = w.Write([]byte(hello))
		}
	}))
	t.Cleanup(ts.Close)

	p := new(Provider)
	p.SetDefaults()
	p.Endpoint.Address = ts.URL
	err := p.Init()
	require.NoError(t, err)

	// fudge client, avoid starting up via Provide
	p.client, err = createClient(p.Namespace, p.Endpoint)
	require.NoError(t, err)

	// make the query for services
	items, err := p.getNomadServiceData(context.TODO())
	require.NoError(t, err)
	require.Len(t, items, 2)
}

const services = `
[
  {
    "Namespace": "default",
    "Services": [
      {
        "ServiceName": "redis",
        "Tags": [
          "traefik.enable=true"
        ]
      },
      {
        "ServiceName": "hello-nomad",
        "Tags": [
          "traefik.enable=true",
          "traefik.http.routers.hellon.entrypoints=web",
          "traefik.http.routers.hellon.service=hello-nomad"
        ]
      }
    ]
  }
]
`

const redis = `
[
  {
    "Address": "127.0.0.1",
    "AllocID": "07501480-8175-8071-7da6-133bd1ff890f",
    "CreateIndex": 46,
    "Datacenter": "dc1",
    "ID": "_nomad-task-07501480-8175-8071-7da6-133bd1ff890f-group-redis-redis-redis",
    "JobID": "echo",
    "ModifyIndex": 46,
    "Namespace": "default",
    "NodeID": "6d7f412e-e7ff-2e66-d47b-867b0e9d8726",
    "Port": 30826,
    "ServiceName": "redis",
    "Tags": [
      "traefik.enable=true"
    ]
  }
]
`

const hello = `
[
  {
    "Address": "127.0.0.1",
    "AllocID": "71a63a80-a98a-93ee-4fd7-73b808577c20",
    "CreateIndex": 18,
    "Datacenter": "dc1",
    "ID": "_nomad-task-71a63a80-a98a-93ee-4fd7-73b808577c20-group-hello-nomad-hello-nomad-http",
    "JobID": "echo",
    "ModifyIndex": 18,
    "Namespace": "default",
    "NodeID": "6d7f412e-e7ff-2e66-d47b-867b0e9d8726",
    "Port": 20627,
    "ServiceName": "hello-nomad",
    "Tags": [
      "traefik.enable=true",
      "traefik.http.routers.hellon.entrypoints=web",
      "traefik.http.routers.hellon.service=hello-nomad"
    ]
  }
]
`
