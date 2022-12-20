package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
)

func TestSortByName(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByName
		expected  []orderedByName
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByName{
				routerRepresentation{
					Name: "b",
				},
				routerRepresentation{
					Name: "a",
				},
			},
			expected: []orderedByName{
				routerRepresentation{
					Name: "a",
				},
				routerRepresentation{
					Name: "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByName{
				routerRepresentation{
					Name: "a",
				},
				routerRepresentation{
					Name: "b",
				},
			},
			expected: []orderedByName{
				routerRepresentation{
					Name: "b",
				},
				routerRepresentation{
					Name: "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByName(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByType(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByType
		expected  []orderedByType
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByType{
				middlewareRepresentation{
					Type: "b",
				},
				middlewareRepresentation{
					Type: "a",
				},
			},
			expected: []orderedByType{
				middlewareRepresentation{
					Type: "a",
				},
				middlewareRepresentation{
					Type: "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByType{
				middlewareRepresentation{
					Type: "a",
				},
				middlewareRepresentation{
					Type: "b",
				},
			},
			expected: []orderedByType{
				middlewareRepresentation{
					Type: "b",
				},
				middlewareRepresentation{
					Type: "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByType(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByPriority(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByPriority
		expected  []orderedByPriority
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByPriority{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 2,
						},
					},
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 1,
						},
					},
				},
			},
			expected: []orderedByPriority{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 1,
						},
					},
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 2,
						},
					},
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByPriority{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 1,
						},
					},
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 2,
						},
					},
				},
			},
			expected: []orderedByPriority{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 2,
						},
					},
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Priority: 1,
						},
					},
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByPriority(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByStatus(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByStatus
		expected  []orderedByStatus
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByStatus{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "a",
				},
			},
			expected: []orderedByStatus{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByStatus{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "b",
				},
			},
			expected: []orderedByStatus{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "b",
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Status: "a",
					},
					Name: "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByStatus(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByRule(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByRule
		expected  []orderedByRule
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByRule{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "a",
				},
			},
			expected: []orderedByRule{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByRule{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "b",
				},
			},
			expected: []orderedByRule{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Rule: "a",
						},
					},
					Name: "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByRule(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByProvider(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByProvider
		expected  []orderedByProvider
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByProvider{
				routerRepresentation{
					Provider: "b",
					Name:     "b",
				},
				routerRepresentation{
					Provider: "b",
					Name:     "a",
				},
				routerRepresentation{
					Provider: "a",
					Name:     "b",
				},
				routerRepresentation{
					Provider: "a",
					Name:     "a",
				},
			},
			expected: []orderedByProvider{
				routerRepresentation{
					Provider: "a",
					Name:     "a",
				},
				routerRepresentation{
					Provider: "a",
					Name:     "b",
				},
				routerRepresentation{
					Provider: "b",
					Name:     "a",
				},
				routerRepresentation{
					Provider: "b",
					Name:     "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByProvider{
				routerRepresentation{
					Provider: "a",
					Name:     "a",
				},
				routerRepresentation{
					Provider: "a",
					Name:     "b",
				},
				routerRepresentation{
					Provider: "b",
					Name:     "a",
				},
				routerRepresentation{
					Provider: "b",
					Name:     "b",
				},
			},
			expected: []orderedByProvider{
				routerRepresentation{
					Provider: "b",
					Name:     "b",
				},
				routerRepresentation{
					Provider: "b",
					Name:     "a",
				},
				routerRepresentation{
					Provider: "a",
					Name:     "b",
				},
				routerRepresentation{
					Provider: "a",
					Name:     "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByProvider(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByServers(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByServers
		expected  []orderedByServers
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByServers{
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "b",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "a",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "b",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "a",
				},
			},
			expected: []orderedByServers{
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "a",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "b",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "a",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByServers{
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "a",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "b",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "a",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "b",
				},
			},
			expected: []orderedByServers{
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "b",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
									{},
								},
							},
						},
					},
					Name: "a",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "b",
				},
				tcpServiceRepresentation{
					TCPServiceInfo: &runtime.TCPServiceInfo{
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{},
								},
							},
						},
					},
					Name: "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByServers(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByEntryPoints(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByEntryPoints
		expected  []orderedByEntryPoints
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByEntryPoints{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "a",
				},
			},
			expected: []orderedByEntryPoints{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByEntryPoints{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "b",
				},
			},
			expected: []orderedByEntryPoints{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a", "b"},
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							EntryPoints: []string{"a"},
						},
					},
					Name: "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByEntryPoints(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}

func TestSortByService(t *testing.T) {
	testCases := []struct {
		desc      string
		direction string
		elements  []orderedByService
		expected  []orderedByService
	}{
		{
			desc:      "Ascending",
			direction: ascendantSorting,
			elements: []orderedByService{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "a",
				},
			},
			expected: []orderedByService{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "b",
				},
			},
		},
		{
			desc:      "Descending",
			direction: descendantSorting,
			elements: []orderedByService{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "b",
				},
			},
			expected: []orderedByService{
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "b",
						},
					},
					Name: "a",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "b",
				},
				routerRepresentation{
					RouterInfo: &runtime.RouterInfo{
						Router: &dynamic.Router{
							Service: "a",
						},
					},
					Name: "a",
				},
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sortByService(test.direction, test.elements)
			assert.Equal(t, test.expected, test.elements)
		})
	}
}
