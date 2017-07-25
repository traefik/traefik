package config

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/internal/test"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	// Import builders to get the builder function as package function
	. "github.com/docker/cli/cli/internal/test/builders"
	"github.com/docker/docker/pkg/testutil"
	"github.com/docker/docker/pkg/testutil/golden"
	"github.com/stretchr/testify/assert"
)

func TestConfigListErrors(t *testing.T) {
	testCases := []struct {
		args           []string
		configListFunc func(types.ConfigListOptions) ([]swarm.Config, error)
		expectedError  string
	}{
		{
			args:          []string{"foo"},
			expectedError: "accepts no argument",
		},
		{
			configListFunc: func(options types.ConfigListOptions) ([]swarm.Config, error) {
				return []swarm.Config{}, errors.Errorf("error listing configs")
			},
			expectedError: "error listing configs",
		},
	}
	for _, tc := range testCases {
		buf := new(bytes.Buffer)
		cmd := newConfigListCommand(
			test.NewFakeCli(&fakeClient{
				configListFunc: tc.configListFunc,
			}, buf),
		)
		cmd.SetArgs(tc.args)
		cmd.SetOutput(ioutil.Discard)
		testutil.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestConfigList(t *testing.T) {
	buf := new(bytes.Buffer)
	cli := test.NewFakeCli(&fakeClient{
		configListFunc: func(options types.ConfigListOptions) ([]swarm.Config, error) {
			return []swarm.Config{
				*Config(ConfigID("ID-foo"),
					ConfigName("foo"),
					ConfigVersion(swarm.Version{Index: 10}),
					ConfigCreatedAt(time.Now().Add(-2*time.Hour)),
					ConfigUpdatedAt(time.Now().Add(-1*time.Hour)),
				),
				*Config(ConfigID("ID-bar"),
					ConfigName("bar"),
					ConfigVersion(swarm.Version{Index: 11}),
					ConfigCreatedAt(time.Now().Add(-2*time.Hour)),
					ConfigUpdatedAt(time.Now().Add(-1*time.Hour)),
				),
			}, nil
		},
	}, buf)
	cli.SetConfigfile(&configfile.ConfigFile{})
	cmd := newConfigListCommand(cli)
	cmd.SetOutput(buf)
	assert.NoError(t, cmd.Execute())
	actual := buf.String()
	expected := golden.Get(t, []byte(actual), "config-list.golden")
	testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
}

func TestConfigListWithQuietOption(t *testing.T) {
	buf := new(bytes.Buffer)
	cli := test.NewFakeCli(&fakeClient{
		configListFunc: func(options types.ConfigListOptions) ([]swarm.Config, error) {
			return []swarm.Config{
				*Config(ConfigID("ID-foo"), ConfigName("foo")),
				*Config(ConfigID("ID-bar"), ConfigName("bar"), ConfigLabels(map[string]string{
					"label": "label-bar",
				})),
			}, nil
		},
	}, buf)
	cli.SetConfigfile(&configfile.ConfigFile{})
	cmd := newConfigListCommand(cli)
	cmd.Flags().Set("quiet", "true")
	assert.NoError(t, cmd.Execute())
	actual := buf.String()
	expected := golden.Get(t, []byte(actual), "config-list-with-quiet-option.golden")
	testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
}

func TestConfigListWithConfigFormat(t *testing.T) {
	buf := new(bytes.Buffer)
	cli := test.NewFakeCli(&fakeClient{
		configListFunc: func(options types.ConfigListOptions) ([]swarm.Config, error) {
			return []swarm.Config{
				*Config(ConfigID("ID-foo"), ConfigName("foo")),
				*Config(ConfigID("ID-bar"), ConfigName("bar"), ConfigLabels(map[string]string{
					"label": "label-bar",
				})),
			}, nil
		},
	}, buf)
	cli.SetConfigfile(&configfile.ConfigFile{
		ConfigFormat: "{{ .Name }} {{ .Labels }}",
	})
	cmd := newConfigListCommand(cli)
	assert.NoError(t, cmd.Execute())
	actual := buf.String()
	expected := golden.Get(t, []byte(actual), "config-list-with-config-format.golden")
	testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
}

func TestConfigListWithFormat(t *testing.T) {
	buf := new(bytes.Buffer)
	cli := test.NewFakeCli(&fakeClient{
		configListFunc: func(options types.ConfigListOptions) ([]swarm.Config, error) {
			return []swarm.Config{
				*Config(ConfigID("ID-foo"), ConfigName("foo")),
				*Config(ConfigID("ID-bar"), ConfigName("bar"), ConfigLabels(map[string]string{
					"label": "label-bar",
				})),
			}, nil
		},
	}, buf)
	cmd := newConfigListCommand(cli)
	cmd.Flags().Set("format", "{{ .Name }} {{ .Labels }}")
	assert.NoError(t, cmd.Execute())
	actual := buf.String()
	expected := golden.Get(t, []byte(actual), "config-list-with-format.golden")
	testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
}

func TestConfigListWithFilter(t *testing.T) {
	buf := new(bytes.Buffer)
	cli := test.NewFakeCli(&fakeClient{
		configListFunc: func(options types.ConfigListOptions) ([]swarm.Config, error) {
			assert.Equal(t, "foo", options.Filters.Get("name")[0])
			assert.Equal(t, "lbl1=Label-bar", options.Filters.Get("label")[0])
			return []swarm.Config{
				*Config(ConfigID("ID-foo"),
					ConfigName("foo"),
					ConfigVersion(swarm.Version{Index: 10}),
					ConfigCreatedAt(time.Now().Add(-2*time.Hour)),
					ConfigUpdatedAt(time.Now().Add(-1*time.Hour)),
				),
				*Config(ConfigID("ID-bar"),
					ConfigName("bar"),
					ConfigVersion(swarm.Version{Index: 11}),
					ConfigCreatedAt(time.Now().Add(-2*time.Hour)),
					ConfigUpdatedAt(time.Now().Add(-1*time.Hour)),
				),
			}, nil
		},
	}, buf)
	cli.SetConfigfile(&configfile.ConfigFile{})
	cmd := newConfigListCommand(cli)
	cmd.Flags().Set("filter", "name=foo")
	cmd.Flags().Set("filter", "label=lbl1=Label-bar")
	assert.NoError(t, cmd.Execute())
	actual := buf.String()
	expected := golden.Get(t, []byte(actual), "config-list-with-filter.golden")
	testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
}
