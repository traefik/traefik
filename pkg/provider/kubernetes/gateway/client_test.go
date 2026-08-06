package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatefake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"
)

func TestUpdateGatewayClassStatusSupportedFeatures(t *testing.T) {
	acceptedCondition := metav1.Condition{
		Type:   string(gatev1.GatewayClassConditionStatusAccepted),
		Status: metav1.ConditionTrue,
		Reason: "Handled",
	}

	gatewayClass := &gatev1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{Name: "traefik"},
		Status: gatev1.GatewayClassStatus{
			Conditions:        []metav1.Condition{acceptedCondition},
			SupportedFeatures: []gatev1.SupportedFeature{{Name: "TCPRoute"}},
		},
	}

	gatewayClient := gatefake.NewSimpleClientset(gatewayClass)
	client := newClientImpl(kubefake.NewClientset(), gatewayClient)
	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })

	_, err := client.WatchAll(nil, stopCh)
	require.NoError(t, err)

	desiredStatus := gatev1.GatewayClassStatus{
		Conditions:        []metav1.Condition{acceptedCondition},
		SupportedFeatures: []gatev1.SupportedFeature{{Name: "TCPRoute"}, {Name: "UDPRoute"}},
	}
	require.NoError(t, client.UpdateGatewayClassStatus(t.Context(), gatewayClass.Name, desiredStatus))

	updatedGatewayClass, err := gatewayClient.GatewayV1().GatewayClasses().Get(t.Context(), gatewayClass.Name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, desiredStatus, updatedGatewayClass.Status)
}

func Test_gatewayStatusEquals(t *testing.T) {
	testCases := []struct {
		desc     string
		statusA  gatev1.GatewayStatus
		statusB  gatev1.GatewayStatus
		expected bool
	}{
		{
			desc:     "Empty",
			statusA:  gatev1.GatewayStatus{},
			statusB:  gatev1.GatewayStatus{},
			expected: true,
		},
		{
			desc: "Same status",
			statusA: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "foobar",
						Reason: "foobar",
					},
				},
				Listeners: []gatev1.ListenerStatus{
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
			statusB: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "foobar",
						Reason: "foobar",
					},
				},
				Listeners: []gatev1.ListenerStatus{
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
			statusA: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{},
			},
			statusB: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
					{},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway conditions length not equal",
			statusA: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{},
			},
			statusB: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{
					{},
				},
			},
			expected: false,
		},
		{
			desc: "Gateway conditions different types",
			statusA: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type: "foobar",
					},
				},
			},
			statusB: gatev1.GatewayStatus{
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
			statusA: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{
					{
						Type: "foobar",
					},
				},
			},
			statusB: gatev1.GatewayStatus{
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
			statusA: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
					{
						Name:       "foo",
						Conditions: []metav1.Condition{},
					},
				},
			},
			statusB: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
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
			statusA: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
					{
						Conditions: []metav1.Condition{
							{
								Type: "foobar",
							},
						},
					},
				},
			},
			statusB: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
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
			statusA: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
					{
						Conditions: []metav1.Condition{
							{
								Type: "foobar",
							},
						},
					},
				},
			},
			statusB: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
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
			statusA: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
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
			statusB: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
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
		{
			desc: "Gateway listeners with same conditions but different number of attached routes",
			statusA: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
					{
						Name:           "foo",
						AttachedRoutes: 1,
						Conditions: []metav1.Condition{
							{
								Type:   "foobar",
								Reason: "foobar",
							},
						},
					},
				},
			},
			statusB: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{
					{
						Name:           "foo",
						AttachedRoutes: 2,
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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := gatewayStatusEqual(test.statusA, test.statusB)

			assert.Equal(t, test.expected, result)
		})
	}
}
