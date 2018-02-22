package consulcatalog

import (
	"sort"
	"testing"

	"github.com/BurntSushi/ty/fun"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestNodeSorter(t *testing.T) {
	testCases := []struct {
		desc     string
		nodes    []*api.ServiceEntry
		expected []*api.ServiceEntry
	}{
		{
			desc:     "Should sort nothing",
			nodes:    []*api.ServiceEntry{},
			expected: []*api.ServiceEntry{},
		},
		{
			desc: "Should sort by node address",
			nodes: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
			},
			expected: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
			},
		},
		{
			desc: "Should sort by service name",
			nodes: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    81,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
			},
			expected: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    81,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
			},
		},
		{
			desc: "Should sort by node address",
			nodes: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
			},
			expected: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sort.Sort(nodeSorter(test.nodes))
			actual := test.nodes
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetChangedKeys(t *testing.T) {
	type Input struct {
		currState map[string]Service
		prevState map[string]Service
	}

	type Output struct {
		addedKeys   []string
		removedKeys []string
	}

	testCases := []struct {
		desc   string
		input  Input
		output Output
	}{
		{
			desc: "Should add 0 services and removed 0",
			input: Input{
				currState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
				prevState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
			},
			output: Output{
				addedKeys:   []string{},
				removedKeys: []string{},
			},
		},
		{
			desc: "Should add 3 services and removed 0",
			input: Input{
				currState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
				prevState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
			},
			output: Output{
				addedKeys:   []string{"qux-service", "quux-service", "quuz-service"},
				removedKeys: []string{},
			},
		},
		{
			desc: "Should add 2 services and removed 2",
			input: Input{
				currState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
				prevState: map[string]Service{
					"foo-service":   {Name: "v1"},
					"bar-service":   {Name: "v1"},
					"baz-service":   {Name: "v1"},
					"qux-service":   {Name: "v1"},
					"quux-service":  {Name: "v1"},
					"quuz-service":  {Name: "v1"},
					"corge-service": {Name: "v1"},
					"waldo-service": {Name: "v1"},
					"fred-service":  {Name: "v1"},
					"plugh-service": {Name: "v1"},
					"xyzzy-service": {Name: "v1"},
					"thud-service":  {Name: "v1"},
				},
			},
			output: Output{
				addedKeys:   []string{"grault-service", "garply-service"},
				removedKeys: []string{"bar-service", "baz-service"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			addedKeys, removedKeys := getChangedServiceKeys(test.input.currState, test.input.prevState)
			assert.Equal(t, fun.Set(test.output.addedKeys), fun.Set(addedKeys), "Added keys comparison results: got %q, want %q", addedKeys, test.output.addedKeys)
			assert.Equal(t, fun.Set(test.output.removedKeys), fun.Set(removedKeys), "Removed keys comparison results: got %q, want %q", removedKeys, test.output.removedKeys)
		})
	}
}

func TestFilterEnabled(t *testing.T) {
	testCases := []struct {
		desc             string
		exposedByDefault bool
		node             *api.ServiceEntry
		expected         bool
	}{
		{
			desc:             "exposed",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{""},
				},
			},
			expected: true,
		},
		{
			desc:             "exposed and tolerated by valid label value",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=true"},
				},
			},
			expected: true,
		},
		{
			desc:             "exposed and tolerated by invalid label value",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=bad"},
				},
			},
			expected: true,
		},
		{
			desc:             "exposed but overridden by label",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=false"},
				},
			},
			expected: false,
		},
		{
			desc:             "non-exposed",
			exposedByDefault: false,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{""},
				},
			},
			expected: false,
		},
		{
			desc:             "non-exposed but overridden by label",
			exposedByDefault: false,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=true"},
				},
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{
				Domain:           "localhost",
				Prefix:           "traefik",
				ExposedByDefault: test.exposedByDefault,
			}
			actual := provider.nodeFilter("test", test.node)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetChangedStringKeys(t *testing.T) {
	testCases := []struct {
		desc            string
		current         []string
		previous        []string
		expectedAdded   []string
		expectedRemoved []string
	}{
		{
			desc:            "1 element added, 0 removed",
			current:         []string{"chou"},
			previous:        []string{},
			expectedAdded:   []string{"chou"},
			expectedRemoved: []string{},
		}, {
			desc:            "0 element added, 0 removed",
			current:         []string{"chou"},
			previous:        []string{"chou"},
			expectedAdded:   []string{},
			expectedRemoved: []string{},
		},
		{
			desc:            "0 element added, 1 removed",
			current:         []string{},
			previous:        []string{"chou"},
			expectedAdded:   []string{},
			expectedRemoved: []string{"chou"},
		},
		{
			desc:            "1 element added, 1 removed",
			current:         []string{"carotte"},
			previous:        []string{"chou"},
			expectedAdded:   []string{"carotte"},
			expectedRemoved: []string{"chou"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actualAdded, actualRemoved := getChangedStringKeys(test.current, test.previous)
			assert.Equal(t, test.expectedAdded, actualAdded)
			assert.Equal(t, test.expectedRemoved, actualRemoved)
		})
	}
}

func TestHasServiceChanged(t *testing.T) {
	testCases := []struct {
		desc     string
		current  map[string]Service
		previous map[string]Service
		expected bool
	}{
		{
			desc: "Change detected due to change of nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node2"},
					Tags:  []string{},
				},
			},
			expected: true,
		},
		{
			desc:    "No change missing current service",
			current: make(map[string]Service),
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: false,
		},
		{
			desc: "No change on nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: false,
		},
		{
			desc: "No change on nodes and tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			expected: false,
		},
		{
			desc: "Change detected on tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
				},
			},
			expected: true,
		},
		{
			desc: "Change detected on ports",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
					Ports: []int{80},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
					Ports: []int{81},
				},
			},
			expected: true,
		},
		{
			desc: "Change detected on ports",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
					Ports: []int{80},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
					Ports: []int{81, 82},
				},
			},
			expected: true,
		},
		{
			desc: "Change detected on addresses",
			current: map[string]Service{
				"foo-service": {
					Name:      "foo",
					Nodes:     []string{"node1"},
					Addresses: []string{"127.0.0.1"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:      "foo",
					Nodes:     []string{"node1"},
					Addresses: []string{"127.0.0.2"},
				},
			},
			expected: true,
		},
		{
			desc: "No Change detected",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
					Ports: []int{80},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
					Ports: []int{80},
				},
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasServiceChanged(test.current, test.previous)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasChanged(t *testing.T) {
	testCases := []struct {
		desc     string
		current  map[string]Service
		previous map[string]Service
		expected bool
	}{
		{
			desc: "Change detected due to change new service",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: make(map[string]Service),
			expected: true,
		},
		{
			desc:    "Change detected due to change service removed",
			current: make(map[string]Service),
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: true,
		},
		{
			desc: "Change detected due to change of nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node2"},
					Tags:  []string{},
				},
			},
			expected: true,
		},
		{
			desc: "No change on nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: false,
		},
		{
			desc: "No change on nodes and tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			expected: false,
		},
		{
			desc: "Change detected on tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
				},
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasChanged(test.current, test.previous)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetConstraintTags(t *testing.T) {
	provider := &Provider{
		Domain: "localhost",
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected []string
	}{
		{
			desc: "nil tags",
		},
		{
			desc:     "invalid tag",
			tags:     []string{"tags=foobar"},
			expected: nil,
		},
		{
			desc:     "wrong tag",
			tags:     []string{"traefik_tags=foobar"},
			expected: nil,
		},
		{
			desc:     "empty value",
			tags:     []string{"traefik.tags="},
			expected: nil,
		},
		{
			desc:     "simple tag",
			tags:     []string{"traefik.tags=foobar "},
			expected: []string{"foobar"},
		},
		{
			desc:     "multiple values tag",
			tags:     []string{"traefik.tags=foobar, fiibir"},
			expected: []string{"foobar", "fiibir"},
		},
		{
			desc:     "multiple tags",
			tags:     []string{"traefik.tags=foobar", "traefik.tags=foobor"},
			expected: []string{"foobar", "foobor"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			constraints := provider.getConstraintTags(test.tags)
			assert.EqualValues(t, test.expected, constraints)
		})
	}
}
