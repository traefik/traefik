package integration

import (
	"fmt"
	"os/exec"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestVolumeFromService(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/regression/60-volume_from.yml", "-p", p, "create")
	err := cmd.Run()
	c.Assert(err, IsNil)

	volumeFromContainer := fmt.Sprintf("%s_%s_1", p, "first")
	secondContainerName := p + "_second_1"

	cn := s.GetContainerByName(c, secondContainerName)
	c.Assert(cn, NotNil)

	c.Assert(len(cn.HostConfig.VolumesFrom), Equals, 1)
	c.Assert(cn.HostConfig.VolumesFrom[0], Equals, volumeFromContainer)
}

func (s *CliSuite) TestVolumeFromServiceWithContainerName(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/regression/volume_from_container_name.yml", "-p", p, "create")
	err := cmd.Run()
	c.Assert(err, IsNil)

	volumeFromContainer := "first_container_name"
	secondContainerName := p + "_second_1"

	cn := s.GetContainerByName(c, secondContainerName)
	c.Assert(cn, NotNil)

	c.Assert(len(cn.HostConfig.VolumesFrom), Equals, 1)
	c.Assert(cn.HostConfig.VolumesFrom[0], Equals, volumeFromContainer)
}

func (s *CliSuite) TestRelativeVolume(c *C) {
	p := s.ProjectFromText(c, "up", `
	server:
	  image: busybox
	  volumes:
	    - .:/path
	`)

	absPath, err := filepath.Abs(".")
	c.Assert(err, IsNil)
	serverName := fmt.Sprintf("%s_%s_1", p, "server")
	cn := s.GetContainerByName(c, serverName)

	c.Assert(cn, NotNil)
	c.Assert(len(cn.Mounts), DeepEquals, 1)
	c.Assert(cn.Mounts[0].Source, DeepEquals, absPath)
	c.Assert(cn.Mounts[0].Destination, DeepEquals, "/path")
}

func (s *CliSuite) TestNamedVolume(c *C) {
	p := s.ProjectFromText(c, "up", `
	server:
	  image: busybox
	  volumes:
	    - vol:/path
	`)

	serverName := fmt.Sprintf("%s_%s_1", p, "server")
	cn := s.GetContainerByName(c, serverName)

	c.Assert(cn, NotNil)
	c.Assert(len(cn.Mounts), DeepEquals, 1)
	c.Assert(cn.Mounts[0].Name, DeepEquals, "vol")
	c.Assert(cn.Mounts[0].Destination, DeepEquals, "/path")
}

func (s *CliSuite) TestV2Volume(c *C) {
	testRequires(c, not(DaemonVersionIs("1.9")))
	p := s.ProjectFromText(c, "up", `version: "2"
services:
  with_volume:
    image: busybox
    volumes:
    - test:/test

volumes:
  test: {}
  test2: {}
`)

	v := s.GetVolumeByName(c, p+"_test")
	c.Assert(v, NotNil)

	v = s.GetVolumeByName(c, p+"_test2")
	c.Assert(v, NotNil)
}
