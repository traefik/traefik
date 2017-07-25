package consul

import (
	"context"
	"log"
	"sync"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/consul/consul/structs"
)

// StatsFetcher has two functions for autopilot. First, lets us fetch all the
// stats in parallel so we are taking a sample as close to the same time as
// possible, since we are comparing time-sensitive info for the health check.
// Second, it bounds the time so that one slow RPC can't hold up the health
// check loop; as a side effect of how it implements this, it also limits to
// a single in-flight RPC to any given server, so goroutines don't accumulate
// as we run the health check fairly frequently.
type StatsFetcher struct {
	logger       *log.Logger
	pool         *ConnPool
	datacenter   string
	inflight     map[string]struct{}
	inflightLock sync.Mutex
}

// NewStatsFetcher returns a stats fetcher.
func NewStatsFetcher(logger *log.Logger, pool *ConnPool, datacenter string) *StatsFetcher {
	return &StatsFetcher{
		logger:     logger,
		pool:       pool,
		datacenter: datacenter,
		inflight:   make(map[string]struct{}),
	}
}

// fetch does the RPC to fetch the server stats from a single server. We don't
// cancel this when the context is canceled because we only want one in-flight
// RPC to each server, so we let it finish and then clean up the in-flight
// tracking.
func (f *StatsFetcher) fetch(server *agent.Server, replyCh chan *structs.ServerStats) {
	var args struct{}
	var reply structs.ServerStats
	err := f.pool.RPC(f.datacenter, server.Addr, server.Version, "Status.RaftStats", &args, &reply)
	if err != nil {
		f.logger.Printf("[WARN] consul: error getting server health from %q: %v",
			server.Name, err)
	} else {
		replyCh <- &reply
	}

	f.inflightLock.Lock()
	delete(f.inflight, server.ID)
	f.inflightLock.Unlock()
}

// Fetch will attempt to query all the servers in parallel.
func (f *StatsFetcher) Fetch(ctx context.Context, servers []*agent.Server) map[string]*structs.ServerStats {
	type workItem struct {
		server  *agent.Server
		replyCh chan *structs.ServerStats
	}
	var work []*workItem

	// Skip any servers that have inflight requests.
	f.inflightLock.Lock()
	for _, server := range servers {
		if _, ok := f.inflight[server.ID]; ok {
			f.logger.Printf("[WARN] consul: error getting server health from %q: last request still outstanding",
				server.Name)
		} else {
			workItem := &workItem{
				server:  server,
				replyCh: make(chan *structs.ServerStats, 1),
			}
			work = append(work, workItem)
			f.inflight[server.ID] = struct{}{}
			go f.fetch(workItem.server, workItem.replyCh)
		}
	}
	f.inflightLock.Unlock()

	// Now wait for the results to come in, or for the context to be
	// canceled.
	replies := make(map[string]*structs.ServerStats)
	for _, workItem := range work {
		select {
		case reply := <-workItem.replyCh:
			replies[workItem.server.ID] = reply

		case <-ctx.Done():
			f.logger.Printf("[WARN] consul: error getting server health from %q: %v",
				workItem.server.Name, ctx.Err())
		}
	}
	return replies
}
