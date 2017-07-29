package config

import (
	"testing"

	yamlTypes "github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestConfig struct {
	SystemContainers map[string]*ServiceConfig
}

func newTestConfig() TestConfig {
	return TestConfig{
		SystemContainers: map[string]*ServiceConfig{
			"udev": {
				Image:       "udev",
				Restart:     "always",
				NetworkMode: "host",
				Privileged:  true,
				DNS:         []string{"8.8.8.8", "8.8.4.4"},
				Environment: yamlTypes.MaporEqualSlice([]string{
					"DAEMON=true",
				}),
				Labels: yamlTypes.SliceorMap{
					"io.rancher.os.detach": "true",
					"io.rancher.os.scope":  "system",
				},
				VolumesFrom: []string{
					"system-volumes",
				},
				Ulimits: yamlTypes.Ulimits{
					Elements: []yamlTypes.Ulimit{
						yamlTypes.NewUlimit("nproc", 65557, 65557),
					},
				},
			},
			"system-volumes": {
				Image:       "state",
				NetworkMode: "none",
				ReadOnly:    true,
				Privileged:  true,
				Labels: yamlTypes.SliceorMap{
					"io.rancher.os.createonly": "true",
					"io.rancher.os.scope":      "system",
				},
				Volumes: &yamlTypes.Volumes{
					Volumes: []*yamlTypes.Volume{
						{
							Source:      "/dev",
							Destination: "/host/dev",
						},
						{
							Source:      "/var/lib/rancher/conf",
							Destination: "/var/lib/rancher/conf",
						},
						{
							Source:      "/etc/ssl/certs/ca-certificates.crt",
							Destination: "/etc/ssl/certs/ca-certificates.crt.rancher",
						},
						{
							Source:      "/lib/modules",
							Destination: "lib/modules",
						},
						{
							Source:      "/lib/firmware",
							Destination: "/lib/firmware",
						},
						{
							Source:      "/var/run",
							Destination: "/var/run",
						},
						{
							Source:      "/var/log",
							Destination: "/var/log",
						},
					},
				},
				Logging: Log{
					Driver: "json-file",
				},
			},
		},
	}
}

func TestMarshalConfig(t *testing.T) {
	config := newTestConfig()
	bytes, err := yaml.Marshal(config)
	assert.Nil(t, err)

	config2 := TestConfig{}

	err = yaml.Unmarshal(bytes, &config2)
	assert.Nil(t, err)

	assert.Equal(t, config, config2)
}

func TestMarshalServiceConfig(t *testing.T) {
	configPtr := newTestConfig().SystemContainers["udev"]
	bytes, err := yaml.Marshal(configPtr)
	assert.Nil(t, err)

	configPtr2 := &ServiceConfig{}

	err = yaml.Unmarshal(bytes, configPtr2)
	assert.Nil(t, err)

	assert.Equal(t, configPtr, configPtr2)
}
