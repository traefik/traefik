package snippet

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
	"github.com/tufanbarisyildirim/gonginx/config"
)

// context holds variables set during request processing.
type actionContext struct {
	vars                    map[string]string
	nonMergeablePostActions map[string][]action
	mergeablePostActions    []action

	statusCode int
	body       string
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
		if finish {
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
		case "location":
			action, err := createLocationAction(d)
			if err != nil {
				return nil, fmt.Errorf("creating location action: %w", err)
			}
			acts.actions = append(acts.actions, action)
		default:
			return nil, fmt.Errorf("unsupported directive %q", d.GetName())
		}
	}

	return acts, nil
}

func createAddHeaderAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) < 2 {
		return nil, errors.New("add_header directive must have at least 2 parameters (header and value)")
	}

	key := params[0].String()
	val := trimQuote(params[1].String())

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		rw.Header().Add(key, ingressnginx.ReplaceVariables(val, req, ctx.vars))
		return false, nil
	}, nil
}

func createMoreSetHeadersAction(d config.IDirective) (action, error) {
	key, val, err := parseMoreSetDirective(d, "more_set_headers")
	if err != nil {
		return nil, fmt.Errorf("parsing more_set_headers directive: %w", err)
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		rw.Header().Set(key, ingressnginx.ReplaceVariables(val, req, ctx.vars))
		return false, nil
	}, nil
}

func createMoreSetInputHeadersAction(d config.IDirective) (action, error) {
	key, val, err := parseMoreSetDirective(d, "more_set_input_headers")
	if err != nil {
		return nil, fmt.Errorf("parsing more_set_input_headers directive: %w", err)
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		req.Header.Set(key, ingressnginx.ReplaceVariables(val, req, ctx.vars))
		return false, nil
	}, nil
}

func parseMoreSetDirective(d config.IDirective, directiveName string) (string, string, error) {
	params := d.GetParameters()
	if len(params) != 1 {
		return "", "", fmt.Errorf("%s directive must have 1 parameter", directiveName)
	}

	trimmedVal := trimQuote(params[0].String())
	parts := strings.SplitN(trimmedVal, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%s directive must have a single ':' separator", directiveName)
	}

	return parts[0], strings.TrimSpace(parts[1]), nil
}

func createProxySetHeaderAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) < 2 {
		return nil, errors.New("proxy_set_header directive requires 2 parameters (header and value)")
	}

	key := params[0].String()
	val := trimQuote(params[1].String())

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		req.Header.Set(key, ingressnginx.ReplaceVariables(val, req, ctx.vars))
		return false, nil
	}, nil
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

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		if evaluateCondition(condition, req, ctx) {
			return blockActions.Execute(rw, req, ctx)
		}
		return false, nil
	}, nil
}

// createReturnAction is a simplified implementation of NGINX return, it assumes code [text].
func createReturnAction(d config.IDirective) (action, error) {
	params := d.GetParameters()
	if len(params) == 0 {
		return nil, errors.New("return directive requires parameters")
	}

	code, err := strconv.Atoi(params[0].String())
	if err != nil {
		return nil, fmt.Errorf("invalid return code: %w", err)
	}

	var text string
	if len(params) > 1 {
		text = trimQuote(params[1].String())
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		ctx.statusCode = code
		if text != "" {
			ctx.body = text
		}
		return true, nil
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

	// Compile regex if it's a regex pattern (starts with ~).
	var re *regexp.Regexp
	if pattern, ok := strings.CutPrefix(pathPattern, "~"); ok {
		pattern = strings.TrimSpace(pattern)

		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling location regex: %w", err)
		}
	}

	return func(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
		var matches bool
		switch {
		case re != nil:
			// Regex match
			matches = re.MatchString(req.URL.Path)
		case strings.HasPrefix(pathPattern, "="):
			// Exact match
			exactPath := strings.TrimSpace(strings.TrimPrefix(pathPattern, "="))
			matches = req.URL.Path == exactPath
		default:
			// Prefix match
			matches = strings.HasPrefix(req.URL.Path, pathPattern)
		}

		if matches {
			stop, err := blockActions.Execute(rw, req, ctx)
			if err != nil {
				return false, fmt.Errorf("executing location block: %w", err)
			}

			if !stop {
				ctx.statusCode = 503
				return true, nil
			}

			return stop, nil
		}
		return false, nil
	}, nil
}

// evaluateCondition is a simplified condition evaluation.
// It supports $var = value, $var != value, $var ~ value, $var !~ value, $var ~* value, $var !~* value.
func evaluateCondition(condition string, req *http.Request, ctx *actionContext) bool {
	parts := strings.Fields(condition)
	if len(parts) == 0 {
		return false
	}

	// Simple variable check: if ($var).
	if len(parts) == 1 {
		if val, ok := ctx.vars[parts[0]]; ok {
			return val != "" && val != "0"
		}
		return false
	}

	// Comparison: $var = value or $var != value.
	if len(parts) >= 3 {
		varName := ingressnginx.ReplaceVariables(strings.Trim(parts[0], `"`), req, ctx.vars)
		operator := parts[1]
		expectedValue := ingressnginx.ReplaceVariables(strings.Trim(parts[2], `"`), req, ctx.vars)

		switch operator {
		case "=":
			return varName == expectedValue
		case "!=":
			return varName != expectedValue
		case "~":
			// Regex match.
			re, err := regexp.Compile(expectedValue)
			if err != nil {
				return false
			}
			return re.MatchString(varName)
		case "!~":
			// Negative regex match.
			re, err := regexp.Compile(expectedValue)
			if err != nil {
				return false
			}
			return !re.MatchString(varName)
		case "~*":
			// Case-insensitive regex match.
			re, err := regexp.Compile("(?i)" + expectedValue)
			if err != nil {
				return false
			}
			return re.MatchString(varName)
		case "!~*":
			// Negative case-insensitive regex match.
			re, err := regexp.Compile("(?i)" + expectedValue)
			if err != nil {
				return false
			}
			return !re.MatchString(varName)
		}
	}

	return false
}

func trimQuote(val string) string {
	if len(val) > 1 {
		if val[0] == '"' && val[len(val)-1] == '"' {
			return val[1 : len(val)-1]
		}
	}
	return val
}
