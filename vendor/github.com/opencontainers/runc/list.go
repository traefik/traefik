// +build linux

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"encoding/json"

	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/urfave/cli"
)

const formatOptions = `table or json`

// containerState represents the platform agnostic pieces relating to a
// running container's status and state
type containerState struct {
	// ID is the container ID
	ID string `json:"id"`
	// InitProcessPid is the init process id in the parent namespace
	InitProcessPid int `json:"pid"`
	// Status is the current status of the container, running, paused, ...
	Status string `json:"status"`
	// Bundle is the path on the filesystem to the bundle
	Bundle string `json:"bundle"`
	// Created is the unix timestamp for the creation time of the container in UTC
	Created time.Time `json:"created"`
	// Annotations is the user defined annotations added to the config.
	Annotations map[string]string `json:"annotations,omitempty"`
}

var listCommand = cli.Command{
	Name:  "list",
	Usage: "lists containers started by runc with the given root",
	ArgsUsage: `

Where the given root is specified via the global option "--root"
(default: "/run/runc").

EXAMPLE 1:
To list containers created via the default "--root":
       # runc list

EXAMPLE 2:
To list containers created using a non-default value for "--root":
       # runc --root value list`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "format, f",
			Value: "table",
			Usage: `select one of: ` + formatOptions,
		},
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "display only container IDs",
		},
	},
	Action: func(context *cli.Context) error {
		s, err := getContainers(context)
		if err != nil {
			return err
		}

		if context.Bool("quiet") {
			for _, item := range s {
				fmt.Println(item.ID)
			}
			return nil
		}

		switch context.String("format") {
		case "table":
			w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
			fmt.Fprint(w, "ID\tPID\tSTATUS\tBUNDLE\tCREATED\n")
			for _, item := range s {
				fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
					item.ID,
					item.InitProcessPid,
					item.Status,
					item.Bundle,
					item.Created.Format(time.RFC3339Nano))
			}
			if err := w.Flush(); err != nil {
				return err
			}
		case "json":
			if err := json.NewEncoder(os.Stdout).Encode(s); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid format option")
		}
		return nil
	},
}

func getContainers(context *cli.Context) ([]containerState, error) {
	factory, err := loadFactory(context)
	if err != nil {
		return nil, err
	}
	root := context.GlobalString("root")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	list, err := ioutil.ReadDir(absRoot)
	if err != nil {
		fatal(err)
	}

	var s []containerState
	for _, item := range list {
		if item.IsDir() {
			container, err := factory.Load(item.Name())
			if err != nil {
				return nil, err
			}
			containerStatus, err := container.Status()
			if err != nil {
				return nil, err
			}
			state, err := container.State()
			if err != nil {
				return nil, err
			}
			bundle, annotations := utils.Annotations(state.Config.Labels)
			s = append(s, containerState{
				ID:             state.BaseState.ID,
				InitProcessPid: state.BaseState.InitProcessPid,
				Status:         containerStatus.String(),
				Bundle:         bundle,
				Created:        state.BaseState.Created,
				Annotations:    annotations,
			})
		}
	}
	return s, nil
}
