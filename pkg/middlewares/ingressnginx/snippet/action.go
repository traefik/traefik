package snippet

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
	"github.com/tufanbarisyildirim/gonginx/config"
)

// context holds variables set during request processing.
type actionContext struct {
	vars                    map[string]string
	nonMergeablePostActions map[string][]action
	mergeablePostActions    []action

	statusCode     int
	body           string
	redirectURL    string
	accessResolved bool

	stopCurrentBlock  bool
	stopAllDirectives bool
}

func newContext(previousCtx *actionContext, a *actions) *actionContext {
	unmergeablePostActions := a.nonMergeablePostActions
	if a.nonMergeablePostActions == nil {
		unmergeablePostActions = map[string][]action{}
	}

	return &actionContext{
		vars:                    previousCtx.vars,
		statusCode:              previousCtx.statusCode,
		body:                    previousCtx.body,
		redirectURL:             previousCtx.redirectURL,
		nonMergeablePostActions: unmergeablePostActions,
		mergeablePostActions:    a.mergeablePostActions,
	}
}

func (c *actionContext) mergeWithSubContext(subCtx *actionContext) {
	for directive, actions := range subCtx.nonMergeablePostActions {
		if len(actions) > 0 {
			c.nonMergeablePostActions[directive] = actions
		}
	}

	c.mergeablePostActions = append(c.mergeablePostActions, subCtx.mergeablePostActions...)
	c.statusCode = subCtx.statusCode
	c.body = subCtx.body
	c.redirectURL = subCtx.redirectURL

	if subCtx.stopCurrentBlock {
		c.stopCurrentBlock = true
	}
	if subCtx.stopAllDirectives {
		c.stopAllDirectives = true
	}
}

type actions struct {
	actions                 []action
	mergeablePostActions    []action
	nonMergeablePostActions map[string][]action
}

func (a *actions) Execute(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
	// The actions struct can be nil if the middleware does not have a server snippet or a configuration snippet.
	if a == nil {
		return false, nil
	}

	subCtx := newContext(ctx, a)
	finish := false
	for _, act := range a.actions {
		var err error
		finish, err = act(rw, req, subCtx)
		if err != nil {
			return false, err
		}
		if finish || subCtx.stopCurrentBlock || subCtx.stopAllDirectives {
			break
		}
	}

	ctx.mergeWithSubContext(subCtx)

	return finish, nil
}

// action is a function that applies a directive to the request/response.
// It returns true if the request should be interrupted (e.g. return directive).
// The context parameter allows actions to share state (e.g., variables set by 'set' directive).
type action func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error)

func buildActions(block config.IBlock) (*actions, error) {
	acts := &actions{
		nonMergeablePostActions: make(map[string][]action),
	}
	for _, d := range block.GetDirectives() {
		if err := isAllowedInContext(d); err != nil {
			return nil, err
		}

		switch d.GetName() {
		case "add_header":
			action, err := createAddHeaderAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating add_header action: %w", err)
			}
			acts.nonMergeablePostActions[d.GetName()] = append(acts.nonMergeablePostActions[d.GetName()], action)
		case "more_set_headers":
			action, err := createMoreSetHeadersAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating more_set_headers action: %w", err)
			}
			acts.mergeablePostActions = append(acts.mergeablePostActions, action)
		case "proxy_set_header":
			action, err := createProxySetHeaderAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating proxy_set_header action: %w", err)
			}
			acts.nonMergeablePostActions[d.GetName()] = append(acts.nonMergeablePostActions[d.GetName()], action)
		case "more_set_input_headers":
			action, err := createMoreSetInputHeadersAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating more_set_input_headers action: %w", err)
			}
			acts.mergeablePostActions = append(acts.mergeablePostActions, action)
		case "more_clear_headers":
			action, err := createMoreClearHeadersAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating more_clear_headers action: %w", err)
			}
			acts.mergeablePostActions = append(acts.mergeablePostActions, action)
		case "more_clear_input_headers":
			action, err := createMoreClearInputHeadersAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating more_clear_input_headers action: %w", err)
			}
			acts.mergeablePostActions = append(acts.mergeablePostActions, action)
		case "if":
			action, err := createIfAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating if action: %w", err)
			}
			acts.actions = append(acts.actions, action)
		case "set":
			action, err := createSetAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating set action: %w", err)
			}
			acts.actions = append(acts.actions, action)
		case "return":
			action, err := createReturnAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating return action: %w", err)
			}
			acts.actions = append(acts.actions, action)
		case "rewrite":
			action, err := createRewriteAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating rewrite action: %w", err)
			}
			acts.actions = append(acts.actions, action)
		case "location":
			action, err := createLocationAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating location action: %w", err)
			}
			acts.actions = append(acts.actions, action)
		case "allow", "deny":
			action, err := createAccessAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating %s action: %w", d.GetName(), err)
			}
			acts.actions = append(acts.actions, action)
		case "proxy_hide_header":
			action, err := createProxyHideHeaderAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating proxy_hide_header action: %w", err)
			}
			acts.mergeablePostActions = append(acts.mergeablePostActions, action)
		case "expires":
			action, err := createExpiresAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating expires action: %w", err)
			}
			acts.mergeablePostActions = append(acts.mergeablePostActions, action)
		default:
			return nil, fmt.Errorf("unsupported directive %q", d.GetName())
		}
	}

	return acts, nil
}

// addHeaderStatusCodes lists the status codes for which add_header is effective
// when the "always" parameter is NOT specified.
var addHeaderStatusCodes = []int{
	200, 201, 204, 206,
	301, 302, 303, 304,
	307, 308,
}

func createAddHeaderAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) < 2 {
		return nil, errors.New("add_header directive must have at least 2 parameters (header and value)")
	}

	key := params[0].String()
	val := trimQuote(params[1].String())

	// Check for the "always" flag (third parameter).
	var always bool
	if len(params) >= 3 && params[2].String() == "always" {
		always = true
	}

	if always {
		// With "always", the header is added regardless of response status code.
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			rw.Header().Add(key, ingressnginx.ReplaceVariables(val, req, ctx.vars))
			return false, nil
		}, nil
	}

	// Without "always", register a deferred operation that checks the status code.
	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		if wrapper, ok := rw.(*snippetResponseWriter); ok {
			resolvedVal := ingressnginx.ReplaceVariables(val, req, ctx.vars)
			wrapper.onWriteHeader = append(wrapper.onWriteHeader, func(code int, h http.Header) {
				if slices.Contains(addHeaderStatusCodes, code) {
					h.Add(key, resolvedVal)
				}
			})
		} else {
			// Fallback: add unconditionally.
			rw.Header().Add(key, ingressnginx.ReplaceVariables(val, req, ctx.vars))
		}
		return false, nil
	}, nil
}

// directiveFlags holds parsed flags for headers-more directives (-s, -t, -a, -r).
type directiveFlags struct {
	statusCodes  []int
	contentTypes []string
	appendMode   bool
	restrictOnly bool
}

// parseDirectiveFlags parses -s, -t, -a, -r flags from directive parameters.
// It returns the flags and the remaining (non-flag) parameters.
// The allowStatusFilter parameter controls whether the -s flag is accepted;
// in the real headers-more module, -s is only valid on output header directives.
func parseDirectiveFlags(params []config.Parameter, allowStatusFilter bool) (directiveFlags, []config.Parameter, error) {
	var flags directiveFlags
	var remaining []config.Parameter

	i := 0
	for i < len(params) {
		p := params[i].String()
		switch p {
		case "-s":
			if !allowStatusFilter {
				return flags, nil, errors.New("flag -s is not supported for this directive")
			}
			i++
			if i < len(params) {
				for s := range strings.FieldsSeq(trimQuote(params[i].String())) {
					code, err := strconv.Atoi(s)
					if err == nil && code > 0 {
						flags.statusCodes = append(flags.statusCodes, code)
					}
				}
			}
		case "-t":
			i++
			if i < len(params) {
				flags.contentTypes = append(flags.contentTypes, strings.Fields(trimQuote(params[i].String()))...)
			}
		case "-a":
			flags.appendMode = true
		case "-r":
			flags.restrictOnly = true
		default:
			remaining = append(remaining, params[i])
		}
		i++
	}

	return flags, remaining, nil
}

// matchesStatusFilter returns true if the code matches the filter (or filter is empty).
func matchesStatusFilter(filter []int, code int) bool {
	if len(filter) == 0 {
		return true
	}
	return slices.Contains(filter, code)
}

// matchesContentTypeFilter returns true if the contentType matches the filter (or filter is empty).
func matchesContentTypeFilter(filter []string, contentType string) bool {
	if len(filter) == 0 {
		return true
	}
	// Extract base MIME type (without parameters like charset).
	base, _, _ := strings.Cut(contentType, ";")
	base = strings.TrimSpace(base)
	for _, f := range filter {
		if strings.EqualFold(base, f) {
			return true
		}
	}
	return false
}

func createMoreSetHeadersAction(d config.IDirective) (action, error) {
	flags, remaining, err := parseDirectiveFlags(d.GetParameters(), true)
	if err != nil {
		return nil, fmt.Errorf("parsing more_set_headers flags: %w", err)
	}
	ops, err := parseMoreSetParams(remaining, "more_set_headers")
	if err != nil {
		return nil, fmt.Errorf("parsing more_set_headers directive: %w", err)
	}

	hasFilters := len(flags.statusCodes) > 0 || len(flags.contentTypes) > 0

	if hasFilters {
		// Deferred: register a hook on the response writer to execute when
		// the status code is known.
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			if wrapper, ok := rw.(*snippetResponseWriter); ok {
				wrapper.onWriteHeader = append(wrapper.onWriteHeader, func(code int, h http.Header) {
					ct := h.Get("Content-Type")
					if !matchesStatusFilter(flags.statusCodes, code) || !matchesContentTypeFilter(flags.contentTypes, ct) {
						return
					}
					applyHeaderOps(h, nil, ops, flags, req, ctx)
				})
			}
			return false, nil
		}, nil
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		applyHeaderOps(rw.Header(), nil, ops, flags, req, ctx)
		return false, nil
	}, nil
}

func createMoreSetInputHeadersAction(d config.IDirective) (action, error) {
	flags, remaining, err := parseDirectiveFlags(d.GetParameters(), false)
	if err != nil {
		return nil, fmt.Errorf("parsing more_set_input_headers flags: %w", err)
	}
	ops, err := parseMoreSetParams(remaining, "more_set_input_headers")
	if err != nil {
		return nil, fmt.Errorf("parsing more_set_input_headers directive: %w", err)
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		// -t filter on request-side: check request Content-Type.
		if len(flags.contentTypes) > 0 && !matchesContentTypeFilter(flags.contentTypes, req.Header.Get("Content-Type")) {
			return false, nil
		}
		applyHeaderOps(nil, req, ops, flags, req, ctx)
		return false, nil
	}, nil
}

// applyHeaderOps applies header operations to either response headers (h) or request headers (req).
// If h is non-nil, operations apply to response headers. Otherwise, operations apply to req.Header.
func applyHeaderOps(h http.Header, r *http.Request, ops []headerOp, flags directiveFlags, req *http.Request, ctx *actionContext) {
	target := h
	if target == nil && r != nil {
		target = r.Header
	}
	if target == nil {
		return
	}

	for _, op := range ops {
		if op.clear {
			target.Del(op.key)
			continue
		}

		// -r flag: only set if the header already exists.
		if flags.restrictOnly && target.Get(op.key) == "" {
			continue
		}

		resolvedVal := ingressnginx.ReplaceVariables(op.value, req, ctx.vars)
		if flags.appendMode {
			target.Add(op.key, resolvedVal)
		} else {
			target.Set(op.key, resolvedVal)
		}
	}
}

// headerClearOp represents a header to clear, with optional wildcard matching.
type headerClearOp struct {
	name     string // exact header name or prefix (without trailing *)
	wildcard bool   // true if the original name ended with *
}

func createMoreClearHeadersAction(d config.IDirective) (action, error) {
	flags, remaining, err := parseDirectiveFlags(d.GetParameters(), true)
	if err != nil {
		return nil, fmt.Errorf("parsing more_clear_headers flags: %w", err)
	}
	ops, err := parseMoreClearParams(remaining, "more_clear_headers")
	if err != nil {
		return nil, fmt.Errorf("parsing more_clear_headers directive: %w", err)
	}

	hasFilters := len(flags.statusCodes) > 0 || len(flags.contentTypes) > 0

	if hasFilters {
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			if wrapper, ok := rw.(*snippetResponseWriter); ok {
				wrapper.onWriteHeader = append(wrapper.onWriteHeader, func(code int, h http.Header) {
					ct := h.Get("Content-Type")
					if !matchesStatusFilter(flags.statusCodes, code) || !matchesContentTypeFilter(flags.contentTypes, ct) {
						return
					}
					clearHeaders(h, ops)
				})
			}
			return false, nil
		}, nil
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		clearHeaders(rw.Header(), ops)
		return false, nil
	}, nil
}

func createMoreClearInputHeadersAction(d config.IDirective) (action, error) {
	flags, remaining, err := parseDirectiveFlags(d.GetParameters(), false)
	if err != nil {
		return nil, fmt.Errorf("parsing more_clear_input_headers flags: %w", err)
	}
	ops, err := parseMoreClearParams(remaining, "more_clear_input_headers")
	if err != nil {
		return nil, fmt.Errorf("parsing more_clear_input_headers directive: %w", err)
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		if len(flags.contentTypes) > 0 && !matchesContentTypeFilter(flags.contentTypes, req.Header.Get("Content-Type")) {
			return false, nil
		}
		clearHeaders(req.Header, ops)
		return false, nil
	}, nil
}

// parseMoreClearParams parses header names from the remaining (non-flag) parameters.
func parseMoreClearParams(params []config.Parameter, directiveName string) ([]headerClearOp, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("%s directive must have at least 1 header name", directiveName)
	}

	ops := make([]headerClearOp, 0, len(params))
	for _, p := range params {
		name := trimQuote(p.String())
		if name == "" {
			return nil, fmt.Errorf("%s directive has an empty header name", directiveName)
		}

		if prefix, ok := strings.CutSuffix(name, "*"); ok {
			if prefix == "" {
				return nil, fmt.Errorf("%s directive has an invalid wildcard pattern %q", directiveName, name)
			}
			ops = append(ops, headerClearOp{name: prefix, wildcard: true})
		} else {
			ops = append(ops, headerClearOp{name: name})
		}
	}

	return ops, nil
}

// clearHeaders removes headers from the given http.Header map based on the
// provided clear operations. For exact matches it uses Del; for wildcard
// matches it iterates all headers and removes those whose name starts with
// the given prefix (case-insensitive).
func clearHeaders(h http.Header, ops []headerClearOp) {
	for _, op := range ops {
		if op.wildcard {
			lowerPrefix := strings.ToLower(op.name)
			for name := range h {
				if strings.HasPrefix(strings.ToLower(name), lowerPrefix) {
					h.Del(name)
				}
			}
		} else {
			h.Del(op.name)
		}
	}
}

// headerOp represents a single header operation: either setting a header to a
// value or clearing (deleting) a header.
type headerOp struct {
	key   string
	value string
	clear bool
}

// parseMoreSetParams parses one or more quoted "Key: Value" parameters
// from the remaining (non-flag) parameters. It supports:
//   - Multiple parameters: more_set_headers "H1: v1" "H2: v2";
//   - Header clearing: "Key:", "Key: ", or "Key" (no colon) all result in a
//     clear operation that deletes the header.
func parseMoreSetParams(params []config.Parameter, directiveName string) ([]headerOp, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("%s directive must have at least 1 parameter", directiveName)
	}

	ops := make([]headerOp, 0, len(params))
	for _, p := range params {
		trimmed := trimQuote(p.String())
		if trimmed == "" {
			return nil, fmt.Errorf("%s directive has an empty parameter", directiveName)
		}

		parts := strings.SplitN(trimmed, ":", 2)
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("%s directive has an empty header name", directiveName)
		}

		if len(parts) == 1 {
			// No colon found — this is a clear operation.
			ops = append(ops, headerOp{key: key, clear: true})
			continue
		}

		value := strings.TrimSpace(parts[1])
		if value == "" {
			// Colon present but value is empty — clear operation.
			ops = append(ops, headerOp{key: key, clear: true})
			continue
		}

		ops = append(ops, headerOp{key: key, value: value})
	}

	return ops, nil
}

func createProxySetHeaderAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) < 2 {
		return nil, errors.New("proxy_set_header directive requires 2 parameters (header and value)")
	}

	key := params[0].String()
	val := trimQuote(params[1].String())

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		resolved := ingressnginx.ReplaceVariables(val, req, ctx.vars)
		if resolved == "" {
			req.Header.Del(key)
		} else {
			req.Header.Set(key, resolved)
		}
		return false, nil
	}, nil
}

// createProxyHideHeaderAction hides the specified header from the upstream response.
// It registers a deferred hook on the response writer to delete the header when
// the response status is being written.
func createProxyHideHeaderAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("proxy_hide_header directive requires 1 parameter (header name)")
	}

	headerName := trimQuote(params[0].String())

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		if wrapper, ok := rw.(*snippetResponseWriter); ok {
			wrapper.onWriteHeader = append(wrapper.onWriteHeader, func(_ int, h http.Header) {
				h.Del(headerName)
			})
		}
		return false, nil
	}, nil
}

// createAccessAction creates an action for the allow or deny directive.
// Rules are evaluated in order until the first match is found (first match wins).
// If a deny rule matches, it returns 403. If an allow rule matches, processing continues.
func createAccessAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, fmt.Errorf("%s directive requires 1 parameter", d.GetName())
	}

	isAllow := d.GetName() == "allow"
	addr := trimQuote(params[0].String())

	if addr == "all" {
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			if ctx.accessResolved {
				return false, nil
			}
			ctx.accessResolved = true
			if !isAllow {
				ctx.statusCode = http.StatusForbidden
				ctx.body = "403 Forbidden"
				return true, nil
			}
			return false, nil
		}, nil
	}

	// Parse IP or CIDR.
	var ipNet *net.IPNet
	if strings.Contains(addr, "/") {
		_, parsed, err := net.ParseCIDR(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR in %s directive: %w", d.GetName(), err)
		}
		ipNet = parsed
	} else {
		ip := net.ParseIP(addr)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address in %s directive: %q", d.GetName(), addr)
		}
		// Single IP: create a /32 or /128 mask.
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		ipNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)}
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		if ctx.accessResolved {
			return false, nil
		}

		remoteIP := extractIP(req.RemoteAddr)
		if remoteIP == nil {
			return false, nil
		}

		if !ipNet.Contains(remoteIP) {
			return false, nil
		}

		// First matching rule wins.
		ctx.accessResolved = true
		if !isAllow {
			ctx.statusCode = http.StatusForbidden
			ctx.body = "403 Forbidden"
			return true, nil
		}
		return false, nil
	}, nil
}

// extractIP parses a remote address (potentially with port) and returns the IP.
func extractIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// Try parsing as bare IP.
		return net.ParseIP(remoteAddr)
	}
	return net.ParseIP(host)
}

// createExpiresAction implements the NGINX expires directive.
// Syntax: expires time | epoch | max | off;
// Sets "Expires" and "Cache-Control" response headers.
func createExpiresAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("expires directive requires 1 parameter")
	}

	value := trimQuote(params[0].String())

	switch value {
	case "off":
		// No-op: don't set any cache headers.
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			return false, nil
		}, nil

	case "epoch":
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			rw.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:01 GMT")
			rw.Header().Set("Cache-Control", "no-cache")
			return false, nil
		}, nil

	case "max":
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			rw.Header().Set("Expires", "Thu, 31 Dec 2037 23:55:55 GMT")
			rw.Header().Set("Cache-Control", "max-age=315360000")
			return false, nil
		}, nil
	}

	// Parse as duration. NGINX uses a custom format but common values are like "24h", "30d", "1y", "-1".
	dur, err := parseNginxDuration(value)
	if err != nil {
		return nil, fmt.Errorf("invalid expires value %q: %w", value, err)
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		if dur < 0 {
			rw.Header().Set("Expires", time.Now().Add(dur).UTC().Format(http.TimeFormat))
			rw.Header().Set("Cache-Control", "no-cache")
		} else {
			rw.Header().Set("Expires", time.Now().Add(dur).UTC().Format(http.TimeFormat))
			rw.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(dur.Seconds())))
		}
		return false, nil
	}, nil
}

// parseNginxDuration parses NGINX-style duration strings: "30s", "5m", "24h", "7d", "1y", "-1".
func parseNginxDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, errors.New("empty duration")
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	// Try standard Go duration first (handles "5s", "10m", "24h" etc.).
	dur, err := time.ParseDuration(s)
	if err == nil {
		if negative {
			dur = -dur
		}
		return dur, nil
	}

	// Handle NGINX-specific suffixes: "d" (days), "M" (months≈30d), "y" (years≈365d).
	if len(s) > 1 {
		numStr := s[:len(s)-1]
		suffix := s[len(s)-1]
		num, numErr := strconv.Atoi(numStr)
		if numErr == nil {
			var d time.Duration
			switch suffix {
			case 'd':
				d = time.Duration(num) * 24 * time.Hour
			case 'M':
				d = time.Duration(num) * 30 * 24 * time.Hour
			case 'y':
				d = time.Duration(num) * 365 * 24 * time.Hour
			default:
				return 0, fmt.Errorf("unsupported duration suffix %q", string(suffix))
			}
			if negative {
				d = -d
			}
			return d, nil
		}
	}

	// Try bare number (seconds in NGINX).
	num, numErr := strconv.Atoi(s)
	if numErr == nil {
		d := time.Duration(num) * time.Second
		if negative {
			d = -d
		}
		return d, nil
	}

	return 0, fmt.Errorf("cannot parse duration %q", s)
}

func createIfAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("if directive requires a condition")
	}

	// Parse condition - simplified implementation.
	// Supports: ($var = value), ($var != value), ($var), ($request_method = METHOD).
	var paramStrs []string
	for _, p := range params {
		paramStrs = append(paramStrs, p.String())
	}
	condition := strings.Join(paramStrs, " ")
	condition = strings.Trim(condition, "()")

	// Build actions from the if block.
	block := d.GetBlock()
	if block == nil {
		return nil, errors.New("if directive requires a block")
	}

	blockActions, err := buildActions(block)
	if err != nil {
		return nil, fmt.Errorf("building if block actions: %w", err)
	}

	eval, err := buildCondition(condition)
	if err != nil {
		return nil, fmt.Errorf("building if condition: %w", err)
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		if eval(req, ctx) {
			return blockActions.Execute(rw, req, ctx)
		}
		return false, nil
	}, nil
}

// isRedirectCode returns true if the given HTTP status code is a redirect code
// that NGINX treats as requiring a Location header (301, 302, 303, 307, 308).
func isRedirectCode(code int) bool {
	switch code {
	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	}
	return false
}

// parseIntSimple returns the parsed integer and true, or zero and false.
func parseIntSimple(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

func createReturnURLAction(u string) action {
	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		resolvedURL := ingressnginx.ReplaceVariables(u, req, ctx.vars)
		ctx.statusCode = http.StatusFound
		ctx.redirectURL = resolvedURL
		return true, nil
	}
}

// createReturnAction implements the NGINX return directive.
// Syntax: return code [text]; | return code URL; | return URL;
// For redirect codes (301, 302, 303, 307, 308), the second parameter is treated as a URL
// and a Location header is set. For other codes, it is treated as response body text.
func createReturnAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("return directive requires parameters")
	}

	first := params[0].String()

	// If the first parameter is not a number, it's a URL (return URL; syntax = 302 redirect).
	code, isNumeric := parseIntSimple(first)
	if !isNumeric {
		return createReturnURLAction(trimQuote(first)), nil
	}

	var text string
	if len(params) > 1 {
		text = trimQuote(params[1].String())
	}

	if isRedirectCode(code) {
		return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
			ctx.statusCode = code
			if text != "" {
				ctx.redirectURL = ingressnginx.ReplaceVariables(text, req, ctx.vars)
			}
			return true, nil
		}, nil
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		ctx.statusCode = code
		if text != "" {
			ctx.body = ingressnginx.ReplaceVariables(text, req, ctx.vars)
		}
		return true, nil
	}, nil
}

// captureGroupRegexp matches $1 through $9 capture group references in replacement strings.
var captureGroupRegexp = regexp.MustCompile(`\$([1-9])`)

// replaceCaptureGroups replaces $1-$9 references in the replacement string with
// the corresponding capture group values from the regex match.
func replaceCaptureGroups(replacement string, matches []string) string {
	return captureGroupRegexp.ReplaceAllStringFunc(replacement, func(ref string) string {
		idx := int(ref[1] - '0')
		if idx < len(matches) {
			return matches[idx]
		}
		return ref
	})
}

// createRewriteAction implements the NGINX rewrite directive.
// Syntax: rewrite regex replacement [flag];
// Flags: last, break, redirect, permanent.
// If the replacement starts with http://, https://, or $scheme, a redirect is returned.
// If the replacement ends with ?, the original query string is not appended.
func createRewriteAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) < 2 {
		return nil, errors.New("rewrite directive requires at least 2 parameters (regex and replacement)")
	}

	pattern := params[0].String()
	replacement := params[1].String()

	var flag string
	if len(params) >= 3 {
		flag = params[2].String()
	}

	// Validate the flag.
	switch flag {
	case "", "last", "break", "redirect", "permanent":
	default:
		return nil, fmt.Errorf("rewrite directive has invalid flag %q", flag)
	}

	// Compile the regex at build time.
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compiling rewrite regex: %w", err)
	}

	// Determine if the ? suffix is present to suppress query string appending.
	suppressQueryString := strings.HasSuffix(replacement, "?")
	if suppressQueryString {
		replacement = replacement[:len(replacement)-1]
	}

	// Determine if this is a redirect based on the replacement or flag.
	isURLRedirect := strings.HasPrefix(replacement, "http://") ||
		strings.HasPrefix(replacement, "https://") ||
		strings.HasPrefix(replacement, "$scheme")

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		matches := re.FindStringSubmatch(req.URL.Path)
		if matches == nil {
			return false, nil
		}

		// Build the replacement string: first replace capture groups, then NGINX variables.
		result := replaceCaptureGroups(replacement, matches)
		result = ingressnginx.ReplaceVariables(result, req, ctx.vars)

		// Determine redirect behavior.
		switch {
		case flag == "redirect":
			ctx.statusCode = http.StatusFound
			ctx.redirectURL = result
			return true, nil

		case flag == "permanent":
			ctx.statusCode = http.StatusMovedPermanently
			ctx.redirectURL = result
			return true, nil

		case isURLRedirect:
			// Replacement starts with http://, https://, or $scheme: default to 302.
			ctx.statusCode = http.StatusFound
			ctx.redirectURL = result
			return true, nil
		}

		// Not a redirect: rewrite the request URI in place.
		if path, query, ok := strings.Cut(result, "?"); ok {
			// Replacement contains a query string.
			req.URL.Path = path
			req.URL.RawQuery = query
		} else {
			req.URL.Path = result
			if suppressQueryString {
				req.URL.RawQuery = ""
			}
			// Otherwise, keep the original query string.
		}

		req.RequestURI = req.URL.RequestURI()

		// In NGINX, last restarts location matching while break stays in the
		// current location. In Traefik's middleware model, last stops
		// processing the current block (allowing subsequent blocks to run),
		// while break stops all remaining directive processing.
		// Both forward the rewritten request to the upstream.
		if flag == "last" {
			ctx.stopCurrentBlock = true
			return false, nil
		}
		if flag == "break" {
			ctx.stopAllDirectives = true
			return false, nil
		}

		return false, nil
	}, nil
}

func createSetAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) < 2 {
		return nil, errors.New("set directive requires 2 parameters (variable and value)")
	}

	varName := params[0].String()
	value := trimQuote(params[1].String())

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		ctx.vars[varName] = ingressnginx.ReplaceVariables(value, req, ctx.vars)
		return false, nil
	}, nil
}

func createLocationAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("location directive requires a path pattern")
	}

	pathPattern := params[0].String()
	if len(params) == 2 {
		pathPattern += params[1].String()
	}

	// Build actions from the location block.
	block := d.GetBlock()
	if block == nil {
		return nil, errors.New("location directive requires a block")
	}

	blockActions, err := buildActions(block)
	if err != nil {
		return nil, fmt.Errorf("building location block actions: %w", err)
	}

	// Determine location type and compile regex if needed.
	// Check ~* (case-insensitive regex) before ~ (case-sensitive regex).
	var re *regexp.Regexp
	var isExact bool
	if pattern, ok := strings.CutPrefix(pathPattern, "~*"); ok {
		pattern = strings.TrimSpace(pattern)
		re, err = regexp.Compile("(?i)" + pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling location regex: %w", err)
		}
	} else if pattern, ok := strings.CutPrefix(pathPattern, "~"); ok {
		pattern = strings.TrimSpace(pattern)
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling location regex: %w", err)
		}
	} else if exact, ok := strings.CutPrefix(pathPattern, "="); ok {
		pathPattern = strings.TrimSpace(exact)
		isExact = true
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		var matches bool
		switch {
		case re != nil:
			matches = re.MatchString(req.URL.Path)
		case isExact:
			matches = req.URL.Path == pathPattern
		default:
			matches = strings.HasPrefix(req.URL.Path, pathPattern)
		}

		if matches {
			stop, err := blockActions.Execute(rw, req, ctx)
			if err != nil {
				return false, fmt.Errorf("executing location block: %w", err)
			}

			return stop, nil
		}
		return false, nil
	}, nil
}

// conditionEvaluator is a function that evaluates a parsed if-condition at request time.
type conditionEvaluator func(req *http.Request, ctx *actionContext) bool

// buildCondition pre-parses an if-condition string and returns a conditionEvaluator.
// It pre-compiles regexes at build time instead of on every request.
// Supports: ($var), ($var = value), ($var != value), ($var ~ regex), ($var !~ regex),
// ($var ~* regex), ($var !~* regex).
func buildCondition(condition string) (conditionEvaluator, error) {
	parts := strings.Fields(condition)
	if len(parts) == 0 {
		return nil, errors.New("empty condition")
	}

	// Simple variable check: if ($var).
	// Uses ReplaceVariables so both custom and built-in variables work.
	if len(parts) == 1 {
		varExpr := parts[0]
		return func(req *http.Request, ctx *actionContext) bool {
			val := ingressnginx.ReplaceVariables(varExpr, req, ctx.vars)
			// If the variable was not resolved, treat as undefined.
			if val == varExpr {
				return false
			}
			return val != "" && val != "0"
		}, nil
	}

	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid condition format: %s", condition)
	}

	varExpr := strings.Trim(parts[0], `"`)
	operator := parts[1]
	expectedExpr := strings.Trim(parts[2], `"`)

	switch operator {
	case "=":
		return func(req *http.Request, ctx *actionContext) bool {
			varVal := ingressnginx.ReplaceVariables(varExpr, req, ctx.vars)
			expected := ingressnginx.ReplaceVariables(expectedExpr, req, ctx.vars)
			return varVal == expected
		}, nil

	case "!=":
		return func(req *http.Request, ctx *actionContext) bool {
			varVal := ingressnginx.ReplaceVariables(varExpr, req, ctx.vars)
			expected := ingressnginx.ReplaceVariables(expectedExpr, req, ctx.vars)
			return varVal != expected
		}, nil

	case "~":
		re, err := regexp.Compile(expectedExpr)
		if err != nil {
			return nil, fmt.Errorf("compiling regex in condition: %w", err)
		}
		return func(req *http.Request, ctx *actionContext) bool {
			varVal := ingressnginx.ReplaceVariables(varExpr, req, ctx.vars)
			matches := re.FindStringSubmatch(varVal)
			if matches == nil {
				return false
			}
			storeCaptureGroups(ctx, matches)
			return true
		}, nil

	case "!~":
		re, err := regexp.Compile(expectedExpr)
		if err != nil {
			return nil, fmt.Errorf("compiling regex in condition: %w", err)
		}
		return func(req *http.Request, ctx *actionContext) bool {
			varVal := ingressnginx.ReplaceVariables(varExpr, req, ctx.vars)
			return !re.MatchString(varVal)
		}, nil

	case "~*":
		re, err := regexp.Compile("(?i)" + expectedExpr)
		if err != nil {
			return nil, fmt.Errorf("compiling case-insensitive regex in condition: %w", err)
		}
		return func(req *http.Request, ctx *actionContext) bool {
			varVal := ingressnginx.ReplaceVariables(varExpr, req, ctx.vars)
			matches := re.FindStringSubmatch(varVal)
			if matches == nil {
				return false
			}
			storeCaptureGroups(ctx, matches)
			return true
		}, nil

	case "!~*":
		re, err := regexp.Compile("(?i)" + expectedExpr)
		if err != nil {
			return nil, fmt.Errorf("compiling case-insensitive regex in condition: %w", err)
		}
		return func(req *http.Request, ctx *actionContext) bool {
			varVal := ingressnginx.ReplaceVariables(varExpr, req, ctx.vars)
			return !re.MatchString(varVal)
		}, nil

	default:
		return nil, fmt.Errorf("unsupported operator in condition: %s", operator)
	}
}

// storeCaptureGroups stores regex capture groups ($1-$9) in the action context.
func storeCaptureGroups(ctx *actionContext, matches []string) {
	for i := 1; i < len(matches); i++ {
		ctx.vars[fmt.Sprintf("$%d", i)] = matches[i]
	}
}

func trimQuote(val string) string {
	if len(val) > 1 {
		if val[0] == '"' && val[len(val)-1] == '"' {
			return val[1 : len(val)-1]
		}
	}
	return val
}
