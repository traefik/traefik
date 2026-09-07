package udp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeHandler struct {
	name string
}

func (f fakeHandler) ServeUDP(conn *Conn) {}

func TestLoadBalancing(t *testing.T) {
	testCases := []struct {
		desc                 string
		serversWeight        map[string]int
		totalCall            int
		expectedServe        map[string]int
		expectedErrorMessage string
	}{
		{
			desc:                 "no server in the pool",
			totalCall:            1,
			expectedServe:        map[string]int{},
			expectedErrorMessage: "no servers in the pool",
		},
		{
			desc: "RoundRobin",
			serversWeight: map[string]int{
				"h1": 1,
				"h2": 1,
			},
			totalCall: 4,
			expectedServe: map[string]int{
				"h1": 2,
				"h2": 2,
			},
		},
		{
			desc: "WeighedRoundRobin",
			serversWeight: map[string]int{
				"h1": 3,
				"h2": 1,
			},
			totalCall: 16,
			expectedServe: map[string]int{
				"h1": 12,
				"h2": 4,
			},
		},
		{
			desc: "WeighedRoundRobin with one 0 weight server",
			serversWeight: map[string]int{
				"h1": 3,
				"h2": 0,
			},
			totalCall: 16,
			expectedServe: map[string]int{
				"h1": 16,
			},
		},
		{
			desc: "WeighedRoundRobin with all servers with 0 weight",
			serversWeight: map[string]int{
				"h1": 0,
				"h2": 0,
				"h3": 0,
			},
			totalCall:            10,
			expectedServe:        map[string]int{},
			expectedErrorMessage: "no server with a positive weight",
		},
		{
			desc: "WeighedRoundRobin with all servers with a negative weight",
			serversWeight: map[string]int{
				"h1": -2,
				"h2": -2,
			},
			totalCall:            10,
			expectedServe:        map[string]int{},
			expectedErrorMessage: "no server with a positive weight",
		},
		{
			desc: "WeighedRoundRobin with one negative weight server",
			serversWeight: map[string]int{
				"h1": 3,
				"h2": -2,
			},
			totalCall: 16,
			expectedServe: map[string]int{
				"h1": 16,
			},
		},
		{
			desc: "WeighedRoundRobin with a negative weight greater in magnitude than the positive one",
			serversWeight: map[string]int{
				"h1": 2,
				"h2": -3,
			},
			totalCall: 16,
			expectedServe: map[string]int{
				"h1": 16,
			},
		},
		{
			desc: "WeighedRoundRobin with a negative weight and a non unit gcd",
			serversWeight: map[string]int{
				"h1": 4,
				"h2": -6,
			},
			totalCall: 16,
			expectedServe: map[string]int{
				"h1": 16,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			balancer := NewWRRLoadBalancer()
			for name, weight := range test.serversWeight {
				balancer.AddWeightedServer(fakeHandler{name: name}, &weight)
			}

			served := map[string]int{}
			for range test.totalCall {
				next, err := balancer.next()
				if test.expectedErrorMessage != "" {
					require.EqualError(t, err, test.expectedErrorMessage)
					continue
				}

				require.NoError(t, err)
				served[next.(server).Handler.(fakeHandler).name]++
			}

			assert.Equal(t, test.expectedServe, served)
		})
	}
}
