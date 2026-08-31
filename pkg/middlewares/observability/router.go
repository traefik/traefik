package observability

import (
	"context"
	"net/http"
	"regexp"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	routerTypeName = "TracingRouter"
)

type routerTracing struct {
	router     string
	routerRule string
	service    string
	next       http.Handler
}

// WrapRouterHandler Wraps tracing to alice.Constructor.
func WrapRouterHandler(ctx context.Context, router, routerRule, service string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return newRouter(ctx, router, routerRule, service, next), nil
	}
}

// newRouter creates a new tracing middleware that traces the internal requests.
func newRouter(ctx context.Context, router, routerRule, service string, next http.Handler) http.Handler {
	middlewares.GetLogger(ctx, "tracing", routerTypeName).
		Debug().Str(logs.RouterName, router).Str(logs.ServiceName, service).Msg("Added outgoing tracing middleware")

	return &routerTracing{
		router:     router,
		routerRule: routerRule,
		service:    service,
		next:       next,
	}
}

func (f *routerTracing) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Enrich the root server span with http.route and update its name.
	// This runs regardless of DetailedTracingEnabled because the root
	// server span always exists (created by the entrypoint middleware
	// when TracingEnabled is true). The router span (internal) is only
	// created when DetailedTracingEnabled is true.
	if tracer := tracing.TracerFromContext(req.Context()); tracer != nil {
		serverSpan := trace.SpanFromContext(req.Context())

		if DetailedTracingEnabled(req.Context()) {
			tracingCtx, span := tracer.Start(req.Context(), "Router", trace.WithSpanKind(trace.SpanKindInternal))
			defer span.End()

			req = req.WithContext(tracingCtx)

			span.SetAttributes(attribute.String("traefik.service.name", f.service))
			span.SetAttributes(attribute.String("traefik.router.name", f.router))
			span.SetAttributes(semconv.HTTPRoute(f.routerRule))
		}

		// Set http.route on the root server span and update its name to
		// "{method} {route}" per OTel HTTP semantic conventions:
		// https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name
		//
		// The root server span is created by the entrypoint middleware with
		// just the HTTP method as its name (e.g. "GET"). When a router
		// matches, we enrich that span with the http.route attribute and
		// update its name to include the route (e.g. "GET /api/v1/ml-scribe").
		//
		// The route value is extracted from the router rule. For Path and
		// PathPrefix matchers, the path argument is used directly. For more
		// complex rules (PathRegexp, Host, etc.), the raw rule string is
		// used as a fallback.
		route := extractRouteFromRule(f.routerRule)
		if route != "" && serverSpan != nil && serverSpan.IsRecording() {
			serverSpan.SetAttributes(semconv.HTTPRoute(route))
			serverSpan.SetName(req.Method + " " + route)
		}
	}

	f.next.ServeHTTP(rw, req)
}

// pathPattern captures the path argument from Path(`/foo`) or PathPrefix(`/foo`).
var pathPattern = regexp.MustCompile(`(?:Path|PathPrefix)\(\s*\x60([^\x60]+)\x60\s*\)`)

// extractRouteFromRule parses a Traefik router rule and extracts a low-cardinality
// route string suitable for the http.route OTel semantic convention attribute.
//
// For rules using Path or PathPrefix matchers, the path argument is returned
// directly (e.g. PathPrefix(`/api/v1/ml-scribe`) → /api/v1/ml-scribe).
//
// For complex rules that cannot be reduced to a simple path (PathRegexp,
// Host, Header, etc.), the raw rule string is returned as a fallback so
// the http.route attribute is always populated when a router matches.
//
// When multiple Path/PathPrefix matchers are present (OR'd with ||), the
// first one is used. This is consistent with how Traefik evaluates rules:
// the first match wins.
func extractRouteFromRule(rule string) string {
	if rule == "" {
		return ""
	}

	matches := pathPattern.FindStringSubmatch(rule)
	if len(matches) >= 2 {
		return matches[1]
	}

	// Fallback: use the raw rule string. This covers cases like
	// Host(`example.com`) && PathRegexp(`/api/.*`) where no simple
	// path can be extracted.
	return rule
}
