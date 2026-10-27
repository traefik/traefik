package snippet

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/tufanbarisyildirim/gonginx/config"
)

// PhaseCollector holds actions collected for each phase during directive evaluation.
// Directives inside conditional blocks (if) are only added when the condition matches.
type PhaseCollector struct {
	// InputFilter contains request header manipulation actions.
	InputFilter []action

	// Access contains access control actions (allow/deny).
	Access []action

	// Content contains upstream request configuration actions (proxy_set_header, proxy_method).
	Content []action

	// HeaderFilterOverride contains override header directives (add_header).
	// Uses stack-based override: when a nested block (location/if) has add_header,
	// it completely replaces parent block's add_header directives.
	HeaderFilterOverride []addHeaderEntry

	// HeaderFilterAdditive contains additive header directives (more_*, proxy_hide_header, expires).
	// Both server and location apply.
	HeaderFilterAdditive []action

	// blockDepth tracks the current nested block depth for override behavior.
	blockDepth int

	// overrideDepth tracks at which depth add_header was last added.
	// When entering a deeper block that has add_header, we clear previous and update depth.
	overrideDepth int
}

// CollectableAction is an action that processes a directive during the collection phase.
// It may either:
// - Execute immediately (set, rewrite, return) and potentially terminate
// - Collect actions into phase buckets for later execution (add_header, proxy_set_header, etc.)
// - Recursively process nested directives for conditional blocks (if, location)
//
// Returns (terminated, error) where terminated=true means request processing should stop.
type CollectableAction func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (terminated bool, err error)

// SnippetActions holds the parsed collectable actions from a snippet.
type SnippetActions struct {
	actions []CollectableAction
}

// BuildSnippetActions parses directives from a block and builds collectable actions.
func BuildSnippetActions(block config.IBlock) (*SnippetActions, error) {
	sa := &SnippetActions{}

	for _, d := range block.GetDirectives() {
		if err := isAllowedInContext(d); err != nil {
			return nil, err
		}

		action, err := buildCollectableAction(d)
		if err != nil {
			return nil, fmt.Errorf("building %s action: %w", d.GetName(), err)
		}

		sa.actions = append(sa.actions, action)
	}

	return sa, nil
}

// Collect evaluates all directives with the given request and collects phase actions.
// Rewrite-phase directives (set, rewrite, return) execute immediately.
// Other directives are collected into the appropriate phase bucket.
// After a `return` directive terminates, we continue collecting header-phase
// directives (they're in a different NGINX phase and should still apply).
// Returns (terminated, skipToAccess, error).
func (sa *SnippetActions) Collect(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (terminated bool, skipToAccess bool, err error) {
	if sa == nil {
		return false, false, nil
	}

	for _, act := range sa.actions {
		termResult, err := act(rw, req, ctx, pc)
		if err != nil {
			return false, false, err
		}
		if termResult {
			terminated = true
			// Continue to collect header-phase directives
			continue
		}
		// Skip rewrite-phase directives after termination
		// (collection directives always return false, so they still run)
		if terminated {
			continue
		}
		if ctx.stopAllDirectives {
			// rewrite...break: stop processing, skip to access phase
			return terminated, true, nil
		}
		if ctx.stopCurrentBlock {
			// rewrite...last: stop current block
			return terminated, false, nil
		}
	}

	return terminated, false, nil
}

// HasAccess returns true if there are any access control actions collected.
func (pc *PhaseCollector) HasAccess() bool {
	return pc != nil && len(pc.Access) > 0
}

// HasContent returns true if there are any content phase actions collected.
func (pc *PhaseCollector) HasContent() bool {
	return pc != nil && len(pc.Content) > 0
}

// HasHeaderOverride returns true if there are any override header actions collected.
func (pc *PhaseCollector) HasHeaderOverride() bool {
	return pc != nil && len(pc.HeaderFilterOverride) > 0
}

// addHeaderEntry represents an add_header directive with its "always" flag.
type addHeaderEntry struct {
	key    string
	value  string
	always bool
}

// buildCollectableAction creates a CollectableAction for a directive.
func buildCollectableAction(d config.IDirective) (CollectableAction, error) {
	switch d.GetName() {
	// Rewrite phase - execute immediately
	case "set":
		return createSetCollectable(d)
	case "return":
		return createReturnCollectable(d)
	case "rewrite":
		return createRewriteCollectable(d)
	case "if":
		return createIfCollectable(d)
	case "location":
		return createLocationCollectable(d)

	// Access phase - collect
	case "allow", "deny":
		return createAccessCollectable(d)

	// Input filter phase - collect
	case "more_set_input_headers":
		return createMoreSetInputHeadersCollectable(d)
	case "more_clear_input_headers":
		return createMoreClearInputHeadersCollectable(d)

	// Content phase - collect
	case "auth_request_set":
		return createAuthRequestSetCollectable(d)
	case "proxy_set_header":
		return createProxySetHeaderCollectable(d)
	case "proxy_method":
		return createProxyMethodCollectable(d)

	// Header filter phase - collect
	case "add_header":
		return createAddHeaderCollectable(d)
	case "more_set_headers":
		return createMoreSetHeadersCollectable(d)
	case "more_clear_headers":
		return createMoreClearHeadersCollectable(d)
	case "proxy_hide_header":
		return createProxyHideHeaderCollectable(d)
	case "expires":
		return createExpiresCollectable(d)

	default:
		return nil, fmt.Errorf("unsupported directive %q", d.GetName())
	}
}

// === Rewrite phase directives (execute immediately) ===

func createSetCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createSetAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		return act(rw, req, ctx)
	}, nil
}

func createReturnCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createReturnAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		return act(rw, req, ctx)
	}, nil
}

func createRewriteCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createRewriteAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		return act(rw, req, ctx)
	}, nil
}

func createIfCollectable(d config.IDirective) (CollectableAction, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("if directive requires a condition")
	}

	var paramStrs []string
	for _, p := range params {
		paramStrs = append(paramStrs, p.String())
	}
	condition := strings.Join(paramStrs, " ")
	condition = strings.Trim(condition, "()")

	block := d.GetBlock()
	if block == nil {
		return nil, errors.New("if directive requires a block")
	}

	// Build nested collectable actions
	nestedActions, err := BuildSnippetActions(block)
	if err != nil {
		return nil, fmt.Errorf("building if block actions: %w", err)
	}

	eval, err := buildCondition(condition)
	if err != nil {
		return nil, fmt.Errorf("building if condition: %w", err)
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		if eval(req, ctx) {
			// Increase block depth for override behavior
			pc.blockDepth++
			// Recursively collect nested actions
			terminated, _, err := nestedActions.Collect(rw, req, ctx, pc)
			pc.blockDepth--
			return terminated, err
		}
		return false, nil
	}, nil
}

func createLocationCollectable(d config.IDirective) (CollectableAction, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("location directive requires a path pattern")
	}

	pathPattern := params[0].String()
	if len(params) == 2 {
		pathPattern += params[1].String()
	}

	block := d.GetBlock()
	if block == nil {
		return nil, errors.New("location directive requires a block")
	}

	// Build nested collectable actions
	nestedActions, err := BuildSnippetActions(block)
	if err != nil {
		return nil, fmt.Errorf("building location block actions: %w", err)
	}

	// Build location matcher
	matcher, err := buildLocationMatcher(pathPattern)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		if matcher(req) {
			// Increase block depth for override behavior
			pc.blockDepth++
			// Recursively collect nested actions
			terminated, _, err := nestedActions.Collect(rw, req, ctx, pc)
			pc.blockDepth--
			return terminated, err
		}
		return false, nil
	}, nil
}

// === Access phase directives (collect) ===

func createAccessCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createAccessAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.Access = append(pc.Access, act)
		return false, nil
	}, nil
}

// === Input filter phase directives (collect) ===

func createMoreSetInputHeadersCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createMoreSetInputHeadersAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.InputFilter = append(pc.InputFilter, act)
		return false, nil
	}, nil
}

func createMoreClearInputHeadersCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createMoreClearInputHeadersAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.InputFilter = append(pc.InputFilter, act)
		return false, nil
	}, nil
}

// === Content phase directives (collect) ===

func createAuthRequestSetCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createAuthRequestSetAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.Content = append(pc.Content, act)
		return false, nil
	}, nil
}

func createProxySetHeaderCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createProxySetHeaderAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.Content = append(pc.Content, act)
		return false, nil
	}, nil
}

func createProxyMethodCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createProxyMethodAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.Content = append(pc.Content, act)
		return false, nil
	}, nil
}

// === Header filter phase directives (collect) ===

func createAddHeaderCollectable(d config.IDirective) (CollectableAction, error) {
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

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		// Override behavior: if this add_header is at a deeper block than previous,
		// clear previous add_headers and update depth
		if pc.blockDepth > pc.overrideDepth {
			pc.HeaderFilterOverride = nil
			pc.overrideDepth = pc.blockDepth
		}
		// Only add if we're at the current override depth (ignore parent-level add_headers)
		if pc.blockDepth >= pc.overrideDepth {
			pc.HeaderFilterOverride = append(pc.HeaderFilterOverride, addHeaderEntry{
				key:    key,
				value:  val,
				always: always,
			})
		}
		return false, nil
	}, nil
}

func createMoreSetHeadersCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createMoreSetHeadersAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.HeaderFilterAdditive = append(pc.HeaderFilterAdditive, act)
		return false, nil
	}, nil
}

func createMoreClearHeadersCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createMoreClearHeadersAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.HeaderFilterAdditive = append(pc.HeaderFilterAdditive, act)
		return false, nil
	}, nil
}

func createProxyHideHeaderCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createProxyHideHeaderAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.HeaderFilterAdditive = append(pc.HeaderFilterAdditive, act)
		return false, nil
	}, nil
}

func createExpiresCollectable(d config.IDirective) (CollectableAction, error) {
	act, err := createExpiresAction(d)
	if err != nil {
		return nil, err
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext, pc *PhaseCollector) (bool, error) {
		pc.HeaderFilterAdditive = append(pc.HeaderFilterAdditive, act)
		return false, nil
	}, nil
}

// locationMatcher is a function that checks if a request matches a location pattern.
type locationMatcher func(req *http.Request) bool

// buildLocationMatcher creates a matcher function for a location pattern.
func buildLocationMatcher(pathPattern string) (locationMatcher, error) {
	// Check ~* (case-insensitive regex) before ~ (case-sensitive regex)
	if pattern, ok := strings.CutPrefix(pathPattern, "~*"); ok {
		pattern = strings.TrimSpace(pattern)
		re, err := compileRegex("(?i)" + pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling location regex: %w", err)
		}
		return func(req *http.Request) bool {
			return re.MatchString(req.URL.Path)
		}, nil
	}

	if pattern, ok := strings.CutPrefix(pathPattern, "~"); ok {
		pattern = strings.TrimSpace(pattern)
		re, err := compileRegex(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling location regex: %w", err)
		}
		return func(req *http.Request) bool {
			return re.MatchString(req.URL.Path)
		}, nil
	}

	if exact, ok := strings.CutPrefix(pathPattern, "="); ok {
		exactPath := strings.TrimSpace(exact)
		return func(req *http.Request) bool {
			return req.URL.Path == exactPath
		}, nil
	}

	// Prefix match
	return func(req *http.Request) bool {
		return strings.HasPrefix(req.URL.Path, pathPattern)
	}, nil
}

// compileRegex compiles a regex pattern, used by location matcher.
func compileRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}
