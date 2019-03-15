package types

import (
	"github.com/containous/mux"
)

// RouteAppender appends routes on a router (/api, /ping ...)
type RouteAppender interface {
	Append(systemRouter *mux.Router)
}
