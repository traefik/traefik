package config

import (
	"reflect"
	"testing"

	"github.com/docker/libcompose/yaml"
)

func TestBuild(t *testing.T) {
	v2Services, err := ConvertServices(map[string]*ServiceConfigV1{
		"test": {
			Build:      ".",
			Dockerfile: "Dockerfile",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	v2Config := v2Services["test"]

	expectedBuild := yaml.Build{
		Context:    ".",
		Dockerfile: "Dockerfile",
	}

	if !reflect.DeepEqual(v2Config.Build, expectedBuild) {
		t.Fatal("Failed to convert build", v2Config.Build)
	}
}

func TestLogging(t *testing.T) {
	v2Services, err := ConvertServices(map[string]*ServiceConfigV1{
		"test": {
			LogDriver: "json-file",
			LogOpt: map[string]string{
				"key": "value",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	v2Config := v2Services["test"]

	expectedLogging := Log{
		Driver: "json-file",
		Options: map[string]string{
			"key": "value",
		},
	}

	if !reflect.DeepEqual(v2Config.Logging, expectedLogging) {
		t.Fatal("Failed to convert logging", v2Config.Logging)
	}
}

func TestNetworkMode(t *testing.T) {
	v2Services, err := ConvertServices(map[string]*ServiceConfigV1{
		"test": {
			Net: "host",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	v2Config := v2Services["test"]

	if v2Config.NetworkMode != "host" {
		t.Fatal("Failed to convert network mode", v2Config.NetworkMode)
	}
}
