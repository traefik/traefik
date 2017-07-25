package integration

import (
	"bytes"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	lclient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/docker/container"
	"github.com/docker/libcompose/labels"

	. "gopkg.in/check.v1"
)

const (
	SimpleTemplate = `
        hello:
          image: busybox
          stdin_open: true
          tty: true
        `
	SimpleTemplateWithVols = `
        hello:
          image: busybox
          stdin_open: true
          tty: true
          volumes:
          - /root:/root
          - /home:/home
          - /var/lib/vol1
          - /var/lib/vol2
          - /var/lib/vol4
        `

	SimpleTemplateWithVols2 = `
        hello:
          image: busybox
          stdin_open: true
          tty: true
          volumes:
          - /tmp/tmp-root:/root
          - /var/lib/vol1
          - /var/lib/vol3
          - /var/lib/vol4
        `
)

func Test(t *testing.T) { TestingT(t) }

func init() {
	Suite(&CliSuite{
		command: "../bundles/libcompose-cli",
	})
}

type CliSuite struct {
	command  string
	projects []string
}

func (s *CliSuite) TearDownTest(c *C) {
	// Delete all containers
	client := GetClient(c)

	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	c.Assert(err, IsNil)
	for _, container := range containers {
		// Unpause container (if paused) and ignore error (if wasn't paused)
		client.ContainerUnpause(context.Background(), container.ID)
		// And remove force \o/
		err := client.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
		c.Assert(err, IsNil)
	}
}

func (s *CliSuite) CreateProjectFromText(c *C, input string) string {
	return s.ProjectFromText(c, "create", input)
}

func (s *CliSuite) RandomProject() string {
	return "testproject" + RandStr(7)
}

func (s *CliSuite) ProjectFromText(c *C, command, input string) string {
	projectName := s.RandomProject()
	return s.FromText(c, projectName, command, input)
}

func (s *CliSuite) FromText(c *C, projectName, command string, argsAndInput ...string) string {
	command, args, input := s.createCommand(c, projectName, command, argsAndInput)

	cmd := exec.Command(s.command, args...)
	cmd.Stdin = bytes.NewBufferString(strings.Replace(input, "\t", "  ", -1))
	if os.Getenv("TESTVERBOSE") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}

	err := cmd.Run()
	c.Assert(err, IsNil, Commentf("Failed to run %s %v: %v\n with input:\n%s", s.command, err, args, input))

	return projectName
}

// Doesn't assert that command runs successfully
func (s *CliSuite) FromTextCaptureOutput(c *C, projectName, command string, argsAndInput ...string) (string, string) {
	command, args, input := s.createCommand(c, projectName, command, argsAndInput)

	cmd := exec.Command(s.command, args...)
	cmd.Stdin = bytes.NewBufferString(strings.Replace(input, "\t", "  ", -1))

	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to run %s %v: %v\n with input:\n%s", s.command, err, args, input)
	}

	return projectName, string(output[:])
}

func (s *CliSuite) createCommand(c *C, projectName, command string, argsAndInput []string) (string, []string, string) {
	args := []string{"--verbose", "-p", projectName, "-f", "-", command}
	args = append(args, argsAndInput[0:len(argsAndInput)-1]...)

	input := argsAndInput[len(argsAndInput)-1]

	if command == "up" {
		args = append(args, "-d")
	} else if command == "restart" {
		args = append(args, "--timeout", "0")
	} else if command == "stop" {
		args = append(args, "--timeout", "0")
	}

	logrus.Infof("Running %s %v", command, args)

	return command, args, input
}

func GetClient(c *C) client.APIClient {
	client, err := lclient.Create(lclient.Options{})

	c.Assert(err, IsNil)

	return client
}

func (s *CliSuite) GetContainerByName(c *C, name string) *types.ContainerJSON {
	client := GetClient(c)
	container, err := container.Get(context.Background(), client, name)

	c.Assert(err, IsNil)

	return container
}

func (s *CliSuite) GetVolumeByName(c *C, name string) *types.Volume {
	client := GetClient(c)
	volume, err := client.VolumeInspect(context.Background(), name)

	c.Assert(err, IsNil)

	return &volume
}

func (s *CliSuite) GetContainersByProject(c *C, project string) []types.Container {
	client := GetClient(c)
	containers, err := container.ListByFilter(context.Background(), client, labels.PROJECT.Eq(project))

	c.Assert(err, IsNil)

	return containers
}

func asMap(items []string) map[string]bool {
	result := map[string]bool{}
	for _, item := range items {
		result[item] = true
	}
	return result
}

var random = rand.New(rand.NewSource(time.Now().Unix()))

func RandStr(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}
	return string(b)
}
