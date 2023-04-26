package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestStatusEquals(t *testing.T) {
	testCases := []struct {
		desc     string
		statusA  gatev1alpha2.GatewayStatus
		statusB  gatev1alpha2.GatewayStatus
		expected bool
	}{
		{
			desc:     "Empty",
			statusA:  gatev1alpha2.GatewayStatus{},
			statusB:  gatev1alpha2.GatewayStatus{},
			expected: true,
		},
		{
			desc: "Same status",
			statusA: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "foobar",
						Reason: "foobar",
					},
				},
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Name: "foo",
						Conditions: []metav1.Condition{
							{
								Type:   "foobar",
								Reason: "foobar",
							},
						},
					},
				},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "foobar",
						Reason: "foobar",
					},
				},
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Name: "foo",
						Conditions: []metav1.Condition{
							{
								Type:   "foobar",
								Reason: "foobar",
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			desc: "Listeners length not equal",
			statusA: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway conditions length not equal",
			statusA: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{
					{},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway conditions different types",
			statusA: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type: "foobar",
					},
				},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type: "foobir",
					},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway conditions same types but different reason",
			statusA: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type: "foobar",
					},
				},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "foobar",
						Reason: "Another reason",
					},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway listeners conditions length",
			statusA: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Name:       "foo",
						Conditions: []metav1.Condition{},
					},
				},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Name: "foo",
						Conditions: []metav1.Condition{
							{},
						},
					},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway listeners conditions same types but different status",
			statusA: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Conditions: []metav1.Condition{
							{
								Type: "foobar",
							},
						},
					},
				},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Conditions: []metav1.Condition{
							{
								Type:   "foobar",
								Status: "Another status",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway listeners conditions same types but different message",
			statusA: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Conditions: []metav1.Condition{
							{
								Type: "foobar",
							},
						},
					},
				},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Conditions: []metav1.Condition{
							{
								Type:    "foobar",
								Message: "Another status",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway listeners conditions same types/reason but different names",
			statusA: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Name: "foo",
						Conditions: []metav1.Condition{
							{
								Type:   "foobar",
								Reason: "foobar",
							},
						},
					},
				},
			},
			statusB: gatev1alpha2.GatewayStatus{
				Listeners: []gatev1alpha2.ListenerStatus{
					{
						Name: "bar",
						Conditions: []metav1.Condition{
							{
								Type:   "foobar",
								Reason: "foobar",
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := statusEquals(test.statusA, test.statusB)

			assert.Equal(t, test.expected, result)
		})
	}
}
