package servers

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/serf/serf"
)

// FloodPortFn gets the port to use for a given server when flood-joining. This
// will return false if it doesn't have one.
type FloodPortFn func(*agent.Server) (int, bool)

// FloodJoins attempts to make sure all Consul servers in the local Serf
// instance are joined in the global Serf instance. It assumes names in the
// local area are of the form <node> and those in the global area are of the
// form <node>.<dc> as is done for WAN and general network areas in Consul
// Enterprise.
func FloodJoins(logger *log.Logger, portFn FloodPortFn,
	localDatacenter string, localSerf *serf.Serf, globalSerf *serf.Serf) {

	// Names in the global Serf have the datacenter suffixed.
	suffix := fmt.Sprintf(".%s", localDatacenter)

	// Index the global side so we can do one pass through the local side
	// with cheap lookups.
	index := make(map[string]*agent.Server)
	for _, m := range globalSerf.Members() {
		ok, server := agent.IsConsulServer(m)
		if !ok {
			continue
		}

		if server.Datacenter != localDatacenter {
			continue
		}

		localName := strings.TrimSuffix(server.Name, suffix)
		index[localName] = server
	}

	// Now run through the local side and look for joins.
	for _, m := range localSerf.Members() {
		if m.Status != serf.StatusAlive {
			continue
		}

		ok, server := agent.IsConsulServer(m)
		if !ok {
			continue
		}

		if _, ok := index[server.Name]; ok {
			continue
		}

		// We can't use the port number from the local Serf, so we just
		// get the host part.
		addr, _, err := net.SplitHostPort(server.Addr.String())
		if err != nil {
			logger.Printf("[DEBUG] consul: Failed to flood-join %q (bad address %q): %v",
				server.Name, server.Addr.String(), err)
		}

		// Let the callback see if it can get the port number, otherwise
		// leave it blank to behave as if we just supplied an address.
		if port, ok := portFn(server); ok {
			addr = net.JoinHostPort(addr, fmt.Sprintf("%d", port))
		}

		// Do the join!
		n, err := globalSerf.Join([]string{addr}, true)
		if err != nil {
			logger.Printf("[DEBUG] consul: Failed to flood-join %q at %s: %v",
				server.Name, addr, err)
		} else if n > 0 {
			logger.Printf("[DEBUG] consul: Successfully performed flood-join for %q at %s",
				server.Name, addr)
		}
	}
}
