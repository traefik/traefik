// This is the main file that sets up integration tests using go-check.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/emilevauge/traefik/integration/utils"

	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

func init() {
	check.Suite(&SimpleSuite{})
	check.Suite(&FileSuite{})
	check.Suite(&DockerSuite{})
	check.Suite(&ConsulSuite{})
	check.Suite(&MarathonSuite{})
}

var traefikBinary = "../dist/traefik"

// SimpleSuite
type SimpleSuite struct{ BaseSuite }

// File test suites
type FileSuite struct{ BaseSuite }

func (s *FileSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "file")

	s.composeProject.Up()
}

// Docker test suites
type DockerSuite struct{ BaseSuite }

func (s *DockerSuite) SetUpSuite(c *check.C) {
	// Make sure we can speak to docker
}

func (s *DockerSuite) TearDownSuite(c *check.C) {
	// Clean the mess
}

// Consul test suites (using libcompose)
type ConsulSuite struct{ BaseSuite }

func (s *ConsulSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "consul")
}

// Marathon test suites (using libcompose)
type MarathonSuite struct{ BaseSuite }

func (s *MarathonSuite) SetUpSuite(c *check.C) {
	s.createComposeProject(c, "marathon")
}

type BaseSuite struct {
	composeProject *project.Project
	listenChan     chan project.ProjectEvent
	started        chan bool
	stopped        chan bool
	deleted        chan bool
}

func (s *BaseSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Down()
		// Waiting for libcompose#55 to be merged
		// <-s.stopped
		time.Sleep(2 * time.Second)

		s.composeProject.Delete()
		// Waiting for libcompose#55 to be merged
		// <-s.deleted
		time.Sleep(2 * time.Second)
	}
}

func (s *BaseSuite) createComposeProject(c *check.C, name string) {
	composeProject, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeFile: fmt.Sprintf("resources/compose/%s.yml", name),
			ProjectName: fmt.Sprintf("integration-test-%s", name),
		},
	})
	c.Assert(err, checker.IsNil)
	s.composeProject = composeProject

	s.listenChan = make(chan project.ProjectEvent)
	go s.startListening(c)

	composeProject.AddListener(s.listenChan)

	composeProject.Start()

	// FIXME Wait for compose to start
	// Waiting for libcompose#55 to be merged
	// <-s.started
	time.Sleep(2 * time.Second)

}

func (s *BaseSuite) startListening(c *check.C) {
	for event := range s.listenChan {
		// FIXME Remove this when it's working (libcompose#55)
		// fmt.Fprintf(os.Stdout, "Event: %s (%v)\n", event.Event, event)
		// FIXME Add a timeout on event
		if event.Event == project.PROJECT_UP_DONE {
			s.started <- true
		}
		if event.Event == project.PROJECT_DOWN_DONE {
			s.stopped <- true
		}
		if event.Event == project.PROJECT_DELETE_DONE {
			s.deleted <- true
		}
	}
}

func (s *BaseSuite) traefikCmd(c *check.C, args ...string) (*exec.Cmd, string) {
	cmd, out, err := utils.RunCommand(traefikBinary, args...)
	c.Assert(err, checker.IsNil, check.Commentf("Fail to run %s with %v", traefikBinary, args))
	return cmd, out
}

func (s *BaseSuite) adaptFileForHost(c *check.C, path string) string {
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		// Default docker socket
		dockerHost = "unix:///var/run/docker.sock"
	}

	// Load file
	tmpl, err := template.ParseFiles(path)
	c.Assert(err, checker.IsNil)

	folder, prefix := filepath.Split(path)
	tmpFile, err := ioutil.TempFile(folder, prefix)
	c.Assert(err, checker.IsNil)
	defer tmpFile.Close()

	err = tmpl.ExecuteTemplate(tmpFile, prefix, struct{ DockerHost string }{dockerHost})
	c.Assert(err, checker.IsNil)
	err = tmpFile.Sync()

	c.Assert(err, checker.IsNil)

	return tmpFile.Name()
}
