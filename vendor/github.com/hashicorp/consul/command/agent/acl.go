package agent

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/golang-lru"
	"github.com/hashicorp/serf/serf"
)

// There's enough behavior difference with client-side ACLs that we've
// intentionally kept this code separate from the server-side ACL code in
// consul/acl.go. We may refactor some of the caching logic in the future,
// but for now we are developing this separately to see how things shake out.

// These must be kept in sync with the constants in consul/acl.go.
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

	// Maximum number of cached ACL entries.
	aclCacheSize = 10 * 1024
)

var (
	permissionDeniedErr = errors.New(permissionDenied)
)

// aclCacheEntry is used to cache ACL tokens.
type aclCacheEntry struct {
	// ACL is the cached ACL.
	ACL acl.ACL

	// Expires is set based on the TTL for the ACL.
	Expires time.Time

	// ETag is used as an optimization when fetching ACLs from servers to
	// avoid transmitting data back when the agent has a good copy, which is
	// usually the case when refreshing a TTL.
	ETag string
}

// aclManager is used by the agent to keep track of state related to ACLs,
// including caching tokens from the servers. This has some internal state that
// we don't want to dump into the agent itself.
type aclManager struct {
	// acls is a cache mapping ACL tokens to compiled policies.
	acls *lru.TwoQueueCache

	// master is the ACL to use when the agent master token is supplied.
	// This may be nil if that option isn't set in the agent config.
	master acl.ACL

	// down is the ACL to use when the servers are down. This may be nil
	// which means to try and use the cached policy if there is one (or
	// deny if there isn't a policy in the cache).
	down acl.ACL

	// disabled is used to keep track of feedback from the servers that ACLs
	// are disabled. If the manager discovers that ACLs are disabled, this
	// will be set to the next time we should check to see if they have been
	// enabled. This helps cut useless traffic, but allows us to turn on ACL
	// support at the servers without having to restart the whole cluster.
	disabled     time.Time
	disabledLock sync.RWMutex
}

// newACLManager returns an ACL manager based on the given config.
func newACLManager(config *Config) (*aclManager, error) {
	// Set up the cache from ID to ACL (we don't cache policies like the
	// servers; only one level).
	acls, err := lru.New2Q(aclCacheSize)
	if err != nil {
		return nil, err
	}

	// If an agent master token is configured, build a policy and ACL for
	// it, otherwise leave it nil.
	var master acl.ACL
	if len(config.ACLAgentMasterToken) > 0 {
		policy := &acl.Policy{
			Agents: []*acl.AgentPolicy{
				&acl.AgentPolicy{
					Node:   config.NodeName,
					Policy: acl.PolicyWrite,
				},
			},
		}
		acl, err := acl.New(acl.DenyAll(), policy)
		if err != nil {
			return nil, err
		}
		master = acl
	}

	var down acl.ACL
	switch config.ACLDownPolicy {
	case "allow":
		down = acl.AllowAll()
	case "deny":
		down = acl.DenyAll()
	case "extend-cache":
		// Leave the down policy as nil to signal this.
	default:
		return nil, fmt.Errorf("invalid ACL down policy %q", config.ACLDownPolicy)
	}

	// Give back a manager.
	return &aclManager{
		acls:   acls,
		master: master,
		down:   down,
	}, nil
}

// isDisabled returns true if the manager has discovered that ACLs are disabled
// on the servers.
func (m *aclManager) isDisabled() bool {
	m.disabledLock.RLock()
	defer m.disabledLock.RUnlock()
	return time.Now().Before(m.disabled)
}

// lookupACL attempts to locate the compiled policy associated with the given
// token. The agent may be used to perform RPC calls to the servers to fetch
// policies that aren't in the cache.
func (m *aclManager) lookupACL(agent *Agent, id string) (acl.ACL, error) {
	// Handle some special cases for the ID.
	if len(id) == 0 {
		id = anonymousToken
	} else if acl.RootACL(id) != nil {
		return nil, errors.New(rootDenied)
	} else if m.master != nil && id == agent.config.ACLAgentMasterToken {
		return m.master, nil
	}

	// Try the cache first.
	var cached *aclCacheEntry
	if raw, ok := m.acls.Get(id); ok {
		cached = raw.(*aclCacheEntry)
	}
	if cached != nil && time.Now().Before(cached.Expires) {
		metrics.IncrCounter([]string{"consul", "acl", "cache_hit"}, 1)
		return cached.ACL, nil
	} else {
		metrics.IncrCounter([]string{"consul", "acl", "cache_miss"}, 1)
	}

	// At this point we might have a stale cached ACL, or none at all, so
	// try to contact the servers.
	args := structs.ACLPolicyRequest{
		Datacenter: agent.config.ACLDatacenter,
		ACL:        id,
	}
	if cached != nil {
		args.ETag = cached.ETag
	}
	var reply structs.ACLPolicy
	err := agent.RPC(agent.getEndpoint("ACL")+".GetPolicy", &args, &reply)
	if err != nil {
		if strings.Contains(err.Error(), aclDisabled) {
			agent.logger.Printf("[DEBUG] agent: ACLs disabled on servers, will check again after %s", agent.config.ACLDisabledTTL)
			m.disabledLock.Lock()
			m.disabled = time.Now().Add(agent.config.ACLDisabledTTL)
			m.disabledLock.Unlock()
			return nil, nil
		} else if strings.Contains(err.Error(), aclNotFound) {
			return nil, errors.New(aclNotFound)
		} else {
			agent.logger.Printf("[DEBUG] agent: Failed to get policy for ACL from servers: %v", err)
			if m.down != nil {
				return m.down, nil
			} else if cached != nil {
				return cached.ACL, nil
			} else {
				return acl.DenyAll(), nil
			}
		}
	}

	// Use the old cached compiled ACL if we can, otherwise compile it and
	// resolve any parents.
	var compiled acl.ACL
	if cached != nil && cached.ETag == reply.ETag {
		compiled = cached.ACL
	} else {
		parent := acl.RootACL(reply.Parent)
		if parent == nil {
			parent, err = m.lookupACL(agent, reply.Parent)
			if err != nil {
				return nil, err
			}
		}

		acl, err := acl.New(parent, reply.Policy)
		if err != nil {
			return nil, err
		}
		compiled = acl
	}

	// Update the cache.
	cached = &aclCacheEntry{
		ACL:  compiled,
		ETag: reply.ETag,
	}
	if reply.TTL > 0 {
		cached.Expires = time.Now().Add(reply.TTL)
	}
	m.acls.Add(id, cached)
	return compiled, nil
}

// resolveToken is the primary interface used by ACL-checkers in the agent
// endpoints, which is the one place where we do some ACL enforcement on
// clients. Some of the enforcement is normative (e.g. self and monitor)
// and some is informative (e.g. catalog and health).
func (a *Agent) resolveToken(id string) (acl.ACL, error) {
	// Disable ACLs if version 8 enforcement isn't enabled.
	if !(*a.config.ACLEnforceVersion8) {
		return nil, nil
	}

	// Bail if there's no ACL datacenter configured. This means that agent
	// enforcement isn't on.
	if a.config.ACLDatacenter == "" {
		return nil, nil
	}

	// Bail if the ACL manager is disabled. This happens if it gets feedback
	// from the servers that ACLs are disabled.
	if a.acls.isDisabled() {
		return nil, nil
	}

	// This will look in the cache and fetch from the servers if necessary.
	return a.acls.lookupACL(a, id)
}

// vetServiceRegister makes sure the service registration action is allowed by
// the given token.
func (a *Agent) vetServiceRegister(token string, service *structs.NodeService) error {
	// Resolve the token and bail if ACLs aren't enabled.
	acl, err := a.resolveToken(token)
	if err != nil {
		return err
	}
	if acl == nil {
		return nil
	}

	// Vet the service itself.
	if !acl.ServiceWrite(service.Service) {
		return permissionDeniedErr
	}

	// Vet any service that might be getting overwritten.
	services := a.state.Services()
	if existing, ok := services[service.ID]; ok {
		if !acl.ServiceWrite(existing.Service) {
			return permissionDeniedErr
		}
	}

	return nil
}

// vetServiceUpdate makes sure the service update action is allowed by the given
// token.
func (a *Agent) vetServiceUpdate(token string, serviceID string) error {
	// Resolve the token and bail if ACLs aren't enabled.
	acl, err := a.resolveToken(token)
	if err != nil {
		return err
	}
	if acl == nil {
		return nil
	}

	// Vet any changes based on the existing services's info.
	services := a.state.Services()
	if existing, ok := services[serviceID]; ok {
		if !acl.ServiceWrite(existing.Service) {
			return permissionDeniedErr
		}
	} else {
		return fmt.Errorf("Unknown service %q", serviceID)
	}

	return nil
}

// vetCheckRegister makes sure the check registration action is allowed by the
// given token.
func (a *Agent) vetCheckRegister(token string, check *structs.HealthCheck) error {
	// Resolve the token and bail if ACLs aren't enabled.
	acl, err := a.resolveToken(token)
	if err != nil {
		return err
	}
	if acl == nil {
		return nil
	}

	// Vet the check itself.
	if len(check.ServiceName) > 0 {
		if !acl.ServiceWrite(check.ServiceName) {
			return permissionDeniedErr
		}
	} else {
		if !acl.NodeWrite(a.config.NodeName) {
			return permissionDeniedErr
		}
	}

	// Vet any check that might be getting overwritten.
	checks := a.state.Checks()
	if existing, ok := checks[check.CheckID]; ok {
		if len(existing.ServiceName) > 0 {
			if !acl.ServiceWrite(existing.ServiceName) {
				return permissionDeniedErr
			}
		} else {
			if !acl.NodeWrite(a.config.NodeName) {
				return permissionDeniedErr
			}
		}
	}

	return nil
}

// vetCheckUpdate makes sure that a check update is allowed by the given token.
func (a *Agent) vetCheckUpdate(token string, checkID types.CheckID) error {
	// Resolve the token and bail if ACLs aren't enabled.
	acl, err := a.resolveToken(token)
	if err != nil {
		return err
	}
	if acl == nil {
		return nil
	}

	// Vet any changes based on the existing check's info.
	checks := a.state.Checks()
	if existing, ok := checks[checkID]; ok {
		if len(existing.ServiceName) > 0 {
			if !acl.ServiceWrite(existing.ServiceName) {
				return permissionDeniedErr
			}
		} else {
			if !acl.NodeWrite(a.config.NodeName) {
				return permissionDeniedErr
			}
		}
	} else {
		return fmt.Errorf("Unknown check %q", checkID)
	}

	return nil
}

// filterMembers redacts members that the token doesn't have access to.
func (a *Agent) filterMembers(token string, members *[]serf.Member) error {
	// Resolve the token and bail if ACLs aren't enabled.
	acl, err := a.resolveToken(token)
	if err != nil {
		return err
	}
	if acl == nil {
		return nil
	}

	// Filter out members based on the node policy.
	m := *members
	for i := 0; i < len(m); i++ {
		node := m[i].Name
		if acl.NodeRead(node) {
			continue
		}
		a.logger.Printf("[DEBUG] agent: dropping node %q from result due to ACLs", node)
		m = append(m[:i], m[i+1:]...)
		i--
	}
	*members = m
	return nil
}

// filterServices redacts services that the token doesn't have access to.
func (a *Agent) filterServices(token string, services *map[string]*structs.NodeService) error {
	// Resolve the token and bail if ACLs aren't enabled.
	acl, err := a.resolveToken(token)
	if err != nil {
		return err
	}
	if acl == nil {
		return nil
	}

	// Filter out services based on the service policy.
	for id, service := range *services {
		if acl.ServiceRead(service.Service) {
			continue
		}
		a.logger.Printf("[DEBUG] agent: dropping service %q from result due to ACLs", id)
		delete(*services, id)
	}
	return nil
}

// filterChecks redacts checks that the token doesn't have access to.
func (a *Agent) filterChecks(token string, checks *map[types.CheckID]*structs.HealthCheck) error {
	// Resolve the token and bail if ACLs aren't enabled.
	acl, err := a.resolveToken(token)
	if err != nil {
		return err
	}
	if acl == nil {
		return nil
	}

	// Filter out checks based on the node or service policy.
	for id, check := range *checks {
		if len(check.ServiceName) > 0 {
			if acl.ServiceRead(check.ServiceName) {
				continue
			}
		} else {
			if acl.NodeRead(a.config.NodeName) {
				continue
			}
		}
		a.logger.Printf("[DEBUG] agent: dropping check %q from result due to ACLs", id)
		delete(*checks, id)
	}
	return nil
}
