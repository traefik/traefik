package types

import (
	"github.com/gorilla/mux"
)

// RouteAppender appends routes on a router (/api, /ping ...).
type RouteAppender interface {
	Append(systemRouter *mux.Router)
}
