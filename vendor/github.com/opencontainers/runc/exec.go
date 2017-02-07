// +build linux

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/opencontainers/runc/libcontainer/utils"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "execute new process inside the container",
	ArgsUsage: `<container-id> <container command>

Where "<container-id>" is the name for the instance of the container and
"<container command>" is the command to be executed in the container.

EXAMPLE:
For example, if the container is configured to run the linux ps command the
following will output a list of processes running in the container:
	 
       # runc exec <container-id> ps`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "console",
			Usage: "specify the pty slave path for use with the container",
		},
		cli.StringFlag{
			Name:  "cwd",
			Usage: "current working directory in the container",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "set environment variables",
		},
		cli.BoolFlag{
			Name:  "tty, t",
			Usage: "allocate a pseudo-TTY",
		},
		cli.StringFlag{
			Name:  "user, u",
			Usage: "UID (format: <uid>[:<gid>])",
		},
		cli.StringFlag{
			Name:  "process, p",
			Usage: "path to the process.json",
		},
		cli.BoolFlag{
			Name:  "detach,d",
			Usage: "detach from the container's process",
		},
		cli.StringFlag{
			Name:  "pid-file",
			Value: "",
			Usage: "specify the file to write the process id to",
		},
		cli.StringFlag{
			Name:  "process-label",
			Usage: "set the asm process label for the process commonly used with selinux",
		},
		cli.StringFlag{
			Name:  "apparmor",
			Usage: "set the apparmor profile for the process",
		},
		cli.BoolFlag{
			Name:  "no-new-privs",
			Usage: "set the no new privileges value for the process",
		},
		cli.StringSliceFlag{
			Name:  "cap, c",
			Value: &cli.StringSlice{},
			Usage: "add a capability to the bounding set for the process",
		},
		cli.BoolFlag{
			Name:  "no-subreaper",
			Usage: "disable the use of the subreaper used to reap reparented processes",
		},
	},
	Action: func(context *cli.Context) error {
		if os.Geteuid() != 0 {
			return fmt.Errorf("runc should be run as root")
		}
		status, err := execProcess(context)
		if err == nil {
			os.Exit(status)
		}
		return fmt.Errorf("exec failed: %v", err)
	},
}

func execProcess(context *cli.Context) (int, error) {
	container, err := getContainer(context)
	if err != nil {
		return -1, err
	}
	path := context.String("process")
	if path == "" && len(context.Args()) == 1 {
		return -1, fmt.Errorf("process args cannot be empty")
	}
	detach := context.Bool("detach")
	state, err := container.State()
	if err != nil {
		return -1, err
	}
	bundle := utils.SearchLabels(state.Config.Labels, "bundle")
	p, err := getProcess(context, bundle)
	if err != nil {
		return -1, err
	}
	r := &runner{
		enableSubreaper: !context.Bool("no-subreaper"),
		shouldDestroy:   false,
		container:       container,
		console:         context.String("console"),
		detach:          detach,
		pidFile:         context.String("pid-file"),
	}
	return r.run(p)
}

func getProcess(context *cli.Context, bundle string) (*specs.Process, error) {
	if path := context.String("process"); path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var p specs.Process
		if err := json.NewDecoder(f).Decode(&p); err != nil {
			return nil, err
		}
		return &p, validateProcessSpec(&p)
	}
	// process via cli flags
	if err := os.Chdir(bundle); err != nil {
		return nil, err
	}
	spec, err := loadSpec(specConfig)
	if err != nil {
		return nil, err
	}
	p := spec.Process
	p.Args = context.Args()[1:]
	// override the cwd, if passed
	if context.String("cwd") != "" {
		p.Cwd = context.String("cwd")
	}
	if ap := context.String("apparmor"); ap != "" {
		p.ApparmorProfile = ap
	}
	if l := context.String("process-label"); l != "" {
		p.SelinuxLabel = l
	}
	if caps := context.StringSlice("cap"); len(caps) > 0 {
		p.Capabilities = caps
	}
	// append the passed env variables
	for _, e := range context.StringSlice("env") {
		p.Env = append(p.Env, e)
	}
	// set the tty
	if context.IsSet("tty") {
		p.Terminal = context.Bool("tty")
	}
	if context.IsSet("no-new-privs") {
		p.NoNewPrivileges = context.Bool("no-new-privs")
	}
	// override the user, if passed
	if context.String("user") != "" {
		u := strings.SplitN(context.String("user"), ":", 2)
		if len(u) > 1 {
			gid, err := strconv.Atoi(u[1])
			if err != nil {
				return nil, fmt.Errorf("parsing %s as int for gid failed: %v", u[1], err)
			}
			p.User.GID = uint32(gid)
		}
		uid, err := strconv.Atoi(u[0])
		if err != nil {
			return nil, fmt.Errorf("parsing %s as int for uid failed: %v", u[0], err)
		}
		p.User.UID = uint32(uid)
	}
	return &p, nil
}
