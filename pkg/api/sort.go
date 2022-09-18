package api

import (
	"net/url"
	"sort"
)

func sortHTTPMiddlewares(query url.Values, results []middlewareRepresentation) []middlewareRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "type":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Type < results[j].Type
			}

			// Descending
			return results[i].Type > results[j].Type
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	}

	return results
}

func sortHTTPServices(query url.Values, results []serviceRepresentation) []serviceRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "type":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Type < results[j].Type
			}

			// Descending
			return results[i].Type > results[j].Type
		})

	case "servers":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return len(results[i].ServerStatus) < len(results[j].ServerStatus)
			}

			// Descending
			return len(results[i].ServerStatus) > len(results[j].ServerStatus)
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	}

	return results
}

func sortHTTPRouters(query url.Values, results []routerRepresentation) []routerRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "priority":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Priority < results[j].Priority
			}

			// Descending
			return results[i].Priority > results[j].Priority
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	case "rule":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Rule == results[j].Rule {
					return results[i].Name < results[j].Name
				}
				return results[i].Rule < results[j].Rule
			}

			// Descending
			if results[i].Rule == results[j].Rule {
				return results[i].Name > results[j].Name
			}
			return results[i].Rule > results[j].Rule
		})

	case "service":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Service < results[j].Service
			}

			// Descending
			return results[i].Service > results[j].Service
		})

	case "entryPoints":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if len(results[i].EntryPoints) == len(results[j].EntryPoints) {
					return results[i].Name[0] < results[j].Name[0]
				}
				return len(results[i].EntryPoints) < len(results[j].EntryPoints)
			}

			// Descending
			if len(results[i].EntryPoints) == len(results[j].EntryPoints) {
				return results[i].Name[0] > results[j].Name[0]
			}
			return len(results[i].EntryPoints) > len(results[j].EntryPoints)
		})
	}

	return results
}

func sortTCPMiddlewares(query url.Values, results []tcpMiddlewareRepresentation) []tcpMiddlewareRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "type":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Type < results[j].Type
			}

			// Descending
			return results[i].Type > results[j].Type
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	}

	return results
}

func sortTCPServices(query url.Values, results []tcpServiceRepresentation) []tcpServiceRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "type":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Type < results[j].Type
			}

			// Descending
			return results[i].Type > results[j].Type
		})

	case "servers":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return len(results[i].LoadBalancer.Servers) < len(results[j].LoadBalancer.Servers)
			}

			// Descending
			return len(results[i].LoadBalancer.Servers) > len(results[j].LoadBalancer.Servers)
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	}

	return results
}

func sortTCPRouters(query url.Values, results []tcpRouterRepresentation) []tcpRouterRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "priority":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Priority < results[j].Priority
			}

			// Descending
			return results[i].Priority > results[j].Priority
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	case "rule":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Rule == results[j].Rule {
					return results[i].Name < results[j].Name
				}
				return results[i].Rule < results[j].Rule
			}

			// Descending
			if results[i].Rule == results[j].Rule {
				return results[i].Name > results[j].Name
			}
			return results[i].Rule > results[j].Rule
		})

	case "service":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Service < results[j].Service
			}

			// Descending
			return results[i].Service > results[j].Service
		})

	case "entryPoints":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if len(results[i].EntryPoints) == len(results[j].EntryPoints) {
					return results[i].Name[0] < results[j].Name[0]
				}
				return len(results[i].EntryPoints) < len(results[j].EntryPoints)
			}

			// Descending
			if len(results[i].EntryPoints) == len(results[j].EntryPoints) {
				return results[i].Name[0] > results[j].Name[0]
			}
			return len(results[i].EntryPoints) > len(results[j].EntryPoints)
		})
	}

	return results
}

func sortUDPServices(query url.Values, results []udpServiceRepresentation) []udpServiceRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "type":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Type < results[j].Type
			}

			// Descending
			return results[i].Type > results[j].Type
		})

	case "servers":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return len(results[i].LoadBalancer.Servers) < len(results[j].LoadBalancer.Servers)
			}

			// Descending
			return len(results[i].LoadBalancer.Servers) > len(results[j].LoadBalancer.Servers)
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	}

	return results
}

func sortUDPRouters(query url.Values, results []udpRouterRepresentation) []udpRouterRepresentation {
	sortBy := query.Get("sortBy")
	direction := query.Get("direction")

	if len(query) == 0 || sortBy == "" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}

	switch sortBy {
	case "name":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Name < results[j].Name
			}

			// Descending
			return results[i].Name > results[j].Name
		})

	case "provider":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Provider == results[j].Provider {
					return results[i].Name < results[j].Name
				}
				return results[i].Provider < results[j].Provider
			}

			// Descending
			if results[i].Provider == results[j].Provider {
				return results[i].Name > results[j].Name
			}
			return results[i].Provider > results[j].Provider
		})

	case "status":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if results[i].Status == results[j].Status {
					return results[i].Name < results[j].Name
				}
				return results[i].Status < results[j].Status
			}

			// Descending
			if results[i].Status == results[j].Status {
				return results[i].Name > results[j].Name
			}
			return results[i].Status > results[j].Status
		})

	case "service":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				return results[i].Service < results[j].Service
			}

			// Descending
			return results[i].Service > results[j].Service
		})

	case "entryPoints":
		sort.Slice(results, func(i, j int) bool {
			// Ascending
			if direction == "asc" {
				if len(results[i].EntryPoints) == len(results[j].EntryPoints) {
					return results[i].Name[0] < results[j].Name[0]
				}
				return len(results[i].EntryPoints) < len(results[j].EntryPoints)
			}

			// Descending
			if len(results[i].EntryPoints) == len(results[j].EntryPoints) {
				return results[i].Name[0] > results[j].Name[0]
			}
			return len(results[i].EntryPoints) > len(results[j].EntryPoints)
		})
	}

	return results
}
