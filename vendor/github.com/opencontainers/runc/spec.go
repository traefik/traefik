// +build linux

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

var specCommand = cli.Command{
	Name:      "spec",
	Usage:     "create a new specification file",
	ArgsUsage: "",
	Description: `The spec command creates the new specification file named "` + specConfig + `" for
the bundle.

The spec generated is just a starter file. Editing of the spec is required to
achieve desired results. For example, the newly generated spec includes an args
parameter that is initially set to call the "sh" command when the container is
started. Calling "sh" may work for an ubuntu container or busybox, but will not
work for containers that do not include the "sh" program.

EXAMPLE:
  To run docker's hello-world container one needs to set the args parameter
in the spec to call hello. This can be done using the sed command or a text
editor. The following commands create a bundle for hello-world, change the
default args parameter in the spec from "sh" to "/hello", then run the hello
command in a new hello-world container named container1:

    mkdir hello
    cd hello
    docker pull hello-world
    docker export $(docker create hello-world) > hello-world.tar
    mkdir rootfs
    tar -C rootfs -xf hello-world.tar
    runc spec
    sed -i 's;"sh";"/hello";' ` + specConfig + `
    runc run container1

In the run command above, "container1" is the name for the instance of the
container that you are starting. The name you provide for the container instance
must be unique on your host.

An alternative for generating a customized spec config is to use "ocitools", the
sub-command "ocitools generate" has lots of options that can be used to do any
customizations as you want, see [ocitools](https://github.com/opencontainers/ocitools)
to get more information.

When starting a container through runc, runc needs root privilege. If not
already running as root, you can use sudo to give runc root privilege. For
example: "sudo runc start container1" will give runc root privilege to start the
container on your host.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: "path to the root of the bundle directory",
		},
	},
	Action: func(context *cli.Context) error {
		spec := specs.Spec{
			Version: specs.Version,
			Platform: specs.Platform{
				OS:   runtime.GOOS,
				Arch: runtime.GOARCH,
			},
			Root: specs.Root{
				Path:     "rootfs",
				Readonly: true,
			},
			Process: specs.Process{
				Terminal: true,
				User:     specs.User{},
				Args: []string{
					"sh",
				},
				Env: []string{
					"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
					"TERM=xterm",
				},
				Cwd:             "/",
				NoNewPrivileges: true,
				Capabilities: []string{
					"CAP_AUDIT_WRITE",
					"CAP_KILL",
					"CAP_NET_BIND_SERVICE",
				},
				Rlimits: []specs.Rlimit{
					{
						Type: "RLIMIT_NOFILE",
						Hard: uint64(1024),
						Soft: uint64(1024),
					},
				},
			},
			Hostname: "runc",
			Mounts: []specs.Mount{
				{
					Destination: "/proc",
					Type:        "proc",
					Source:      "proc",
					Options:     nil,
				},
				{
					Destination: "/dev",
					Type:        "tmpfs",
					Source:      "tmpfs",
					Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
				},
				{
					Destination: "/dev/pts",
					Type:        "devpts",
					Source:      "devpts",
					Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
				},
				{
					Destination: "/dev/shm",
					Type:        "tmpfs",
					Source:      "shm",
					Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
				},
				{
					Destination: "/dev/mqueue",
					Type:        "mqueue",
					Source:      "mqueue",
					Options:     []string{"nosuid", "noexec", "nodev"},
				},
				{
					Destination: "/sys",
					Type:        "sysfs",
					Source:      "sysfs",
					Options:     []string{"nosuid", "noexec", "nodev", "ro"},
				},
				{
					Destination: "/sys/fs/cgroup",
					Type:        "cgroup",
					Source:      "cgroup",
					Options:     []string{"nosuid", "noexec", "nodev", "relatime", "ro"},
				},
			},
			Linux: specs.Linux{
				MaskedPaths: []string{
					"/proc/kcore",
					"/proc/latency_stats",
					"/proc/timer_stats",
					"/proc/sched_debug",
				},
				ReadonlyPaths: []string{
					"/proc/asound",
					"/proc/bus",
					"/proc/fs",
					"/proc/irq",
					"/proc/sys",
					"/proc/sysrq-trigger",
				},
				Resources: &specs.Resources{
					Devices: []specs.DeviceCgroup{
						{
							Allow:  false,
							Access: sPtr("rwm"),
						},
					},
				},
				Namespaces: []specs.Namespace{
					{
						Type: "pid",
					},
					{
						Type: "network",
					},
					{
						Type: "ipc",
					},
					{
						Type: "uts",
					},
					{
						Type: "mount",
					},
				},
			},
		}

		checkNoFile := func(name string) error {
			_, err := os.Stat(name)
			if err == nil {
				return fmt.Errorf("File %s exists. Remove it first", name)
			}
			if !os.IsNotExist(err) {
				return err
			}
			return nil
		}
		bundle := context.String("bundle")
		if bundle != "" {
			if err := os.Chdir(bundle); err != nil {
				return err
			}
		}
		if err := checkNoFile(specConfig); err != nil {
			return err
		}
		data, err := json.MarshalIndent(&spec, "", "\t")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(specConfig, data, 0666); err != nil {
			return err
		}
		return nil
	},
}

func sPtr(s string) *string      { return &s }
func rPtr(r rune) *rune          { return &r }
func iPtr(i int64) *int64        { return &i }
func u32Ptr(i int64) *uint32     { u := uint32(i); return &u }
func fmPtr(i int64) *os.FileMode { fm := os.FileMode(i); return &fm }

// loadSpec loads the specification from the provided path.
func loadSpec(cPath string) (spec *specs.Spec, err error) {
	cf, err := os.Open(cPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file %s not found", cPath)
		}
		return nil, err
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&spec); err != nil {
		return nil, err
	}
	return spec, validateProcessSpec(&spec.Process)
}

func createLibContainerRlimit(rlimit specs.Rlimit) (configs.Rlimit, error) {
	rl, err := strToRlimit(rlimit.Type)
	if err != nil {
		return configs.Rlimit{}, err
	}
	return configs.Rlimit{
		Type: rl,
		Hard: uint64(rlimit.Hard),
		Soft: uint64(rlimit.Soft),
	}, nil
}
