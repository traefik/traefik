package command

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/consul"
	"github.com/hashicorp/consul/logger"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/consul/version"
	"github.com/hashicorp/go-uuid"
	"github.com/mitchellh/cli"
)

var offset uint64

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	version.Version = "0.8.0"
}

type agentWrapper struct {
	dir      string
	config   *agent.Config
	agent    *agent.Agent
	http     *agent.HTTPServer
	addr     string
	httpAddr string
}

func (a *agentWrapper) Shutdown() {
	a.agent.Shutdown()
	a.http.Shutdown()
	os.RemoveAll(a.dir)
}

func testAgent(t *testing.T) *agentWrapper {
	return testAgentWithConfig(t, nil)
}

func testAgentWithAPIClient(t *testing.T) (*agentWrapper, *api.Client) {
	agent := testAgentWithConfig(t, func(c *agent.Config) {})
	client, err := api.NewClient(&api.Config{Address: agent.httpAddr})
	if err != nil {
		t.Fatalf("consul client: %#v", err)
	}
	return agent, client
}

func testAgentWithConfig(t *testing.T, cb func(c *agent.Config)) *agentWrapper {
	return testAgentWithConfigReload(t, cb, nil)
}

func testAgentWithConfigReload(t *testing.T, cb func(c *agent.Config), reloadCh chan chan error) *agentWrapper {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	lw := logger.NewLogWriter(512)

	conf := nextConfig()
	if cb != nil {
		cb(conf)
	}

	dir, err := ioutil.TempDir("", "agent")
	if err != nil {
		t.Fatalf(fmt.Sprintf("err: %v", err))
	}
	conf.DataDir = dir

	a, err := agent.Create(conf, lw, nil, reloadCh)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf(fmt.Sprintf("err: %v", err))
	}

	conf.Addresses.HTTP = "127.0.0.1"
	httpAddr := fmt.Sprintf("127.0.0.1:%d", conf.Ports.HTTP)
	http, err := agent.NewHTTPServers(a, conf, os.Stderr)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf(fmt.Sprintf("err: %v", err))
	}

	if http == nil || len(http) == 0 {
		os.RemoveAll(dir)
		t.Fatalf(fmt.Sprintf("Could not create HTTP server to listen on: %s", httpAddr))
	}

	return &agentWrapper{
		dir:      dir,
		config:   conf,
		agent:    a,
		http:     http[0],
		addr:     l.Addr().String(),
		httpAddr: httpAddr,
	}
}

func nextConfig() *agent.Config {
	idx := int(atomic.AddUint64(&offset, 1))
	conf := agent.DefaultConfig()

	nodeID, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}

	conf.Bootstrap = true
	conf.Datacenter = "dc1"
	conf.NodeName = fmt.Sprintf("Node %d", idx)
	conf.NodeID = types.NodeID(nodeID)
	conf.BindAddr = "127.0.0.1"
	conf.Server = true

	conf.Version = version.Version

	conf.Ports.HTTP = 10000 + 10*idx
	conf.Ports.HTTPS = 10401 + 10*idx
	conf.Ports.SerfLan = 10201 + 10*idx
	conf.Ports.SerfWan = 10202 + 10*idx
	conf.Ports.Server = 10300 + 10*idx

	cons := consul.DefaultConfig()
	conf.ConsulConfig = cons

	cons.SerfLANConfig.MemberlistConfig.ProbeTimeout = 100 * time.Millisecond
	cons.SerfLANConfig.MemberlistConfig.ProbeInterval = 100 * time.Millisecond
	cons.SerfLANConfig.MemberlistConfig.GossipInterval = 100 * time.Millisecond

	cons.SerfWANConfig.MemberlistConfig.ProbeTimeout = 100 * time.Millisecond
	cons.SerfWANConfig.MemberlistConfig.ProbeInterval = 100 * time.Millisecond
	cons.SerfWANConfig.MemberlistConfig.GossipInterval = 100 * time.Millisecond

	cons.RaftConfig.LeaderLeaseTimeout = 20 * time.Millisecond
	cons.RaftConfig.HeartbeatTimeout = 40 * time.Millisecond
	cons.RaftConfig.ElectionTimeout = 40 * time.Millisecond

	return conf
}

func assertNoTabs(t *testing.T, c cli.Command) {
	if strings.ContainsRune(c.Help(), '\t') {
		t.Errorf("%#v help output contains tabs", c)
	}
}
