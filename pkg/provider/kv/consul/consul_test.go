package consul

import "testing"

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
			ns := extractNSFromProvider(pb.BuildProviders())

			checkNS(t, ns, test.expectedNamespaces)
			checkNS(t, test.expectedNamespaces, ns)
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

func checkNS(t *testing.T, nsA, nsB []string) {
	t.Helper()

	for _, nA := range nsA {
		var nsFound bool
		for _, nB := range nsB {
			if nA == nB {
				nsFound = true
				break
			}
		}

		if !nsFound {
			t.Errorf("found nothing to handle %q namespace", nA)
		}
	}
}
