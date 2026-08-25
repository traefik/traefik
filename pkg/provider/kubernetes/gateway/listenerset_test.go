package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_hasProtocolConflict(t *testing.T) {
	testCases := []struct {
		desc      string
		allocated map[gatev1.ProtocolType]struct{}
		protocol  gatev1.ProtocolType
		expected  bool
	}{
		{
			desc:     "empty port has no conflict",
			protocol: gatev1.HTTPProtocolType,
			expected: false,
		},
		{
			desc:      "same protocol as the one allocated",
			allocated: map[gatev1.ProtocolType]struct{}{gatev1.HTTPSProtocolType: {}},
			protocol:  gatev1.HTTPSProtocolType,
			expected:  false,
		},
		{
			// Traefik maps a port to a single entryPoint, so a TLS listener cannot join
			// a port an HTTPS listener already claimed.
			desc:      "TLS conflicts with an allocated HTTPS",
			allocated: map[gatev1.ProtocolType]struct{}{gatev1.HTTPSProtocolType: {}},
			protocol:  gatev1.TLSProtocolType,
			expected:  true,
		},
		{
			desc:      "HTTPS conflicts with an allocated TLS",
			allocated: map[gatev1.ProtocolType]struct{}{gatev1.TLSProtocolType: {}},
			protocol:  gatev1.HTTPSProtocolType,
			expected:  true,
		},
		{
			desc:      "TCP conflicts with an allocated HTTP",
			allocated: map[gatev1.ProtocolType]struct{}{gatev1.HTTPProtocolType: {}},
			protocol:  gatev1.TCPProtocolType,
			expected:  true,
		},
		{
			desc:      "conflicts with one of several allocated protocols",
			allocated: map[gatev1.ProtocolType]struct{}{gatev1.TLSProtocolType: {}, gatev1.HTTPProtocolType: {}},
			protocol:  gatev1.TLSProtocolType,
			expected:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, hasProtocolConflict(test.allocated, test.protocol))
		})
	}
}

func Test_dedupeConditionsByType(t *testing.T) {
	testCases := []struct {
		desc       string
		conditions []metav1.Condition
		expected   []metav1.Condition
	}{
		{
			desc:     "empty",
			expected: []metav1.Condition{},
		},
		{
			desc: "no duplicates preserves order",
			conditions: []metav1.Condition{
				{Type: "Accepted", Reason: "a"},
				{Type: "Programmed", Reason: "b"},
			},
			expected: []metav1.Condition{
				{Type: "Accepted", Reason: "a"},
				{Type: "Programmed", Reason: "b"},
			},
		},
		{
			desc: "keeps the first condition of each type",
			conditions: []metav1.Condition{
				{Type: "ResolvedRefs", Reason: "first"},
				{Type: "Accepted", Reason: "a"},
				{Type: "ResolvedRefs", Reason: "second"},
			},
			expected: []metav1.Condition{
				{Type: "ResolvedRefs", Reason: "first"},
				{Type: "Accepted", Reason: "a"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, dedupeConditionsByType(test.conditions))
		})
	}
}
