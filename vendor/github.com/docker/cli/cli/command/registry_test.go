package command_test

import (
	"bytes"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	// Prevents a circular import with "github.com/docker/cli/cli/internal/test"
	. "github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/internal/test"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type fakeClient struct {
	client.Client
	infoFunc func() (types.Info, error)
}

func (cli *fakeClient) Info(_ context.Context) (types.Info, error) {
	if cli.infoFunc != nil {
		return cli.infoFunc()
	}
	return types.Info{}, nil
}

func TestElectAuthServer(t *testing.T) {
	testCases := []struct {
		expectedAuthServer string
		expectedWarning    string
		infoFunc           func() (types.Info, error)
	}{
		{
			expectedAuthServer: "https://index.docker.io/v1/",
			expectedWarning:    "",
			infoFunc: func() (types.Info, error) {
				return types.Info{IndexServerAddress: "https://index.docker.io/v1/"}, nil
			},
		},
		{
			expectedAuthServer: "https://index.docker.io/v1/",
			expectedWarning:    "Empty registry endpoint from daemon",
			infoFunc: func() (types.Info, error) {
				return types.Info{IndexServerAddress: ""}, nil
			},
		},
		{
			expectedAuthServer: "https://foo.bar",
			expectedWarning:    "",
			infoFunc: func() (types.Info, error) {
				return types.Info{IndexServerAddress: "https://foo.bar"}, nil
			},
		},
		{
			expectedAuthServer: "https://index.docker.io/v1/",
			expectedWarning:    "failed to get default registry endpoint from daemon",
			infoFunc: func() (types.Info, error) {
				return types.Info{}, errors.Errorf("error getting info")
			},
		},
	}
	for _, tc := range testCases {
		buf := new(bytes.Buffer)
		cli := test.NewFakeCli(&fakeClient{infoFunc: tc.infoFunc}, buf)
		errBuf := new(bytes.Buffer)
		cli.SetErr(errBuf)
		server := ElectAuthServer(context.Background(), cli)
		assert.Equal(t, tc.expectedAuthServer, server)
		actual := errBuf.String()
		if tc.expectedWarning == "" {
			assert.Empty(t, actual)
		} else {
			assert.Contains(t, actual, tc.expectedWarning)
		}
	}
}
