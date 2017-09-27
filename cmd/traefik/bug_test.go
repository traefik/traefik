package main

import (
	"testing"

	"github.com/containous/traefik/cmd/traefik/anonymize"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/provider/file"
	"github.com/stretchr/testify/assert"
)

func Test_createBugReport(t *testing.T) {
	traefikConfiguration := TraefikConfiguration{
		ConfigFile: "FOO",
		GlobalConfiguration: configuration.GlobalConfiguration{
			EntryPoints: configuration.EntryPoints{
				"goo": &configuration.EntryPoint{
					Address: "hoo.bar",
				},
			},
			File: &file.Provider{
				Directory: "BAR",
			},
			RootCAs: configuration.RootCAs{"fllf"},
		},
	}

	report, err := createBugReport(traefikConfiguration)
	assert.NoError(t, err, report)
}

func Test_anonymize_traefikConfiguration(t *testing.T) {
	traefikConfiguration := &TraefikConfiguration{
		ConfigFile: "FOO",
		GlobalConfiguration: configuration.GlobalConfiguration{
			EntryPoints: configuration.EntryPoints{
				"goo": &configuration.EntryPoint{
					Address: "hoo.bar",
				},
			},
			File: &file.Provider{
				Directory: "BAR",
			},
		},
	}
	_, err := anonymize.Do(traefikConfiguration, true)
	assert.NoError(t, err)
	assert.Equal(t, "hoo.bar", traefikConfiguration.GlobalConfiguration.EntryPoints["goo"].Address)
}
