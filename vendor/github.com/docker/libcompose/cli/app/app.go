package app

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/docker/libcompose/version"
	"github.com/urfave/cli"
)

// ProjectAction is an adapter to allow the use of ordinary functions as libcompose actions.
// Any function that has the appropriate signature can be register as an action on a codegansta/cli command.
//
// cli.Command{
//		Name:   "ps",
//		Usage:  "List containers",
//		Action: app.WithProject(factory, app.ProjectPs),
//	}
type ProjectAction func(project project.APIProject, c *cli.Context) error

// BeforeApp is an action that is executed before any cli command.
func BeforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if version.ShowWarning() {
		logrus.Warning("Note: This is an experimental alternate implementation of the Compose CLI (https://github.com/docker/compose)")
	}
	return nil
}

// WithProject is a helper function to create a cli.Command action with a ProjectFactory.
func WithProject(factory ProjectFactory, action ProjectAction) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		p, err := factory.Create(context)
		if err != nil {
			logrus.Fatalf("Failed to read project: %v", err)
		}
		return action(p, context)
	}
}

// ProjectPs lists the containers.
func ProjectPs(p project.APIProject, c *cli.Context) error {
	qFlag := c.Bool("q")
	allInfo, err := p.Ps(context.Background(), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	columns := []string{"Name", "Command", "State", "Ports"}
	if qFlag {
		columns = []string{"Id"}
	}
	os.Stdout.WriteString(allInfo.String(columns, !qFlag))
	return nil
}

// ProjectPort prints the public port for a port binding.
func ProjectPort(p project.APIProject, c *cli.Context) error {
	if len(c.Args()) != 2 {
		return cli.NewExitError("Please pass arguments in the form: SERVICE PORT", 1)
	}

	index := c.Int("index")
	protocol := c.String("protocol")
	serviceName := c.Args()[0]
	privatePort := c.Args()[1]

	port, err := p.Port(context.Background(), index, protocol, serviceName, privatePort)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	fmt.Println(port)
	return nil
}

// ProjectStop stops all services.
func ProjectStop(p project.APIProject, c *cli.Context) error {
	err := p.Stop(context.Background(), c.Int("timeout"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectDown brings all services down (stops and clean containers).
func ProjectDown(p project.APIProject, c *cli.Context) error {
	options := options.Down{
		RemoveVolume:  c.Bool("volumes"),
		RemoveImages:  options.ImageType(c.String("rmi")),
		RemoveOrphans: c.Bool("remove-orphans"),
	}
	err := p.Down(context.Background(), options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectBuild builds or rebuilds services.
func ProjectBuild(p project.APIProject, c *cli.Context) error {
	config := options.Build{
		NoCache:     c.Bool("no-cache"),
		ForceRemove: c.Bool("force-rm"),
		Pull:        c.Bool("pull"),
	}
	err := p.Build(context.Background(), config, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectCreate creates all services but do not start them.
func ProjectCreate(p project.APIProject, c *cli.Context) error {
	options := options.Create{
		NoRecreate:    c.Bool("no-recreate"),
		ForceRecreate: c.Bool("force-recreate"),
		NoBuild:       c.Bool("no-build"),
	}
	err := p.Create(context.Background(), options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectUp brings all services up.
func ProjectUp(p project.APIProject, c *cli.Context) error {
	options := options.Up{
		Create: options.Create{
			NoRecreate:    c.Bool("no-recreate"),
			ForceRecreate: c.Bool("force-recreate"),
			NoBuild:       c.Bool("no-build"),
			ForceBuild:    c.Bool("build"),
		},
	}
	ctx, cancelFun := context.WithCancel(context.Background())
	err := p.Up(ctx, options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if !c.Bool("d") {
		signalChan := make(chan os.Signal, 1)
		cleanupDone := make(chan bool)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		errChan := make(chan error)
		go func() {
			errChan <- p.Log(ctx, true, c.Args()...)
		}()
		go func() {
			select {
			case <-signalChan:
				fmt.Printf("\nGracefully stopping...\n")
				cancelFun()
				ProjectStop(p, c)
				cleanupDone <- true
			case err := <-errChan:
				if err != nil {
					logrus.Fatal(err)
				}
				cleanupDone <- true
			}
		}()
		<-cleanupDone
		return nil
	}
	return nil
}

// ProjectRun runs a given command within a service's container.
func ProjectRun(p project.APIProject, c *cli.Context) error {
	if len(c.Args()) == 0 {
		logrus.Fatal("No service specified")
	}

	serviceName := c.Args()[0]
	commandParts := c.Args()[1:]

	options := options.Run{
		Detached: c.Bool("d"),
	}

	exitCode, err := p.Run(context.Background(), serviceName, commandParts, options)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return cli.NewExitError("", exitCode)
}

// ProjectStart starts services.
func ProjectStart(p project.APIProject, c *cli.Context) error {
	err := p.Start(context.Background(), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectRestart restarts services.
func ProjectRestart(p project.APIProject, c *cli.Context) error {
	err := p.Restart(context.Background(), c.Int("timeout"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectLog gets services logs.
func ProjectLog(p project.APIProject, c *cli.Context) error {
	err := p.Log(context.Background(), c.Bool("follow"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectPull pulls images for services.
func ProjectPull(p project.APIProject, c *cli.Context) error {
	err := p.Pull(context.Background(), c.Args()...)
	if err != nil && !c.Bool("ignore-pull-failures") {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectDelete deletes services.
func ProjectDelete(p project.APIProject, c *cli.Context) error {
	options := options.Delete{
		RemoveVolume: c.Bool("v"),
	}
	if !c.Bool("force") {
		stoppedContainers, err := p.Containers(context.Background(), project.Filter{
			State: project.Stopped,
		}, c.Args()...)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		if len(stoppedContainers) == 0 {
			fmt.Println("No stopped containers")
			return nil
		}
		fmt.Printf("Going to remove %v\nAre you sure? [yN]\n", strings.Join(stoppedContainers, ", "))
		var answer string
		_, err = fmt.Scanln(&answer)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		if answer != "y" && answer != "Y" {
			return nil
		}
	}
	err := p.Delete(context.Background(), options, c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectKill forces stop service containers.
func ProjectKill(p project.APIProject, c *cli.Context) error {
	err := p.Kill(context.Background(), c.String("signal"), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectConfig validates and print the compose file.
func ProjectConfig(p project.APIProject, c *cli.Context) error {
	yaml, err := p.Config()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if !c.Bool("quiet") {
		fmt.Println(yaml)
	}
	return nil
}

// ProjectPause pauses service containers.
func ProjectPause(p project.APIProject, c *cli.Context) error {
	err := p.Pause(context.Background(), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectUnpause unpauses service containers.
func ProjectUnpause(p project.APIProject, c *cli.Context) error {
	err := p.Unpause(context.Background(), c.Args()...)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

// ProjectScale scales services.
func ProjectScale(p project.APIProject, c *cli.Context) error {
	servicesScale := map[string]int{}
	for _, arg := range c.Args() {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			return cli.NewExitError(fmt.Sprintf("Invalid scale parameter: %s", arg), 2)
		}

		name := kv[0]

		count, err := strconv.Atoi(kv[1])
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("Invalid scale parameter: %v", err), 2)
		}

		servicesScale[name] = count
	}

	err := p.Scale(context.Background(), c.Int("timeout"), servicesScale)
	if err != nil {
		return cli.NewExitError(err.Error(), 0)
	}
	return nil
}
