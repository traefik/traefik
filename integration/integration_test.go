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

	"github.com/containous/traefik/integration/utils"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"

	checker "github.com/vdemeester/shakers"
	check "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

func init() {
	check.Suite(&SimpleSuite{})
	check.Suite(&HTTPSSuite{})
	check.Suite(&FileSuite{})
	check.Suite(&DockerSuite{})
	check.Suite(&ConsulSuite{})
	check.Suite(&ConsulCatalogSuite{})
	check.Suite(&EtcdSuite{})
	check.Suite(&MarathonSuite{})
}

var traefikBinary = "../dist/traefik"

type BaseSuite struct {
	composeProject *project.Project
	listenChan     chan project.Event
	started        chan bool
	stopped        chan bool
	deleted        chan bool
}

func (s *BaseSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Down()
		<-s.stopped
		defer close(s.stopped)

		s.composeProject.Delete()
		<-s.deleted
		defer close(s.deleted)
	}
}

func (s *BaseSuite) createComposeProject(c *check.C, name string) {
	composeProject, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeFiles: []string{
				fmt.Sprintf("resources/compose/%s.yml", name),
			},
			ProjectName: fmt.Sprintf("integration-test-%s", name),
		},
	})
	c.Assert(err, checker.IsNil)
	s.composeProject = composeProject

	err = composeProject.Create()
	c.Assert(err, checker.IsNil)

	s.started = make(chan bool)
	s.stopped = make(chan bool)
	s.deleted = make(chan bool)

	s.listenChan = make(chan project.Event)
	go s.startListening(c)

	composeProject.AddListener(s.listenChan)

	err = composeProject.Start()
	c.Assert(err, checker.IsNil)

	// Wait for compose to start
	<-s.started
	defer close(s.started)
}

func (s *BaseSuite) startListening(c *check.C) {
	for event := range s.listenChan {
		// FIXME Add a timeout on event ?
		if event.EventType == project.EventProjectStartDone {
			s.started <- true
		}
		if event.EventType == project.EventProjectDownDone {
			s.stopped <- true
		}
		if event.EventType == project.EventProjectDeleteDone {
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
