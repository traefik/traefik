package config

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/cli/cli/internal/test"
	"github.com/docker/docker/pkg/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestConfigRemoveErrors(t *testing.T) {
	testCases := []struct {
		args             []string
		configRemoveFunc func(string) error
		expectedError    string
	}{
		{
			args:          []string{},
			expectedError: "requires at least 1 argument(s).",
		},
		{
			args: []string{"foo"},
			configRemoveFunc: func(name string) error {
				return errors.Errorf("error removing config")
			},
			expectedError: "error removing config",
		},
	}
	for _, tc := range testCases {
		buf := new(bytes.Buffer)
		cmd := newConfigRemoveCommand(
			test.NewFakeCli(&fakeClient{
				configRemoveFunc: tc.configRemoveFunc,
			}, buf),
		)
		cmd.SetArgs(tc.args)
		cmd.SetOutput(ioutil.Discard)
		testutil.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestConfigRemoveWithName(t *testing.T) {
	names := []string{"foo", "bar"}
	buf := new(bytes.Buffer)
	var removedConfigs []string
	cli := test.NewFakeCli(&fakeClient{
		configRemoveFunc: func(name string) error {
			removedConfigs = append(removedConfigs, name)
			return nil
		},
	}, buf)
	cmd := newConfigRemoveCommand(cli)
	cmd.SetArgs(names)
	assert.NoError(t, cmd.Execute())
	assert.Equal(t, names, strings.Split(strings.TrimSpace(buf.String()), "\n"))
	assert.Equal(t, names, removedConfigs)
}

func TestConfigRemoveContinueAfterError(t *testing.T) {
	names := []string{"foo", "bar"}
	buf := new(bytes.Buffer)
	var removedConfigs []string

	cli := test.NewFakeCli(&fakeClient{
		configRemoveFunc: func(name string) error {
			removedConfigs = append(removedConfigs, name)
			if name == "foo" {
				return errors.Errorf("error removing config: %s", name)
			}
			return nil
		},
	}, buf)

	cmd := newConfigRemoveCommand(cli)
	cmd.SetArgs(names)
	assert.EqualError(t, cmd.Execute(), "error removing config: foo")
	assert.Equal(t, names, removedConfigs)
}
