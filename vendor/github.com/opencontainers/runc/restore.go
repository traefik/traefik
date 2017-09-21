// +build linux

package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

var restoreCommand = cli.Command{
	Name:  "restore",
	Usage: "restore a container from a previous checkpoint",
	ArgsUsage: `<container-id>

Where "<container-id>" is the name for the instance of the container to be
restored.`,
	Description: `Restores the saved state of the container instance that was previously saved
using the runc checkpoint command.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "image-path",
			Value: "",
			Usage: "path to criu image files for restoring",
		},
		cli.StringFlag{
			Name:  "work-path",
			Value: "",
			Usage: "path for saving work files and logs",
		},
		cli.BoolFlag{
			Name:  "tcp-established",
			Usage: "allow open tcp connections",
		},
		cli.BoolFlag{
			Name:  "ext-unix-sk",
			Usage: "allow external unix sockets",
		},
		cli.BoolFlag{
			Name:  "shell-job",
			Usage: "allow shell jobs",
		},
		cli.BoolFlag{
			Name:  "file-locks",
			Usage: "handle file locks, for safety",
		},
		cli.StringFlag{
			Name:  "manage-cgroups-mode",
			Value: "",
			Usage: "cgroups mode: 'soft' (default), 'full' and 'strict'",
		},
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: "path to the root of the bundle directory",
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
		cli.BoolFlag{
			Name:  "no-subreaper",
			Usage: "disable the use of the subreaper used to reap reparented processes",
		},
		cli.BoolFlag{
			Name:  "no-pivot",
			Usage: "do not use pivot root to jail process inside rootfs.  This should be used whenever the rootfs is on top of a ramdisk",
		},
		cli.StringSliceFlag{
			Name:  "empty-ns",
			Usage: "create a namespace, but don't restore its properties",
		},
	},
	Action: func(context *cli.Context) error {
		if err := checkArgs(context, 1, exactArgs); err != nil {
			return err
		}
		// XXX: Currently this is untested with rootless containers.
		if isRootless() {
			return fmt.Errorf("runc restore requires root")
		}

		imagePath := context.String("image-path")
		id := context.Args().First()
		if id == "" {
			return errEmptyID
		}
		if imagePath == "" {
			imagePath = getDefaultImagePath(context)
		}
		bundle := context.String("bundle")
		if bundle != "" {
			if err := os.Chdir(bundle); err != nil {
				return err
			}
		}
		spec, err := loadSpec(specConfig)
		if err != nil {
			return err
		}
		config, err := specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
			CgroupName:       id,
			UseSystemdCgroup: context.GlobalBool("systemd-cgroup"),
			NoPivotRoot:      context.Bool("no-pivot"),
			Spec:             spec,
		})
		if err != nil {
			return err
		}
		status, err := restoreContainer(context, spec, config, imagePath)
		if err == nil {
			os.Exit(status)
		}
		return err
	},
}

func restoreContainer(context *cli.Context, spec *specs.Spec, config *configs.Config, imagePath string) (int, error) {
	var (
		rootuid = 0
		rootgid = 0
		id      = context.Args().First()
	)
	factory, err := loadFactory(context)
	if err != nil {
		return -1, err
	}
	container, err := factory.Load(id)
	if err != nil {
		container, err = factory.Create(id, config)
		if err != nil {
			return -1, err
		}
	}
	options := criuOptions(context)

	status, err := container.Status()
	if err != nil {
		logrus.Error(err)
	}
	if status == libcontainer.Running {
		fatalf("Container with id %s already running", id)
	}

	setManageCgroupsMode(context, options)

	if err = setEmptyNsMask(context, options); err != nil {
		return -1, err
	}

	// ensure that the container is always removed if we were the process
	// that created it.
	detach := context.Bool("detach")
	if !detach {
		defer destroy(container)
	}
	process := &libcontainer.Process{}
	tty, err := setupIO(process, rootuid, rootgid, false, detach, "")
	if err != nil {
		return -1, err
	}

	notifySocket := newNotifySocket(context, os.Getenv("NOTIFY_SOCKET"), id)
	if notifySocket != nil {
		notifySocket.setupSpec(context, spec)
		notifySocket.setupSocket()
	}

	handler := newSignalHandler(!context.Bool("no-subreaper"), notifySocket)
	if err := container.Restore(process, options); err != nil {
		return -1, err
	}
	// We don't need to do a tty.recvtty because config.Terminal is always false.
	defer tty.Close()
	if err := tty.ClosePostStart(); err != nil {
		return -1, err
	}
	if pidFile := context.String("pid-file"); pidFile != "" {
		if err := createPidFile(pidFile, process); err != nil {
			_ = process.Signal(syscall.SIGKILL)
			_, _ = process.Wait()
			return -1, err
		}
	}
	return handler.forward(process, tty, detach)
}

func criuOptions(context *cli.Context) *libcontainer.CriuOpts {
	imagePath := getCheckpointImagePath(context)
	if err := os.MkdirAll(imagePath, 0655); err != nil {
		fatal(err)
	}
	return &libcontainer.CriuOpts{
		ImagesDirectory:         imagePath,
		WorkDirectory:           context.String("work-path"),
		ParentImage:             context.String("parent-path"),
		LeaveRunning:            context.Bool("leave-running"),
		TcpEstablished:          context.Bool("tcp-established"),
		ExternalUnixConnections: context.Bool("ext-unix-sk"),
		ShellJob:                context.Bool("shell-job"),
		FileLocks:               context.Bool("file-locks"),
		PreDump:                 context.Bool("pre-dump"),
	}
}
