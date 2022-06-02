package consul

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaces(t *testing.T) {
	testCases := []struct {
		desc               string
		namespace          string
		namespaces         []string
		expectedNamespaces []string
	}{
		{
			desc:               "no defined namespaces",
			expectedNamespaces: []string{""},
		},
		{
			desc:               "deprecated: use of defined namespace",
			namespace:          "test-ns",
			expectedNamespaces: []string{"test-ns"},
		},
		{
			desc:               "use of 1 defined namespaces",
			namespaces:         []string{"test-ns"},
			expectedNamespaces: []string{"test-ns"},
		},
		{
			desc:               "use of multiple defined namespaces",
			namespaces:         []string{"test-ns1", "test-ns2", "test-ns3", "test-ns4"},
			expectedNamespaces: []string{"test-ns1", "test-ns2", "test-ns3", "test-ns4"},
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pb := &ProviderBuilder{
				Namespace:  test.namespace,
				Namespaces: test.namespaces,
			}

			assert.Equal(t, test.expectedNamespaces, extractNSFromProvider(pb.BuildProviders()))
		})
	}
}

func extractNSFromProvider(providers []*Provider) []string {
	res := make([]string, len(providers))
	for i, p := range providers {
		res[i] = p.namespace
	}
	return res
}
