package service

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/yaml"
	shlex "github.com/flynn/go-shlex"
	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	exp := []string{"sh", "-c", "exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}"}
	cmd, err := shlex.Split("sh -c 'exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}'")
	assert.Nil(t, err)
	assert.Equal(t, exp, cmd)
}

func TestParseBindsAndVolumes(t *testing.T) {
	ctx := &ctx.Context{}
	ctx.ComposeFiles = []string{"foo/docker-compose.yml"}
	ctx.ResourceLookup = &lookup.FileResourceLookup{}

	abs, err := filepath.Abs(".")
	assert.Nil(t, err)
	cfg, hostCfg, err := Convert(&config.ServiceConfig{
		Volumes: &yaml.Volumes{
			Volumes: []*yaml.Volume{
				{
					Destination: "/foo",
				},
				{
					Source:      "/home",
					Destination: "/home",
				},
				{
					Destination: "/bar/baz",
				},
				{
					Source:      ".",
					Destination: "/home",
				},
				{
					Source:      "/usr/lib",
					Destination: "/usr/lib",
					AccessMode:  "ro",
				},
			},
		},
	}, ctx.Context, nil)
	assert.Nil(t, err)
	assert.Equal(t, map[string]struct{}{"/foo": {}, "/bar/baz": {}}, cfg.Volumes)
	assert.Equal(t, []string{"/home:/home", abs + "/foo:/home", "/usr/lib:/usr/lib:ro"}, hostCfg.Binds)
}

func TestParseLabels(t *testing.T) {
	ctx := &ctx.Context{}
	ctx.ComposeFiles = []string{"foo/docker-compose.yml"}
	ctx.ResourceLookup = &lookup.FileResourceLookup{}
	bashCmd := "bash"
	fooLabel := "foo.label"
	fooLabelValue := "service.config.value"
	sc := &config.ServiceConfig{
		Entrypoint: yaml.Command([]string{bashCmd}),
		Labels:     yaml.SliceorMap{fooLabel: "service.config.value"},
	}
	cfg, _, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	cfg.Labels[fooLabel] = "FUN"
	cfg.Entrypoint[0] = "less"

	assert.Equal(t, fooLabelValue, sc.Labels[fooLabel])
	assert.Equal(t, "FUN", cfg.Labels[fooLabel])

	assert.Equal(t, yaml.Command{bashCmd}, sc.Entrypoint)
	assert.Equal(t, []string{"less"}, []string(cfg.Entrypoint))
}

func TestDNSOpt(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		DNSOpts: []string{
			"use-vc",
			"no-tld-query",
		},
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]string{
		"use-vc",
		"no-tld-query",
	}, hostCfg.DNSOptions))
}

func TestGroupAdd(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		GroupAdd: []string{
			"root",
			"1",
		},
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual([]string{
		"root",
		"1",
	}, hostCfg.GroupAdd))
}

func TestIsolation(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		Isolation: "default",
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, container.Isolation("default"), hostCfg.Isolation)
}

func TestMemSwappiness(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		MemSwappiness: yaml.MemStringorInt(10),
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, int64(10), *hostCfg.MemorySwappiness)
}

func TestMemReservation(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		MemReservation: 100000,
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, int64(100000), hostCfg.MemoryReservation)
}

func TestOomKillDisable(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		OomKillDisable: true,
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, true, *hostCfg.OomKillDisable)
}

func TestOomScoreAdj(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		OomScoreAdj: 500,
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, 500, hostCfg.OomScoreAdj)
}

func TestStopGracePeriod(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		StopGracePeriod: "5s",
	}
	cfg, _, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, 5, *cfg.StopTimeout)
}

func TestStopSignal(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		StopSignal: "SIGTERM",
	}
	cfg, _, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.Equal(t, "SIGTERM", cfg.StopSignal)
}

func TestTmpfs(t *testing.T) {
	ctx := &ctx.Context{}
	sc := &config.ServiceConfig{
		Tmpfs: yaml.Stringorslice{"/run"},
	}
	_, hostCfg, err := Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual(map[string]string{
		"/run": "",
	}, hostCfg.Tmpfs))

	sc = &config.ServiceConfig{
		Tmpfs: yaml.Stringorslice{"/run:rw,noexec,nosuid,size=65536k"},
	}
	_, hostCfg, err = Convert(sc, ctx.Context, nil)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual(map[string]string{
		"/run": "rw,noexec,nosuid,size=65536k",
	}, hostCfg.Tmpfs))
}
