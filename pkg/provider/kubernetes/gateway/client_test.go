package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

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
		{
			desc: "Gateway with different attached listener sets count",
			statusA: gatev1.GatewayStatus{
				AttachedListenerSets: new(int32(1)),
			},
			statusB: gatev1.GatewayStatus{
				AttachedListenerSets: new(int32(2)),
			},
			expected: false,
		},
		{
			desc: "Gateway with same attached listener sets count",
			statusA: gatev1.GatewayStatus{
				AttachedListenerSets: new(int32(1)),
			},
			statusB: gatev1.GatewayStatus{
				AttachedListenerSets: new(int32(1)),
			},
			expected: true,
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

func Test_listenerSetStatusEqual(t *testing.T) {
	testCases := []struct {
		desc     string
		statusA  gatev1.ListenerSetStatus
		statusB  gatev1.ListenerSetStatus
		expected bool
	}{
		{
			desc:     "Empty",
			statusA:  gatev1.ListenerSetStatus{},
			statusB:  gatev1.ListenerSetStatus{},
			expected: true,
		},
		{
			desc: "Same status",
			statusA: gatev1.ListenerSetStatus{
				Conditions: []metav1.Condition{{Type: "Accepted", Reason: "Accepted"}},
				Listeners: []gatev1.ListenerEntryStatus{{
					Name:           "foo",
					AttachedRoutes: 1,
					Conditions:     []metav1.Condition{{Type: "Programmed", Reason: "Programmed"}},
				}},
			},
			statusB: gatev1.ListenerSetStatus{
				Conditions: []metav1.Condition{{Type: "Accepted", Reason: "Accepted"}},
				Listeners: []gatev1.ListenerEntryStatus{{
					Name:           "foo",
					AttachedRoutes: 1,
					Conditions:     []metav1.Condition{{Type: "Programmed", Reason: "Programmed"}},
				}},
			},
			expected: true,
		},
		{
			desc: "Top-level conditions differ",
			statusA: gatev1.ListenerSetStatus{
				Conditions: []metav1.Condition{{Type: "Accepted", Reason: "Accepted"}},
			},
			statusB: gatev1.ListenerSetStatus{
				Conditions: []metav1.Condition{{Type: "Accepted", Reason: "NotAllowed"}},
			},
			expected: false,
		},
		{
			desc: "Listener entries length differ",
			statusA: gatev1.ListenerSetStatus{
				Listeners: []gatev1.ListenerEntryStatus{{Name: "foo"}},
			},
			statusB:  gatev1.ListenerSetStatus{},
			expected: false,
		},
		{
			desc: "Listener entry names differ",
			statusA: gatev1.ListenerSetStatus{
				Listeners: []gatev1.ListenerEntryStatus{{Name: "foo"}},
			},
			statusB: gatev1.ListenerSetStatus{
				Listeners: []gatev1.ListenerEntryStatus{{Name: "bar"}},
			},
			expected: false,
		},
		{
			desc: "Listener entry attached routes differ",
			statusA: gatev1.ListenerSetStatus{
				Listeners: []gatev1.ListenerEntryStatus{{Name: "foo", AttachedRoutes: 1}},
			},
			statusB: gatev1.ListenerSetStatus{
				Listeners: []gatev1.ListenerEntryStatus{{Name: "foo", AttachedRoutes: 2}},
			},
			expected: false,
		},
		{
			desc: "Listener entry conditions differ",
			statusA: gatev1.ListenerSetStatus{
				Listeners: []gatev1.ListenerEntryStatus{{
					Name:       "foo",
					Conditions: []metav1.Condition{{Type: "Programmed", Status: metav1.ConditionTrue}},
				}},
			},
			statusB: gatev1.ListenerSetStatus{
				Listeners: []gatev1.ListenerEntryStatus{{
					Name:       "foo",
					Conditions: []metav1.Condition{{Type: "Programmed", Status: metav1.ConditionFalse}},
				}},
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := listenerSetStatusEqual(test.statusA, test.statusB)

			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_mergeGatewayStatus(t *testing.T) {
	accepted := metav1.Condition{Type: string(gatev1.GatewayConditionAccepted), Status: metav1.ConditionTrue}
	staleAccepted := metav1.Condition{Type: string(gatev1.GatewayConditionAccepted), Status: metav1.ConditionFalse}
	listenerAccepted := metav1.Condition{Type: string(gatev1.ListenerConditionAccepted), Status: metav1.ConditionTrue}
	staleConflicted := metav1.Condition{Type: string(gatev1.ListenerConditionConflicted), Status: metav1.ConditionTrue}
	foreign := metav1.Condition{Type: "example.com/Foreign", Status: metav1.ConditionTrue}

	desired := gatev1.GatewayStatus{
		Conditions: []metav1.Condition{accepted},
		Listeners: []gatev1.ListenerStatus{{
			Name:       "http",
			Conditions: []metav1.Condition{listenerAccepted},
		}},
	}

	testCases := []struct {
		desc     string
		current  gatev1.GatewayStatus
		expected gatev1.GatewayStatus
	}{
		{
			desc:     "no current status",
			expected: desired,
		},
		{
			desc: "keeps foreign Gateway condition",
			current: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{staleAccepted, foreign},
			},
			expected: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{accepted, foreign},
				Listeners:  desired.Listeners,
			},
		},
		{
			desc: "keeps foreign listener condition",
			current: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{{
					Name:       "http",
					Conditions: []metav1.Condition{foreign},
				}},
			},
			expected: gatev1.GatewayStatus{
				Conditions: desired.Conditions,
				Listeners: []gatev1.ListenerStatus{{
					Name:       "http",
					Conditions: []metav1.Condition{listenerAccepted, foreign},
				}},
			},
		},
		{
			desc: "drops owned listener condition no longer desired",
			current: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{{
					Name:       "http",
					Conditions: []metav1.Condition{listenerAccepted, staleConflicted},
				}},
			},
			expected: desired,
		},
		{
			desc: "drops status of listener no longer desired",
			current: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{{
					Name:       "removed",
					Conditions: []metav1.Condition{foreign},
				}},
			},
			expected: desired,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := mergeGatewayStatus(test.current, desired)

			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_withoutForeignConditions(t *testing.T) {
	accepted := metav1.Condition{Type: string(gatev1.GatewayConditionAccepted), Status: metav1.ConditionTrue}
	listenerAccepted := metav1.Condition{Type: string(gatev1.ListenerConditionAccepted), Status: metav1.ConditionTrue}
	foreign := metav1.Condition{Type: "example.com/Foreign", Status: metav1.ConditionTrue}

	testCases := []struct {
		desc     string
		status   gatev1.GatewayStatus
		expected gatev1.GatewayStatus
	}{
		{
			desc: "keeps owned conditions",
			status: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{accepted},
				Listeners: []gatev1.ListenerStatus{{
					Name:       "http",
					Conditions: []metav1.Condition{listenerAccepted},
				}},
			},
			expected: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{accepted},
				Listeners: []gatev1.ListenerStatus{{
					Name:       "http",
					Conditions: []metav1.Condition{listenerAccepted},
				}},
			},
		},
		{
			desc: "drops foreign Gateway condition",
			status: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{foreign, accepted},
			},
			expected: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{accepted},
			},
		},
		{
			desc: "drops foreign listener condition",
			status: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{{
					Name:       "http",
					Conditions: []metav1.Condition{foreign, listenerAccepted},
				}},
			},
			expected: gatev1.GatewayStatus{
				Listeners: []gatev1.ListenerStatus{{
					Name:       "http",
					Conditions: []metav1.Condition{listenerAccepted},
				}},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := withoutForeignConditions(test.status)

			assert.Equal(t, test.expected, result)
		})
	}
}
