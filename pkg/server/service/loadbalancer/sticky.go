package loadbalancer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// StickyMode defines the type of sticky session persistence.
type StickyMode string

const (
	// StickyModeCookie uses cookies for session persistence.
	StickyModeCookie StickyMode = "cookie"
	// StickyModeHeader uses headers for session persistence.
	StickyModeHeader StickyMode = "header"
)

// NamedHandler is a http.Handler with a name.
type NamedHandler struct {
	http.Handler

	Name string
}

// stickyCookie represents a sticky cookie configuration.
type stickyCookie struct {
	name     string
	secure   bool
	httpOnly bool
	sameSite http.SameSite
	maxAge   int
	path     string
	domain   string
}

// stickyHeader represents a sticky header configuration.
type stickyHeader struct {
	name string
}

// Sticky ensures that client consistently interacts with the same HTTP handler
// by adding a sticky cookie or header to the response.
// This allows subsequent requests from the same client to be routed to the same handler,
// enabling session persistence across multiple requests.
type Sticky struct {
	// mode defines whether to use cookie or header for sticky sessions.
	mode StickyMode

	// cookie is the sticky cookie configuration (used when mode is StickyModeCookie).
	cookie *stickyCookie

	// header is the sticky header configuration (used when mode is StickyModeHeader).
	header *stickyHeader

	// References all the handlers by name and also by the hashed value of the name.
	handlersMu             sync.RWMutex
	hashMap                map[string]string
	stickyMap              map[string]*NamedHandler
	compatibilityStickyMap map[string]*NamedHandler
}

// NewSticky creates a new Sticky instance from a dynamic.Sticky configuration.
// It returns nil if the configuration is nil or has no cookie/header config.
func NewSticky(cfg *dynamic.Sticky) *Sticky {
	if cfg == nil {
		return nil
	}

	if cfg.Cookie != nil {
		return NewStickyCookie(*cfg.Cookie)
	}

	if cfg.Header != nil {
		return NewStickyHeader(*cfg.Header)
	}

	return nil
}

// NewStickyCookie creates a new Sticky instance configured for cookie-based persistence.
func NewStickyCookie(cookieConfig dynamic.Cookie) *Sticky {
	cookie := &stickyCookie{
		name:     cookieConfig.Name,
		secure:   cookieConfig.Secure,
		httpOnly: cookieConfig.HTTPOnly,
		sameSite: convertSameSite(cookieConfig.SameSite),
		maxAge:   cookieConfig.MaxAge,
		path:     "/",
		domain:   cookieConfig.Domain,
	}
	if cookieConfig.Path != nil {
		cookie.path = *cookieConfig.Path
	}

	return &Sticky{
		mode:                   StickyModeCookie,
		cookie:                 cookie,
		hashMap:                make(map[string]string),
		stickyMap:              make(map[string]*NamedHandler),
		compatibilityStickyMap: make(map[string]*NamedHandler),
	}
}

// NewStickyHeader creates a new Sticky instance configured for header-based persistence.
func NewStickyHeader(headerConfig dynamic.Header) *Sticky {
	name := headerConfig.Name
	if name == "" {
		name = "X-Sticky-Session"
	}

	return &Sticky{
		mode: StickyModeHeader,
		header: &stickyHeader{
			name: name,
		},
		hashMap:                make(map[string]string),
		stickyMap:              make(map[string]*NamedHandler),
		compatibilityStickyMap: make(map[string]*NamedHandler),
	}
}

// AddHandler adds a http.Handler to the sticky pool.
func (s *Sticky) AddHandler(name string, h http.Handler) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	sha256HashedName := sha256Hash(name)
	s.hashMap[name] = sha256HashedName

	handler := &NamedHandler{
		Handler: h,
		Name:    name,
	}

	s.stickyMap[sha256HashedName] = handler
	s.compatibilityStickyMap[name] = handler

	hashedName := fnvHash(name)
	s.compatibilityStickyMap[hashedName] = handler

	// server.URL was fnv hashed in service.Manager
	// so we can have "double" fnv hash in already existing cookies
	hashedName = fnvHash(hashedName)
	s.compatibilityStickyMap[hashedName] = handler
}

// StickyHandler returns the NamedHandler corresponding to the sticky cookie or header.
// It also returns a boolean which indicates if the sticky value has to be overwritten
// because it uses a deprecated hash algorithm.
func (s *Sticky) StickyHandler(req *http.Request) (*NamedHandler, bool, error) {
	switch s.mode {
	case StickyModeHeader:
		return s.stickyHandlerFromHeader(req)
	default:
		return s.stickyHandlerFromCookie(req)
	}
}

// stickyHandlerFromCookie returns the handler based on the sticky cookie value.
func (s *Sticky) stickyHandlerFromCookie(req *http.Request) (*NamedHandler, bool, error) {
	cookie, err := req.Cookie(s.cookie.name)
	if err != nil && errors.Is(err, http.ErrNoCookie) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("reading cookie: %w", err)
	}

	return s.lookupHandler(cookie.Value)
}

// stickyHandlerFromHeader returns the handler based on the sticky header value.
func (s *Sticky) stickyHandlerFromHeader(req *http.Request) (*NamedHandler, bool, error) {
	value := req.Header.Get(s.header.name)
	if value == "" {
		return nil, false, nil
	}

	return s.lookupHandler(value)
}

// lookupHandler finds a handler by its sticky value (hash).
func (s *Sticky) lookupHandler(value string) (*NamedHandler, bool, error) {
	s.handlersMu.RLock()
	handler, ok := s.stickyMap[value]
	s.handlersMu.RUnlock()

	if ok && handler != nil {
		return handler, false, nil
	}

	s.handlersMu.RLock()
	handler, ok = s.compatibilityStickyMap[value]
	s.handlersMu.RUnlock()

	return handler, ok, nil
}

// WriteStickyResponse writes the sticky cookie or header to the response.
func (s *Sticky) WriteStickyResponse(rw http.ResponseWriter, name string) error {
	switch s.mode {
	case StickyModeHeader:
		return s.WriteStickyHeader(rw, name)
	default:
		return s.WriteStickyCookie(rw, name)
	}
}

// WriteStickyCookie writes a sticky cookie to the response to stick the client to the given handler name.
func (s *Sticky) WriteStickyCookie(rw http.ResponseWriter, name string) error {
	s.handlersMu.RLock()
	hash, ok := s.hashMap[name]
	s.handlersMu.RUnlock()
	if !ok {
		return fmt.Errorf("no hash found for handler named %s", name)
	}

	cookie := &http.Cookie{
		Name:     s.cookie.name,
		Value:    hash,
		Path:     s.cookie.path,
		Domain:   s.cookie.domain,
		HttpOnly: s.cookie.httpOnly,
		Secure:   s.cookie.secure,
		SameSite: s.cookie.sameSite,
		MaxAge:   s.cookie.maxAge,
	}
	http.SetCookie(rw, cookie)

	return nil
}

// WriteStickyHeader writes a sticky header to the response to stick the client to the given handler name.
func (s *Sticky) WriteStickyHeader(rw http.ResponseWriter, name string) error {
	s.handlersMu.RLock()
	hash, ok := s.hashMap[name]
	s.handlersMu.RUnlock()
	if !ok {
		return fmt.Errorf("no hash found for handler named %s", name)
	}

	rw.Header().Set(s.header.name, hash)
	return nil
}

func convertSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteDefaultMode
	}
}

// fnvHash returns the FNV-64 hash of the input string.
func fnvHash(input string) string {
	hasher := fnv.New64()
	// We purposely ignore the error because the implementation always returns nil.
	_, _ = hasher.Write([]byte(input))

	return strconv.FormatUint(hasher.Sum64(), 16)
}

// sha256Hash returns the SHA-256 hash, truncated to 16 characters, of the input string.
func sha256Hash(input string) string {
	hash := sha256.New()
	// We purposely ignore the error because the implementation always returns nil.
	_, _ = hash.Write([]byte(input))

	hashedInput := hex.EncodeToString(hash.Sum(nil))
	if len(hashedInput) < 16 {
		return hashedInput
	}
	return hashedInput[:16]
}
