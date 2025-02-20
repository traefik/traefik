package loadbalancer

import (
	"errors"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type Sticky struct {
	stickyCookie *stickyCookie

	handlersMu sync.RWMutex
	hashMap    map[string]string
	// References all the handlers by name and also by the hashed value of the name.
	stickyMap              map[string]*HandlerWithHash
	compatibilityStickyMap map[string]*HandlerWithHash
}

func New(config *dynamic.Sticky) *Sticky {
	var sticky *Sticky
	if config != nil && config.Cookie != nil {
		sticky = &Sticky{
			stickyCookie: &stickyCookie{
				name:     config.Cookie.Name,
				secure:   config.Cookie.Secure,
				httpOnly: config.Cookie.HTTPOnly,
				sameSite: config.Cookie.SameSite,
				maxAge:   config.Cookie.MaxAge,
				path:     "/",
			},
		}
		if config.Cookie.Path != nil {
			sticky.stickyCookie.path = *config.Cookie.Path
		}

		sticky.hashMap = make(map[string]string)
		sticky.stickyMap = make(map[string]*HandlerWithHash)
		sticky.compatibilityStickyMap = make(map[string]*HandlerWithHash)
	}

	return sticky
}

type HandlerWithHash struct {
	http.Handler

	Name string
}

// stickyCookie represents a sticky cookie.
type stickyCookie struct {
	name     string
	secure   bool
	httpOnly bool
	sameSite string
	maxAge   int
	path     string
}

func (s *Sticky) GetStickyHandler(req *http.Request) (*HandlerWithHash, bool) {
	cookie, err := req.Cookie(s.stickyCookie.name)

	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		log.Warn().Err(err).Msg("Error while reading cookie")
	}

	if err == nil && cookie != nil {
		s.handlersMu.RLock()
		handler, ok := s.stickyMap[cookie.Value]
		s.handlersMu.RUnlock()

		if ok && handler != nil {
			return handler, false
		}

		s.handlersMu.RLock()
		handler, ok = s.compatibilityStickyMap[cookie.Value]
		s.handlersMu.RUnlock()

		if ok && handler != nil {
			return handler, true
		}
	}

	return nil, false
}

func (s *Sticky) WriteStickyCookie(w http.ResponseWriter, name string) error {
	hash, ok := s.hashMap[name]
	if !ok {
		return errors.New("hash not found")
	}

	cookie := &http.Cookie{
		Name:     s.stickyCookie.name,
		Value:    hash,
		Path:     s.stickyCookie.path,
		HttpOnly: s.stickyCookie.httpOnly,
		Secure:   s.stickyCookie.secure,
		SameSite: convertSameSite(s.stickyCookie.sameSite),
		MaxAge:   s.stickyCookie.maxAge,
	}
	http.SetCookie(w, cookie)

	return nil
}

func (s *Sticky) Add(name string, h http.Handler) {
	sha256HashedName := Sha256Hash(name)
	s.hashMap[name] = sha256HashedName

	handler := &HandlerWithHash{
		Handler: h,
		Name:    name,
	}

	s.stickyMap[sha256HashedName] = handler
	s.compatibilityStickyMap[name] = handler

	hashedName := FnvHash(name)
	s.compatibilityStickyMap[hashedName] = handler

	// server.URL was fnv hashed in service.Manager
	// so we can have "double" fnv hash in already existing cookies
	hashedName = FnvHash(hashedName)
	s.compatibilityStickyMap[hashedName] = handler
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
