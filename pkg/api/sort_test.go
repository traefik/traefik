package api

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
)

func TestSortRouters(t *testing.T) {
	testCases := []struct {
		direction string
		sortBy    string
		elements  []orderedRouter
		expected  []orderedRouter
	}{
		{
			direction: ascendantSorting,
			sortBy:    "name",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "b",
				},
				routerRepresentation{
					Name: "a",
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "a",
				},
				routerRepresentation{
					Name: "b",
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "name",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "a",
				},
				routerRepresentation{
					Name: "b",
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "b",
				},
				routerRepresentation{
					Name: "a",
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "provider",
			elements: []orderedRouter{
				routerRepresentation{
					Name:     "b",
					Provider: "b",
				},
				routerRepresentation{
					Name:     "b",
					Provider: "a",
				},
				routerRepresentation{
					Name:     "a",
					Provider: "b",
				},
				routerRepresentation{
					Name:     "a",
					Provider: "a",
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name:     "a",
					Provider: "a",
				},
				routerRepresentation{
					Name:     "b",
					Provider: "a",
				},
				routerRepresentation{
					Name:     "a",
					Provider: "b",
				},
				routerRepresentation{
					Name:     "b",
					Provider: "b",
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "provider",
			elements: []orderedRouter{
				routerRepresentation{
					Name:     "a",
					Provider: "a",
				},
				routerRepresentation{
					Name:     "a",
					Provider: "b",
				},
				routerRepresentation{
					Name:     "b",
					Provider: "a",
				},
				routerRepresentation{
					Name:     "b",
					Provider: "b",
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name:     "b",
					Provider: "b",
				},
				routerRepresentation{
					Name:     "a",
					Provider: "b",
				},
				routerRepresentation{
					Name:     "b",
					Provider: "a",
				},
				routerRepresentation{
					Name:     "a",
					Provider: "a",
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "priority",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "priority",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 2,
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						EffectivePriority: 1,
					},
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "status",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "status",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "rule",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "rule",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "service",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "service",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "entryPoints",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "entryPoints",
			elements: []orderedRouter{
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
			},
			expected: []orderedRouter{
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
				},
				routerRepresentation{
					Name: "b",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
				routerRepresentation{
					Name: "a",
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(fmt.Sprintf("%s-%s", test.direction, test.sortBy), func(t *testing.T) {
			t.Parallel()

			url, err := url.Parse(fmt.Sprintf("/?direction=%s&sortBy=%s", test.direction, test.sortBy))
			require.NoError(t, err)

			sortRouters(url.Query(), test.elements)

			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortServices(t *testing.T) {
	testCases := []struct {
		direction string
		sortBy    string
		elements  []orderedService
		expected  []orderedService
	}{
		{
			direction: ascendantSorting,
			sortBy:    "name",
			elements: []orderedService{
				serviceRepresentation{
					Name: "b",
				},
				serviceRepresentation{
					Name: "a",
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "a",
				},
				serviceRepresentation{
					Name: "b",
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "name",
			elements: []orderedService{
				serviceRepresentation{
					Name: "a",
				},
				serviceRepresentation{
					Name: "b",
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "b",
				},
				serviceRepresentation{
					Name: "a",
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "type",
			elements: []orderedService{
				serviceRepresentation{
					Name: "b",
					Type: "b",
				},
				serviceRepresentation{
					Name: "a",
					Type: "b",
				},
				serviceRepresentation{
					Name: "b",
					Type: "a",
				},
				serviceRepresentation{
					Name: "a",
					Type: "a",
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "a",
					Type: "a",
				},
				serviceRepresentation{
					Name: "b",
					Type: "a",
				},
				serviceRepresentation{
					Name: "a",
					Type: "b",
				},
				serviceRepresentation{
					Name: "b",
					Type: "b",
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "type",
			elements: []orderedService{
				serviceRepresentation{
					Name: "a",
					Type: "a",
				},
				serviceRepresentation{
					Name: "b",
					Type: "a",
				},
				serviceRepresentation{
					Name: "a",
					Type: "b",
				},
				serviceRepresentation{
					Name: "b",
					Type: "b",
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "b",
					Type: "b",
				},
				serviceRepresentation{
					Name: "a",
					Type: "b",
				},
				serviceRepresentation{
					Name: "b",
					Type: "a",
				},
				serviceRepresentation{
					Name: "a",
					Type: "a",
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "server",
			elements: []orderedService{
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 2),
							},
						},
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 2),
							},
						},
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 1),
							},
						},
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 1),
							},
						},
					},
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 2),
							},
						},
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 2),
							},
						},
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 2),
							},
						},
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: make([]dynamic.Server, 2),
							},
						},
					},
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "type",
			elements: []orderedService{
				serviceRepresentation{
					Name: "a",
					Type: "a",
				},
				serviceRepresentation{
					Name: "b",
					Type: "a",
				},
				serviceRepresentation{
					Name: "a",
					Type: "b",
				},
				serviceRepresentation{
					Name: "b",
					Type: "b",
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "b",
					Type: "b",
				},
				serviceRepresentation{
					Name: "a",
					Type: "b",
				},
				serviceRepresentation{
					Name: "b",
					Type: "a",
				},
				serviceRepresentation{
					Name: "a",
					Type: "a",
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "provider",
			elements: []orderedService{
				serviceRepresentation{
					Name:     "b",
					Provider: "b",
				},
				serviceRepresentation{
					Name:     "b",
					Provider: "a",
				},
				serviceRepresentation{
					Name:     "a",
					Provider: "b",
				},
				serviceRepresentation{
					Name:     "a",
					Provider: "a",
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name:     "a",
					Provider: "a",
				},
				serviceRepresentation{
					Name:     "b",
					Provider: "a",
				},
				serviceRepresentation{
					Name:     "a",
					Provider: "b",
				},
				serviceRepresentation{
					Name:     "b",
					Provider: "b",
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "provider",
			elements: []orderedService{
				serviceRepresentation{
					Name:     "a",
					Provider: "a",
				},
				serviceRepresentation{
					Name:     "a",
					Provider: "b",
				},
				serviceRepresentation{
					Name:     "b",
					Provider: "a",
				},
				serviceRepresentation{
					Name:     "b",
					Provider: "b",
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name:     "b",
					Provider: "b",
				},
				serviceRepresentation{
					Name:     "a",
					Provider: "b",
				},
				serviceRepresentation{
					Name:     "b",
					Provider: "a",
				},
				serviceRepresentation{
					Name:     "a",
					Provider: "a",
				},
			},
		},
		{
			direction: ascendantSorting,
			sortBy:    "status",
			elements: []orderedService{
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
			},
		},
		{
			direction: descendantSorting,
			sortBy:    "status",
			elements: []orderedService{
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
			},
			expected: []orderedService{
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "b",
					},
				},
				serviceRepresentation{
					Name: "b",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
				serviceRepresentation{
					Name: "a",
					ServiceInfo: &runtime.ServiceInfo{
						Status: "a",
					},
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(fmt.Sprintf("%s-%s", test.direction, test.sortBy), func(t *testing.T) {
			t.Parallel()

			url, err := url.Parse(fmt.Sprintf("/?direction=%s&sortBy=%s", test.direction, test.sortBy))
			require.NoError(t, err)

			sortServices(url.Query(), test.elements)

			assert.Equal(t, test.expected, test.elements)
		})
	}
}
