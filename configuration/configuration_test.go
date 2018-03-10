package configuration

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/file"
)

const defaultConfigFile = "traefik.toml"

func TestSetEffectiveConfigurationGraceTimeout(t *testing.T) {
	tests := []struct {
		desc                  string
		legacyGraceTimeout    time.Duration
		lifeCycleGraceTimeout time.Duration
		wantGraceTimeout      time.Duration
	}{
		{
			desc:               "legacy grace timeout given only",
			legacyGraceTimeout: 5 * time.Second,
			wantGraceTimeout:   5 * time.Second,
		},
		{
			desc:                  "legacy and life cycle grace timeouts given",
			legacyGraceTimeout:    5 * time.Second,
			lifeCycleGraceTimeout: 12 * time.Second,
			wantGraceTimeout:      5 * time.Second,
		},
		{
			desc:                  "legacy grace timeout omitted",
			legacyGraceTimeout:    0,
			lifeCycleGraceTimeout: 12 * time.Second,
			wantGraceTimeout:      12 * time.Second,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			gc := &GlobalConfiguration{
				GraceTimeOut: flaeg.Duration(test.legacyGraceTimeout),
			}
			if test.lifeCycleGraceTimeout > 0 {
				gc.LifeCycle = &LifeCycle{
					GraceTimeOut: flaeg.Duration(test.lifeCycleGraceTimeout),
				}
			}

			gc.SetEffectiveConfiguration(defaultConfigFile)

			gotGraceTimeout := time.Duration(gc.LifeCycle.GraceTimeOut)
			if gotGraceTimeout != test.wantGraceTimeout {
				t.Fatalf("got effective grace timeout %d, want %d", gotGraceTimeout, test.wantGraceTimeout)
			}

		})
	}
}

func TestSetEffectiveConfigurationFileProviderFilename(t *testing.T) {
	tests := []struct {
		desc                     string
		fileProvider             *file.Provider
		wantFileProviderFilename string
	}{
		{
			desc:                     "no filename for file provider given",
			fileProvider:             &file.Provider{},
			wantFileProviderFilename: defaultConfigFile,
		},
		{
			desc:                     "filename for file provider given",
			fileProvider:             &file.Provider{BaseProvider: provider.BaseProvider{Filename: "other.toml"}},
			wantFileProviderFilename: "other.toml",
		},
		{
			desc:                     "directory for file provider given",
			fileProvider:             &file.Provider{Directory: "/"},
			wantFileProviderFilename: "",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			gc := &GlobalConfiguration{
				File: test.fileProvider,
			}

			gc.SetEffectiveConfiguration(defaultConfigFile)

			gotFileProviderFilename := gc.File.Filename
			if gotFileProviderFilename != test.wantFileProviderFilename {
				t.Fatalf("got file provider file name %q, want %q", gotFileProviderFilename, test.wantFileProviderFilename)
			}
		})
	}
}
