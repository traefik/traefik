package servers

import (
	"log"

	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/serf/serf"
)

// routerFn selects one of the router operations to map to incoming Serf events.
type routerFn func(types.AreaID, *agent.Server) error

// handleMemberEvents attempts to apply the given Serf member event to the given
// router function.
func handleMemberEvent(logger *log.Logger, fn routerFn, areaID types.AreaID, e serf.Event) {
	me, ok := e.(serf.MemberEvent)
	if !ok {
		logger.Printf("[ERR] consul: Bad event type %#v", e)
		return
	}

	for _, m := range me.Members {
		ok, parts := agent.IsConsulServer(m)
		if !ok {
			logger.Printf("[WARN]: consul: Non-server %q in server-only area %q",
				m.Name, areaID)
			continue
		}

		if err := fn(areaID, parts); err != nil {
			logger.Printf("[ERR] consul: Failed to process %s event for server %q in area %q: %v",
				me.Type.String(), m.Name, areaID, err)
			continue
		}

		logger.Printf("[INFO] consul: Handled %s event for server %q in area %q",
			me.Type.String(), m.Name, areaID)
	}
}

// HandleSerfEvents is a long-running goroutine that pushes incoming events from
// a Serf manager's channel into the given router. This will return when the
// shutdown channel is closed.
func HandleSerfEvents(logger *log.Logger, router *Router, areaID types.AreaID, shutdownCh <-chan struct{}, eventCh <-chan serf.Event) {
	for {
		select {
		case <-shutdownCh:
			return

		case e := <-eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				handleMemberEvent(logger, router.AddServer, areaID, e)

			case serf.EventMemberLeave:
				handleMemberEvent(logger, router.RemoveServer, areaID, e)

			case serf.EventMemberFailed:
				handleMemberEvent(logger, router.FailServer, areaID, e)

			// All of these event types are ignored.
			case serf.EventMemberUpdate:
			case serf.EventMemberReap:
			case serf.EventUser:
			case serf.EventQuery:

			default:
				logger.Printf("[WARN] consul: Unhandled Serf Event: %#v", e)
			}
		}
	}
}
