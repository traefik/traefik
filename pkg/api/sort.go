package api

import (
	"net/url"
	"sort"
)

const (
	sortByParam       = "sortBy"
	directionParam    = "direction"
	ascendantSorting  = "asc"
	descendantSorting = "desc"
)

type orderedRouter interface {
	name() string
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
		sortByProvider(direction, routers)

	case "priority":
		sortByPriority(direction, routers)

	case "status":
		sortByStatus(direction, routers)

	case "rule":
		sortByRule(direction, routers)

	case "service":
		sortByService(direction, routers)

	case "entryPoints":
		sortByEntryPoints(direction, routers)

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
	name() string
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
		sortByType(direction, services)

	case "servers":
		sortByServers(direction, services)

	case "provider":
		sortByProvider(direction, services)

	case "status":
		sortByStatus(direction, services)

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
	return len(s.LoadBalancer.Servers)
}

func (s udpServiceRepresentation) provider() string {
	return s.Provider
}

func (s udpServiceRepresentation) status() string {
	return s.Status
}

type orderedMiddleware interface {
	name() string
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
		sortByType(direction, middlewares)

	case "provider":
		sortByProvider(direction, middlewares)

	case "status":
		sortByStatus(direction, middlewares)

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

type orderedByName interface {
	name() string
}

func sortByName[T orderedByName](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			return results[i].name() < results[j].name()
		}

		// Descending
		return results[i].name() > results[j].name()
	})
}

type orderedByPriority interface {
	name() string
	priority() int
}

func sortByPriority[T orderedByPriority](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].priority() == results[j].priority() {
				return results[i].name() < results[j].name()
			}
			return results[i].priority() < results[j].priority()
		}

		// Descending
		if results[i].priority() == results[j].priority() {
			return results[i].name() > results[j].name()
		}
		return results[i].priority() > results[j].priority()
	})
}

type orderedByType interface {
	name() string
	resourceType() string
}

func sortByType[T orderedByType](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].resourceType() == results[j].resourceType() {
				return results[i].name() < results[j].name()
			}
			return results[i].resourceType() < results[j].resourceType()
		}

		// Descending
		if results[i].resourceType() == results[j].resourceType() {
			return results[i].name() > results[j].name()
		}
		return results[i].resourceType() > results[j].resourceType()
	})
}

type orderedByServers interface {
	name() string
	serversCount() int
}

func sortByServers[T orderedByServers](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].serversCount() == results[j].serversCount() {
				return results[i].name() < results[j].name()
			}
			return results[i].serversCount() < results[j].serversCount()
		}

		// Descending
		if results[i].serversCount() == results[j].serversCount() {
			return results[i].name() > results[j].name()
		}
		return results[i].serversCount() > results[j].serversCount()
	})
}

type orderedByStatus interface {
	name() string
	status() string
}

func sortByStatus[T orderedByStatus](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].status() == results[j].status() {
				return results[i].name() < results[j].name()
			}
			return results[i].status() < results[j].status()
		}

		// Descending
		if results[i].status() == results[j].status() {
			return results[i].name() > results[j].name()
		}
		return results[i].status() > results[j].status()
	})
}

type orderedByProvider interface {
	provider() string
	name() string
}

func sortByProvider[T orderedByProvider](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].provider() == results[j].provider() {
				return results[i].name() < results[j].name()
			}
			return results[i].provider() < results[j].provider()
		}

		// Descending
		if results[i].provider() == results[j].provider() {
			return results[i].name() > results[j].name()
		}
		return results[i].provider() > results[j].provider()
	})
}

type orderedByRule interface {
	rule() string
	name() string
}

func sortByRule[T orderedByRule](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].rule() == results[j].rule() {
				return results[i].name() < results[j].name()
			}
			return results[i].rule() < results[j].rule()
		}

		// Descending
		if results[i].rule() == results[j].rule() {
			return results[i].name() > results[j].name()
		}
		return results[i].rule() > results[j].rule()
	})
}

type orderedByService interface {
	service() string
	name() string
}

func sortByService[T orderedByService](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].service() == results[j].service() {
				return results[i].name() < results[j].name()
			}
			return results[i].service() < results[j].service()
		}

		// Descending
		if results[i].service() == results[j].service() {
			return results[i].name() > results[j].name()
		}
		return results[i].service() > results[j].service()
	})
}

type orderedByEntryPoints interface {
	entryPointsCount() int
	name() string
}

func sortByEntryPoints[T orderedByEntryPoints](direction string, results []T) {
	sort.Slice(results, func(i, j int) bool {
		// Ascending
		if direction == ascendantSorting {
			if results[i].entryPointsCount() == results[j].entryPointsCount() {
				return results[i].name() < results[j].name()
			}
			return results[i].entryPointsCount() < results[j].entryPointsCount()
		}

		// Descending
		if results[i].entryPointsCount() == results[j].entryPointsCount() {
			return results[i].name() > results[j].name()
		}
		return results[i].entryPointsCount() > results[j].entryPointsCount()
	})
}
