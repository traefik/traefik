package snippet

// Phase represents an NGINX request processing phase.
type Phase int

const (
	// PhaseInputFilter modifies incoming request headers (more_set_input_headers, more_clear_input_headers).
	PhaseInputFilter Phase = iota
	// PhaseRewrite handles set, rewrite, return, if directives.
	PhaseRewrite
	// PhaseLocation handles location directive matching (server context only).
	PhaseLocation
	// PhaseAccess handles allow, deny directives.
	PhaseAccess
	// PhaseContent handles proxy_set_header, proxy_method directives.
	PhaseContent
	// PhaseHeaderFilter handles response header directives (add_header, more_set_headers, etc.).
	PhaseHeaderFilter
)

// PhaseActions groups actions by their NGINX processing phase.
type PhaseActions struct {
	// InputFilter contains request header manipulation actions.
	// Directives: more_set_input_headers, more_clear_input_headers
	InputFilter []action

	// Rewrite contains URI rewriting and control flow actions.
	// Directives: set, rewrite, return, if
	Rewrite []action

	// Location contains location block matching actions (server context only).
	// Directives: location
	Location []action

	// Access contains access control actions.
	// Directives: allow, deny
	Access []action

	// Content contains upstream request configuration actions.
	// Directives: proxy_set_header, proxy_method
	Content []action

	// HeaderFilter contains response header manipulation actions.
	// These are split into overridable and additive categories.
	// Overridable (location replaces server): add_header
	// Additive (both apply): more_set_headers, more_clear_headers, proxy_hide_header, expires
	HeaderFilterOverride []action // add_header only
	HeaderFilterAdditive []action // more_*, proxy_hide_header, expires
}

// HasAccess returns true if there are any access control directives.
func (pa *PhaseActions) HasAccess() bool {
	return pa != nil && len(pa.Access) > 0
}

// HasHeaderOverride returns true if there are any override header directives (add_header).
func (pa *PhaseActions) HasHeaderOverride() bool {
	return pa != nil && len(pa.HeaderFilterOverride) > 0
}

// HasContent returns true if there are any content phase directives.
func (pa *PhaseActions) HasContent() bool {
	return pa != nil && len(pa.Content) > 0
}
