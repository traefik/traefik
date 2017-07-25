package integration

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libcompose/utils"
	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestUp(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	c.Assert(cn.State.Running, Equals, true)
}

func (s *CliSuite) TestUpNotExistService(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "not_exist")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, IsNil)
}

func (s *CliSuite) TestRecreateForceRecreate(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	p = s.FromText(c, p, "up", "--force-recreate", SimpleTemplate)
	cn2 := s.GetContainerByName(c, name)
	c.Assert(cn.ID, Not(Equals), cn2.ID)
}

func mountSet(slice []types.MountPoint) map[string]bool {
	result := map[string]bool{}
	for _, v := range slice {
		result[fmt.Sprint(v.Source, ":", v.Destination)] = true
	}
	return result
}

func (s *CliSuite) TestRecreateVols(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplateWithVols)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	p = s.FromText(c, p, "up", "--force-recreate", SimpleTemplateWithVols2)
	cn2 := s.GetContainerByName(c, name)
	c.Assert(cn.ID, Not(Equals), cn2.ID)

	notHomeRootOrVol2 := func(mount string) bool {
		switch strings.SplitN(mount, ":", 2)[1] {
		case "/home", "/root", "/var/lib/vol2":
			return false
		}
		return true
	}

	shouldMigrate := utils.FilterStringSet(mountSet(cn.Mounts), notHomeRootOrVol2)
	cn2Mounts := mountSet(cn2.Mounts)
	for k := range shouldMigrate {
		c.Assert(cn2Mounts[k], Equals, true)
	}

	almostTheSameButRoot := utils.FilterStringSet(cn2Mounts, notHomeRootOrVol2)
	c.Assert(len(almostTheSameButRoot), Equals, len(cn2Mounts)-1)
	c.Assert(cn2Mounts["/tmp/tmp-root:/root"], Equals, true)
	c.Assert(cn2Mounts["/root:/root"], Equals, false)
}

func (s *CliSuite) TestRecreateNoRecreate(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	p = s.FromText(c, p, "up", "--no-recreate", `
	hello:
	  labels:
	    key: val
	  image: busybox
	  stdin_open: true
	  tty: true
	`)
	cn2 := s.GetContainerByName(c, name)
	c.Assert(cn.ID, Equals, cn2.ID)
	_, ok := cn2.Config.Labels["key"]
	c.Assert(ok, Equals, false)
}

func (s *CliSuite) TestRecreate(c *C) {
	p := s.ProjectFromText(c, "up", SimpleTemplate)

	name := fmt.Sprintf("%s_%s_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	p = s.FromText(c, p, "up", SimpleTemplate)
	cn2 := s.GetContainerByName(c, name)
	c.Assert(cn.ID, Equals, cn2.ID)

	p = s.FromText(c, p, "up", `
	hello:
	  labels:
	    key: val
	  image: busybox
	  stdin_open: true
	  tty: true
	`)
	cn3 := s.GetContainerByName(c, name)
	c.Assert(cn2.ID, Not(Equals), cn3.ID)
	key3 := cn3.Config.Labels["key"]
	c.Assert(key3, Equals, "val")

	// Should still recreate because old has a different label
	p = s.FromText(c, p, "up", `
	hello:
	  image: busybox
	  stdin_open: true
	  tty: true
	`)
	cn4 := s.GetContainerByName(c, name)
	c.Assert(cn3.ID, Not(Equals), cn4.ID)
	_, ok4 := cn4.Config.Labels["key"]
	c.Assert(ok4, Equals, false)

	p = s.FromText(c, p, "up", `
	hello:
	  image: busybox
	  stdin_open: true
	  tty: true
	`)
	cn5 := s.GetContainerByName(c, name)
	c.Assert(cn4.ID, Equals, cn5.ID)
	_, ok5 := cn5.Config.Labels["key"]
	c.Assert(ok5, Equals, false)

	p = s.FromText(c, p, "up", "--force-recreate", `
	hello:
	  image: busybox
	  stdin_open: true
	  tty: true
	`)
	cn6 := s.GetContainerByName(c, name)
	c.Assert(cn5.ID, Not(Equals), cn6.ID)
	_, ok6 := cn6.Config.Labels["key"]
	c.Assert(ok6, Equals, false)

	p = s.FromText(c, p, "up", "--force-recreate", `
	hello:
	  image: busybox
	  stdin_open: true
	  tty: true
	`)
	cn7 := s.GetContainerByName(c, name)
	c.Assert(cn6.ID, Not(Equals), cn7.ID)
	_, ok7 := cn7.Config.Labels["key"]
	c.Assert(ok7, Equals, false)

	c.Assert(cn.State.Running, Equals, true)
}

func (s *CliSuite) TestUpAfterImageTagDeleted(c *C) {
	client := GetClient(c)
	label := RandStr(7)
	repo := "busybox"
	image := fmt.Sprintf("%s:%s", repo, label)

	template := fmt.Sprintf(`
	hello:
	  labels:
	    key: val
	  image: %s
	  stdin_open: true
	  tty: true
	`, image)

	err := client.ImageTag(context.Background(), "busybox:latest", repo+":"+label)
	c.Assert(err, IsNil)

	p := s.ProjectFromText(c, "up", template)
	name := fmt.Sprintf("%s_%s_1", p, "hello")
	firstContainer := s.GetContainerByName(c, name)

	_, err = client.ImageRemove(context.Background(), image, types.ImageRemoveOptions{})
	c.Assert(err, IsNil)

	p = s.FromText(c, p, "up", "--no-recreate", template)
	latestContainer := s.GetContainerByName(c, name)
	c.Assert(firstContainer.ID, Equals, latestContainer.ID)
}

func (s *CliSuite) TestRecreateImageChanging(c *C) {
	client := GetClient(c)
	label := "buildroot-2013.08.1"
	repo := "busybox"
	image := fmt.Sprintf("%s:%s", repo, label)

	template := fmt.Sprintf(`
	hello:
	  labels:
	    key: val
	  image: %s
	  stdin_open: true
	  tty: true
	`, image)

	ctx := context.Background()

	// Ignore error here
	client.ImageRemove(ctx, image, types.ImageRemoveOptions{})

	// Up, pull needed
	p := s.ProjectFromText(c, "up", template)
	name := fmt.Sprintf("%s_%s_1", p, "hello")
	firstContainer := s.GetContainerByName(c, name)

	// Up --no-recreate, no pull needed
	p = s.FromText(c, p, "up", "--no-recreate", template)
	latestContainer := s.GetContainerByName(c, name)
	c.Assert(firstContainer.ID, Equals, latestContainer.ID)

	// Up --no-recreate, no pull needed
	p = s.FromText(c, p, "up", "--no-recreate", template)
	latestContainer = s.GetContainerByName(c, name)
	c.Assert(firstContainer.ID, Equals, latestContainer.ID)

	// Change what tag points to
	// Note: depending on the daemon version it can fail with --force (which is no more possible to pass using engine-api)
	//       thus, the next following lines are a hack…
	err := client.ImageTag(ctx, image, image+"backup")
	c.Assert(err, IsNil)
	_, err = client.ImageRemove(ctx, image, types.ImageRemoveOptions{})
	c.Assert(err, IsNil)
	err = client.ImageTag(ctx, "busybox:latest", image)
	c.Assert(err, IsNil)

	// Up (with recreate - the default), pull is needed and new container is created
	p = s.FromText(c, p, "up", template)
	latestContainer = s.GetContainerByName(c, name)
	c.Assert(firstContainer.ID, Not(Equals), latestContainer.ID)

	s.FromText(c, p, "rm", "-f", template)
}

func (s *CliSuite) TestLink(c *C) {
	p := s.ProjectFromText(c, "up", `
        server:
          image: busybox
          command: cat
          stdin_open: true
          expose:
          - 80
        client:
          image: busybox
          links:
          - server:foo
          - server
        `)

	serverName := fmt.Sprintf("%s_%s_1", p, "server")

	cn := s.GetContainerByName(c, serverName)
	c.Assert(cn, NotNil)
	c.Assert(cn.Config.ExposedPorts, DeepEquals, nat.PortSet{"80/tcp": struct{}{}})

	clientName := fmt.Sprintf("%s_%s_1", p, "client")
	cn = s.GetContainerByName(c, clientName)
	c.Assert(cn, NotNil)
	c.Assert(asMap(cn.HostConfig.Links), DeepEquals, asMap([]string{
		fmt.Sprintf("/%s:/%s/%s", serverName, clientName, "foo"),
		fmt.Sprintf("/%s:/%s/%s", serverName, clientName, "server"),
		fmt.Sprintf("/%s:/%s/%s", serverName, clientName, serverName),
	}))
}

func (s *CliSuite) TestUpNoBuildFailIfImageNotPresent(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/build/docker-compose.yml", "-p", p, "up", "--no-build")
	err := cmd.Run()

	c.Assert(err, NotNil)
}

func (s *CliSuite) TestUpNoBuildShouldWorkIfImageIsPresent(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/simple-build/docker-compose.yml", "-p", p, "build")
	err := cmd.Run()

	c.Assert(err, IsNil)

	cmd = exec.Command(s.command, "-f", "./assets/simple-build/docker-compose.yml", "-p", p, "up", "-d", "--no-build")
	err = cmd.Run()

	c.Assert(err, IsNil)
}
