package bug

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/anonymize"
	"github.com/traefik/traefik/cmd"
	"github.com/traefik/traefik/configuration"
	"github.com/traefik/traefik/provider/file"
	"github.com/traefik/traefik/tls"
	"github.com/traefik/traefik/types"
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
