package consul

import (
	"fmt"
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/consul/consul/state"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/go-uuid"
)

// Catalog endpoint is used to manipulate the service catalog
type Catalog struct {
	srv *Server
}

// Register is used register that a node is providing a given service.
func (c *Catalog) Register(args *structs.RegisterRequest, reply *struct{}) error {
	if done, err := c.srv.forward("Catalog.Register", args, args, reply); done {
		return err
	}
	defer metrics.MeasureSince([]string{"consul", "catalog", "register"}, time.Now())

	// Verify the args.
	if args.Node == "" || args.Address == "" {
		return fmt.Errorf("Must provide node and address")
	}
	if args.ID != "" {
		if _, err := uuid.ParseUUID(string(args.ID)); err != nil {
			return fmt.Errorf("Bad node ID: %v", err)
		}
	}

	// Fetch the ACL token, if any.
	acl, err := c.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}

	// Handle a service registration.
	if args.Service != nil {
		// If no service id, but service name, use default
		if args.Service.ID == "" && args.Service.Service != "" {
			args.Service.ID = args.Service.Service
		}

		// Verify ServiceName provided if ID.
		if args.Service.ID != "" && args.Service.Service == "" {
			return fmt.Errorf("Must provide service name with ID")
		}

		// Apply the ACL policy if any. The 'consul' service is excluded
		// since it is managed automatically internally (that behavior
		// is going away after version 0.8). We check this same policy
		// later if version 0.8 is enabled, so we can eventually just
		// delete this and do all the ACL checks down there.
		if args.Service.Service != ConsulServiceName {
			if acl != nil && !acl.ServiceWrite(args.Service.Service) {
				return permissionDeniedErr
			}
		}
	}

	// Move the old format single check into the slice, and fixup IDs.
	if args.Check != nil {
		args.Checks = append(args.Checks, args.Check)
		args.Check = nil
	}
	for _, check := range args.Checks {
		if check.CheckID == "" && check.Name != "" {
			check.CheckID = types.CheckID(check.Name)
		}
		if check.Node == "" {
			check.Node = args.Node
		}
	}

	// Check the complete register request against the given ACL policy.
	if acl != nil && c.srv.config.ACLEnforceVersion8 {
		state := c.srv.fsm.State()
		_, ns, err := state.NodeServices(nil, args.Node)
		if err != nil {
			return fmt.Errorf("Node lookup failed: %v", err)
		}
		if err := vetRegisterWithACL(acl, args, ns); err != nil {
			return err
		}
	}

	resp, err := c.srv.raftApply(structs.RegisterRequestType, args)
	if err != nil {
		return err
	}
	if respErr, ok := resp.(error); ok {
		return respErr
	}
	return nil
}

// Deregister is used to remove a service registration for a given node.
func (c *Catalog) Deregister(args *structs.DeregisterRequest, reply *struct{}) error {
	if done, err := c.srv.forward("Catalog.Deregister", args, args, reply); done {
		return err
	}
	defer metrics.MeasureSince([]string{"consul", "catalog", "deregister"}, time.Now())

	// Verify the args
	if args.Node == "" {
		return fmt.Errorf("Must provide node")
	}

	// Fetch the ACL token, if any.
	acl, err := c.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}

	// Check the complete deregister request against the given ACL policy.
	if acl != nil && c.srv.config.ACLEnforceVersion8 {
		state := c.srv.fsm.State()

		var ns *structs.NodeService
		if args.ServiceID != "" {
			_, ns, err = state.NodeService(args.Node, args.ServiceID)
			if err != nil {
				return fmt.Errorf("Service lookup failed: %v", err)
			}
		}

		var nc *structs.HealthCheck
		if args.CheckID != "" {
			_, nc, err = state.NodeCheck(args.Node, args.CheckID)
			if err != nil {
				return fmt.Errorf("Check lookup failed: %v", err)
			}
		}

		if err := vetDeregisterWithACL(acl, args, ns, nc); err != nil {
			return err
		}
	}

	if _, err := c.srv.raftApply(structs.DeregisterRequestType, args); err != nil {
		return err
	}
	return nil
}

// ListDatacenters is used to query for the list of known datacenters
func (c *Catalog) ListDatacenters(args *struct{}, reply *[]string) error {
	dcs, err := c.srv.router.GetDatacentersByDistance()
	if err != nil {
		return err
	}

	*reply = dcs
	return nil
}

// ListNodes is used to query the nodes in a DC
func (c *Catalog) ListNodes(args *structs.DCSpecificRequest, reply *structs.IndexedNodes) error {
	if done, err := c.srv.forward("Catalog.ListNodes", args, args, reply); done {
		return err
	}

	return c.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.StateStore) error {
			var index uint64
			var nodes structs.Nodes
			var err error
			if len(args.NodeMetaFilters) > 0 {
				index, nodes, err = state.NodesByMeta(ws, args.NodeMetaFilters)
			} else {
				index, nodes, err = state.Nodes(ws)
			}
			if err != nil {
				return err
			}

			reply.Index, reply.Nodes = index, nodes
			if err := c.srv.filterACL(args.Token, reply); err != nil {
				return err
			}
			return c.srv.sortNodesByDistanceFrom(args.Source, reply.Nodes)
		})
}

// ListServices is used to query the services in a DC
func (c *Catalog) ListServices(args *structs.DCSpecificRequest, reply *structs.IndexedServices) error {
	if done, err := c.srv.forward("Catalog.ListServices", args, args, reply); done {
		return err
	}

	return c.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.StateStore) error {
			var index uint64
			var services structs.Services
			var err error
			if len(args.NodeMetaFilters) > 0 {
				index, services, err = state.ServicesByNodeMeta(ws, args.NodeMetaFilters)
			} else {
				index, services, err = state.Services(ws)
			}
			if err != nil {
				return err
			}

			reply.Index, reply.Services = index, services
			return c.srv.filterACL(args.Token, reply)
		})
}

// ServiceNodes returns all the nodes registered as part of a service
func (c *Catalog) ServiceNodes(args *structs.ServiceSpecificRequest, reply *structs.IndexedServiceNodes) error {
	if done, err := c.srv.forward("Catalog.ServiceNodes", args, args, reply); done {
		return err
	}

	// Verify the arguments
	if args.ServiceName == "" {
		return fmt.Errorf("Must provide service name")
	}

	err := c.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.StateStore) error {
			var index uint64
			var services structs.ServiceNodes
			var err error
			if args.TagFilter {
				index, services, err = state.ServiceTagNodes(ws, args.ServiceName, args.ServiceTag)
			} else {
				index, services, err = state.ServiceNodes(ws, args.ServiceName)
			}
			if err != nil {
				return err
			}
			reply.Index, reply.ServiceNodes = index, services
			if len(args.NodeMetaFilters) > 0 {
				var filtered structs.ServiceNodes
				for _, service := range services {
					if structs.SatisfiesMetaFilters(service.NodeMeta, args.NodeMetaFilters) {
						filtered = append(filtered, service)
					}
				}
				reply.ServiceNodes = filtered
			}
			if err := c.srv.filterACL(args.Token, reply); err != nil {
				return err
			}
			return c.srv.sortNodesByDistanceFrom(args.Source, reply.ServiceNodes)
		})

	// Provide some metrics
	if err == nil {
		metrics.IncrCounter([]string{"consul", "catalog", "service", "query", args.ServiceName}, 1)
		if args.ServiceTag != "" {
			metrics.IncrCounter([]string{"consul", "catalog", "service", "query-tag", args.ServiceName, args.ServiceTag}, 1)
		}
		if len(reply.ServiceNodes) == 0 {
			metrics.IncrCounter([]string{"consul", "catalog", "service", "not-found", args.ServiceName}, 1)
		}
	}
	return err
}

// NodeServices returns all the services registered as part of a node
func (c *Catalog) NodeServices(args *structs.NodeSpecificRequest, reply *structs.IndexedNodeServices) error {
	if done, err := c.srv.forward("Catalog.NodeServices", args, args, reply); done {
		return err
	}

	// Verify the arguments
	if args.Node == "" {
		return fmt.Errorf("Must provide node")
	}

	return c.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.StateStore) error {
			index, services, err := state.NodeServices(ws, args.Node)
			if err != nil {
				return err
			}

			reply.Index, reply.NodeServices = index, services
			return c.srv.filterACL(args.Token, reply)
		})
}
