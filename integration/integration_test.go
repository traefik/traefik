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
	"github.com/go-check/check"

	compose "github.com/libkermit/compose/check"
	checker "github.com/vdemeester/shakers"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

func init() {
	check.Suite(&SimpleSuite{})
	check.Suite(&AccessLogSuite{})
	check.Suite(&HTTPSSuite{})
	check.Suite(&FileSuite{})
	check.Suite(&DockerSuite{})
	check.Suite(&ConsulSuite{})
	check.Suite(&ConsulCatalogSuite{})
	check.Suite(&EtcdSuite{})
	check.Suite(&MarathonSuite{})
	check.Suite(&ConstraintSuite{})
	check.Suite(&MesosSuite{})
}

var traefikBinary = "../dist/traefik"

type BaseSuite struct {
	composeProject *compose.Project
}

func (s *BaseSuite) TearDownSuite(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *BaseSuite) createComposeProject(c *check.C, name string) {
	projectName := fmt.Sprintf("integration-test-%s", name)
	composeFile := fmt.Sprintf("resources/compose/%s.yml", name)
	s.composeProject = compose.CreateProject(c, projectName, composeFile)
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
	tempObjects := struct{ DockerHost string }{dockerHost}
	return s.adaptFile(c, path, tempObjects)
}

func (s *BaseSuite) adaptFile(c *check.C, path string, tempObjects interface{}) string {
	// Load file
	tmpl, err := template.ParseFiles(path)
	c.Assert(err, checker.IsNil)

	folder, prefix := filepath.Split(path)
	tmpFile, err := ioutil.TempFile(folder, prefix)
	c.Assert(err, checker.IsNil)
	defer tmpFile.Close()

	err = tmpl.ExecuteTemplate(tmpFile, prefix, tempObjects)
	c.Assert(err, checker.IsNil)
	err = tmpFile.Sync()

	c.Assert(err, checker.IsNil)

	return tmpFile.Name()
}
