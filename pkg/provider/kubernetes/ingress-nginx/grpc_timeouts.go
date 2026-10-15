package ingressnginx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"k8s.io/utils/ptr"
)

const (
	directiveGRPCConnectTimeout = "grpc_connect_timeout"
	directiveGRPCReadTimeout    = "grpc_read_timeout"
	directiveGRPCSendTimeout    = "grpc_send_timeout"
)

// grpcTimeouts holds the gRPC timeouts set by snippet directives.
// A nil field means that the directive is not set.
type grpcTimeouts struct {
	connect *time.Duration
	read    *time.Duration
	send    *time.Duration
}

// orElse returns t, with the unset fields taken from fallback.
func (t grpcTimeouts) orElse(fallback grpcTimeouts) grpcTimeouts {
	if t.connect == nil {
		t.connect = fallback.connect
	}
	if t.read == nil {
		t.read = fallback.read
	}
	if t.send == nil {
		t.send = fallback.send
	}
	return t
}

// snippetGRPCTimeouts returns the gRPC timeouts that the snippets set for an Ingress.
//
// The grpc_*_timeout directives only change the requests that NGINX sends with grpc_pass.
// ingress-nginx uses grpc_pass only when the backend-protocol is GRPC or GRPCS,
// so the directives are ignored for the other protocols.
//
// A configuration-snippet value wins over a server-snippet value,
// because in NGINX a location value wins over a server value.
//
// A server-snippet is set per host and applies to every Ingress on that host.
// The ServersTransport is per Ingress, so if the Ingress has several hosts,
// the first host in the rules that sets a directive gives its value.
func snippetGRPCTimeouts(ctx context.Context, ing ingressEntry, serverSnippets map[string]string) grpcTimeouts {
	switch strings.ToUpper(ptr.Deref(ing.config.BackendProtocol, "")) {
	case "GRPC", "GRPCS":
	default:
		return grpcTimeouts{}
	}

	location := parseGRPCTimeouts(ctx, "location /", ptr.Deref(ing.config.ConfigurationSnippet, ""))

	var server grpcTimeouts
	seenHosts := make(map[string]struct{})
	for _, rule := range ing.Spec.Rules {
		if _, ok := seenHosts[rule.Host]; ok {
			continue
		}
		seenHosts[rule.Host] = struct{}{}

		hostTimeouts := parseGRPCTimeouts(ctx, "server", serverSnippets[rule.Host])
		if differ(server.read, hostTimeouts.read) || differ(server.send, hostTimeouts.send) || differ(server.connect, hostTimeouts.connect) {
			log.Ctx(ctx).Warn().Msg("The server-snippets of the hosts of this Ingress set different gRPC timeouts, using the first host's values")
		}
		server = server.orElse(hostTimeouts)
	}

	return location.orElse(server)
}

// differ reports whether a and b are both set and are not equal.
func differ(a, b *time.Duration) bool {
	return a != nil && b != nil && *a != *b
}

// parseGRPCTimeouts reads the gRPC timeout directives from a snippet.
// The wrapper is the NGINX block the snippet goes in, such as "server" or "location /".
// Only the top-level directives count.
// Directives in nested blocks apply to those blocks only, and the snippet middleware handles them.
// Errors are logged and never returned, so that a bad value does not drop the Ingress.
func parseGRPCTimeouts(ctx context.Context, wrapper, snippet string) (timeouts grpcTimeouts) {
	if snippet == "" {
		return timeouts
	}

	logger := log.Ctx(ctx)

	// The parser can panic on some inputs.
	defer func() {
		if r := recover(); r != nil {
			logger.Debug().Msgf("Cannot parse snippet to read gRPC timeouts: %v", r)
			timeouts = grpcTimeouts{}
		}
	}()

	// Unknown directives are not an error here.
	// The snippet middleware checks which directives are allowed.
	p := parser.NewStringParser(fmt.Sprintf("%s {%s}", wrapper, snippet),
		parser.WithSkipComments(),
		parser.WithSkipValidDirectivesErr(),
	)

	conf, err := p.Parse()
	if err != nil {
		logger.Debug().Err(err).Msg("Cannot parse snippet to read gRPC timeouts")
		return timeouts
	}

	blocks := conf.GetDirectives()
	if len(blocks) == 0 || blocks[0].GetBlock() == nil {
		return timeouts
	}

	for _, d := range blocks[0].GetBlock().GetDirectives() {
		var target **time.Duration
		switch d.GetName() {
		case directiveGRPCConnectTimeout:
			target = &timeouts.connect
		case directiveGRPCReadTimeout:
			target = &timeouts.read
		case directiveGRPCSendTimeout:
			target = &timeouts.send
		default:
			continue
		}

		params := d.GetParameters()
		if len(params) != 1 {
			logger.Error().Msgf("Invalid %s directive: expected one value, got %d, using the default", d.GetName(), len(params))
			continue
		}

		value, err := parseNginxTime(params[0].GetValue())
		if err != nil {
			logger.Error().Err(err).Msgf("Invalid %s value %q, using the default", d.GetName(), params[0].GetValue())
			continue
		}

		*target = &value
	}

	return timeouts
}

// parseNginxTime parses an NGINX time value, as used by the grpc_*_timeout directives.
// See https://nginx.org/en/docs/syntax.html.
//
// It follows ngx_parse_time in NGINX, for directives that take milliseconds:
//   - The units are ms, s, m, h, d and w.
//     NGINX rejects M (months) and y (years) for these directives, so this function does too.
//   - Units can be combined, from the largest to the smallest, each at most once: "1h30m", "1h 30m".
//   - A number without a unit is in seconds: "30" is 30s, and "1m30" is 90s.
func parseNginxTime(value string) (time.Duration, error) {
	units := []struct {
		suffix string
		scale  time.Duration
	}{
		{"w", 7 * 24 * time.Hour},
		{"d", 24 * time.Hour},
		{"h", time.Hour},
		// "ms" must come before "m".
		{"ms", time.Millisecond},
		{"m", time.Minute},
		{"s", time.Second},
	}

	// Rank of each unit, from the largest to the smallest.
	// "ms" is the smallest unit even if it is checked before "m".
	rank := map[string]int{"w": 0, "d": 1, "h": 2, "m": 3, "s": 4, "ms": 5}

	s := strings.TrimSpace(value)
	if s == "" {
		return 0, errors.New("empty value")
	}

	var total time.Duration
	lastRank := -1

	for s != "" {
		i := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == 0 {
			return 0, fmt.Errorf("unexpected character %q", s[0])
		}

		var n int64
		for _, c := range s[:i] {
			if n > (math.MaxInt64-int64(c-'0'))/10 {
				return 0, errors.New("value too large")
			}
			n = n*10 + int64(c-'0')
		}
		s = s[i:]

		// A number without a unit can only be the last part, and is in seconds.
		// Like in NGINX, it is added without checking the order of the units.
		scale := time.Second
		if s != "" {
			found := false
			for _, u := range units {
				if !strings.HasPrefix(s, u.suffix) {
					continue
				}

				if rank[u.suffix] <= lastRank {
					return 0, errors.New("units must go from the largest to the smallest, each at most once")
				}
				lastRank = rank[u.suffix]

				scale = u.scale
				s = s[len(u.suffix):]
				found = true
				break
			}
			if !found {
				return 0, fmt.Errorf("unknown unit %q", s[0])
			}
		}

		if n > int64(math.MaxInt64/scale) {
			return 0, errors.New("value too large")
		}
		part := time.Duration(n) * scale
		if total > math.MaxInt64-part {
			return 0, errors.New("value too large")
		}
		total += part

		s = strings.TrimLeft(s, " ")
	}

	return total, nil
}
