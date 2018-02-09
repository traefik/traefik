package types

import (
	"github.com/containous/mux"
)

type InternalRouter interface {
	AddRoutes(systemRouter *mux.Router)
}
