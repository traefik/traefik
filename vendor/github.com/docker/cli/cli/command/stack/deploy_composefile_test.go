package stack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/cli/cli/internal/test/network"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/testutil"
	"github.com/docker/docker/pkg/testutil/tempfile"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestGetConfigDetails(t *testing.T) {
	content := `
version: "3.0"
services:
  foo:
    image: alpine:3.5
`
	file := tempfile.NewTempFile(t, "test-get-config-details", content)
	defer file.Remove()

	details, err := getConfigDetails(file.Name())
	require.NoError(t, err)
	assert.Equal(t, filepath.Dir(file.Name()), details.WorkingDir)
	assert.Len(t, details.ConfigFiles, 1)
	assert.Len(t, details.Environment, len(os.Environ()))
}

type notFound struct {
	error
}

func (n notFound) NotFound() bool {
	return true
}

func TestValidateExternalNetworks(t *testing.T) {
	var testcases = []struct {
		inspectResponse types.NetworkResource
		inspectError    error
		expectedMsg     string
		network         string
	}{
		{
			inspectError: notFound{},
			expectedMsg:  "could not be found. You need to create a swarm-scoped network",
		},
		{
			inspectError: errors.New("Unexpected"),
			expectedMsg:  "Unexpected",
		},
		{
			network: "host",
		},
		{
			network:     "user",
			expectedMsg: "is not in the right scope",
		},
		{
			network:         "user",
			inspectResponse: types.NetworkResource{Scope: "swarm"},
		},
	}

	for _, testcase := range testcases {
		fakeClient := &network.FakeClient{
			NetworkInspectFunc: func(_ context.Context, _ string, _ bool) (types.NetworkResource, error) {
				return testcase.inspectResponse, testcase.inspectError
			},
		}
		networks := []string{testcase.network}
		err := validateExternalNetworks(context.Background(), fakeClient, networks)
		if testcase.expectedMsg == "" {
			assert.NoError(t, err)
		} else {
			testutil.ErrorContains(t, err, testcase.expectedMsg)
		}
	}
}
