package volume

import (
	"testing"

	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/config"
)

func TestVolumesFromServices(t *testing.T) {
	cases := []struct {
		volumeConfigs   map[string]*config.VolumeConfig
		services        map[string]*config.ServiceConfig
		volumeEnabled   bool
		expectedVolumes []*Volume
		expectedError   bool
	}{
		{},
		{
			volumeConfigs: map[string]*config.VolumeConfig{
				"vol1": {},
			},
			services: map[string]*config.ServiceConfig{},
			expectedVolumes: []*Volume{
				{
					name:        "vol1",
					projectName: "prj",
				},
			},
			expectedError: false,
		},
		{
			volumeConfigs: map[string]*config.VolumeConfig{
				"vol1": nil,
			},
			services: map[string]*config.ServiceConfig{},
			expectedVolumes: []*Volume{
				{
					name:        "vol1",
					projectName: "prj",
				},
			},
			expectedError: false,
		},
	}

	for index, c := range cases {
		services := config.NewServiceConfigs()
		for name, service := range c.services {
			services.Add(name, service)
		}

		volumes, err := VolumesFromServices(&volumeClient{}, "prj", c.volumeConfigs, services, c.volumeEnabled)
		if c.expectedError {
			if err == nil {
				t.Fatalf("%d: expected an error, got nothing", index)
			}
		} else {
			if err != nil {
				t.Fatalf("%d: didn't expect an error, got one %s", index, err.Error())
			}
		}
		if volumes.volumeEnabled != c.volumeEnabled {
			t.Fatalf("%d: expected volume enabled %v, got %v", index, c.volumeEnabled, volumes.volumeEnabled)
		}
		if len(volumes.volumes) != len(c.expectedVolumes) {
			t.Fatalf("%d: expected %v, got %v", index, c.expectedVolumes, volumes.volumes)
		}
		for _, volume := range volumes.volumes {
			testExpectedContainsVolume(t, index, c.expectedVolumes, volume)
		}
	}
}

func testExpectedContainsVolume(t *testing.T, index int, expected []*Volume, volume *Volume) {
	found := false
	for _, e := range expected {
		if e.name == volume.name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("%d: volume %v not found in %v", index, volume, expected)
	}
}

type volumeClient struct {
	client.Client
	expectedName         string
	expectedVolumeCreate volume.VolumesCreateBody
	inspectError         error
	inspectVolumeDriver  string
	inspectVolumeOptions map[string]string
	removeError          error
}
