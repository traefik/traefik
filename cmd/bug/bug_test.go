package bug

import (
	"testing"

	"github.com/containous/traefik/anonymize"
	"github.com/containous/traefik/cmd"
	"github.com/containous/traefik/config/static"
	"github.com/stretchr/testify/assert"
)

func Test_createReport(t *testing.T) {
	traefikConfiguration := &cmd.TraefikConfiguration{
		ConfigFile: "FOO",
		Configuration: static.Configuration{
			EntryPoints: static.EntryPoints{
				"goo": &static.EntryPoint{
					Address: "hoo.bar",
				},
			},
		},
	}

	report, err := createReport(traefikConfiguration)
	assert.NoError(t, err, report)

	// exported anonymous configuration
	assert.NotContains(t, "hoo.bar", report)
}

func Test_anonymize_traefikConfiguration(t *testing.T) {
	traefikConfiguration := &cmd.TraefikConfiguration{
		ConfigFile: "FOO",
		Configuration: static.Configuration{
			EntryPoints: static.EntryPoints{
				"goo": &static.EntryPoint{
					Address: "hoo.bar",
				},
			},
		},
	}
	_, err := anonymize.Do(traefikConfiguration, true)
	assert.NoError(t, err)
	assert.Equal(t, "hoo.bar", traefikConfiguration.Configuration.EntryPoints["goo"].Address)
}
