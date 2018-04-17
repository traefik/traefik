package types

import (
	"github.com/containous/mux"
)

// InternalRouter router used by server to register internal routes (/api, /ping ...)
type InternalRouter interface {
	AddRoutes(systemRouter *mux.Router)
}
