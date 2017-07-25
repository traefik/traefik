package service

import (
	"io"
	"io/ioutil"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/service/progress"
	"github.com/docker/docker/pkg/jsonmessage"
	"golang.org/x/net/context"
)

// waitOnService waits for the service to converge. It outputs a progress bar,
// if appopriate based on the CLI flags.
func waitOnService(ctx context.Context, dockerCli *command.DockerCli, serviceID string, opts *serviceOptions) error {
	errChan := make(chan error, 1)
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		errChan <- progress.ServiceProgress(ctx, dockerCli.Client(), serviceID, pipeWriter)
	}()

	if opts.quiet {
		go io.Copy(ioutil.Discard, pipeReader)
		return <-errChan
	}

	err := jsonmessage.DisplayJSONMessagesToStream(pipeReader, dockerCli.Out(), nil)
	if err == nil {
		err = <-errChan
	}
	return err
}
