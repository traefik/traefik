package service

import (
	"testing"
	"time"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestMemBytesString(t *testing.T) {
	var mem opts.MemBytes = 1048576
	assert.Equal(t, "1MiB", mem.String())
}

func TestMemBytesSetAndValue(t *testing.T) {
	var mem opts.MemBytes
	assert.NoError(t, mem.Set("5kb"))
	assert.Equal(t, int64(5120), mem.Value())
}

func TestNanoCPUsString(t *testing.T) {
	var cpus opts.NanoCPUs = 6100000000
	assert.Equal(t, "6.100", cpus.String())
}

func TestNanoCPUsSetAndValue(t *testing.T) {
	var cpus opts.NanoCPUs
	assert.NoError(t, cpus.Set("0.35"))
	assert.Equal(t, int64(350000000), cpus.Value())
}

func TestUint64OptString(t *testing.T) {
	value := uint64(2345678)
	opt := Uint64Opt{value: &value}
	assert.Equal(t, "2345678", opt.String())

	opt = Uint64Opt{}
	assert.Equal(t, "", opt.String())
}

func TestUint64OptSetAndValue(t *testing.T) {
	var opt Uint64Opt
	assert.NoError(t, opt.Set("14445"))
	assert.Equal(t, uint64(14445), *opt.Value())
}

func TestHealthCheckOptionsToHealthConfig(t *testing.T) {
	dur := time.Second
	opt := healthCheckOptions{
		cmd:         "curl",
		interval:    opts.PositiveDurationOpt{*opts.NewDurationOpt(&dur)},
		timeout:     opts.PositiveDurationOpt{*opts.NewDurationOpt(&dur)},
		startPeriod: opts.PositiveDurationOpt{*opts.NewDurationOpt(&dur)},
		retries:     10,
	}
	config, err := opt.toHealthConfig()
	assert.NoError(t, err)
	assert.Equal(t, &container.HealthConfig{
		Test:        []string{"CMD-SHELL", "curl"},
		Interval:    time.Second,
		Timeout:     time.Second,
		StartPeriod: time.Second,
		Retries:     10,
	}, config)
}

func TestHealthCheckOptionsToHealthConfigNoHealthcheck(t *testing.T) {
	opt := healthCheckOptions{
		noHealthcheck: true,
	}
	config, err := opt.toHealthConfig()
	assert.NoError(t, err)
	assert.Equal(t, &container.HealthConfig{
		Test: []string{"NONE"},
	}, config)
}

func TestHealthCheckOptionsToHealthConfigConflict(t *testing.T) {
	opt := healthCheckOptions{
		cmd:           "curl",
		noHealthcheck: true,
	}
	_, err := opt.toHealthConfig()
	assert.EqualError(t, err, "--no-healthcheck conflicts with --health-* options")
}
