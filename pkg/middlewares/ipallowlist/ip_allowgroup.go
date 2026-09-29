package ipallowlist

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/ip"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
)

const groupTypeName = "IPAllowGrouper"

// allowEntry holds the parsed allow-checker and strategy for one IPAllowList rule.
type allowEntry struct {
	checker          *ip.Checker
	strategy         ip.Strategy
	rejectStatusCode int
}

// ipAllowGrouper is a middleware that grants access if the client IP matches ANY
// of its configured IPAllowList rules (OR logic).
type ipAllowGrouper struct {
	next    http.Handler
	entries []allowEntry
	name    string
}

// NewGroup builds a new IPAllowGrouper middleware from an IPAllowGroup config.
// It fails if Rules is empty, or if any individual rule has an invalid CIDR or strategy.
func NewGroup(ctx context.Context, next http.Handler, config dynamic.IPAllowGroup, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, groupTypeName)
	logger.Debug().Msg("Creating middleware")

	if len(config.Rules) == 0 {
		return nil, errors.New("ipAllowGroup: rules is empty, at least one rule must be provided")
	}

	entries := make([]allowEntry, 0, len(config.Rules))

	for i, rule := range config.Rules {
		if len(rule.SourceRange) == 0 {
			return nil, fmt.Errorf("ipAllowGroup: rule[%d] has an empty sourceRange", i)
		}

		rejectStatusCode := rule.RejectStatusCode
		if rejectStatusCode == 0 {
			rejectStatusCode = http.StatusForbidden
		} else if http.StatusText(rejectStatusCode) == "" {
			return nil, fmt.Errorf("ipAllowGroup: rule[%d] has invalid HTTP status code %d", i, rejectStatusCode)
		}

		checker, err := ip.NewChecker(rule.SourceRange)
		if err != nil {
			return nil, fmt.Errorf("ipAllowGroup: rule[%d] cannot parse CIDRs %s: %w", i, rule.SourceRange, err)
		}

		strategy, err := rule.IPStrategy.Get()
		if err != nil {
			return nil, fmt.Errorf("ipAllowGroup: rule[%d] invalid IP strategy: %w", i, err)
		}

		entries = append(entries, allowEntry{
			checker:          checker,
			strategy:         strategy,
			rejectStatusCode: rejectStatusCode,
		})
	}

	logger.Debug().Msgf("Setting up IPAllowGrouper with %d rules", len(entries))

	return &ipAllowGrouper{
		next:    next,
		entries: entries,
		name:    name,
	}, nil
}

// GetTracingInformation satisfies the traceable middleware interface.
func (ag *ipAllowGrouper) GetTracingInformation() (string, string) {
	return ag.name, groupTypeName
}

// ServeHTTP checks the client IP against each rule in order.
// If any rule authorizes the IP the request is forwarded (OR logic / short-circuit).
// If no rule matches, the request is rejected with the status code from the last rule
// (defaulting to 403 Forbidden).
func (ag *ipAllowGrouper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), ag.name, groupTypeName)
	ctx := logger.WithContext(req.Context())

	// We use the last entry's rejectStatusCode as the rejection code when no rule matches.
	// All entries always default to 403 unless explicitly overridden.
	rejectCode := http.StatusForbidden
	if len(ag.entries) > 0 {
		rejectCode = ag.entries[len(ag.entries)-1].rejectStatusCode
	}

	for _, entry := range ag.entries {
		clientIP := entry.strategy.GetIP(req)
		if err := entry.checker.IsAuthorized(clientIP); err == nil {
			logger.Debug().Msgf("IPAllowGrouper: accepting IP %s (matched a rule)", clientIP)
			ag.next.ServeHTTP(rw, req)
			return
		}
	}

	// Determine the IP for logging from the first entry's strategy (best effort).
	logIP := req.RemoteAddr
	if len(ag.entries) > 0 {
		logIP = ag.entries[0].strategy.GetIP(req)
	}

	logger.Debug().Msgf("IPAllowGrouper: rejecting IP %s (no rule matched)", logIP)
	observability.SetStatusErrorf(req.Context(), "IPAllowGrouper: rejecting IP %s (no rule matched)", logIP)
	reject(ctx, rejectCode, rw)
}
