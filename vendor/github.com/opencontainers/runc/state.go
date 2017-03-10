// +build linux

package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/urfave/cli"
)

// cState represents the platform agnostic pieces relating to a running
// container's status and state.  Note: The fields in this structure adhere to
// the opencontainers/runtime-spec/specs-go requirement for json fields that must be returned
// in a state command.
type cState struct {
	// Version is the OCI version for the container
	Version string `json:"ociVersion"`
	// ID is the container ID
	ID string `json:"id"`
	// InitProcessPid is the init process id in the parent namespace
	InitProcessPid int `json:"pid"`
	// Bundle is the path on the filesystem to the bundle
	Bundle string `json:"bundlePath"`
	// Rootfs is a path to a directory containing the container's root filesystem.
	Rootfs string `json:"rootfsPath"`
	// Status is the current status of the container, running, paused, ...
	Status string `json:"status"`
	// Created is the unix timestamp for the creation time of the container in UTC
	Created time.Time `json:"created"`
	// Annotations is the user defined annotations added to the config.
	Annotations map[string]string `json:"annotations,omitempty"`
}

var stateCommand = cli.Command{
	Name:  "state",
	Usage: "output the state of a container",
	ArgsUsage: `<container-id>

Where "<container-id>" is your name for the instance of the container.`,
	Description: `The state command outputs current state information for the
instance of a container.`,
	Action: func(context *cli.Context) error {
		container, err := getContainer(context)
		if err != nil {
			return err
		}
		containerStatus, err := container.Status()
		if err != nil {
			return err
		}
		state, err := container.State()
		if err != nil {
			return err
		}
		bundle, annotations := utils.Annotations(state.Config.Labels)
		cs := cState{
			Version:        state.BaseState.Config.Version,
			ID:             state.BaseState.ID,
			InitProcessPid: state.BaseState.InitProcessPid,
			Status:         containerStatus.String(),
			Bundle:         bundle,
			Rootfs:         state.BaseState.Config.Rootfs,
			Created:        state.BaseState.Created,
			Annotations:    annotations,
		}
		data, err := json.MarshalIndent(cs, "", "  ")
		if err != nil {
			return err
		}
		os.Stdout.Write(data)
		return nil
	},
}
