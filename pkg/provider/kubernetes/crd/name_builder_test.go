package crd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeSafeKey(t *testing.T) {
	testCases := []struct {
		desc       string
		components []string
		expected   string
	}{
		{
			desc:       "no component",
			components: nil,
			expected:   "",
		},
		{
			desc:       "single component",
			components: []string{"default"},
			expected:   "default",
		},
		{
			desc:       "several components",
			components: []string{"default", "test.route", "0", "lb"},
			expected:   "default|test.route|0|lb",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, makeSafeKey(test.components...))
		})
	}
}

func TestMakeFlattenedKey(t *testing.T) {
	testCases := []struct {
		desc      string
		namespace string
		name      string
		expected  string
	}{
		{
			desc:      "namespace and name",
			namespace: "default",
			name:      "whoami",
			expected:  "default-whoami",
		},
		{
			desc:      "empty namespace",
			namespace: "",
			name:      "whoami",
			expected:  "whoami",
		},
		{
			desc:      "name is not normalized",
			namespace: "default",
			name:      "test.route",
			expected:  "default-test.route",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, makeFlattenedKey(test.namespace, test.name))
		})
	}
}

func TestMakeServiceKey(t *testing.T) {
	t.Run("deterministic for the same rule and ingress name", func(t *testing.T) {
		t.Parallel()

		key1, err := makeServiceKey("Host(`foo.com`)", "test-route")
		require.NoError(t, err)

		key2, err := makeServiceKey("Host(`foo.com`)", "test-route")
		require.NoError(t, err)

		assert.Equal(t, key1, key2)
		assert.True(t, strings.HasPrefix(key1, "test-route-"))
	})

	t.Run("differs when the rule differs", func(t *testing.T) {
		t.Parallel()

		key1, err := makeServiceKey("Host(`foo.com`)", "test-route")
		require.NoError(t, err)

		key2, err := makeServiceKey("Host(`bar.com`)", "test-route")
		require.NoError(t, err)

		assert.NotEqual(t, key1, key2)
	})

	t.Run("differs when the ingress name differs", func(t *testing.T) {
		t.Parallel()

		key1, err := makeServiceKey("Host(`foo.com`)", "test-route")
		require.NoError(t, err)

		key2, err := makeServiceKey("Host(`foo.com`)", "other-route")
		require.NoError(t, err)

		assert.NotEqual(t, key1, key2)
	})
}

func TestNameBuilder_makeID(t *testing.T) {
	testCases := []struct {
		desc      string
		safe      bool
		namespace string
		name      string
		expected  string
	}{
		{
			desc:      "legacy naming flattens and normalizes",
			safe:      false,
			namespace: "default",
			name:      "test.route",
			expected:  "default-test-route",
		},
		{
			desc:      "legacy naming, empty namespace",
			safe:      false,
			namespace: "",
			name:      "whoami",
			expected:  "whoami",
		},
		{
			desc:      "safe naming keeps the identity, unnormalized",
			safe:      true,
			namespace: "default",
			name:      "test.route",
			expected:  "default|test.route",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			n := nameBuilder{safe: test.safe}
			assert.Equal(t, test.expected, n.makeID(test.namespace, test.name))
		})
	}
}

func TestNameBuilder_makeRawID(t *testing.T) {
	testCases := []struct {
		desc      string
		safe      bool
		namespace string
		name      string
		expected  string
	}{
		{
			desc:      "legacy naming flattens without normalizing",
			safe:      false,
			namespace: "default",
			name:      "test.route",
			expected:  "default-test.route",
		},
		{
			desc:      "legacy naming, empty namespace",
			safe:      false,
			namespace: "",
			name:      "whoami",
			expected:  "whoami",
		},
		{
			desc:      "safe naming keeps the identity",
			safe:      true,
			namespace: "default",
			name:      "test.route",
			expected:  "default|test.route",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			n := nameBuilder{safe: test.safe}
			assert.Equal(t, test.expected, n.makeRawID(test.namespace, test.name))
		})
	}
}

func TestNameBuilder_httpRouter(t *testing.T) {
	testCases := []struct {
		desc        string
		safe        bool
		namespace   string
		ingressName string
		routeIndex  int
		rule        string
		expected    string
	}{
		{
			desc:        "legacy naming derives from the rule, and normalizes",
			safe:        false,
			namespace:   "default",
			ingressName: "test.route",
			routeIndex:  0,
			rule:        "Host(`foo.com`)",
			expected:    "default-test-route-6f97418635c7e18853da",
		},
		{
			desc:        "safe naming derives from the route index, unnormalized",
			safe:        true,
			namespace:   "default",
			ingressName: "test.route",
			routeIndex:  0,
			rule:        "Host(`foo.com`)",
			expected:    "default|test.route|0",
		},
		{
			desc:        "safe naming ignores the rule entirely",
			safe:        true,
			namespace:   "default",
			ingressName: "test.route",
			routeIndex:  0,
			rule:        "Host(`bar.com`)",
			expected:    "default|test.route|0",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			n := nameBuilder{safe: test.safe}
			got, err := n.httpRouter(test.namespace, test.ingressName, test.routeIndex, test.rule)
			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestNameBuilder_tcpRouter(t *testing.T) {
	testCases := []struct {
		desc        string
		safe        bool
		namespace   string
		ingressName string
		routeIndex  int
		rule        string
		expected    string
	}{
		{
			desc:        "legacy naming derives from the rule, and is not normalized",
			safe:        false,
			namespace:   "default",
			ingressName: "test.route",
			routeIndex:  0,
			rule:        "HostSNI(`foo.com`)",
			expected:    "default-test.route-fdd3e9338e47a45efefc",
		},
		{
			desc:        "safe naming derives from the route index, unnormalized",
			safe:        true,
			namespace:   "default",
			ingressName: "test.route",
			routeIndex:  0,
			rule:        "HostSNI(`foo.com`)",
			expected:    "default|test.route|0",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			n := nameBuilder{safe: test.safe}
			got, err := n.tcpRouter(test.namespace, test.ingressName, test.routeIndex, test.rule)
			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestNameBuilder_udpRouter(t *testing.T) {
	testCases := []struct {
		desc        string
		safe        bool
		namespace   string
		ingressName string
		routeIndex  int
		expected    string
	}{
		{
			desc:        "legacy naming derives from the route index, and is not normalized",
			safe:        false,
			namespace:   "default",
			ingressName: "test.route",
			routeIndex:  0,
			expected:    "default-test.route-0",
		},
		{
			desc:        "safe naming derives from the route index, unnormalized",
			safe:        true,
			namespace:   "default",
			ingressName: "test.route",
			routeIndex:  0,
			expected:    "default|test.route|0",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			n := nameBuilder{safe: test.safe}
			assert.Equal(t, test.expected, n.udpRouter(test.namespace, test.ingressName, test.routeIndex))
		})
	}
}
