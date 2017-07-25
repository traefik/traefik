package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/docker/cli/cli/internal/test"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	// Import builders to get the builder function as package function
	. "github.com/docker/cli/cli/internal/test/builders"
	"github.com/docker/docker/pkg/testutil"
	"github.com/docker/docker/pkg/testutil/golden"
	"github.com/stretchr/testify/assert"
)

func TestConfigInspectErrors(t *testing.T) {
	testCases := []struct {
		args              []string
		flags             map[string]string
		configInspectFunc func(configID string) (swarm.Config, []byte, error)
		expectedError     string
	}{
		{
			expectedError: "requires at least 1 argument",
		},
		{
			args: []string{"foo"},
			configInspectFunc: func(configID string) (swarm.Config, []byte, error) {
				return swarm.Config{}, nil, errors.Errorf("error while inspecting the config")
			},
			expectedError: "error while inspecting the config",
		},
		{
			args: []string{"foo"},
			flags: map[string]string{
				"format": "{{invalid format}}",
			},
			expectedError: "Template parsing error",
		},
		{
			args: []string{"foo", "bar"},
			configInspectFunc: func(configID string) (swarm.Config, []byte, error) {
				if configID == "foo" {
					return *Config(ConfigName("foo")), nil, nil
				}
				return swarm.Config{}, nil, errors.Errorf("error while inspecting the config")
			},
			expectedError: "error while inspecting the config",
		},
	}
	for _, tc := range testCases {
		buf := new(bytes.Buffer)
		cmd := newConfigInspectCommand(
			test.NewFakeCli(&fakeClient{
				configInspectFunc: tc.configInspectFunc,
			}, buf),
		)
		cmd.SetArgs(tc.args)
		for key, value := range tc.flags {
			cmd.Flags().Set(key, value)
		}
		cmd.SetOutput(ioutil.Discard)
		testutil.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestConfigInspectWithoutFormat(t *testing.T) {
	testCases := []struct {
		name              string
		args              []string
		configInspectFunc func(configID string) (swarm.Config, []byte, error)
	}{
		{
			name: "single-config",
			args: []string{"foo"},
			configInspectFunc: func(name string) (swarm.Config, []byte, error) {
				if name != "foo" {
					return swarm.Config{}, nil, errors.Errorf("Invalid name, expected %s, got %s", "foo", name)
				}
				return *Config(ConfigID("ID-foo"), ConfigName("foo")), nil, nil
			},
		},
		{
			name: "multiple-configs-with-labels",
			args: []string{"foo", "bar"},
			configInspectFunc: func(name string) (swarm.Config, []byte, error) {
				return *Config(ConfigID("ID-"+name), ConfigName(name), ConfigLabels(map[string]string{
					"label1": "label-foo",
				})), nil, nil
			},
		},
	}
	for _, tc := range testCases {
		buf := new(bytes.Buffer)
		cmd := newConfigInspectCommand(
			test.NewFakeCli(&fakeClient{
				configInspectFunc: tc.configInspectFunc,
			}, buf),
		)
		cmd.SetArgs(tc.args)
		assert.NoError(t, cmd.Execute())
		actual := buf.String()
		expected := golden.Get(t, []byte(actual), fmt.Sprintf("config-inspect-without-format.%s.golden", tc.name))
		testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
	}
}

func TestConfigInspectWithFormat(t *testing.T) {
	configInspectFunc := func(name string) (swarm.Config, []byte, error) {
		return *Config(ConfigName("foo"), ConfigLabels(map[string]string{
			"label1": "label-foo",
		})), nil, nil
	}
	testCases := []struct {
		name              string
		format            string
		args              []string
		configInspectFunc func(name string) (swarm.Config, []byte, error)
	}{
		{
			name:              "simple-template",
			format:            "{{.Spec.Name}}",
			args:              []string{"foo"},
			configInspectFunc: configInspectFunc,
		},
		{
			name:              "json-template",
			format:            "{{json .Spec.Labels}}",
			args:              []string{"foo"},
			configInspectFunc: configInspectFunc,
		},
	}
	for _, tc := range testCases {
		buf := new(bytes.Buffer)
		cmd := newConfigInspectCommand(
			test.NewFakeCli(&fakeClient{
				configInspectFunc: tc.configInspectFunc,
			}, buf),
		)
		cmd.SetArgs(tc.args)
		cmd.Flags().Set("format", tc.format)
		assert.NoError(t, cmd.Execute())
		actual := buf.String()
		expected := golden.Get(t, []byte(actual), fmt.Sprintf("config-inspect-with-format.%s.golden", tc.name))
		testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
	}
}

func TestConfigInspectPretty(t *testing.T) {
	testCases := []struct {
		name              string
		configInspectFunc func(string) (swarm.Config, []byte, error)
	}{
		{
			name: "simple",
			configInspectFunc: func(id string) (swarm.Config, []byte, error) {
				return *Config(
					ConfigLabels(map[string]string{
						"lbl1": "value1",
					}),
					ConfigID("configID"),
					ConfigName("configName"),
					ConfigCreatedAt(time.Time{}),
					ConfigUpdatedAt(time.Time{}),
					ConfigData([]byte("payload here")),
				), []byte{}, nil
			},
		},
	}
	for _, tc := range testCases {
		buf := new(bytes.Buffer)
		cmd := newConfigInspectCommand(
			test.NewFakeCli(&fakeClient{
				configInspectFunc: tc.configInspectFunc,
			}, buf))
		cmd.SetArgs([]string{"configID"})
		cmd.Flags().Set("pretty", "true")
		assert.NoError(t, cmd.Execute())
		actual := buf.String()
		expected := golden.Get(t, []byte(actual), fmt.Sprintf("config-inspect-pretty.%s.golden", tc.name))
		testutil.EqualNormalizedString(t, testutil.RemoveSpace, actual, string(expected))
	}
}
