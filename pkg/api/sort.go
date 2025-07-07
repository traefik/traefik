package api

import (
	"cmp"
	"net/url"
	"sort"
)

const (
	sortByParam    = "sortBy"
	directionParam = "direction"
)

const (
	ascendantSorting  = "asc"
	descendantSorting = "desc"
)

type orderedWithName interface {
	name() string
}

type orderedRouter interface {
	orderedWithName

	provider() string
	priority() int
	status() string
	rule() string
	service() string
	entryPointsCount() int
}

func sortRouters[T orderedRouter](values url.Values, routers []T) {
	sortBy := values.Get(sortByParam)

	direction := values.Get(directionParam)
	if direction == "" {
		direction = ascendantSorting
	}

	switch sortBy {
	case "name":
		sortByName(direction, routers)

	case "provider":
		sortByFunc(direction, routers, func(i int) string { return routers[i].provider() })

	case "priority":
		sortByFunc(direction, routers, func(i int) int { return routers[i].priority() })

	case "status":
		sortByFunc(direction, routers, func(i int) string { return routers[i].status() })

	case "rule":
		sortByFunc(direction, routers, func(i int) string { return routers[i].rule() })

	case "service":
		sortByFunc(direction, routers, func(i int) string { return routers[i].service() })

	case "entryPoints":
		sortByFunc(direction, routers, func(i int) int { return routers[i].entryPointsCount() })

	default:
		sortByName(direction, routers)
	}
}

func (r routerRepresentation) name() string {
	return r.Name
}

func (r routerRepresentation) provider() string {
	return r.Provider
}

func (r routerRepresentation) priority() int {
	return r.Priority
}

func (r routerRepresentation) status() string {
	return r.Status
}

func (r routerRepresentation) rule() string {
	return r.Rule
}

func (r routerRepresentation) service() string {
	return r.Service
}

func (r routerRepresentation) entryPointsCount() int {
	return len(r.EntryPoints)
}

func (r tcpRouterRepresentation) name() string {
	return r.Name
}

func (r tcpRouterRepresentation) provider() string {
	return r.Provider
}

func (r tcpRouterRepresentation) priority() int {
	return r.Priority
}

func (r tcpRouterRepresentation) status() string {
	return r.Status
}

func (r tcpRouterRepresentation) rule() string {
	return r.Rule
}

func (r tcpRouterRepresentation) service() string {
	return r.Service
}

func (r tcpRouterRepresentation) entryPointsCount() int {
	return len(r.EntryPoints)
}

func (r udpRouterRepresentation) name() string {
	return r.Name
}

func (r udpRouterRepresentation) provider() string {
	return r.Provider
}

func (r udpRouterRepresentation) priority() int {
	// noop
	return 0
}

func (r udpRouterRepresentation) status() string {
	return r.Status
}

func (r udpRouterRepresentation) rule() string {
	// noop
	return ""
}

func (r udpRouterRepresentation) service() string {
	return r.Service
}

func (r udpRouterRepresentation) entryPointsCount() int {
	return len(r.EntryPoints)
}

type orderedService interface {
	orderedWithName

	resourceType() string
	serversCount() int
	provider() string
	status() string
}

func sortServices[T orderedService](values url.Values, services []T) {
	sortBy := values.Get(sortByParam)

	direction := values.Get(directionParam)
	if direction == "" {
		direction = ascendantSorting
	}

	switch sortBy {
	case "name":
		sortByName(direction, services)

	case "type":
		sortByFunc(direction, services, func(i int) string { return services[i].resourceType() })

	case "servers":
		sortByFunc(direction, services, func(i int) int { return services[i].serversCount() })

	case "provider":
		sortByFunc(direction, services, func(i int) string { return services[i].provider() })

	case "status":
		sortByFunc(direction, services, func(i int) string { return services[i].status() })

	default:
		sortByName(direction, services)
	}
}

func (s serviceRepresentation) name() string {
	return s.Name
}

func (s serviceRepresentation) resourceType() string {
	return s.Type
}

func (s serviceRepresentation) serversCount() int {
	// TODO: maybe disable that data point altogether,
	// if we can't/won't compute a fully correct (recursive) result.
	// Or "redefine" it as only the top-level count?
	// Note: The current algo is equivalent to the webui one.
	if s.LoadBalancer == nil {
		return 0
	}

	return len(s.LoadBalancer.Servers)
}

func (s serviceRepresentation) provider() string {
	return s.Provider
}

func (s serviceRepresentation) status() string {
	return s.Status
}

func (s tcpServiceRepresentation) name() string {
	return s.Name
}

func (s tcpServiceRepresentation) resourceType() string {
	return s.Type
}

func (s tcpServiceRepresentation) serversCount() int {
	// TODO: maybe disable that data point altogether,
	// if we can't/won't compute a fully correct (recursive) result.
	// Or "redefine" it as only the top-level count?
	// Note: The current algo is equivalent to the webui one.
	if s.LoadBalancer == nil {
		return 0
	}

	return len(s.LoadBalancer.Servers)
}

func (s tcpServiceRepresentation) provider() string {
	return s.Provider
}

func (s tcpServiceRepresentation) status() string {
	return s.Status
}

func (s udpServiceRepresentation) name() string {
	return s.Name
}

func (s udpServiceRepresentation) resourceType() string {
	return s.Type
}

func (s udpServiceRepresentation) serversCount() int {
	// TODO: maybe disable that data point altogether,
	// if we can't/won't compute a fully correct (recursive) result.
	// Or "redefine" it as only the top-level count?
	// Note: The current algo is equivalent to the webui one.
	if s.LoadBalancer == nil {
		return 0
	}

	return len(s.LoadBalancer.Servers)
}

func (s udpServiceRepresentation) provider() string {
	return s.Provider
}

func (s udpServiceRepresentation) status() string {
	return s.Status
}

type orderedMiddleware interface {
	orderedWithName

	resourceType() string
	provider() string
	status() string
}

func sortMiddlewares[T orderedMiddleware](values url.Values, middlewares []T) {
	sortBy := values.Get(sortByParam)

	direction := values.Get(directionParam)
	if direction == "" {
		direction = ascendantSorting
	}

	switch sortBy {
	case "name":
		sortByName(direction, middlewares)

	case "type":
		sortByFunc(direction, middlewares, func(i int) string { return middlewares[i].resourceType() })

	case "provider":
		sortByFunc(direction, middlewares, func(i int) string { return middlewares[i].provider() })

	case "status":
		sortByFunc(direction, middlewares, func(i int) string { return middlewares[i].status() })

	default:
		sortByName(direction, middlewares)
	}
}

func (m middlewareRepresentation) name() string {
	return m.Name
}

func (m middlewareRepresentation) resourceType() string {
	return m.Type
}

func (m middlewareRepresentation) provider() string {
	return m.Provider
}

func (m middlewareRepresentation) status() string {
	return m.Status
}

func (m tcpMiddlewareRepresentation) name() string {
	return m.Name
}

func (m tcpMiddlewareRepresentation) resourceType() string {
	return m.Type
}

func (m tcpMiddlewareRepresentation) provider() string {
	return m.Provider
}

func (m tcpMiddlewareRepresentation) status() string {
	return m.Status
}

func sortByName[T orderedWithName](direction string, results []T) {
	// Ascending
	if direction == ascendantSorting {
		sort.Slice(results, func(i, j int) bool {
			return results[i].name() < results[j].name()
		})

		return
	}

	// Descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].name() > results[j].name()
	})
}

func sortByFunc[T orderedWithName, U cmp.Ordered](direction string, results []T, fn func(int) U) {
	// Ascending
	if direction == ascendantSorting {
		sort.Slice(results, func(i, j int) bool {
			if fn(i) == fn(j) {
				return results[i].name() < results[j].name()
			}

			return fn(i) < fn(j)
		})

		return
	}

	// Descending
	sort.Slice(results, func(i, j int) bool {
		if fn(i) == fn(j) {
			return results[i].name() > results[j].name()
		}

		return fn(i) > fn(j)
	})
}
