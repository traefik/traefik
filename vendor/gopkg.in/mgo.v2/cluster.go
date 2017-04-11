// mgo - MongoDB driver for Go
//
// Copyright (c) 2010-2012 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mgo

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// ---------------------------------------------------------------------------
// Mongo cluster encapsulation.
//
// A cluster enables the communication with one or more servers participating
// in a mongo cluster.  This works with individual servers, a replica set,
// a replica pair, one or multiple mongos routers, etc.

type mongoCluster struct {
	sync.RWMutex
	serverSynced sync.Cond
	userSeeds    []string
	dynaSeeds    []string
	servers      mongoServers
	masters      mongoServers
	references   int
	syncing      bool
	direct       bool
	failFast     bool
	syncCount    uint
	setName      string
	cachedIndex  map[string]bool
	sync         chan bool
	dial         dialer
}

func newCluster(userSeeds []string, direct, failFast bool, dial dialer, setName string) *mongoCluster {
	cluster := &mongoCluster{
		userSeeds:  userSeeds,
		references: 1,
		direct:     direct,
		failFast:   failFast,
		dial:       dial,
		setName:    setName,
	}
	cluster.serverSynced.L = cluster.RWMutex.RLocker()
	cluster.sync = make(chan bool, 1)
	stats.cluster(+1)
	go cluster.syncServersLoop()
	return cluster
}

// Acquire increases the reference count for the cluster.
func (cluster *mongoCluster) Acquire() {
	cluster.Lock()
	cluster.references++
	debugf("Cluster %p acquired (refs=%d)", cluster, cluster.references)
	cluster.Unlock()
}

// Release decreases the reference count for the cluster. Once
// it reaches zero, all servers will be closed.
func (cluster *mongoCluster) Release() {
	cluster.Lock()
	if cluster.references == 0 {
		panic("cluster.Release() with references == 0")
	}
	cluster.references--
	debugf("Cluster %p released (refs=%d)", cluster, cluster.references)
	if cluster.references == 0 {
		for _, server := range cluster.servers.Slice() {
			server.Close()
		}
		// Wake up the sync loop so it can die.
		cluster.syncServers()
		stats.cluster(-1)
	}
	cluster.Unlock()
}

func (cluster *mongoCluster) LiveServers() (servers []string) {
	cluster.RLock()
	for _, serv := range cluster.servers.Slice() {
		servers = append(servers, serv.Addr)
	}
	cluster.RUnlock()
	return servers
}

func (cluster *mongoCluster) removeServer(server *mongoServer) {
	cluster.Lock()
	cluster.masters.Remove(server)
	other := cluster.servers.Remove(server)
	cluster.Unlock()
	if other != nil {
		other.Close()
		log("Removed server ", server.Addr, " from cluster.")
	}
	server.Close()
}

type isMasterResult struct {
	IsMaster       bool
	Secondary      bool
	Primary        string
	Hosts          []string
	Passives       []string
	Tags           bson.D
	Msg            string
	SetName        string `bson:"setName"`
	MaxWireVersion int    `bson:"maxWireVersion"`
}

func (cluster *mongoCluster) isMaster(socket *mongoSocket, result *isMasterResult) error {
	// Monotonic let's it talk to a slave and still hold the socket.
	session := newSession(Monotonic, cluster, 10*time.Second)
	session.setSocket(socket)
	err := session.Run("ismaster", result)
	session.Close()
	return err
}

type possibleTimeout interface {
	Timeout() bool
}

var syncSocketTimeout = 5 * time.Second

func (cluster *mongoCluster) syncServer(server *mongoServer) (info *mongoServerInfo, hosts []string, err error) {
	var syncTimeout time.Duration
	if raceDetector {
		// This variable is only ever touched by tests.
		globalMutex.Lock()
		syncTimeout = syncSocketTimeout
		globalMutex.Unlock()
	} else {
		syncTimeout = syncSocketTimeout
	}

	addr := server.Addr
	log("SYNC Processing ", addr, "...")

	// Retry a few times to avoid knocking a server down for a hiccup.
	var result isMasterResult
	var tryerr error
	for retry := 0; ; retry++ {
		if retry == 3 || retry == 1 && cluster.failFast {
			return nil, nil, tryerr
		}
		if retry > 0 {
			// Don't abuse the server needlessly if there's something actually wrong.
			if err, ok := tryerr.(possibleTimeout); ok && err.Timeout() {
				// Give a chance for waiters to timeout as well.
				cluster.serverSynced.Broadcast()
			}
			time.Sleep(syncShortDelay)
		}

		// It's not clear what would be a good timeout here. Is it
		// better to wait longer or to retry?
		socket, _, err := server.AcquireSocket(0, syncTimeout)
		if err != nil {
			tryerr = err
			logf("SYNC Failed to get socket to %s: %v", addr, err)
			continue
		}
		err = cluster.isMaster(socket, &result)
		socket.Release()
		if err != nil {
			tryerr = err
			logf("SYNC Command 'ismaster' to %s failed: %v", addr, err)
			continue
		}
		debugf("SYNC Result of 'ismaster' from %s: %#v", addr, result)
		break
	}

	if cluster.setName != "" && result.SetName != cluster.setName {
		logf("SYNC Server %s is not a member of replica set %q", addr, cluster.setName)
		return nil, nil, fmt.Errorf("server %s is not a member of replica set %q", addr, cluster.setName)
	}

	if result.IsMaster {
		debugf("SYNC %s is a master.", addr)
		if !server.info.Master {
			// Made an incorrect assumption above, so fix stats.
			stats.conn(-1, false)
			stats.conn(+1, true)
		}
	} else if result.Secondary {
		debugf("SYNC %s is a slave.", addr)
	} else if cluster.direct {
		logf("SYNC %s in unknown state. Pretending it's a slave due to direct connection.", addr)
	} else {
		logf("SYNC %s is neither a master nor a slave.", addr)
		// Let stats track it as whatever was known before.
		return nil, nil, errors.New(addr + " is not a master nor slave")
	}

	info = &mongoServerInfo{
		Master:         result.IsMaster,
		Mongos:         result.Msg == "isdbgrid",
		Tags:           result.Tags,
		SetName:        result.SetName,
		MaxWireVersion: result.MaxWireVersion,
	}

	hosts = make([]string, 0, 1+len(result.Hosts)+len(result.Passives))
	if result.Primary != "" {
		// First in the list to speed up master discovery.
		hosts = append(hosts, result.Primary)
	}
	hosts = append(hosts, result.Hosts...)
	hosts = append(hosts, result.Passives...)

	debugf("SYNC %s knows about the following peers: %#v", addr, hosts)
	return info, hosts, nil
}

type syncKind bool

const (
	completeSync syncKind = true
	partialSync  syncKind = false
)

func (cluster *mongoCluster) addServer(server *mongoServer, info *mongoServerInfo, syncKind syncKind) {
	cluster.Lock()
	current := cluster.servers.Search(server.ResolvedAddr)
	if current == nil {
		if syncKind == partialSync {
			cluster.Unlock()
			server.Close()
			log("SYNC Discarding unknown server ", server.Addr, " due to partial sync.")
			return
		}
		cluster.servers.Add(server)
		if info.Master {
			cluster.masters.Add(server)
			log("SYNC Adding ", server.Addr, " to cluster as a master.")
		} else {
			log("SYNC Adding ", server.Addr, " to cluster as a slave.")
		}
	} else {
		if server != current {
			panic("addServer attempting to add duplicated server")
		}
		if server.Info().Master != info.Master {
			if info.Master {
				log("SYNC Server ", server.Addr, " is now a master.")
				cluster.masters.Add(server)
			} else {
				log("SYNC Server ", server.Addr, " is now a slave.")
				cluster.masters.Remove(server)
			}
		}
	}
	server.SetInfo(info)
	debugf("SYNC Broadcasting availability of server %s", server.Addr)
	cluster.serverSynced.Broadcast()
	cluster.Unlock()
}

func (cluster *mongoCluster) getKnownAddrs() []string {
	cluster.RLock()
	max := len(cluster.userSeeds) + len(cluster.dynaSeeds) + cluster.servers.Len()
	seen := make(map[string]bool, max)
	known := make([]string, 0, max)

	add := func(addr string) {
		if _, found := seen[addr]; !found {
			seen[addr] = true
			known = append(known, addr)
		}
	}

	for _, addr := range cluster.userSeeds {
		add(addr)
	}
	for _, addr := range cluster.dynaSeeds {
		add(addr)
	}
	for _, serv := range cluster.servers.Slice() {
		add(serv.Addr)
	}
	cluster.RUnlock()

	return known
}

// syncServers injects a value into the cluster.sync channel to force
// an iteration of the syncServersLoop function.
func (cluster *mongoCluster) syncServers() {
	select {
	case cluster.sync <- true:
	default:
	}
}

// How long to wait for a checkup of the cluster topology if nothing
// else kicks a synchronization before that.
const syncServersDelay = 30 * time.Second
const syncShortDelay = 500 * time.Millisecond

// syncServersLoop loops while the cluster is alive to keep its idea of
// the server topology up-to-date. It must be called just once from
// newCluster.  The loop iterates once syncServersDelay has passed, or
// if somebody injects a value into the cluster.sync channel to force a
// synchronization.  A loop iteration will contact all servers in
// parallel, ask them about known peers and their own role within the
// cluster, and then attempt to do the same with all the peers
// retrieved.
func (cluster *mongoCluster) syncServersLoop() {
	for {
		debugf("SYNC Cluster %p is starting a sync loop iteration.", cluster)

		cluster.Lock()
		if cluster.references == 0 {
			cluster.Unlock()
			break
		}
		cluster.references++ // Keep alive while syncing.
		direct := cluster.direct
		cluster.Unlock()

		cluster.syncServersIteration(direct)

		// We just synchronized, so consume any outstanding requests.
		select {
		case <-cluster.sync:
		default:
		}

		cluster.Release()

		// Hold off before allowing another sync. No point in
		// burning CPU looking for down servers.
		if !cluster.failFast {
			time.Sleep(syncShortDelay)
		}

		cluster.Lock()
		if cluster.references == 0 {
			cluster.Unlock()
			break
		}
		cluster.syncCount++
		// Poke all waiters so they have a chance to timeout or
		// restart syncing if they wish to.
		cluster.serverSynced.Broadcast()
		// Check if we have to restart immediately either way.
		restart := !direct && cluster.masters.Empty() || cluster.servers.Empty()
		cluster.Unlock()

		if restart {
			log("SYNC No masters found. Will synchronize again.")
			time.Sleep(syncShortDelay)
			continue
		}

		debugf("SYNC Cluster %p waiting for next requested or scheduled sync.", cluster)

		// Hold off until somebody explicitly requests a synchronization
		// or it's time to check for a cluster topology change again.
		select {
		case <-cluster.sync:
		case <-time.After(syncServersDelay):
		}
	}
	debugf("SYNC Cluster %p is stopping its sync loop.", cluster)
}

func (cluster *mongoCluster) server(addr string, tcpaddr *net.TCPAddr) *mongoServer {
	cluster.RLock()
	server := cluster.servers.Search(tcpaddr.String())
	cluster.RUnlock()
	if server != nil {
		return server
	}
	return newServer(addr, tcpaddr, cluster.sync, cluster.dial)
}

func resolveAddr(addr string) (*net.TCPAddr, error) {
	// Simple cases that do not need actual resolution. Works with IPv4 and v6.
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if port, _ := strconv.Atoi(port); port > 0 {
			zone := ""
			if i := strings.LastIndex(host, "%"); i >= 0 {
				zone = host[i+1:]
				host = host[:i]
			}
			ip := net.ParseIP(host)
			if ip != nil {
				return &net.TCPAddr{IP: ip, Port: port, Zone: zone}, nil
			}
		}
	}

	// Attempt to resolve IPv4 and v6 concurrently.
	addrChan := make(chan *net.TCPAddr, 2)
	for _, network := range []string{"udp4", "udp6"} {
		network := network
		go func() {
			// The unfortunate UDP dialing hack allows having a timeout on address resolution.
			conn, err := net.DialTimeout(network, addr, 10*time.Second)
			if err != nil {
				addrChan <- nil
			} else {
				addrChan <- (*net.TCPAddr)(conn.RemoteAddr().(*net.UDPAddr))
				conn.Close()
			}
		}()
	}

	// Wait for the result of IPv4 and v6 resolution. Use IPv4 if available.
	tcpaddr := <-addrChan
	if tcpaddr == nil || len(tcpaddr.IP) != 4 {
		var timeout <-chan time.Time
		if tcpaddr != nil {
			// Don't wait too long if an IPv6 address is known.
			timeout = time.After(50 * time.Millisecond)
		}
		select {
		case <-timeout:
		case tcpaddr2 := <-addrChan:
			if tcpaddr == nil || tcpaddr2 != nil {
				// It's an IPv4 address or the only known address. Use it.
				tcpaddr = tcpaddr2
			}
		}
	}

	if tcpaddr == nil {
		log("SYNC Failed to resolve server address: ", addr)
		return nil, errors.New("failed to resolve server address: " + addr)
	}
	if tcpaddr.String() != addr {
		debug("SYNC Address ", addr, " resolved as ", tcpaddr.String())
	}
	return tcpaddr, nil
}

type pendingAdd struct {
	server *mongoServer
	info   *mongoServerInfo
}

func (cluster *mongoCluster) syncServersIteration(direct bool) {
	log("SYNC Starting full topology synchronization...")

	var wg sync.WaitGroup
	var m sync.Mutex
	notYetAdded := make(map[string]pendingAdd)
	addIfFound := make(map[string]bool)
	seen := make(map[string]bool)
	syncKind := partialSync

	var spawnSync func(addr string, byMaster bool)
	spawnSync = func(addr string, byMaster bool) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			tcpaddr, err := resolveAddr(addr)
			if err != nil {
				log("SYNC Failed to start sync of ", addr, ": ", err.Error())
				return
			}
			resolvedAddr := tcpaddr.String()

			m.Lock()
			if byMaster {
				if pending, ok := notYetAdded[resolvedAddr]; ok {
					delete(notYetAdded, resolvedAddr)
					m.Unlock()
					cluster.addServer(pending.server, pending.info, completeSync)
					return
				}
				addIfFound[resolvedAddr] = true
			}
			if seen[resolvedAddr] {
				m.Unlock()
				return
			}
			seen[resolvedAddr] = true
			m.Unlock()

			server := cluster.server(addr, tcpaddr)
			info, hosts, err := cluster.syncServer(server)
			if err != nil {
				cluster.removeServer(server)
				return
			}

			m.Lock()
			add := direct || info.Master || addIfFound[resolvedAddr]
			if add {
				syncKind = completeSync
			} else {
				notYetAdded[resolvedAddr] = pendingAdd{server, info}
			}
			m.Unlock()
			if add {
				cluster.addServer(server, info, completeSync)
			}
			if !direct {
				for _, addr := range hosts {
					spawnSync(addr, info.Master)
				}
			}
		}()
	}

	knownAddrs := cluster.getKnownAddrs()
	for _, addr := range knownAddrs {
		spawnSync(addr, false)
	}
	wg.Wait()

	if syncKind == completeSync {
		logf("SYNC Synchronization was complete (got data from primary).")
		for _, pending := range notYetAdded {
			cluster.removeServer(pending.server)
		}
	} else {
		logf("SYNC Synchronization was partial (cannot talk to primary).")
		for _, pending := range notYetAdded {
			cluster.addServer(pending.server, pending.info, partialSync)
		}
	}

	cluster.Lock()
	mastersLen := cluster.masters.Len()
	logf("SYNC Synchronization completed: %d master(s) and %d slave(s) alive.", mastersLen, cluster.servers.Len()-mastersLen)

	// Update dynamic seeds, but only if we have any good servers. Otherwise,
	// leave them alone for better chances of a successful sync in the future.
	if syncKind == completeSync {
		dynaSeeds := make([]string, cluster.servers.Len())
		for i, server := range cluster.servers.Slice() {
			dynaSeeds[i] = server.Addr
		}
		cluster.dynaSeeds = dynaSeeds
		debugf("SYNC New dynamic seeds: %#v\n", dynaSeeds)
	}
	cluster.Unlock()
}

// AcquireSocket returns a socket to a server in the cluster.  If slaveOk is
// true, it will attempt to return a socket to a slave server.  If it is
// false, the socket will necessarily be to a master server.
func (cluster *mongoCluster) AcquireSocket(mode Mode, slaveOk bool, syncTimeout time.Duration, socketTimeout time.Duration, serverTags []bson.D, poolLimit int) (s *mongoSocket, err error) {
	var started time.Time
	var syncCount uint
	warnedLimit := false
	for {
		cluster.RLock()
		for {
			mastersLen := cluster.masters.Len()
			slavesLen := cluster.servers.Len() - mastersLen
			debugf("Cluster has %d known masters and %d known slaves.", mastersLen, slavesLen)
			if mastersLen > 0 && !(slaveOk && mode == Secondary) || slavesLen > 0 && slaveOk {
				break
			}
			if mastersLen > 0 && mode == Secondary && cluster.masters.HasMongos() {
				break
			}
			if started.IsZero() {
				// Initialize after fast path above.
				started = time.Now()
				syncCount = cluster.syncCount
			} else if syncTimeout != 0 && started.Before(time.Now().Add(-syncTimeout)) || cluster.failFast && cluster.syncCount != syncCount {
				cluster.RUnlock()
				return nil, errors.New("no reachable servers")
			}
			log("Waiting for servers to synchronize...")
			cluster.syncServers()

			// Remember: this will release and reacquire the lock.
			cluster.serverSynced.Wait()
		}

		var server *mongoServer
		if slaveOk {
			server = cluster.servers.BestFit(mode, serverTags)
		} else {
			server = cluster.masters.BestFit(mode, nil)
		}
		cluster.RUnlock()

		if server == nil {
			// Must have failed the requested tags. Sleep to avoid spinning.
			time.Sleep(1e8)
			continue
		}

		s, abended, err := server.AcquireSocket(poolLimit, socketTimeout)
		if err == errPoolLimit {
			if !warnedLimit {
				warnedLimit = true
				log("WARNING: Per-server connection limit reached.")
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err != nil {
			cluster.removeServer(server)
			cluster.syncServers()
			continue
		}
		if abended && !slaveOk {
			var result isMasterResult
			err := cluster.isMaster(s, &result)
			if err != nil || !result.IsMaster {
				logf("Cannot confirm server %s as master (%v)", server.Addr, err)
				s.Release()
				cluster.syncServers()
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		return s, nil
	}
	panic("unreached")
}

func (cluster *mongoCluster) CacheIndex(cacheKey string, exists bool) {
	cluster.Lock()
	if cluster.cachedIndex == nil {
		cluster.cachedIndex = make(map[string]bool)
	}
	if exists {
		cluster.cachedIndex[cacheKey] = true
	} else {
		delete(cluster.cachedIndex, cacheKey)
	}
	cluster.Unlock()
}

func (cluster *mongoCluster) HasCachedIndex(cacheKey string) (result bool) {
	cluster.RLock()
	if cluster.cachedIndex != nil {
		result = cluster.cachedIndex[cacheKey]
	}
	cluster.RUnlock()
	return
}

func (cluster *mongoCluster) ResetIndexCache() {
	cluster.Lock()
	cluster.cachedIndex = make(map[string]bool)
	cluster.Unlock()
}
