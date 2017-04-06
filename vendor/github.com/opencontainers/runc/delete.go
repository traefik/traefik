// +build !solaris

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/urfave/cli"
)

func killContainer(container libcontainer.Container) error {
	container.Signal(syscall.SIGKILL)
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		if err := container.Signal(syscall.Signal(0)); err != nil {
			destroy(container)
			return nil
		}
	}
	return fmt.Errorf("container init still running")
}

var deleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete any resources held by the container often used with detached containers",
	ArgsUsage: `<container-id>

Where "<container-id>" is the name for the instance of the container.

EXAMPLE:
For example, if the container id is "ubuntu01" and runc list currently shows the
status of "ubuntu01" as "stopped" the following will delete resources held for
"ubuntu01" removing "ubuntu01" from the runc list of containers:

       # runc delete ubuntu01`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Forcibly kills the container if it is still running",
		},
	},
	Action: func(context *cli.Context) error {
		container, err := getContainer(context)
		if err != nil {
			if lerr, ok := err.(libcontainer.Error); ok && lerr.Code() == libcontainer.ContainerNotExists {
				// if there was an aborted start or something of the sort then the container's directory could exist but
				// libcontainer does not see it because the state.json file inside that directory was never created.
				path := filepath.Join(context.GlobalString("root"), context.Args().First())
				if err := os.RemoveAll(path); err != nil {
					return err
				}
			}
			return nil
		}
		s, err := container.Status()
		if err != nil {
			return err
		}
		switch s {
		case libcontainer.Stopped:
			destroy(container)
		case libcontainer.Created:
			return killContainer(container)
		default:
			if context.Bool("force") {
				return killContainer(container)
			}
			return fmt.Errorf("cannot delete container that is not stopped: %s", s)
		}
		return nil
	},
}
