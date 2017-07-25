package command

import (
	"os"
	"strings"

	"github.com/docker/libcompose/cli/app"
	"github.com/docker/libcompose/project"
	"github.com/urfave/cli"
)

// Populate updates the specified project context based on command line arguments and subcommands.
func Populate(context *project.Context, c *cli.Context) {
	// urfave/cli does not distinguish whether the first string in the slice comes from the envvar
	// or is from a flag. Worse off, it appends the flag values to the envvar value instead of
	// overriding it. To ensure the multifile envvar case is always handled, the first string
	// must always be split. It gives a more consistent behavior, then, to split each string in
	// the slice.
	for _, v := range c.GlobalStringSlice("file") {
		context.ComposeFiles = append(context.ComposeFiles, strings.Split(v, string(os.PathListSeparator))...)
	}

	if len(context.ComposeFiles) == 0 {
		context.ComposeFiles = []string{"docker-compose.yml"}
		if _, err := os.Stat("docker-compose.override.yml"); err == nil {
			context.ComposeFiles = append(context.ComposeFiles, "docker-compose.override.yml")
		}
	}

	context.ProjectName = c.GlobalString("project-name")
}

// CreateCommand defines the libcompose create subcommand.
func CreateCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "Create all services but do not start",
		Action: app.WithProject(factory, app.ProjectCreate),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-recreate",
				Usage: "If containers already exist, don't recreate them. Incompatible with --force-recreate.",
			},
			cli.BoolFlag{
				Name:  "force-recreate",
				Usage: "Recreate containers even if their configuration and image haven't changed. Incompatible with --no-recreate.",
			},
			cli.BoolFlag{
				Name:  "no-build",
				Usage: "Don't build an image, even if it's missing.",
			},
		},
	}
}

// ConfigCommand defines the libcompose config subcommand
func ConfigCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "config",
		Usage:  "Validate and view the compose file.",
		Action: app.WithProject(factory, app.ProjectConfig),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet,q",
				Usage: "Only validate the configuration, don't print anything.",
			},
		},
	}
}

// BuildCommand defines the libcompose build subcommand.
func BuildCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "build",
		Usage:  "Build or rebuild services.",
		Action: app.WithProject(factory, app.ProjectBuild),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-cache",
				Usage: "Do not use cache when building the image",
			},
			cli.BoolFlag{
				Name:  "force-rm",
				Usage: "Always remove intermediate containers",
			},
			cli.BoolFlag{
				Name:  "pull",
				Usage: "Always attempt to pull a newer version of the image",
			},
		},
	}
}

// PsCommand defines the libcompose ps subcommand.
func PsCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "List containers",
		Action: app.WithProject(factory, app.ProjectPs),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "q",
				Usage: "Only display IDs",
			},
		},
	}
}

// PortCommand defines the libcompose port subcommand.
func PortCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "port",
		Usage:  "Print the public port for a port binding",
		Action: app.WithProject(factory, app.ProjectPort),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "protocol",
				Usage: "tcp or udp ",
				Value: "tcp",
			},
			cli.IntFlag{
				Name:  "index",
				Usage: "index of the container if there are multiple instances of a service",
				Value: 1,
			},
		},
	}
}

// UpCommand defines the libcompose up subcommand.
func UpCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Bring all services up",
		Action: app.WithProject(factory, app.ProjectUp),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "d",
				Usage: "Do not block and log",
			},
			cli.BoolFlag{
				Name:  "no-build",
				Usage: "Don't build an image, even if it's missing.",
			},
			cli.BoolFlag{
				Name:  "no-recreate",
				Usage: "If containers already exist, don't recreate them. Incompatible with --force-recreate.",
			},
			cli.BoolFlag{
				Name:  "force-recreate",
				Usage: "Recreate containers even if their configuration and image haven't changed. Incompatible with --no-recreate.",
			},
			cli.BoolFlag{
				Name:  "build",
				Usage: "Build images before starting containers.",
			},
		},
	}
}

// StartCommand defines the libcompose start subcommand.
func StartCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "Start services",
		Action: app.WithProject(factory, app.ProjectStart),
		Flags: []cli.Flag{
			cli.BoolTFlag{
				Name:  "d",
				Usage: "Do not block and log",
			},
		},
	}
}

// RunCommand defines the libcompose run subcommand.
func RunCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "run",
		Usage:  "Run a one-off command",
		Action: app.WithProject(factory, app.ProjectRun),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "d",
				Usage: "Detached mode: Run container in the background, print new container name.",
			},
		},
	}
}

// PullCommand defines the libcompose pull subcommand.
func PullCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "pull",
		Usage:  "Pulls images for services",
		Action: app.WithProject(factory, app.ProjectPull),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "ignore-pull-failures",
				Usage: "Pull what it can and ignores images with pull failures.",
			},
		},
	}
}

// LogsCommand defines the libcompose logs subcommand.
func LogsCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "logs",
		Usage:  "Get service logs",
		Action: app.WithProject(factory, app.ProjectLog),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "follow",
				Usage: "Follow log output.",
			},
		},
	}
}

// RestartCommand defines the libcompose restart subcommand.
func RestartCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "restart",
		Usage:  "Restart services",
		Action: app.WithProject(factory, app.ProjectRestart),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
			},
		},
	}
}

// StopCommand defines the libcompose stop subcommand.
func StopCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "stop",
		Usage:  "Stop services",
		Action: app.WithProject(factory, app.ProjectStop),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
			},
		},
	}
}

// DownCommand defines the libcompose stop subcommand.
func DownCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "down",
		Usage:  "Stop and remove containers, networks, images, and volumes",
		Action: app.WithProject(factory, app.ProjectDown),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "volumes,v",
				Usage: "Remove data volumes",
			},
			cli.StringFlag{
				Name:  "rmi",
				Usage: "Remove images, type may be one of: 'all' to remove all images, or 'local' to remove only images that don't have an custom name set by the `image` field",
			},
			cli.BoolFlag{
				Name:  "remove-orphans",
				Usage: "Remove containers for services not defined in the Compose file",
			},
		},
	}
}

// ScaleCommand defines the libcompose scale subcommand.
func ScaleCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "scale",
		Usage:  "Scale services",
		Action: app.WithProject(factory, app.ProjectScale),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
			},
		},
	}
}

// RmCommand defines the libcompose rm subcommand.
func RmCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "rm",
		Usage:  "Delete services",
		Action: app.WithProject(factory, app.ProjectDelete),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force,f",
				Usage: "Allow deletion of all services",
			},
			cli.BoolFlag{
				Name:  "v",
				Usage: "Remove volumes associated with containers",
			},
		},
	}
}

// KillCommand defines the libcompose kill subcommand.
func KillCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "kill",
		Usage:  "Force stop service containers",
		Action: app.WithProject(factory, app.ProjectKill),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "signal,s",
				Usage: "SIGNAL to send to the container",
				Value: "SIGKILL",
			},
		},
	}
}

// PauseCommand defines the libcompose pause subcommand.
func PauseCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:  "pause",
		Usage: "Pause services.",
		// ArgsUsage: "[SERVICE...]",
		Action: app.WithProject(factory, app.ProjectPause),
	}
}

// UnpauseCommand defines the libcompose unpause subcommand.
func UnpauseCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:  "unpause",
		Usage: "Unpause services.",
		// ArgsUsage: "[SERVICE...]",
		Action: app.WithProject(factory, app.ProjectUnpause),
	}
}

// EventsCommand defines the libcompose events subcommand
func EventsCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "events",
		Usage:  "Receive real time events from containers.",
		Action: app.WithProject(factory, app.ProjectEvents),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "json",
				Usage: "Output events as a stream of json objects",
			},
		},
	}
}

// VersionCommand defines the libcompose version subcommand.
func VersionCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "version",
		Usage:  "Show version informations",
		Action: app.Version,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "short",
				Usage: "Shows only Compose's version number.",
			},
		},
	}
}

// CommonFlags defines the flags that are in common for all subcommands.
func CommonFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name: "verbose,debug",
		},
		cli.StringSliceFlag{
			Name:   "file,f",
			Usage:  "Specify one or more alternate compose files (default: docker-compose.yml)",
			Value:  &cli.StringSlice{},
			EnvVar: "COMPOSE_FILE",
		},
		cli.StringFlag{
			Name:   "project-name,p",
			Usage:  "Specify an alternate project name (default: directory name)",
			EnvVar: "COMPOSE_PROJECT_NAME",
		},
	}
}
