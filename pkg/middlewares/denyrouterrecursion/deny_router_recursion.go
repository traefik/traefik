package denyrouterrecursion

import (
	"errors"
	"hash/fnv"
	"net/http"
	"strconv"

	"github.com/baqupio/baqup/v3/pkg/observability/logs"
	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
)

const xBaqupRouter = "X-Baqup-Router"

type DenyRouterRecursion struct {
	routerName     string
	routerNameHash string
	next           http.Handler
}

// WrapHandler Wraps router to alice.Constructor.
func WrapHandler(routerName string) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return New(routerName, next)
	}
}

// New creates a new DenyRouterRecursion.
// DenyRouterRecursion middleware is an internal middleware used to avoid infinite requests loop on the same router.
func New(routerName string, next http.Handler) (*DenyRouterRecursion, error) {
	if routerName == "" {
		return nil, errors.New("routerName cannot be empty")
	}

	return &DenyRouterRecursion{
		routerName:     routerName,
		routerNameHash: makeHash(routerName),
		next:           next,
	}, nil
}

// ServeHTTP implements http.Handler.
func (l *DenyRouterRecursion) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get(xBaqupRouter) == l.routerNameHash {
		logger := log.With().Str(logs.MiddlewareType, "DenyRouterRecursion").Logger()
		logger.Debug().Msgf("Rejecting request in provenance of the same router (%q) to stop potential infinite loop.", l.routerName)

		rw.WriteHeader(http.StatusBadRequest)

		return
	}

	req.Header.Set(xBaqupRouter, l.routerNameHash)

	l.next.ServeHTTP(rw, req)
}

func makeHash(routerName string) string {
	hasher := fnv.New64()
	// purposely ignoring the error, as no error can be returned from the implementation.
	_, _ = hasher.Write([]byte(routerName))
	return strconv.FormatUint(hasher.Sum64(), 16)
}
