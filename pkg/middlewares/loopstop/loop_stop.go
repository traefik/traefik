package loopstop

import (
	"crypto/sha1"
	"fmt"
	"net/http"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v2/pkg/log"
)

const xTraefikRouter = "X-Traefik-Router"

type LoopStop struct {
	routerName     string
	routerNameHash string
	next           http.Handler
}

// WrapHandler Wraps router to alice.Constructor.
func WrapHandler(routerName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return NewLoopStop(routerName, next)
	}
}

// NewLoopStop creates a new LoopStop.
// LoopStop middleware is an internal middleware used to avoid infinite requests loop on the same router.
func NewLoopStop(routerName string, next http.Handler) (*LoopStop, error) {
	hash, err := makeHash(routerName)
	if err != nil {
		return nil, err
	}

	return &LoopStop{
		routerName:     routerName,
		routerNameHash: hash,
		next:           next,
	}, nil
}

// ServeHTTP implements http.Handler.
func (l *LoopStop) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(xTraefikRouter) == l.routerNameHash {
		log.WithoutContext().Debugf("Rejecting request in provenance of the same default rule's router (%q) to stop potential infinite loop.", l.routerName)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	r.Header.Set(xTraefikRouter, l.routerNameHash)

	l.next.ServeHTTP(w, r)
}

func makeHash(routerName string) (string, error) {
	h := sha1.New()
	if _, err := h.Write([]byte(routerName)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
