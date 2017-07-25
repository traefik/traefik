package servers_test

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/consul/consul/servers"
)

type fauxConnPool struct {
	// failPct between 0.0 and 1.0 == pct of time a Ping should fail
	failPct float64
}

func (cp *fauxConnPool) PingConsulServer(server *agent.Server) (bool, error) {
	var success bool
	successProb := rand.Float64()
	if successProb > cp.failPct {
		success = true
	}
	return success, nil
}

type fauxSerf struct {
}

func (s *fauxSerf) NumNodes() int {
	return 16384
}

func testManager() (m *servers.Manager) {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	shutdownCh := make(chan struct{})
	m = servers.New(logger, shutdownCh, &fauxSerf{}, &fauxConnPool{})
	return m
}

func testManagerFailProb(failPct float64) (m *servers.Manager) {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	shutdownCh := make(chan struct{})
	m = servers.New(logger, shutdownCh, &fauxSerf{}, &fauxConnPool{failPct: failPct})
	return m
}

// func (m *Manager) AddServer(server *agent.Server) {
func TestServers_AddServer(t *testing.T) {
	m := testManager()
	var num int
	num = m.NumServers()
	if num != 0 {
		t.Fatalf("Expected zero servers to start")
	}

	s1 := &agent.Server{Name: "s1"}
	m.AddServer(s1)
	num = m.NumServers()
	if num != 1 {
		t.Fatalf("Expected one server")
	}

	m.AddServer(s1)
	num = m.NumServers()
	if num != 1 {
		t.Fatalf("Expected one server (still)")
	}

	s2 := &agent.Server{Name: "s2"}
	m.AddServer(s2)
	num = m.NumServers()
	if num != 2 {
		t.Fatalf("Expected two servers")
	}
}

// func (m *Manager) IsOffline() bool {
func TestServers_IsOffline(t *testing.T) {
	m := testManager()
	if !m.IsOffline() {
		t.Fatalf("bad")
	}

	s1 := &agent.Server{Name: "s1"}
	m.AddServer(s1)
	if m.IsOffline() {
		t.Fatalf("bad")
	}
	m.RebalanceServers()
	if m.IsOffline() {
		t.Fatalf("bad")
	}
	m.RemoveServer(s1)
	m.RebalanceServers()
	if !m.IsOffline() {
		t.Fatalf("bad")
	}

	const failPct = 0.5
	m = testManagerFailProb(failPct)
	m.AddServer(s1)
	var on, off int
	for i := 0; i < 100; i++ {
		m.RebalanceServers()
		if m.IsOffline() {
			off++
		} else {
			on++
		}
	}
	if on == 0 || off == 0 {
		t.Fatalf("bad: %d %d", on, off)
	}
}

// func (m *Manager) FindServer() (server *agent.Server) {
func TestServers_FindServer(t *testing.T) {
	m := testManager()

	if m.FindServer() != nil {
		t.Fatalf("Expected nil return")
	}

	m.AddServer(&agent.Server{Name: "s1"})
	if m.NumServers() != 1 {
		t.Fatalf("Expected one server")
	}

	s1 := m.FindServer()
	if s1 == nil {
		t.Fatalf("Expected non-nil server")
	}
	if s1.Name != "s1" {
		t.Fatalf("Expected s1 server")
	}

	s1 = m.FindServer()
	if s1 == nil || s1.Name != "s1" {
		t.Fatalf("Expected s1 server (still)")
	}

	m.AddServer(&agent.Server{Name: "s2"})
	if m.NumServers() != 2 {
		t.Fatalf("Expected two servers")
	}
	s1 = m.FindServer()
	if s1 == nil || s1.Name != "s1" {
		t.Fatalf("Expected s1 server (still)")
	}

	m.NotifyFailedServer(s1)
	s2 := m.FindServer()
	if s2 == nil || s2.Name != "s2" {
		t.Fatalf("Expected s2 server")
	}

	m.NotifyFailedServer(s2)
	s1 = m.FindServer()
	if s1 == nil || s1.Name != "s1" {
		t.Fatalf("Expected s1 server")
	}
}

// func New(logger *log.Logger, shutdownCh chan struct{}) (m *Manager) {
func TestServers_New(t *testing.T) {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	shutdownCh := make(chan struct{})
	m := servers.New(logger, shutdownCh, &fauxSerf{}, &fauxConnPool{})
	if m == nil {
		t.Fatalf("Manager nil")
	}
}

// func (m *Manager) NotifyFailedServer(server *agent.Server) {
func TestServers_NotifyFailedServer(t *testing.T) {
	m := testManager()

	if m.NumServers() != 0 {
		t.Fatalf("Expected zero servers to start")
	}

	s1 := &agent.Server{Name: "s1"}
	s2 := &agent.Server{Name: "s2"}

	// Try notifying for a server that is not managed by Manager
	m.NotifyFailedServer(s1)
	if m.NumServers() != 0 {
		t.Fatalf("Expected zero servers to start")
	}
	m.AddServer(s1)

	// Test again w/ a server not in the list
	m.NotifyFailedServer(s2)
	if m.NumServers() != 1 {
		t.Fatalf("Expected one server")
	}

	m.AddServer(s2)
	if m.NumServers() != 2 {
		t.Fatalf("Expected two servers")
	}

	s1 = m.FindServer()
	if s1 == nil || s1.Name != "s1" {
		t.Fatalf("Expected s1 server")
	}

	m.NotifyFailedServer(s2)
	s1 = m.FindServer()
	if s1 == nil || s1.Name != "s1" {
		t.Fatalf("Expected s1 server (still)")
	}

	m.NotifyFailedServer(s1)
	s2 = m.FindServer()
	if s2 == nil || s2.Name != "s2" {
		t.Fatalf("Expected s2 server")
	}

	m.NotifyFailedServer(s2)
	s1 = m.FindServer()
	if s1 == nil || s1.Name != "s1" {
		t.Fatalf("Expected s1 server")
	}
}

// func (m *Manager) NumServers() (numServers int) {
func TestServers_NumServers(t *testing.T) {
	m := testManager()
	var num int
	num = m.NumServers()
	if num != 0 {
		t.Fatalf("Expected zero servers to start")
	}

	s := &agent.Server{}
	m.AddServer(s)
	num = m.NumServers()
	if num != 1 {
		t.Fatalf("Expected one server after AddServer")
	}
}

// func (m *Manager) RebalanceServers() {
func TestServers_RebalanceServers(t *testing.T) {
	const failPct = 0.5
	m := testManagerFailProb(failPct)
	const maxServers = 100
	const numShuffleTests = 100
	const uniquePassRate = 0.5

	// Make a huge list of nodes.
	for i := 0; i < maxServers; i++ {
		nodeName := fmt.Sprintf("s%02d", i)
		m.AddServer(&agent.Server{Name: nodeName})
	}

	// Keep track of how many unique shuffles we get.
	uniques := make(map[string]struct{}, maxServers)
	for i := 0; i < numShuffleTests; i++ {
		m.RebalanceServers()

		var names []string
		for j := 0; j < maxServers; j++ {
			server := m.FindServer()
			m.NotifyFailedServer(server)
			names = append(names, server.Name)
		}
		key := strings.Join(names, "|")
		uniques[key] = struct{}{}
	}

	// We have to allow for the fact that there won't always be a unique
	// shuffle each pass, so we just look for smell here without the test
	// being flaky.
	if len(uniques) < int(maxServers*uniquePassRate) {
		t.Fatalf("unique shuffle ratio too low: %d/%d", len(uniques), maxServers)
	}
}

// func (m *Manager) RemoveServer(server *agent.Server) {
func TestManager_RemoveServer(t *testing.T) {
	const nodeNameFmt = "s%02d"
	m := testManager()

	if m.NumServers() != 0 {
		t.Fatalf("Expected zero servers to start")
	}

	// Test removing server before its added
	nodeName := fmt.Sprintf(nodeNameFmt, 1)
	s1 := &agent.Server{Name: nodeName}
	m.RemoveServer(s1)
	m.AddServer(s1)

	nodeName = fmt.Sprintf(nodeNameFmt, 2)
	s2 := &agent.Server{Name: nodeName}
	m.RemoveServer(s2)
	m.AddServer(s2)

	const maxServers = 19
	servers := make([]*agent.Server, maxServers)
	// Already added two servers above
	for i := maxServers; i > 2; i-- {
		nodeName := fmt.Sprintf(nodeNameFmt, i)
		server := &agent.Server{Name: nodeName}
		servers = append(servers, server)
		m.AddServer(server)
	}
	if m.NumServers() != maxServers {
		t.Fatalf("Expected %d servers, received %d", maxServers, m.NumServers())
	}

	m.RebalanceServers()

	if m.NumServers() != maxServers {
		t.Fatalf("Expected %d servers, received %d", maxServers, m.NumServers())
	}

	findServer := func(server *agent.Server) bool {
		for i := m.NumServers(); i > 0; i-- {
			s := m.FindServer()
			if s == server {
				return true
			}
		}
		return false
	}

	expectedNumServers := maxServers
	removedServers := make([]*agent.Server, 0, maxServers)

	// Remove servers from the front of the list
	for i := 3; i > 0; i-- {
		server := m.FindServer()
		if server == nil {
			t.Fatalf("FindServer returned nil")
		}
		m.RemoveServer(server)
		expectedNumServers--
		if m.NumServers() != expectedNumServers {
			t.Fatalf("Expected %d servers (got %d)", expectedNumServers, m.NumServers())
		}
		if findServer(server) == true {
			t.Fatalf("Did not expect to find server %s after removal from the front", server.Name)
		}
		removedServers = append(removedServers, server)
	}

	// Remove server from the end of the list
	for i := 3; i > 0; i-- {
		server := m.FindServer()
		m.NotifyFailedServer(server)
		m.RemoveServer(server)
		expectedNumServers--
		if m.NumServers() != expectedNumServers {
			t.Fatalf("Expected %d servers (got %d)", expectedNumServers, m.NumServers())
		}
		if findServer(server) == true {
			t.Fatalf("Did not expect to find server %s", server.Name)
		}
		removedServers = append(removedServers, server)
	}

	// Remove server from the middle of the list
	for i := 3; i > 0; i-- {
		server := m.FindServer()
		m.NotifyFailedServer(server)
		server2 := m.FindServer()
		m.NotifyFailedServer(server2) // server2 now at end of the list

		m.RemoveServer(server)
		expectedNumServers--
		if m.NumServers() != expectedNumServers {
			t.Fatalf("Expected %d servers (got %d)", expectedNumServers, m.NumServers())
		}
		if findServer(server) == true {
			t.Fatalf("Did not expect to find server %s", server.Name)
		}
		removedServers = append(removedServers, server)
	}

	if m.NumServers()+len(removedServers) != maxServers {
		t.Fatalf("Expected %d+%d=%d servers", m.NumServers(), len(removedServers), maxServers)
	}

	// Drain the remaining servers from the middle
	for i := m.NumServers(); i > 0; i-- {
		server := m.FindServer()
		m.NotifyFailedServer(server)
		server2 := m.FindServer()
		m.NotifyFailedServer(server2) // server2 now at end of the list
		m.RemoveServer(server)
		removedServers = append(removedServers, server)
	}

	if m.NumServers() != 0 {
		t.Fatalf("Expected an empty server list")
	}
	if len(removedServers) != maxServers {
		t.Fatalf("Expected all servers to be in removed server list")
	}
}

// func (m *Manager) Start() {
