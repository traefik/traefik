package swarm

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/context"

	"io/ioutil"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/swarm/progress"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type caOptions struct {
	swarmOptions
	rootCACert PEMFile
	rootCAKey  PEMFile
	rotate     bool
	detach     bool
	quiet      bool
}

func newRotateCACommand(dockerCli command.Cli) *cobra.Command {
	opts := caOptions{}

	cmd := &cobra.Command{
		Use:   "ca [OPTIONS]",
		Short: "Manage root CA",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRotateCA(dockerCli, cmd.Flags(), opts)
		},
		Tags: map[string]string{"version": "1.30"},
	}

	flags := cmd.Flags()
	addSwarmCAFlags(flags, &opts.swarmOptions)
	flags.BoolVar(&opts.rotate, flagRotate, false, "Rotate the swarm CA - if no certificate or key are provided, new ones will be generated")
	flags.Var(&opts.rootCACert, flagCACert, "Path to the PEM-formatted root CA certificate to use for the new cluster")
	flags.Var(&opts.rootCAKey, flagCAKey, "Path to the PEM-formatted root CA key to use for the new cluster")

	flags.BoolVarP(&opts.detach, "detach", "d", false, "Exit immediately instead of waiting for the root rotation to converge")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Suppress progress output")
	return cmd
}

func runRotateCA(dockerCli command.Cli, flags *pflag.FlagSet, opts caOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	swarmInspect, err := client.SwarmInspect(ctx)
	if err != nil {
		return err
	}

	if !opts.rotate {
		if swarmInspect.ClusterInfo.TLSInfo.TrustRoot == "" {
			fmt.Fprintln(dockerCli.Out(), "No CA information available")
		} else {
			fmt.Fprintln(dockerCli.Out(), strings.TrimSpace(swarmInspect.ClusterInfo.TLSInfo.TrustRoot))
		}
		return nil
	}

	genRootCA := true
	spec := &swarmInspect.Spec
	opts.mergeSwarmSpec(spec, flags)
	if flags.Changed(flagCACert) {
		spec.CAConfig.SigningCACert = opts.rootCACert.Contents()
		genRootCA = false
	}
	if flags.Changed(flagCAKey) {
		spec.CAConfig.SigningCAKey = opts.rootCAKey.Contents()
		genRootCA = false
	}
	if genRootCA {
		spec.CAConfig.ForceRotate++
		spec.CAConfig.SigningCACert = ""
		spec.CAConfig.SigningCAKey = ""
	}

	if err := client.SwarmUpdate(ctx, swarmInspect.Version, swarmInspect.Spec, swarm.UpdateFlags{}); err != nil {
		return err
	}

	if opts.detach {
		return nil
	}

	errChan := make(chan error, 1)
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		errChan <- progress.RootRotationProgress(ctx, client, pipeWriter)
	}()

	if opts.quiet {
		go io.Copy(ioutil.Discard, pipeReader)
		return <-errChan
	}

	err = jsonmessage.DisplayJSONMessagesToStream(pipeReader, dockerCli.Out(), nil)
	if err == nil {
		err = <-errChan
	}
	if err != nil {
		return err
	}

	swarmInspect, err = client.SwarmInspect(ctx)
	if err != nil {
		return err
	}

	if swarmInspect.ClusterInfo.TLSInfo.TrustRoot == "" {
		fmt.Fprintln(dockerCli.Out(), "No CA information available")
	} else {
		fmt.Fprintln(dockerCli.Out(), strings.TrimSpace(swarmInspect.ClusterInfo.TLSInfo.TrustRoot))
	}
	return nil
}
