package command

import (
	"fmt"
	"strings"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/consul/testutil"
	"github.com/mitchellh/cli"
)

func testExecCommand(t *testing.T) (*cli.MockUi, *ExecCommand) {
	ui := new(cli.MockUi)
	return ui, &ExecCommand{
		Command: base.Command{
			Ui:    ui,
			Flags: base.FlagSetHTTP,
		},
	}
}

func TestExecCommand_implements(t *testing.T) {
	var _ cli.Command = &ExecCommand{}
}

func TestExecCommandRun(t *testing.T) {
	a1 := testAgentWithConfig(t, func(c *agent.Config) {
		c.DisableRemoteExec = agent.Bool(false)
	})
	defer a1.Shutdown()
	waitForLeader(t, a1.httpAddr)

	ui, c := testExecCommand(t)
	args := []string{"-http-addr=" + a1.httpAddr, "-wait=10s", "uptime"}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. Error:%#v  (std)Output:%#v", code, ui.ErrorWriter.String(), ui.OutputWriter.String())
	}

	if !strings.Contains(ui.OutputWriter.String(), "load") {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}

func TestExecCommandRun_CrossDC(t *testing.T) {
	a1 := testAgentWithConfig(t, func(c *agent.Config) {
		c.DisableRemoteExec = agent.Bool(false)
	})
	defer a1.Shutdown()

	a2 := testAgentWithConfig(t, func(c *agent.Config) {
		c.Datacenter = "dc2"
		c.DisableRemoteExec = agent.Bool(false)
	})
	defer a2.Shutdown()

	// Join over the WAN
	wanAddr := fmt.Sprintf("%s:%d", a1.config.BindAddr, a1.config.Ports.SerfWan)
	n, err := a2.agent.JoinWAN([]string{wanAddr})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("bad %d", n)
	}

	waitForLeader(t, a1.httpAddr)
	waitForLeader(t, a2.httpAddr)

	ui, c := testExecCommand(t)
	args := []string{"-http-addr=" + a1.httpAddr,
		"-wait=400ms", "-datacenter=dc2", "uptime"}

	code := c.Run(args)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	if !strings.Contains(ui.OutputWriter.String(), "load") {
		t.Fatalf("bad: %#v", ui.OutputWriter.String())
	}
}

func waitForLeader(t *testing.T, httpAddr string) {
	client, err := httpClient(httpAddr)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := testutil.WaitForResult(func() (bool, error) {
		_, qm, err := client.Catalog().Nodes(nil)
		return err == nil && qm.KnownLeader && qm.LastIndex > 0, err
	}); err != nil {
		t.Fatal(err)
	}
}

func httpClient(addr string) (*consulapi.Client, error) {
	conf := consulapi.DefaultConfig()
	conf.Address = addr
	return consulapi.NewClient(conf)
}

func TestExecCommand_Validate(t *testing.T) {
	conf := &rExecConf{}
	err := conf.validate()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	conf.node = "("
	err = conf.validate()
	if err == nil {
		t.Fatalf("err: %v", err)
	}

	conf.node = ""
	conf.service = "("
	err = conf.validate()
	if err == nil {
		t.Fatalf("err: %v", err)
	}

	conf.service = "()"
	conf.tag = "("
	err = conf.validate()
	if err == nil {
		t.Fatalf("err: %v", err)
	}

	conf.service = ""
	conf.tag = "foo"
	err = conf.validate()
	if err == nil {
		t.Fatalf("err: %v", err)
	}
}

func TestExecCommand_Sessions(t *testing.T) {
	a1 := testAgentWithConfig(t, func(c *agent.Config) {
		c.DisableRemoteExec = agent.Bool(false)
	})
	defer a1.Shutdown()
	waitForLeader(t, a1.httpAddr)

	client, err := httpClient(a1.httpAddr)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	_, c := testExecCommand(t)
	c.client = client

	id, err := c.createSession()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	se, _, err := client.Session().Info(id, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if se == nil || se.Name != "Remote Exec" {
		t.Fatalf("bad: %v", se)
	}

	c.sessionID = id
	err = c.destroySession()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	se, _, err = client.Session().Info(id, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if se != nil {
		t.Fatalf("bad: %v", se)
	}
}

func TestExecCommand_Sessions_Foreign(t *testing.T) {
	a1 := testAgentWithConfig(t, func(c *agent.Config) {
		c.DisableRemoteExec = agent.Bool(false)
	})
	defer a1.Shutdown()
	waitForLeader(t, a1.httpAddr)

	client, err := httpClient(a1.httpAddr)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	_, c := testExecCommand(t)
	c.client = client

	c.conf.foreignDC = true
	c.conf.localDC = "dc1"
	c.conf.localNode = "foo"

	var id string
	if err := testutil.WaitForResult(func() (bool, error) {
		id, err = c.createSession()
		if err != nil && strings.Contains(err.Error(), "Failed to find Consul server") {
			err = nil
		}
		return id != "", err
	}); err != nil {
		t.Fatal(err)
	}

	se, _, err := client.Session().Info(id, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if se == nil || se.Name != "Remote Exec via foo@dc1" {
		t.Fatalf("bad: %v", se)
	}

	c.sessionID = id
	err = c.destroySession()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	se, _, err = client.Session().Info(id, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if se != nil {
		t.Fatalf("bad: %v", se)
	}
}

func TestExecCommand_UploadDestroy(t *testing.T) {
	a1 := testAgentWithConfig(t, func(c *agent.Config) {
		c.DisableRemoteExec = agent.Bool(false)
	})
	defer a1.Shutdown()
	waitForLeader(t, a1.httpAddr)

	client, err := httpClient(a1.httpAddr)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	_, c := testExecCommand(t)
	c.client = client

	id, err := c.createSession()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	c.sessionID = id

	c.conf.prefix = "_rexec"
	c.conf.cmd = "uptime"
	c.conf.wait = time.Second

	buf, err := c.makeRExecSpec()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	err = c.uploadPayload(buf)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pair, _, err := client.KV().Get("_rexec/"+id+"/job", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if pair == nil || len(pair.Value) == 0 {
		t.Fatalf("missing job spec")
	}

	err = c.destroyData()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pair, _, err = client.KV().Get("_rexec/"+id+"/job", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if pair != nil {
		t.Fatalf("should be destroyed")
	}
}

func TestExecCommand_StreamResults(t *testing.T) {
	a1 := testAgentWithConfig(t, func(c *agent.Config) {
		c.DisableRemoteExec = agent.Bool(false)
	})
	defer a1.Shutdown()
	waitForLeader(t, a1.httpAddr)

	client, err := httpClient(a1.httpAddr)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	_, c := testExecCommand(t)
	c.client = client
	c.conf.prefix = "_rexec"

	id, err := c.createSession()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	c.sessionID = id

	ackCh := make(chan rExecAck, 128)
	heartCh := make(chan rExecHeart, 128)
	outputCh := make(chan rExecOutput, 128)
	exitCh := make(chan rExecExit, 128)
	doneCh := make(chan struct{})
	errCh := make(chan struct{}, 1)
	defer close(doneCh)
	go c.streamResults(doneCh, ackCh, heartCh, outputCh, exitCh, errCh)

	prefix := "_rexec/" + id + "/"
	ok, _, err := client.KV().Acquire(&consulapi.KVPair{
		Key:     prefix + "foo/ack",
		Session: id,
	}, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("should be ok bro")
	}

	select {
	case a := <-ackCh:
		if a.Node != "foo" {
			t.Fatalf("bad: %#v", a)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("timeout")
	}

	ok, _, err = client.KV().Acquire(&consulapi.KVPair{
		Key:     prefix + "foo/exit",
		Value:   []byte("127"),
		Session: id,
	}, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("should be ok bro")
	}

	select {
	case e := <-exitCh:
		if e.Node != "foo" || e.Code != 127 {
			t.Fatalf("bad: %#v", e)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("timeout")
	}

	// Random key, should ignore
	ok, _, err = client.KV().Acquire(&consulapi.KVPair{
		Key:     prefix + "foo/random",
		Session: id,
	}, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("should be ok bro")
	}

	// Output heartbeat
	ok, _, err = client.KV().Acquire(&consulapi.KVPair{
		Key:     prefix + "foo/out/00000",
		Session: id,
	}, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("should be ok bro")
	}

	select {
	case h := <-heartCh:
		if h.Node != "foo" {
			t.Fatalf("bad: %#v", h)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("timeout")
	}

	// Output value
	ok, _, err = client.KV().Acquire(&consulapi.KVPair{
		Key:     prefix + "foo/out/00001",
		Value:   []byte("test"),
		Session: id,
	}, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("should be ok bro")
	}

	select {
	case o := <-outputCh:
		if o.Node != "foo" || string(o.Output) != "test" {
			t.Fatalf("bad: %#v", o)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("timeout")
	}
}
