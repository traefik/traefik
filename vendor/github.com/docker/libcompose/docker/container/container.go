package container

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/labels"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project"
)

// Container holds information about a docker container and the service it is tied on.
type Container struct {
	client    client.ContainerAPIClient
	id        string
	container *types.ContainerJSON
}

// Create creates a container and return a Container struct (and an error if any)
func Create(ctx context.Context, client client.ContainerAPIClient, name string, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) (*Container, error) {
	container, err := client.ContainerCreate(ctx, config, hostConfig, networkingConfig, name)
	if err != nil {
		return nil, err
	}
	return New(ctx, client, container.ID)
}

// New creates a container struct with the specified client, id and name
func New(ctx context.Context, client client.ContainerAPIClient, id string) (*Container, error) {
	container, err := Get(ctx, client, id)
	if err != nil {
		return nil, err
	}
	return &Container{
		client:    client,
		id:        id,
		container: container,
	}, nil
}

// NewInspected creates a container struct from an inspected container
func NewInspected(client client.ContainerAPIClient, container *types.ContainerJSON) *Container {
	return &Container{
		client:    client,
		id:        container.ID,
		container: container,
	}
}

// Info returns info about the container, like name, command, state or ports.
func (c *Container) Info(ctx context.Context) (project.Info, error) {
	infos, err := ListByFilter(ctx, c.client, map[string][]string{
		"name": {c.container.Name},
	})
	if err != nil || len(infos) == 0 {
		return nil, err
	}
	info := infos[0]

	result := project.Info{}
	result["Id"] = c.container.ID
	result["Name"] = name(info.Names)
	result["Command"] = info.Command
	result["State"] = info.Status
	result["Ports"] = portString(info.Ports)

	return result, nil
}

func portString(ports []types.Port) string {
	result := []string{}

	for _, port := range ports {
		if port.PublicPort > 0 {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		} else {
			result = append(result, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		}
	}

	return strings.Join(result, ", ")
}

func name(names []string) string {
	max := math.MaxInt32
	var current string

	for _, v := range names {
		if len(v) < max {
			max = len(v)
			current = v
		}
	}

	return current[1:]
}

// Rename rename the container.
func (c *Container) Rename(ctx context.Context, newName string) error {
	return c.client.ContainerRename(ctx, c.container.ID, newName)
}

// Remove removes the container.
func (c *Container) Remove(ctx context.Context, removeVolume bool) error {
	return c.client.ContainerRemove(ctx, c.container.ID, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: removeVolume,
	})
}

// Stop stops the container.
func (c *Container) Stop(ctx context.Context, timeout int) error {
	timeoutDuration := time.Duration(timeout) * time.Second
	return c.client.ContainerStop(ctx, c.container.ID, &timeoutDuration)
}

// Pause pauses the container. If the containers are already paused, don't fail.
func (c *Container) Pause(ctx context.Context) error {
	if !c.container.State.Paused {
		if err := c.client.ContainerPause(ctx, c.container.ID); err != nil {
			return err
		}
		return c.updateInnerContainer(ctx)
	}
	return nil
}

// Unpause unpauses the container. If the containers are not paused, don't fail.
func (c *Container) Unpause(ctx context.Context) error {
	if c.container.State.Paused {
		if err := c.client.ContainerUnpause(ctx, c.container.ID); err != nil {
			return err
		}
		return c.updateInnerContainer(ctx)
	}
	return nil
}

func (c *Container) updateInnerContainer(ctx context.Context) error {
	container, err := Get(ctx, c.client, c.container.ID)
	if err != nil {
		return err
	}
	c.container = container
	return nil
}

// Kill kill the container.
func (c *Container) Kill(ctx context.Context, signal string) error {
	return c.client.ContainerKill(ctx, c.container.ID, signal)
}

// IsRunning returns the running state of the container.
func (c *Container) IsRunning(ctx context.Context) bool {
	return c.container.State.Running
}

// Run creates, start and attach to the container based on the image name,
// the specified configuration.
// It will always create a new container.
func (c *Container) Run(ctx context.Context, configOverride *config.ServiceConfig) (int, error) {
	var (
		errCh       chan error
		out, stderr io.Writer
		in          io.ReadCloser
	)

	if configOverride.StdinOpen {
		in = os.Stdin
	}
	if configOverride.Tty {
		out = os.Stdout
		stderr = os.Stderr
	}

	options := types.ContainerAttachOptions{
		Stream: true,
		Stdin:  configOverride.StdinOpen,
		Stdout: configOverride.Tty,
		Stderr: configOverride.Tty,
	}

	resp, err := c.client.ContainerAttach(ctx, c.container.ID, options)
	if err != nil {
		return -1, err
	}

	// set raw terminal
	inFd, _ := term.GetFdInfo(in)
	state, err := term.SetRawTerminal(inFd)
	if err != nil {
		return -1, err
	}
	// restore raw terminal
	defer term.RestoreTerminal(inFd, state)
	// holdHijackedConnection (in goroutine)
	errCh = promise.Go(func() error {
		return holdHijackedConnection(configOverride.Tty, in, out, stderr, resp)
	})

	if err := c.client.ContainerStart(ctx, c.container.ID, types.ContainerStartOptions{}); err != nil {
		return -1, err
	}

	if configOverride.Tty {
		ws, err := term.GetWinsize(inFd)
		if err != nil {
			return -1, err
		}

		resizeOpts := types.ResizeOptions{
			Height: uint(ws.Height),
			Width:  uint(ws.Width),
		}

		if err := c.client.ContainerResize(ctx, c.container.ID, resizeOpts); err != nil {
			return -1, err
		}
	}

	if err := <-errCh; err != nil {
		logrus.Debugf("Error hijack: %s", err)
		return -1, err
	}

	exitedContainer, err := c.client.ContainerInspect(ctx, c.container.ID)
	if err != nil {
		return -1, err
	}

	return exitedContainer.State.ExitCode, nil
}

func holdHijackedConnection(tty bool, inputStream io.ReadCloser, outputStream, errorStream io.Writer, resp types.HijackedResponse) error {
	var err error
	receiveStdout := make(chan error, 1)
	if outputStream != nil || errorStream != nil {
		go func() {
			// When TTY is ON, use regular copy
			if tty && outputStream != nil {
				_, err = io.Copy(outputStream, resp.Reader)
			} else {
				_, err = stdcopy.StdCopy(outputStream, errorStream, resp.Reader)
			}
			logrus.Debugf("[hijack] End of stdout")
			receiveStdout <- err
		}()
	}

	stdinDone := make(chan struct{})
	go func() {
		if inputStream != nil {
			io.Copy(resp.Conn, inputStream)
			logrus.Debugf("[hijack] End of stdin")
		}

		if err := resp.CloseWrite(); err != nil {
			logrus.Debugf("Couldn't send EOF: %s", err)
		}
		close(stdinDone)
	}()

	select {
	case err := <-receiveStdout:
		if err != nil {
			logrus.Debugf("Error receiveStdout: %s", err)
			return err
		}
	case <-stdinDone:
		if outputStream != nil || errorStream != nil {
			if err := <-receiveStdout; err != nil {
				logrus.Debugf("Error receiveStdout: %s", err)
				return err
			}
		}
	}

	return nil
}

// Start the specified container with the specified host config
func (c *Container) Start(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{"container.ID": c.container.ID, "container.Name": c.container.Name}).Debug("Starting container")
	if err := c.client.ContainerStart(ctx, c.container.ID, types.ContainerStartOptions{}); err != nil {
		logrus.WithFields(logrus.Fields{"container.ID": c.container.ID, "container.Name": c.container.Name}).Debug("Failed to start container")
		return err
	}
	return nil
}

// Restart restarts the container if existing, does nothing otherwise.
func (c *Container) Restart(ctx context.Context, timeout int) error {
	timeoutDuration := time.Duration(timeout) * time.Second
	return c.client.ContainerRestart(ctx, c.container.ID, &timeoutDuration)
}

// Log forwards container logs to the project configured logger.
func (c *Container) Log(ctx context.Context, l logger.Logger, follow bool) error {
	info, err := c.client.ContainerInspect(ctx, c.container.ID)
	if err != nil {
		return err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       "all",
	}
	responseBody, err := c.client.ContainerLogs(ctx, c.container.ID, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if info.Config.Tty {
		_, err = io.Copy(&logger.Wrapper{Logger: l}, responseBody)
	} else {
		_, err = stdcopy.StdCopy(&logger.Wrapper{Logger: l}, &logger.Wrapper{Logger: l, Err: true}, responseBody)
	}
	logrus.WithFields(logrus.Fields{"Logger": l, "err": err}).Debug("c.client.Logs() returned error")

	return err
}

// Port returns the host port the specified port is mapped on.
func (c *Container) Port(ctx context.Context, port string) (string, error) {
	if bindings, ok := c.container.NetworkSettings.Ports[nat.Port(port)]; ok {
		result := []string{}
		for _, binding := range bindings {
			result = append(result, binding.HostIP+":"+binding.HostPort)
		}

		return strings.Join(result, "\n"), nil
	}
	return "", nil
}

// Networks returns the containers network
func (c *Container) Networks() (map[string]*network.EndpointSettings, error) {
	return c.container.NetworkSettings.Networks, nil
}

// ID returns the container Id.
func (c *Container) ID() string {
	return c.container.ID
}

// ShortID return the container Id in its short form
func (c *Container) ShortID() string {
	return c.container.ID[:12]
}

// Name returns the container name.
func (c *Container) Name() string {
	return c.container.Name
}

// Image returns the container image. Depending on the engine version its either
// the complete id or the digest reference the image.
func (c *Container) Image() string {
	return c.container.Image
}

// ImageConfig returns the container image stored in the config. It's the
// human-readable name of the image.
func (c *Container) ImageConfig() string {
	return c.container.Config.Image
}

// Hash returns the container hash stored as label.
func (c *Container) Hash() string {
	return c.container.Config.Labels[labels.HASH.Str()]
}

// Number returns the container number stored as label.
func (c *Container) Number() (int, error) {
	numberStr := c.container.Config.Labels[labels.NUMBER.Str()]
	return strconv.Atoi(numberStr)
}
