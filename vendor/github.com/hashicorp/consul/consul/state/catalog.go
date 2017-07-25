package state

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/go-memdb"
)

const (
	// minUUIDLookupLen is used as a minimum length of a node name required before
	// we test to see if the name is actually a UUID and perform an ID-based node
	// lookup.
	minUUIDLookupLen = 2
)

func resizeNodeLookupKey(s string) string {
	l := len(s)

	if l%2 != 0 {
		return s[0 : l-1]
	}

	return s
}

// Nodes is used to pull the full list of nodes for use during snapshots.
func (s *StateSnapshot) Nodes() (memdb.ResultIterator, error) {
	iter, err := s.tx.Get("nodes", "id")
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// Services is used to pull the full list of services for a given node for use
// during snapshots.
func (s *StateSnapshot) Services(node string) (memdb.ResultIterator, error) {
	iter, err := s.tx.Get("services", "node", node)
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// Checks is used to pull the full list of checks for a given node for use
// during snapshots.
func (s *StateSnapshot) Checks(node string) (memdb.ResultIterator, error) {
	iter, err := s.tx.Get("checks", "node", node)
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// Registration is used to make sure a node, service, and check registration is
// performed within a single transaction to avoid race conditions on state
// updates.
func (s *StateRestore) Registration(idx uint64, req *structs.RegisterRequest) error {
	if err := s.store.ensureRegistrationTxn(s.tx, idx, req); err != nil {
		return err
	}
	return nil
}

// EnsureRegistration is used to make sure a node, service, and check
// registration is performed within a single transaction to avoid race
// conditions on state updates.
func (s *StateStore) EnsureRegistration(idx uint64, req *structs.RegisterRequest) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	if err := s.ensureRegistrationTxn(tx, idx, req); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// ensureRegistrationTxn is used to make sure a node, service, and check
// registration is performed within a single transaction to avoid race
// conditions on state updates.
func (s *StateStore) ensureRegistrationTxn(tx *memdb.Txn, idx uint64, req *structs.RegisterRequest) error {
	// Create a node structure.
	node := &structs.Node{
		ID:              req.ID,
		Node:            req.Node,
		Address:         req.Address,
		TaggedAddresses: req.TaggedAddresses,
		Meta:            req.NodeMeta,
	}

	// Since this gets called for all node operations (service and check
	// updates) and churn on the node itself is basically none after the
	// node updates itself the first time, it's worth seeing if we need to
	// modify the node at all so we prevent watch churn and useless writes
	// and modify index bumps on the node.
	{
		existing, err := tx.First("nodes", "id", node.Node)
		if err != nil {
			return fmt.Errorf("node lookup failed: %s", err)
		}
		if existing == nil || req.ChangesNode(existing.(*structs.Node)) {
			if err := s.ensureNodeTxn(tx, idx, node); err != nil {
				return fmt.Errorf("failed inserting node: %s", err)
			}
		}
	}

	// Add the service, if any. We perform a similar check as we do for the
	// node info above to make sure we actually need to update the service
	// definition in order to prevent useless churn if nothing has changed.
	if req.Service != nil {
		existing, err := tx.First("services", "id", req.Node, req.Service.ID)
		if err != nil {
			return fmt.Errorf("failed service lookup: %s", err)
		}
		if existing == nil || !(existing.(*structs.ServiceNode).ToNodeService()).IsSame(req.Service) {
			if err := s.ensureServiceTxn(tx, idx, req.Node, req.Service); err != nil {
				return fmt.Errorf("failed inserting service: %s", err)

			}
		}
	}

	// Add the checks, if any.
	if req.Check != nil {
		if req.Check.Node != req.Node {
			return fmt.Errorf("check node %q does not match node %q",
				req.Check.Node, req.Node)
		}
		if err := s.ensureCheckTxn(tx, idx, req.Check); err != nil {
			return fmt.Errorf("failed inserting check: %s", err)
		}
	}
	for _, check := range req.Checks {
		if check.Node != req.Node {
			return fmt.Errorf("check node %q does not match node %q",
				check.Node, req.Node)
		}
		if err := s.ensureCheckTxn(tx, idx, check); err != nil {
			return fmt.Errorf("failed inserting check: %s", err)
		}
	}

	return nil
}

// EnsureNode is used to upsert node registration or modification.
func (s *StateStore) EnsureNode(idx uint64, node *structs.Node) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the node upsert
	if err := s.ensureNodeTxn(tx, idx, node); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// ensureNodeTxn is the inner function called to actually create a node
// registration or modify an existing one in the state store. It allows
// passing in a memdb transaction so it may be part of a larger txn.
func (s *StateStore) ensureNodeTxn(tx *memdb.Txn, idx uint64, node *structs.Node) error {
	// See if there's an existing node with this UUID, and make sure the
	// name is the same.
	var n *structs.Node
	if node.ID != "" {
		existing, err := tx.First("nodes", "uuid", string(node.ID))
		if err != nil {
			return fmt.Errorf("node lookup failed: %s", err)
		}
		if existing != nil {
			n = existing.(*structs.Node)
			if n.Node != node.Node {
				return fmt.Errorf("node ID %q for node %q aliases existing node %q",
					node.ID, node.Node, n.Node)
			}
		}
	}

	// Check for an existing node by name to support nodes with no IDs.
	if n == nil {
		existing, err := tx.First("nodes", "id", node.Node)
		if err != nil {
			return fmt.Errorf("node name lookup failed: %s", err)
		}
		if existing != nil {
			n = existing.(*structs.Node)
		}
	}

	// Get the indexes.
	if n != nil {
		node.CreateIndex = n.CreateIndex
		node.ModifyIndex = idx
	} else {
		node.CreateIndex = idx
		node.ModifyIndex = idx
	}

	// Insert the node and update the index.
	if err := tx.Insert("nodes", node); err != nil {
		return fmt.Errorf("failed inserting node: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"nodes", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// GetNode is used to retrieve a node registration by node name ID.
func (s *StateStore) GetNode(id string) (uint64, *structs.Node, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes")

	// Retrieve the node from the state store
	node, err := tx.First("nodes", "id", id)
	if err != nil {
		return 0, nil, fmt.Errorf("node lookup failed: %s", err)
	}
	if node != nil {
		return idx, node.(*structs.Node), nil
	}
	return idx, nil, nil
}

// GetNodeID is used to retrieve a node registration by node ID.
func (s *StateStore) GetNodeID(id types.NodeID) (uint64, *structs.Node, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes")

	// Retrieve the node from the state store
	node, err := tx.First("nodes", "uuid", string(id))
	if err != nil {
		return 0, nil, fmt.Errorf("node lookup failed: %s", err)
	}
	if node != nil {
		return idx, node.(*structs.Node), nil
	}
	return idx, nil, nil
}

// Nodes is used to return all of the known nodes.
func (s *StateStore) Nodes(ws memdb.WatchSet) (uint64, structs.Nodes, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes")

	// Retrieve all of the nodes
	nodes, err := tx.Get("nodes", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed nodes lookup: %s", err)
	}
	ws.Add(nodes.WatchCh())

	// Create and return the nodes list.
	var results structs.Nodes
	for node := nodes.Next(); node != nil; node = nodes.Next() {
		results = append(results, node.(*structs.Node))
	}
	return idx, results, nil
}

// NodesByMeta is used to return all nodes with the given metadata key/value pairs.
func (s *StateStore) NodesByMeta(ws memdb.WatchSet, filters map[string]string) (uint64, structs.Nodes, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes")

	// Retrieve all of the nodes
	var args []interface{}
	for key, value := range filters {
		args = append(args, key, value)
		break
	}
	nodes, err := tx.Get("nodes", "meta", args...)
	if err != nil {
		return 0, nil, fmt.Errorf("failed nodes lookup: %s", err)
	}
	ws.Add(nodes.WatchCh())

	// Create and return the nodes list.
	var results structs.Nodes
	for node := nodes.Next(); node != nil; node = nodes.Next() {
		n := node.(*structs.Node)
		if len(filters) <= 1 || structs.SatisfiesMetaFilters(n.Meta, filters) {
			results = append(results, n)
		}
	}
	return idx, results, nil
}

// DeleteNode is used to delete a given node by its ID.
func (s *StateStore) DeleteNode(idx uint64, nodeName string) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the node deletion.
	if err := s.deleteNodeTxn(tx, idx, nodeName); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// deleteNodeTxn is the inner method used for removing a node from
// the store within a given transaction.
func (s *StateStore) deleteNodeTxn(tx *memdb.Txn, idx uint64, nodeName string) error {
	// Look up the node.
	node, err := tx.First("nodes", "id", nodeName)
	if err != nil {
		return fmt.Errorf("node lookup failed: %s", err)
	}
	if node == nil {
		return nil
	}

	// Delete all services associated with the node and update the service index.
	services, err := tx.Get("services", "node", nodeName)
	if err != nil {
		return fmt.Errorf("failed service lookup: %s", err)
	}
	var sids []string
	for service := services.Next(); service != nil; service = services.Next() {
		sids = append(sids, service.(*structs.ServiceNode).ServiceID)
	}

	// Do the delete in a separate loop so we don't trash the iterator.
	for _, sid := range sids {
		if err := s.deleteServiceTxn(tx, idx, nodeName, sid); err != nil {
			return err
		}
	}

	// Delete all checks associated with the node. This will invalidate
	// sessions as necessary.
	checks, err := tx.Get("checks", "node", nodeName)
	if err != nil {
		return fmt.Errorf("failed check lookup: %s", err)
	}
	var cids []types.CheckID
	for check := checks.Next(); check != nil; check = checks.Next() {
		cids = append(cids, check.(*structs.HealthCheck).CheckID)
	}

	// Do the delete in a separate loop so we don't trash the iterator.
	for _, cid := range cids {
		if err := s.deleteCheckTxn(tx, idx, nodeName, cid); err != nil {
			return err
		}
	}

	// Delete any coordinate associated with this node.
	coord, err := tx.First("coordinates", "id", nodeName)
	if err != nil {
		return fmt.Errorf("failed coordinate lookup: %s", err)
	}
	if coord != nil {
		if err := tx.Delete("coordinates", coord); err != nil {
			return fmt.Errorf("failed deleting coordinate: %s", err)
		}
		if err := tx.Insert("index", &IndexEntry{"coordinates", idx}); err != nil {
			return fmt.Errorf("failed updating index: %s", err)
		}
	}

	// Delete the node and update the index.
	if err := tx.Delete("nodes", node); err != nil {
		return fmt.Errorf("failed deleting node: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"nodes", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	// Invalidate any sessions for this node.
	sessions, err := tx.Get("sessions", "node", nodeName)
	if err != nil {
		return fmt.Errorf("failed session lookup: %s", err)
	}
	var ids []string
	for sess := sessions.Next(); sess != nil; sess = sessions.Next() {
		ids = append(ids, sess.(*structs.Session).ID)
	}

	// Do the delete in a separate loop so we don't trash the iterator.
	for _, id := range ids {
		if err := s.deleteSessionTxn(tx, idx, id); err != nil {
			return fmt.Errorf("failed session delete: %s", err)
		}
	}

	return nil
}

// EnsureService is called to upsert creation of a given NodeService.
func (s *StateStore) EnsureService(idx uint64, node string, svc *structs.NodeService) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the service registration upsert
	if err := s.ensureServiceTxn(tx, idx, node, svc); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// ensureServiceTxn is used to upsert a service registration within an
// existing memdb transaction.
func (s *StateStore) ensureServiceTxn(tx *memdb.Txn, idx uint64, node string, svc *structs.NodeService) error {
	// Check for existing service
	existing, err := tx.First("services", "id", node, svc.ID)
	if err != nil {
		return fmt.Errorf("failed service lookup: %s", err)
	}

	// Create the service node entry and populate the indexes. Note that
	// conversion doesn't populate any of the node-specific information.
	// That's always populated when we read from the state store.
	entry := svc.ToServiceNode(node)
	if existing != nil {
		entry.CreateIndex = existing.(*structs.ServiceNode).CreateIndex
		entry.ModifyIndex = idx
	} else {
		entry.CreateIndex = idx
		entry.ModifyIndex = idx
	}

	// Get the node
	n, err := tx.First("nodes", "id", node)
	if err != nil {
		return fmt.Errorf("failed node lookup: %s", err)
	}
	if n == nil {
		return ErrMissingNode
	}

	// Insert the service and update the index
	if err := tx.Insert("services", entry); err != nil {
		return fmt.Errorf("failed inserting service: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"services", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// Services returns all services along with a list of associated tags.
func (s *StateStore) Services(ws memdb.WatchSet) (uint64, structs.Services, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "services")

	// List all the services.
	services, err := tx.Get("services", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed querying services: %s", err)
	}
	ws.Add(services.WatchCh())

	// Rip through the services and enumerate them and their unique set of
	// tags.
	unique := make(map[string]map[string]struct{})
	for service := services.Next(); service != nil; service = services.Next() {
		svc := service.(*structs.ServiceNode)
		tags, ok := unique[svc.ServiceName]
		if !ok {
			unique[svc.ServiceName] = make(map[string]struct{})
			tags = unique[svc.ServiceName]
		}
		for _, tag := range svc.ServiceTags {
			tags[tag] = struct{}{}
		}
	}

	// Generate the output structure.
	var results = make(structs.Services)
	for service, tags := range unique {
		results[service] = make([]string, 0)
		for tag, _ := range tags {
			results[service] = append(results[service], tag)
		}
	}
	return idx, results, nil
}

// ServicesByNodeMeta returns all services, filtered by the given node metadata.
func (s *StateStore) ServicesByNodeMeta(ws memdb.WatchSet, filters map[string]string) (uint64, structs.Services, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "services", "nodes")

	// Retrieve all of the nodes with the meta k/v pair
	var args []interface{}
	for key, value := range filters {
		args = append(args, key, value)
		break
	}
	nodes, err := tx.Get("nodes", "meta", args...)
	if err != nil {
		return 0, nil, fmt.Errorf("failed nodes lookup: %s", err)
	}
	ws.Add(nodes.WatchCh())

	// We don't want to track an unlimited number of services, so we pull a
	// top-level watch to use as a fallback.
	allServices, err := tx.Get("services", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed services lookup: %s", err)
	}
	allServicesCh := allServices.WatchCh()

	// Populate the services map
	unique := make(map[string]map[string]struct{})
	for node := nodes.Next(); node != nil; node = nodes.Next() {
		n := node.(*structs.Node)
		if len(filters) > 1 && !structs.SatisfiesMetaFilters(n.Meta, filters) {
			continue
		}

		// List all the services on the node
		services, err := tx.Get("services", "node", n.Node)
		if err != nil {
			return 0, nil, fmt.Errorf("failed querying services: %s", err)
		}
		ws.AddWithLimit(watchLimit, services.WatchCh(), allServicesCh)

		// Rip through the services and enumerate them and their unique set of
		// tags.
		for service := services.Next(); service != nil; service = services.Next() {
			svc := service.(*structs.ServiceNode)
			tags, ok := unique[svc.ServiceName]
			if !ok {
				unique[svc.ServiceName] = make(map[string]struct{})
				tags = unique[svc.ServiceName]
			}
			for _, tag := range svc.ServiceTags {
				tags[tag] = struct{}{}
			}
		}
	}

	// Generate the output structure.
	var results = make(structs.Services)
	for service, tags := range unique {
		results[service] = make([]string, 0)
		for tag, _ := range tags {
			results[service] = append(results[service], tag)
		}
	}
	return idx, results, nil
}

// ServiceNodes returns the nodes associated with a given service name.
func (s *StateStore) ServiceNodes(ws memdb.WatchSet, serviceName string) (uint64, structs.ServiceNodes, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "services")

	// List all the services.
	services, err := tx.Get("services", "service", serviceName)
	if err != nil {
		return 0, nil, fmt.Errorf("failed service lookup: %s", err)
	}
	ws.Add(services.WatchCh())

	var results structs.ServiceNodes
	for service := services.Next(); service != nil; service = services.Next() {
		results = append(results, service.(*structs.ServiceNode))
	}

	// Fill in the node details.
	results, err = s.parseServiceNodes(tx, ws, results)
	if err != nil {
		return 0, nil, fmt.Errorf("failed parsing service nodes: %s", err)
	}
	return idx, results, nil
}

// ServiceTagNodes returns the nodes associated with a given service, filtering
// out services that don't contain the given tag.
func (s *StateStore) ServiceTagNodes(ws memdb.WatchSet, service string, tag string) (uint64, structs.ServiceNodes, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "services")

	// List all the services.
	services, err := tx.Get("services", "service", service)
	if err != nil {
		return 0, nil, fmt.Errorf("failed service lookup: %s", err)
	}
	ws.Add(services.WatchCh())

	// Gather all the services and apply the tag filter.
	var results structs.ServiceNodes
	for service := services.Next(); service != nil; service = services.Next() {
		svc := service.(*structs.ServiceNode)
		if !serviceTagFilter(svc, tag) {
			results = append(results, svc)
		}
	}

	// Fill in the node details.
	results, err = s.parseServiceNodes(tx, ws, results)
	if err != nil {
		return 0, nil, fmt.Errorf("failed parsing service nodes: %s", err)
	}
	return idx, results, nil
}

// serviceTagFilter returns true (should filter) if the given service node
// doesn't contain the given tag.
func serviceTagFilter(sn *structs.ServiceNode, tag string) bool {
	tag = strings.ToLower(tag)

	// Look for the lower cased version of the tag.
	for _, t := range sn.ServiceTags {
		if strings.ToLower(t) == tag {
			return false
		}
	}

	// If we didn't hit the tag above then we should filter.
	return true
}

// parseServiceNodes iterates over a services query and fills in the node details,
// returning a ServiceNodes slice.
func (s *StateStore) parseServiceNodes(tx *memdb.Txn, ws memdb.WatchSet, services structs.ServiceNodes) (structs.ServiceNodes, error) {
	// We don't want to track an unlimited number of nodes, so we pull a
	// top-level watch to use as a fallback.
	allNodes, err := tx.Get("nodes", "id")
	if err != nil {
		return nil, fmt.Errorf("failed nodes lookup: %s", err)
	}
	allNodesCh := allNodes.WatchCh()

	// Fill in the node data for each service instance.
	var results structs.ServiceNodes
	for _, sn := range services {
		// Note that we have to clone here because we don't want to
		// modify the node-related fields on the object in the database,
		// which is what we are referencing.
		s := sn.PartialClone()

		// Grab the corresponding node record.
		watchCh, n, err := tx.FirstWatch("nodes", "id", sn.Node)
		if err != nil {
			return nil, fmt.Errorf("failed node lookup: %s", err)
		}
		ws.AddWithLimit(watchLimit, watchCh, allNodesCh)

		// Populate the node-related fields. The tagged addresses may be
		// used by agents to perform address translation if they are
		// configured to do that.
		node := n.(*structs.Node)
		s.ID = node.ID
		s.Address = node.Address
		s.TaggedAddresses = node.TaggedAddresses
		s.NodeMeta = node.Meta

		results = append(results, s)
	}
	return results, nil
}

// NodeService is used to retrieve a specific service associated with the given
// node.
func (s *StateStore) NodeService(nodeName string, serviceID string) (uint64, *structs.NodeService, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "services")

	// Query the service
	service, err := tx.First("services", "id", nodeName, serviceID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed querying service for node %q: %s", nodeName, err)
	}

	if service != nil {
		return idx, service.(*structs.ServiceNode).ToNodeService(), nil
	} else {
		return idx, nil, nil
	}
}

// NodeServices is used to query service registrations by node name or UUID.
func (s *StateStore) NodeServices(ws memdb.WatchSet, nodeNameOrID string) (uint64, *structs.NodeServices, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "services")

	// Query the node by node name
	watchCh, n, err := tx.FirstWatch("nodes", "id", nodeNameOrID)
	if err != nil {
		return 0, nil, fmt.Errorf("node lookup failed: %s", err)
	}

	if n != nil {
		ws.Add(watchCh)
	} else {
		if len(nodeNameOrID) < minUUIDLookupLen {
			ws.Add(watchCh)
			return 0, nil, nil
		}

		// Attempt to lookup the node by its node ID
		iter, err := tx.Get("nodes", "uuid_prefix", resizeNodeLookupKey(nodeNameOrID))
		if err != nil {
			ws.Add(watchCh)
			// TODO(sean@): We could/should log an error re: the uuid_prefix lookup
			// failing once a logger has been introduced to the catalog.
			return 0, nil, nil
		}

		n = iter.Next()
		if n == nil {
			// No nodes matched, even with the Node ID: add a watch on the node name.
			ws.Add(watchCh)
			return 0, nil, nil
		}

		idWatchCh := iter.WatchCh()
		if iter.Next() != nil {
			// More than one match present: Watch on the node name channel and return
			// an empty result (node lookups can not be ambiguous).
			ws.Add(watchCh)
			return 0, nil, nil
		}

		ws.Add(idWatchCh)
	}

	node := n.(*structs.Node)
	nodeName := node.Node

	// Read all of the services
	services, err := tx.Get("services", "node", nodeName)
	if err != nil {
		return 0, nil, fmt.Errorf("failed querying services for node %q: %s", nodeName, err)
	}
	ws.Add(services.WatchCh())

	// Initialize the node services struct
	ns := &structs.NodeServices{
		Node:     node,
		Services: make(map[string]*structs.NodeService),
	}

	// Add all of the services to the map.
	for service := services.Next(); service != nil; service = services.Next() {
		svc := service.(*structs.ServiceNode).ToNodeService()
		ns.Services[svc.ID] = svc
	}

	return idx, ns, nil
}

// DeleteService is used to delete a given service associated with a node.
func (s *StateStore) DeleteService(idx uint64, nodeName, serviceID string) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the service deletion
	if err := s.deleteServiceTxn(tx, idx, nodeName, serviceID); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// deleteServiceTxn is the inner method called to remove a service
// registration within an existing transaction.
func (s *StateStore) deleteServiceTxn(tx *memdb.Txn, idx uint64, nodeName, serviceID string) error {
	// Look up the service.
	service, err := tx.First("services", "id", nodeName, serviceID)
	if err != nil {
		return fmt.Errorf("failed service lookup: %s", err)
	}
	if service == nil {
		return nil
	}

	// Delete any checks associated with the service. This will invalidate
	// sessions as necessary.
	checks, err := tx.Get("checks", "node_service", nodeName, serviceID)
	if err != nil {
		return fmt.Errorf("failed service check lookup: %s", err)
	}
	var cids []types.CheckID
	for check := checks.Next(); check != nil; check = checks.Next() {
		cids = append(cids, check.(*structs.HealthCheck).CheckID)
	}

	// Do the delete in a separate loop so we don't trash the iterator.
	for _, cid := range cids {
		if err := s.deleteCheckTxn(tx, idx, nodeName, cid); err != nil {
			return err
		}
	}

	// Update the index.
	if err := tx.Insert("index", &IndexEntry{"checks", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	// Delete the service and update the index
	if err := tx.Delete("services", service); err != nil {
		return fmt.Errorf("failed deleting service: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"services", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// EnsureCheck is used to store a check registration in the db.
func (s *StateStore) EnsureCheck(idx uint64, hc *structs.HealthCheck) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the check registration
	if err := s.ensureCheckTxn(tx, idx, hc); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// ensureCheckTransaction is used as the inner method to handle inserting
// a health check into the state store. It ensures safety against inserting
// checks with no matching node or service.
func (s *StateStore) ensureCheckTxn(tx *memdb.Txn, idx uint64, hc *structs.HealthCheck) error {
	// Check if we have an existing health check
	existing, err := tx.First("checks", "id", hc.Node, string(hc.CheckID))
	if err != nil {
		return fmt.Errorf("failed health check lookup: %s", err)
	}

	// Set the indexes
	if existing != nil {
		hc.CreateIndex = existing.(*structs.HealthCheck).CreateIndex
		hc.ModifyIndex = idx
	} else {
		hc.CreateIndex = idx
		hc.ModifyIndex = idx
	}

	// Use the default check status if none was provided
	if hc.Status == "" {
		hc.Status = structs.HealthCritical
	}

	// Get the node
	node, err := tx.First("nodes", "id", hc.Node)
	if err != nil {
		return fmt.Errorf("failed node lookup: %s", err)
	}
	if node == nil {
		return ErrMissingNode
	}

	// If the check is associated with a service, check that we have
	// a registration for the service.
	if hc.ServiceID != "" {
		service, err := tx.First("services", "id", hc.Node, hc.ServiceID)
		if err != nil {
			return fmt.Errorf("failed service lookup: %s", err)
		}
		if service == nil {
			return ErrMissingService
		}

		// Copy in the service name
		hc.ServiceName = service.(*structs.ServiceNode).ServiceName
	}

	// Delete any sessions for this check if the health is critical.
	if hc.Status == structs.HealthCritical {
		mappings, err := tx.Get("session_checks", "node_check", hc.Node, string(hc.CheckID))
		if err != nil {
			return fmt.Errorf("failed session checks lookup: %s", err)
		}

		var ids []string
		for mapping := mappings.Next(); mapping != nil; mapping = mappings.Next() {
			ids = append(ids, mapping.(*sessionCheck).Session)
		}

		// Delete the session in a separate loop so we don't trash the
		// iterator.
		for _, id := range ids {
			if err := s.deleteSessionTxn(tx, idx, id); err != nil {
				return fmt.Errorf("failed deleting session: %s", err)
			}
		}
	}

	// Persist the check registration in the db.
	if err := tx.Insert("checks", hc); err != nil {
		return fmt.Errorf("failed inserting check: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"checks", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// NodeCheck is used to retrieve a specific check associated with the given
// node.
func (s *StateStore) NodeCheck(nodeName string, checkID types.CheckID) (uint64, *structs.HealthCheck, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "checks")

	// Return the check.
	check, err := tx.First("checks", "id", nodeName, string(checkID))
	if err != nil {
		return 0, nil, fmt.Errorf("failed check lookup: %s", err)
	}

	if check != nil {
		return idx, check.(*structs.HealthCheck), nil
	} else {
		return idx, nil, nil
	}
}

// NodeChecks is used to retrieve checks associated with the
// given node from the state store.
func (s *StateStore) NodeChecks(ws memdb.WatchSet, nodeName string) (uint64, structs.HealthChecks, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "checks")

	// Return the checks.
	iter, err := tx.Get("checks", "node", nodeName)
	if err != nil {
		return 0, nil, fmt.Errorf("failed check lookup: %s", err)
	}
	ws.Add(iter.WatchCh())

	var results structs.HealthChecks
	for check := iter.Next(); check != nil; check = iter.Next() {
		results = append(results, check.(*structs.HealthCheck))
	}
	return idx, results, nil
}

// ServiceChecks is used to get all checks associated with a
// given service ID. The query is performed against a service
// _name_ instead of a service ID.
func (s *StateStore) ServiceChecks(ws memdb.WatchSet, serviceName string) (uint64, structs.HealthChecks, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "checks")

	// Return the checks.
	iter, err := tx.Get("checks", "service", serviceName)
	if err != nil {
		return 0, nil, fmt.Errorf("failed check lookup: %s", err)
	}
	ws.Add(iter.WatchCh())

	var results structs.HealthChecks
	for check := iter.Next(); check != nil; check = iter.Next() {
		results = append(results, check.(*structs.HealthCheck))
	}
	return idx, results, nil
}

// ServiceChecksByNodeMeta is used to get all checks associated with a
// given service ID, filtered by the given node metadata values. The query
// is performed against a service _name_ instead of a service ID.
func (s *StateStore) ServiceChecksByNodeMeta(ws memdb.WatchSet, serviceName string,
	filters map[string]string) (uint64, structs.HealthChecks, error) {

	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "checks")

	// Return the checks.
	iter, err := tx.Get("checks", "service", serviceName)
	if err != nil {
		return 0, nil, fmt.Errorf("failed check lookup: %s", err)
	}
	ws.Add(iter.WatchCh())

	return s.parseChecksByNodeMeta(tx, ws, idx, iter, filters)
}

// ChecksInState is used to query the state store for all checks
// which are in the provided state.
func (s *StateStore) ChecksInState(ws memdb.WatchSet, state string) (uint64, structs.HealthChecks, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "checks")

	// Query all checks if HealthAny is passed, otherwise use the index.
	var iter memdb.ResultIterator
	var err error
	if state == structs.HealthAny {
		iter, err = tx.Get("checks", "status")
		if err != nil {
			return 0, nil, fmt.Errorf("failed check lookup: %s", err)
		}
	} else {
		iter, err = tx.Get("checks", "status", state)
		if err != nil {
			return 0, nil, fmt.Errorf("failed check lookup: %s", err)
		}
	}
	ws.Add(iter.WatchCh())

	var results structs.HealthChecks
	for check := iter.Next(); check != nil; check = iter.Next() {
		results = append(results, check.(*structs.HealthCheck))
	}
	return idx, results, nil
}

// ChecksInStateByNodeMeta is used to query the state store for all checks
// which are in the provided state, filtered by the given node metadata values.
func (s *StateStore) ChecksInStateByNodeMeta(ws memdb.WatchSet, state string, filters map[string]string) (uint64, structs.HealthChecks, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "checks")

	// Query all checks if HealthAny is passed, otherwise use the index.
	var iter memdb.ResultIterator
	var err error
	if state == structs.HealthAny {
		iter, err = tx.Get("checks", "status")
		if err != nil {
			return 0, nil, fmt.Errorf("failed check lookup: %s", err)
		}
	} else {
		iter, err = tx.Get("checks", "status", state)
		if err != nil {
			return 0, nil, fmt.Errorf("failed check lookup: %s", err)
		}
	}
	ws.Add(iter.WatchCh())

	return s.parseChecksByNodeMeta(tx, ws, idx, iter, filters)
}

// parseChecksByNodeMeta is a helper function used to deduplicate some
// repetitive code for returning health checks filtered by node metadata fields.
func (s *StateStore) parseChecksByNodeMeta(tx *memdb.Txn, ws memdb.WatchSet,
	idx uint64, iter memdb.ResultIterator, filters map[string]string) (uint64, structs.HealthChecks, error) {

	// We don't want to track an unlimited number of nodes, so we pull a
	// top-level watch to use as a fallback.
	allNodes, err := tx.Get("nodes", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed nodes lookup: %s", err)
	}
	allNodesCh := allNodes.WatchCh()

	// Only take results for nodes that satisfy the node metadata filters.
	var results structs.HealthChecks
	for check := iter.Next(); check != nil; check = iter.Next() {
		healthCheck := check.(*structs.HealthCheck)
		watchCh, node, err := tx.FirstWatch("nodes", "id", healthCheck.Node)
		if err != nil {
			return 0, nil, fmt.Errorf("failed node lookup: %s", err)
		}
		if node == nil {
			return 0, nil, ErrMissingNode
		}

		// Add even the filtered nodes so we wake up if the node metadata
		// changes.
		ws.AddWithLimit(watchLimit, watchCh, allNodesCh)
		if structs.SatisfiesMetaFilters(node.(*structs.Node).Meta, filters) {
			results = append(results, healthCheck)
		}
	}
	return idx, results, nil
}

// DeleteCheck is used to delete a health check registration.
func (s *StateStore) DeleteCheck(idx uint64, node string, checkID types.CheckID) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the check deletion
	if err := s.deleteCheckTxn(tx, idx, node, checkID); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// deleteCheckTxn is the inner method used to call a health
// check deletion within an existing transaction.
func (s *StateStore) deleteCheckTxn(tx *memdb.Txn, idx uint64, node string, checkID types.CheckID) error {
	// Try to retrieve the existing health check.
	hc, err := tx.First("checks", "id", node, string(checkID))
	if err != nil {
		return fmt.Errorf("check lookup failed: %s", err)
	}
	if hc == nil {
		return nil
	}

	// Delete the check from the DB and update the index.
	if err := tx.Delete("checks", hc); err != nil {
		return fmt.Errorf("failed removing check: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"checks", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	// Delete any sessions for this check.
	mappings, err := tx.Get("session_checks", "node_check", node, string(checkID))
	if err != nil {
		return fmt.Errorf("failed session checks lookup: %s", err)
	}
	var ids []string
	for mapping := mappings.Next(); mapping != nil; mapping = mappings.Next() {
		ids = append(ids, mapping.(*sessionCheck).Session)
	}

	// Do the delete in a separate loop so we don't trash the iterator.
	for _, id := range ids {
		if err := s.deleteSessionTxn(tx, idx, id); err != nil {
			return fmt.Errorf("failed deleting session: %s", err)
		}
	}

	return nil
}

// CheckServiceNodes is used to query all nodes and checks for a given service.
func (s *StateStore) CheckServiceNodes(ws memdb.WatchSet, serviceName string) (uint64, structs.CheckServiceNodes, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "services", "checks")

	// Query the state store for the service.
	iter, err := tx.Get("services", "service", serviceName)
	if err != nil {
		return 0, nil, fmt.Errorf("failed service lookup: %s", err)
	}
	ws.Add(iter.WatchCh())

	// Return the results.
	var results structs.ServiceNodes
	for service := iter.Next(); service != nil; service = iter.Next() {
		results = append(results, service.(*structs.ServiceNode))
	}
	return s.parseCheckServiceNodes(tx, ws, idx, serviceName, results, err)
}

// CheckServiceTagNodes is used to query all nodes and checks for a given
// service, filtering out services that don't contain the given tag.
func (s *StateStore) CheckServiceTagNodes(ws memdb.WatchSet, serviceName, tag string) (uint64, structs.CheckServiceNodes, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "services", "checks")

	// Query the state store for the service.
	iter, err := tx.Get("services", "service", serviceName)
	if err != nil {
		return 0, nil, fmt.Errorf("failed service lookup: %s", err)
	}
	ws.Add(iter.WatchCh())

	// Return the results, filtering by tag.
	var results structs.ServiceNodes
	for service := iter.Next(); service != nil; service = iter.Next() {
		svc := service.(*structs.ServiceNode)
		if !serviceTagFilter(svc, tag) {
			results = append(results, svc)
		}
	}
	return s.parseCheckServiceNodes(tx, ws, idx, serviceName, results, err)
}

// parseCheckServiceNodes is used to parse through a given set of services,
// and query for an associated node and a set of checks. This is the inner
// method used to return a rich set of results from a more simple query.
func (s *StateStore) parseCheckServiceNodes(
	tx *memdb.Txn, ws memdb.WatchSet, idx uint64,
	serviceName string, services structs.ServiceNodes,
	err error) (uint64, structs.CheckServiceNodes, error) {
	if err != nil {
		return 0, nil, err
	}

	// Special-case the zero return value to nil, since this ends up in
	// external APIs.
	if len(services) == 0 {
		return idx, nil, nil
	}

	// We don't want to track an unlimited number of nodes, so we pull a
	// top-level watch to use as a fallback.
	allNodes, err := tx.Get("nodes", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed nodes lookup: %s", err)
	}
	allNodesCh := allNodes.WatchCh()

	// We need a similar fallback for checks. Since services need the
	// status of node + service-specific checks, we pull in a top-level
	// watch over all checks.
	allChecks, err := tx.Get("checks", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed checks lookup: %s", err)
	}
	allChecksCh := allChecks.WatchCh()

	results := make(structs.CheckServiceNodes, 0, len(services))
	for _, sn := range services {
		// Retrieve the node.
		watchCh, n, err := tx.FirstWatch("nodes", "id", sn.Node)
		if err != nil {
			return 0, nil, fmt.Errorf("failed node lookup: %s", err)
		}
		ws.AddWithLimit(watchLimit, watchCh, allNodesCh)

		if n == nil {
			return 0, nil, ErrMissingNode
		}
		node := n.(*structs.Node)

		// First add the node-level checks. These always apply to any
		// service on the node.
		var checks structs.HealthChecks
		iter, err := tx.Get("checks", "node_service_check", sn.Node, false)
		if err != nil {
			return 0, nil, err
		}
		ws.AddWithLimit(watchLimit, iter.WatchCh(), allChecksCh)
		for check := iter.Next(); check != nil; check = iter.Next() {
			checks = append(checks, check.(*structs.HealthCheck))
		}

		// Now add the service-specific checks.
		iter, err = tx.Get("checks", "node_service", sn.Node, sn.ServiceID)
		if err != nil {
			return 0, nil, err
		}
		ws.AddWithLimit(watchLimit, iter.WatchCh(), allChecksCh)
		for check := iter.Next(); check != nil; check = iter.Next() {
			checks = append(checks, check.(*structs.HealthCheck))
		}

		// Append to the results.
		results = append(results, structs.CheckServiceNode{
			Node:    node,
			Service: sn.ToNodeService(),
			Checks:  checks,
		})
	}

	return idx, results, nil
}

// NodeInfo is used to generate a dump of a single node. The dump includes
// all services and checks which are registered against the node.
func (s *StateStore) NodeInfo(ws memdb.WatchSet, node string) (uint64, structs.NodeDump, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "services", "checks")

	// Query the node by the passed node
	nodes, err := tx.Get("nodes", "id", node)
	if err != nil {
		return 0, nil, fmt.Errorf("failed node lookup: %s", err)
	}
	ws.Add(nodes.WatchCh())
	return s.parseNodes(tx, ws, idx, nodes)
}

// NodeDump is used to generate a dump of all nodes. This call is expensive
// as it has to query every node, service, and check. The response can also
// be quite large since there is currently no filtering applied.
func (s *StateStore) NodeDump(ws memdb.WatchSet) (uint64, structs.NodeDump, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "nodes", "services", "checks")

	// Fetch all of the registered nodes
	nodes, err := tx.Get("nodes", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed node lookup: %s", err)
	}
	ws.Add(nodes.WatchCh())
	return s.parseNodes(tx, ws, idx, nodes)
}

// parseNodes takes an iterator over a set of nodes and returns a struct
// containing the nodes along with all of their associated services
// and/or health checks.
func (s *StateStore) parseNodes(tx *memdb.Txn, ws memdb.WatchSet, idx uint64,
	iter memdb.ResultIterator) (uint64, structs.NodeDump, error) {

	// We don't want to track an unlimited number of services, so we pull a
	// top-level watch to use as a fallback.
	allServices, err := tx.Get("services", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed services lookup: %s", err)
	}
	allServicesCh := allServices.WatchCh()

	// We need a similar fallback for checks.
	allChecks, err := tx.Get("checks", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed checks lookup: %s", err)
	}
	allChecksCh := allChecks.WatchCh()

	var results structs.NodeDump
	for n := iter.Next(); n != nil; n = iter.Next() {
		node := n.(*structs.Node)

		// Create the wrapped node
		dump := &structs.NodeInfo{
			ID:              node.ID,
			Node:            node.Node,
			Address:         node.Address,
			TaggedAddresses: node.TaggedAddresses,
			Meta:            node.Meta,
		}

		// Query the node services
		services, err := tx.Get("services", "node", node.Node)
		if err != nil {
			return 0, nil, fmt.Errorf("failed services lookup: %s", err)
		}
		ws.AddWithLimit(watchLimit, services.WatchCh(), allServicesCh)
		for service := services.Next(); service != nil; service = services.Next() {
			ns := service.(*structs.ServiceNode).ToNodeService()
			dump.Services = append(dump.Services, ns)
		}

		// Query the node checks
		checks, err := tx.Get("checks", "node", node.Node)
		if err != nil {
			return 0, nil, fmt.Errorf("failed node lookup: %s", err)
		}
		ws.AddWithLimit(watchLimit, checks.WatchCh(), allChecksCh)
		for check := checks.Next(); check != nil; check = checks.Next() {
			hc := check.(*structs.HealthCheck)
			dump.Checks = append(dump.Checks, hc)
		}

		// Add the result to the slice
		results = append(results, dump)
	}
	return idx, results, nil
}
