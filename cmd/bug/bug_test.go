package bug

import (
	"testing"

	"github.com/containous/traefik/anonymize"
	"github.com/containous/traefik/cmd"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func Test_createReport(t *testing.T) {
	traefikConfiguration := &cmd.TraefikConfiguration{
		ConfigFile: "FOO",
		GlobalConfiguration: configuration.GlobalConfiguration{
			EntryPoints: configuration.EntryPoints{
				"goo": &configuration.EntryPoint{
					Address: "hoo.bar",
					Auth: &types.Auth{
						Basic: &types.Basic{
							UsersFile: "foo Basic UsersFile",
							Users:     types.Users{"foo Basic Users 1", "foo Basic Users 2", "foo Basic Users 3"},
						},
						Digest: &types.Digest{
							UsersFile: "foo Digest UsersFile",
							Users:     types.Users{"foo Digest Users 1", "foo Digest Users 2", "foo Digest Users 3"},
						},
					},
				},
			},
			File: &file.Provider{
				Directory: "BAR",
			},
			RootCAs: tls.FilesOrContents{"fllf"},
		},
	}

	report, err := createReport(traefikConfiguration)
	assert.NoError(t, err, report)

	// exported anonymous configuration
	assert.NotContains(t, "web Basic Users ", report)
	assert.NotContains(t, "foo Digest Users ", report)
	assert.NotContains(t, "hoo.bar", report)
}

func Test_anonymize_traefikConfiguration(t *testing.T) {
	traefikConfiguration := &cmd.TraefikConfiguration{
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
