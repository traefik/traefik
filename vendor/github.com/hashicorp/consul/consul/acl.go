package consul

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/golang-lru"
)

// These must be kept in sync with the constants in command/agent/acl.go.
const (
	// aclNotFound indicates there is no matching ACL.
	aclNotFound = "ACL not found"

	// rootDenied is returned when attempting to resolve a root ACL.
	rootDenied = "Cannot resolve root ACL"

	// permissionDenied is returned when an ACL based rejection happens.
	permissionDenied = "Permission denied"

	// aclDisabled is returned when ACL changes are not permitted since they
	// are disabled.
	aclDisabled = "ACL support disabled"

	// anonymousToken is the token ID we re-write to if there is no token ID
	// provided.
	anonymousToken = "anonymous"

	// redactedToken is shown in structures with embedded tokens when they
	// are not allowed to be displayed.
	redactedToken = "<hidden>"

	// Maximum number of cached ACL entries.
	aclCacheSize = 10 * 1024
)

var (
	permissionDeniedErr = errors.New(permissionDenied)
)

// aclCacheEntry is used to cache non-authoritative ACLs
// If non-authoritative, then we must respect a TTL
type aclCacheEntry struct {
	ACL     acl.ACL
	Expires time.Time
	ETag    string
}

// aclLocalFault is used by the authoritative ACL cache to fault in the rules
// for an ACL if we take a miss. This goes directly to the state store, so it
// assumes its running in the ACL datacenter, or in a non-ACL datacenter when
// using its replicated ACLs during an outage.
func (s *Server) aclLocalFault(id string) (string, string, error) {
	defer metrics.MeasureSince([]string{"consul", "acl", "fault"}, time.Now())

	// Query the state store.
	state := s.fsm.State()
	_, acl, err := state.ACLGet(nil, id)
	if err != nil {
		return "", "", err
	}
	if acl == nil {
		return "", "", errors.New(aclNotFound)
	}

	// Management tokens have no policy and inherit from the 'manage' root
	// policy.
	if acl.Type == structs.ACLTypeManagement {
		return "manage", "", nil
	}

	// Otherwise use the default policy.
	return s.config.ACLDefaultPolicy, acl.Rules, nil
}

// resolveToken is the primary interface used by ACL-checkers (such as an
// endpoint handling a request) to resolve a token. If ACLs aren't enabled
// then this will return a nil token, otherwise it will attempt to use local
// cache and ultimately the ACL datacenter to get the policy associated with the
// token.
func (s *Server) resolveToken(id string) (acl.ACL, error) {
	// Check if there is no ACL datacenter (ACLs disabled)
	authDC := s.config.ACLDatacenter
	if len(authDC) == 0 {
		return nil, nil
	}
	defer metrics.MeasureSince([]string{"consul", "acl", "resolveToken"}, time.Now())

	// Handle the anonymous token
	if len(id) == 0 {
		id = anonymousToken
	} else if acl.RootACL(id) != nil {
		return nil, errors.New(rootDenied)
	}

	// Check if we are the ACL datacenter and the leader, use the
	// authoritative cache
	if s.config.Datacenter == authDC && s.IsLeader() {
		return s.aclAuthCache.GetACL(id)
	}

	// Use our non-authoritative cache
	return s.aclCache.lookupACL(id, authDC)
}

// rpcFn is used to make an RPC call to the client or server.
type rpcFn func(string, interface{}, interface{}) error

// aclCache is used to cache ACLs and policies.
type aclCache struct {
	config *Config
	logger *log.Logger

	// acls is a non-authoritative ACL cache.
	acls *lru.TwoQueueCache

	// aclPolicyCache is a non-authoritative policy cache.
	policies *lru.TwoQueueCache

	// rpc is a function used to talk to the client/server.
	rpc rpcFn

	// local is a function used to look for an ACL locally if replication is
	// enabled. This will be nil if replication isn't enabled.
	local acl.FaultFunc
}

// newAclCache returns a new non-authoritative cache for ACLs. This is used for
// performance, and is used inside the ACL datacenter on non-leader servers, and
// outside the ACL datacenter everywhere.
func newAclCache(conf *Config, logger *log.Logger, rpc rpcFn, local acl.FaultFunc) (*aclCache, error) {
	var err error
	cache := &aclCache{
		config: conf,
		logger: logger,
		rpc:    rpc,
		local:  local,
	}

	// Initialize the non-authoritative ACL cache
	cache.acls, err = lru.New2Q(aclCacheSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ACL cache: %v", err)
	}

	// Initialize the ACL policy cache
	cache.policies, err = lru.New2Q(aclCacheSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ACL policy cache: %v", err)
	}

	return cache, nil
}

// lookupACL is used when we are non-authoritative, and need to resolve an ACL.
func (c *aclCache) lookupACL(id, authDC string) (acl.ACL, error) {
	// Check the cache for the ACL.
	var cached *aclCacheEntry
	raw, ok := c.acls.Get(id)
	if ok {
		cached = raw.(*aclCacheEntry)
	}

	// Check for live cache.
	if cached != nil && time.Now().Before(cached.Expires) {
		metrics.IncrCounter([]string{"consul", "acl", "cache_hit"}, 1)
		return cached.ACL, nil
	} else {
		metrics.IncrCounter([]string{"consul", "acl", "cache_miss"}, 1)
	}

	// Attempt to refresh the policy from the ACL datacenter via an RPC.
	args := structs.ACLPolicyRequest{
		Datacenter: authDC,
		ACL:        id,
	}
	if cached != nil {
		args.ETag = cached.ETag
	}
	var reply structs.ACLPolicy
	err := c.rpc("ACL.GetPolicy", &args, &reply)
	if err == nil {
		return c.useACLPolicy(id, authDC, cached, &reply)
	}

	// Check for not-found, which will cause us to bail immediately. For any
	// other error we report it in the logs but can continue.
	if strings.Contains(err.Error(), aclNotFound) {
		return nil, errors.New(aclNotFound)
	} else {
		c.logger.Printf("[ERR] consul.acl: Failed to get policy from ACL datacenter: %v", err)
	}

	// TODO (slackpad) - We could do a similar thing *within* the ACL
	// datacenter if the leader isn't available. We have a local state
	// store of the ACLs, so by populating the local member in this cache,
	// it would fall back to the state store if there was a leader loss and
	// the extend-cache policy was true. This feels subtle to explain and
	// configure, and leader blips should be paved over by cache already, so
	// we won't do this for now but should consider for the future. This is
	// a lot different than the replication story where you might be cut off
	// from the ACL datacenter for an extended period of time and need to
	// carry on operating with the full set of ACLs as they were known
	// before the partition.

	// At this point we might have an expired cache entry and we know that
	// there was a problem getting the ACL from the ACL datacenter. If a
	// local ACL fault function is registered to query replicated ACL data,
	// and the user's policy allows it, we will try locally before we give
	// up.
	if c.local != nil && c.config.ACLDownPolicy == "extend-cache" {
		parent, rules, err := c.local(id)
		if err != nil {
			// We don't make an exception here for ACLs that aren't
			// found locally. It seems more robust to use an expired
			// cached entry (if we have one) rather than ignore it
			// for the case that replication was a bit behind and
			// didn't have the ACL yet.
			c.logger.Printf("[DEBUG] consul.acl: Failed to get policy from replicated ACLs: %v", err)
			goto ACL_DOWN
		}

		policy, err := acl.Parse(rules)
		if err != nil {
			c.logger.Printf("[DEBUG] consul.acl: Failed to parse policy for replicated ACL: %v", err)
			goto ACL_DOWN
		}
		policy.ID = acl.RuleID(rules)

		// Fake up an ACL datacenter reply and inject it into the cache.
		// Note we use the local TTL here, so this'll be used for that
		// amount of time even once the ACL datacenter becomes available.
		metrics.IncrCounter([]string{"consul", "acl", "replication_hit"}, 1)
		reply.ETag = makeACLETag(parent, policy)
		reply.TTL = c.config.ACLTTL
		reply.Parent = parent
		reply.Policy = policy
		return c.useACLPolicy(id, authDC, cached, &reply)
	}

ACL_DOWN:
	// Unable to refresh, apply the down policy.
	switch c.config.ACLDownPolicy {
	case "allow":
		return acl.AllowAll(), nil
	case "extend-cache":
		if cached != nil {
			return cached.ACL, nil
		}
		fallthrough
	default:
		return acl.DenyAll(), nil
	}
}

// useACLPolicy handles an ACLPolicy response
func (c *aclCache) useACLPolicy(id, authDC string, cached *aclCacheEntry, p *structs.ACLPolicy) (acl.ACL, error) {
	// Check if we can used the cached policy
	if cached != nil && cached.ETag == p.ETag {
		if p.TTL > 0 {
			// TODO (slackpad) - This seems like it's an unsafe
			// write.
			cached.Expires = time.Now().Add(p.TTL)
		}
		return cached.ACL, nil
	}

	// Check for a cached compiled policy
	var compiled acl.ACL
	raw, ok := c.policies.Get(p.ETag)
	if ok {
		compiled = raw.(acl.ACL)
	} else {
		// Resolve the parent policy
		parent := acl.RootACL(p.Parent)
		if parent == nil {
			var err error
			parent, err = c.lookupACL(p.Parent, authDC)
			if err != nil {
				return nil, err
			}
		}

		// Compile the ACL
		acl, err := acl.New(parent, p.Policy)
		if err != nil {
			return nil, err
		}

		// Cache the policy
		c.policies.Add(p.ETag, acl)
		compiled = acl
	}

	// Cache the ACL
	cached = &aclCacheEntry{
		ACL:  compiled,
		ETag: p.ETag,
	}
	if p.TTL > 0 {
		cached.Expires = time.Now().Add(p.TTL)
	}
	c.acls.Add(id, cached)
	return compiled, nil
}

// aclFilter is used to filter results from our state store based on ACL rules
// configured for the provided token.
type aclFilter struct {
	acl             acl.ACL
	logger          *log.Logger
	enforceVersion8 bool
}

// newAclFilter constructs a new aclFilter.
func newAclFilter(acl acl.ACL, logger *log.Logger, enforceVersion8 bool) *aclFilter {
	if logger == nil {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}
	return &aclFilter{
		acl:             acl,
		logger:          logger,
		enforceVersion8: enforceVersion8,
	}
}

// allowNode is used to determine if a node is accessible for an ACL.
func (f *aclFilter) allowNode(node string) bool {
	if !f.enforceVersion8 {
		return true
	}
	return f.acl.NodeRead(node)
}

// allowService is used to determine if a service is accessible for an ACL.
func (f *aclFilter) allowService(service string) bool {
	if service == "" {
		return true
	}

	if !f.enforceVersion8 && service == ConsulServiceID {
		return true
	}

	return f.acl.ServiceRead(service)
}

// allowSession is used to determine if a session for a node is accessible for
// an ACL.
func (f *aclFilter) allowSession(node string) bool {
	if !f.enforceVersion8 {
		return true
	}
	return f.acl.SessionRead(node)
}

// filterHealthChecks is used to filter a set of health checks down based on
// the configured ACL rules for a token.
func (f *aclFilter) filterHealthChecks(checks *structs.HealthChecks) {
	hc := *checks
	for i := 0; i < len(hc); i++ {
		check := hc[i]
		if f.allowNode(check.Node) && f.allowService(check.ServiceName) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping check %q from result due to ACLs", check.CheckID)
		hc = append(hc[:i], hc[i+1:]...)
		i--
	}
	*checks = hc
}

// filterServices is used to filter a set of services based on ACLs.
func (f *aclFilter) filterServices(services structs.Services) {
	for svc, _ := range services {
		if f.allowService(svc) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping service %q from result due to ACLs", svc)
		delete(services, svc)
	}
}

// filterServiceNodes is used to filter a set of nodes for a given service
// based on the configured ACL rules.
func (f *aclFilter) filterServiceNodes(nodes *structs.ServiceNodes) {
	sn := *nodes
	for i := 0; i < len(sn); i++ {
		node := sn[i]
		if f.allowNode(node.Node) && f.allowService(node.ServiceName) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping node %q from result due to ACLs", node.Node)
		sn = append(sn[:i], sn[i+1:]...)
		i--
	}
	*nodes = sn
}

// filterNodeServices is used to filter services on a given node base on ACLs.
func (f *aclFilter) filterNodeServices(services **structs.NodeServices) {
	if *services == nil {
		return
	}

	if !f.allowNode((*services).Node.Node) {
		*services = nil
		return
	}

	for svc, _ := range (*services).Services {
		if f.allowService(svc) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping service %q from result due to ACLs", svc)
		delete((*services).Services, svc)
	}
}

// filterCheckServiceNodes is used to filter nodes based on ACL rules.
func (f *aclFilter) filterCheckServiceNodes(nodes *structs.CheckServiceNodes) {
	csn := *nodes
	for i := 0; i < len(csn); i++ {
		node := csn[i]
		if f.allowNode(node.Node.Node) && f.allowService(node.Service.Service) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping node %q from result due to ACLs", node.Node.Node)
		csn = append(csn[:i], csn[i+1:]...)
		i--
	}
	*nodes = csn
}

// filterSessions is used to filter a set of sessions based on ACLs.
func (f *aclFilter) filterSessions(sessions *structs.Sessions) {
	s := *sessions
	for i := 0; i < len(s); i++ {
		session := s[i]
		if f.allowSession(session.Node) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping session %q from result due to ACLs", session.ID)
		s = append(s[:i], s[i+1:]...)
		i--
	}
	*sessions = s
}

// filterCoordinates is used to filter nodes in a coordinate dump based on ACL
// rules.
func (f *aclFilter) filterCoordinates(coords *structs.Coordinates) {
	c := *coords
	for i := 0; i < len(c); i++ {
		node := c[i].Node
		if f.allowNode(node) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping node %q from result due to ACLs", node)
		c = append(c[:i], c[i+1:]...)
		i--
	}
	*coords = c
}

// filterNodeDump is used to filter through all parts of a node dump and
// remove elements the provided ACL token cannot access.
func (f *aclFilter) filterNodeDump(dump *structs.NodeDump) {
	nd := *dump
	for i := 0; i < len(nd); i++ {
		info := nd[i]

		// Filter nodes
		if node := info.Node; !f.allowNode(node) {
			f.logger.Printf("[DEBUG] consul: dropping node %q from result due to ACLs", node)
			nd = append(nd[:i], nd[i+1:]...)
			i--
			continue
		}

		// Filter services
		for j := 0; j < len(info.Services); j++ {
			svc := info.Services[j].Service
			if f.allowService(svc) {
				continue
			}
			f.logger.Printf("[DEBUG] consul: dropping service %q from result due to ACLs", svc)
			info.Services = append(info.Services[:j], info.Services[j+1:]...)
			j--
		}

		// Filter checks
		for j := 0; j < len(info.Checks); j++ {
			chk := info.Checks[j]
			if f.allowService(chk.ServiceName) {
				continue
			}
			f.logger.Printf("[DEBUG] consul: dropping check %q from result due to ACLs", chk.CheckID)
			info.Checks = append(info.Checks[:j], info.Checks[j+1:]...)
			j--
		}
	}
	*dump = nd
}

// filterNodes is used to filter through all parts of a node list and remove
// elements the provided ACL token cannot access.
func (f *aclFilter) filterNodes(nodes *structs.Nodes) {
	n := *nodes
	for i := 0; i < len(n); i++ {
		node := n[i].Node
		if f.allowNode(node) {
			continue
		}
		f.logger.Printf("[DEBUG] consul: dropping node %q from result due to ACLs", node)
		n = append(n[:i], n[i+1:]...)
		i--
	}
	*nodes = n
}

// redactPreparedQueryTokens will redact any tokens unless the client has a
// management token. This eases the transition to delegated authority over
// prepared queries, since it was easy to capture management tokens in Consul
// 0.6.3 and earlier, and we don't want to willy-nilly show those. This does
// have the limitation of preventing delegated non-management users from seeing
// captured tokens, but they can at least see whether or not a token is set.
func (f *aclFilter) redactPreparedQueryTokens(query **structs.PreparedQuery) {
	// Management tokens can see everything with no filtering.
	if f.acl.ACLList() {
		return
	}

	// Let the user see if there's a blank token, otherwise we need
	// to redact it, since we know they don't have a management
	// token.
	if (*query).Token != "" {
		// Redact the token, using a copy of the query structure
		// since we could be pointed at a live instance from the
		// state store so it's not safe to modify it. Note that
		// this clone will still point to things like underlying
		// arrays in the original, but for modifying just the
		// token it will be safe to use.
		clone := *(*query)
		clone.Token = redactedToken
		*query = &clone
	}
}

// filterPreparedQueries is used to filter prepared queries based on ACL rules.
// We prune entries the user doesn't have access to, and we redact any tokens
// if the user doesn't have a management token.
func (f *aclFilter) filterPreparedQueries(queries *structs.PreparedQueries) {
	// Management tokens can see everything with no filtering.
	if f.acl.ACLList() {
		return
	}

	// Otherwise, we need to see what the token has access to.
	ret := make(structs.PreparedQueries, 0, len(*queries))
	for _, query := range *queries {
		// If no prefix ACL applies to this query then filter it, since
		// we know at this point the user doesn't have a management
		// token, otherwise see what the policy says.
		prefix, ok := query.GetACLPrefix()
		if !ok || !f.acl.PreparedQueryRead(prefix) {
			f.logger.Printf("[DEBUG] consul: dropping prepared query %q from result due to ACLs", query.ID)
			continue
		}

		// Redact any tokens if necessary. We make a copy of just the
		// pointer so we don't mess with the caller's slice.
		final := query
		f.redactPreparedQueryTokens(&final)
		ret = append(ret, final)
	}
	*queries = ret
}

// filterACL is used to filter results from our service catalog based on the
// rules configured for the provided token. The subject is scrubbed and
// modified in-place, leaving only resources the token can access.
func (s *Server) filterACL(token string, subj interface{}) error {
	// Get the ACL from the token
	acl, err := s.resolveToken(token)
	if err != nil {
		return err
	}

	// Fast path if ACLs are not enabled
	if acl == nil {
		return nil
	}

	// Create the filter
	filt := newAclFilter(acl, s.logger, s.config.ACLEnforceVersion8)

	switch v := subj.(type) {
	case *structs.CheckServiceNodes:
		filt.filterCheckServiceNodes(v)

	case *structs.IndexedCheckServiceNodes:
		filt.filterCheckServiceNodes(&v.Nodes)

	case *structs.IndexedCoordinates:
		filt.filterCoordinates(&v.Coordinates)

	case *structs.IndexedHealthChecks:
		filt.filterHealthChecks(&v.HealthChecks)

	case *structs.IndexedNodeDump:
		filt.filterNodeDump(&v.Dump)

	case *structs.IndexedNodes:
		filt.filterNodes(&v.Nodes)

	case *structs.IndexedNodeServices:
		filt.filterNodeServices(&v.NodeServices)

	case *structs.IndexedServiceNodes:
		filt.filterServiceNodes(&v.ServiceNodes)

	case *structs.IndexedServices:
		filt.filterServices(v.Services)

	case *structs.IndexedSessions:
		filt.filterSessions(&v.Sessions)

	case *structs.IndexedPreparedQueries:
		filt.filterPreparedQueries(&v.Queries)

	case **structs.PreparedQuery:
		filt.redactPreparedQueryTokens(v)

	default:
		panic(fmt.Errorf("Unhandled type passed to ACL filter: %#v", subj))
	}

	return nil
}

// vetRegisterWithACL applies the given ACL's policy to the catalog update and
// determines if it is allowed. Since the catalog register request is so
// dynamic, this is a pretty complex algorithm and was worth breaking out of the
// endpoint. The NodeServices record for the node must be supplied, and can be
// nil.
//
// This is a bit racy because we have to check the state store outside of a
// transaction. It's the best we can do because we don't want to flow ACL
// checking down there. The node information doesn't change in practice, so this
// will be fine. If we expose ways to change node addresses in a later version,
// then we should split the catalog API at the node and service level so we can
// address this race better (even then it would be super rare, and would at
// worst let a service update revert a recent node update, so it doesn't open up
// too much abuse).
func vetRegisterWithACL(acl acl.ACL, subj *structs.RegisterRequest,
	ns *structs.NodeServices) error {
	// Fast path if ACLs are not enabled.
	if acl == nil {
		return nil
	}

	// Vet the node info. This allows service updates to re-post the required
	// node info for each request without having to have node "write"
	// privileges.
	needsNode := ns == nil || subj.ChangesNode(ns.Node)
	if needsNode && !acl.NodeWrite(subj.Node) {
		return permissionDeniedErr
	}

	// Vet the service change. This includes making sure they can register
	// the given service, and that we can write to any existing service that
	// is being modified by id (if any).
	if subj.Service != nil {
		if !acl.ServiceWrite(subj.Service.Service) {
			return permissionDeniedErr
		}

		if ns != nil {
			other, ok := ns.Services[subj.Service.ID]
			if ok && !acl.ServiceWrite(other.Service) {
				return permissionDeniedErr
			}
		}
	}

	// Make sure that the member was flattened before we got there. This
	// keeps us from having to verify this check as well.
	if subj.Check != nil {
		return fmt.Errorf("check member must be nil")
	}

	// Vet the checks. Node-level checks require node write, and
	// service-level checks require service write.
	for _, check := range subj.Checks {
		// Make sure that the node matches - we don't allow you to mix
		// checks from other nodes because we'd have to pull a bunch
		// more state store data to check this. If ACLs are enabled then
		// we simply require them to match in a given request. There's a
		// note in state_store.go to ban this down there in Consul 0.8,
		// but it's good to leave this here because it's required for
		// correctness wrt. ACLs.
		if check.Node != subj.Node {
			return fmt.Errorf("Node '%s' for check '%s' doesn't match register request node '%s'",
				check.Node, check.CheckID, subj.Node)
		}

		// Node-level check.
		if check.ServiceID == "" {
			if !acl.NodeWrite(subj.Node) {
				return permissionDeniedErr
			}
			continue
		}

		// Service-level check, check the common case where it
		// matches the service part of this request, which has
		// already been vetted above, and might be being registered
		// along with its checks.
		if subj.Service != nil && subj.Service.ID == check.ServiceID {
			continue
		}

		// Service-level check for some other service. Make sure they've
		// got write permissions for that service.
		if ns == nil {
			return fmt.Errorf("Unknown service '%s' for check '%s'",
				check.ServiceID, check.CheckID)
		} else {
			other, ok := ns.Services[check.ServiceID]
			if !ok {
				return fmt.Errorf("Unknown service '%s' for check '%s'",
					check.ServiceID, check.CheckID)
			}
			if !acl.ServiceWrite(other.Service) {
				return permissionDeniedErr
			}
		}
	}

	return nil
}

// vetDeregisterWithACL applies the given ACL's policy to the catalog update and
// determines if it is allowed. Since the catalog deregister request is so
// dynamic, this is a pretty complex algorithm and was worth breaking out of the
// endpoint. The NodeService for the referenced service must be supplied, and can
// be nil; similar for the HealthCheck for the referenced health check.
func vetDeregisterWithACL(acl acl.ACL, subj *structs.DeregisterRequest,
	ns *structs.NodeService, nc *structs.HealthCheck) error {
	// Fast path if ACLs are not enabled.
	if acl == nil {
		return nil
	}

	// This order must match the code in applyRegister() in fsm.go since it
	// also evaluates things in this order, and will ignore fields based on
	// this precedence. This lets us also ignore them from an ACL perspective.
	if subj.ServiceID != "" {
		if ns == nil {
			return fmt.Errorf("Unknown service '%s'", subj.ServiceID)
		}
		if !acl.ServiceWrite(ns.Service) {
			return permissionDeniedErr
		}
	} else if subj.CheckID != "" {
		if nc == nil {
			return fmt.Errorf("Unknown check '%s'", subj.CheckID)
		}
		if nc.ServiceID != "" {
			if !acl.ServiceWrite(nc.ServiceName) {
				return permissionDeniedErr
			}
		} else {
			if !acl.NodeWrite(subj.Node) {
				return permissionDeniedErr
			}
		}
	} else {
		if !acl.NodeWrite(subj.Node) {
			return permissionDeniedErr
		}
	}

	return nil
}
