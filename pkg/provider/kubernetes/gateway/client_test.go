package gateway

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestWatchAll_gatewayAPIInformers(t *testing.T) {
	client := newClientImpl(kubefake.NewClientset(), newGatewaySimpleClientSet(t))

	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })

	_, err := client.WatchAll(nil, stopCh)
	require.NoError(t, err)

	assert.Equal(t, map[reflect.Type]bool{
		reflect.TypeFor[*gatev1.GatewayClass](): true,
	}, client.factoryGatewayClass.WaitForCacheSync(stopCh))

	assert.Equal(t, map[reflect.Type]bool{
		reflect.TypeFor[*gatev1.Gateway]():          true,
		reflect.TypeFor[*gatev1.HTTPRoute]():        true,
		reflect.TypeFor[*gatev1.GRPCRoute]():        true,
		reflect.TypeFor[*gatev1.TCPRoute]():         true,
		reflect.TypeFor[*gatev1.TLSRoute]():         true,
		reflect.TypeFor[*gatev1.ReferenceGrant]():   true,
		reflect.TypeFor[*gatev1.BackendTLSPolicy](): true,
		reflect.TypeFor[*gatev1.ListenerSet]():      true,
	}, client.factoriesGateway[metav1.NamespaceAll].WaitForCacheSync(stopCh))
}

func Test_checkGatewayAPIResources(t *testing.T) {
	var servedResources []metav1.APIResource
	for _, resource := range gatewayAPIResources {
		servedResources = append(servedResources, metav1.APIResource{Name: resource.name})
	}

	errAccessReview := errors.New("access review error")

	testCases := []struct {
		desc         string
		resources    []metav1.APIResource
		namespaces   []string
		gatewayClass string
		denied       []string
		apiError     error
		expectedErr  error
	}{
		{
			desc:      "Resources served and allowed",
			resources: servedResources,
		},
		{
			desc:        "Gateway API v1 not served",
			expectedErr: errGatewayAPIResourceUnavailable,
		},
		{
			desc:        "Gateway API v1 served without resources",
			resources:   []metav1.APIResource{},
			expectedErr: errGatewayAPIResourceUnavailable,
		},
		{
			desc: "TCPRoute not served",
			resources: slices.DeleteFunc(slices.Clone(servedResources), func(r metav1.APIResource) bool {
				return r.Name == "tcproutes"
			}),
			expectedErr: errGatewayAPIResourceUnavailable,
		},
		{
			desc:        "List not allowed cluster-wide",
			resources:   servedResources,
			denied:      []string{"list/tcproutes/"},
			expectedErr: errGatewayAPIResourceUnavailable,
		},
		{
			desc:       "Cluster-wide permission covers watched namespaces",
			resources:  servedResources,
			namespaces: []string{"foo"},
			denied:     []string{"watch/listenersets/foo"},
		},
		{
			desc:       "Permission granted per watched namespace",
			resources:  servedResources,
			namespaces: []string{"foo", "bar"},
			denied:     []string{"watch/listenersets/"},
		},
		{
			desc:        "Watch not allowed in a watched namespace",
			resources:   servedResources,
			namespaces:  []string{"foo", "bar"},
			denied:      []string{"watch/listenersets/", "watch/listenersets/bar"},
			expectedErr: errGatewayAPIResourceUnavailable,
		},
		{
			desc:        "GatewayClass permission checked cluster-wide with watched namespaces",
			resources:   servedResources,
			namespaces:  []string{"foo"},
			denied:      []string{"watch/gatewayclasses/"},
			expectedErr: errGatewayAPIResourceUnavailable,
		},
		{
			desc:         "GatewayClass permission checked by name",
			resources:    servedResources,
			gatewayClass: "traefik",
			denied:       []string{"update/gatewayclasses/status:traefik/"},
			expectedErr:  errGatewayAPIResourceUnavailable,
		},
		{
			desc:        "Status update not allowed",
			resources:   servedResources,
			denied:      []string{"update/httproutes/status/"},
			expectedErr: errGatewayAPIResourceUnavailable,
		},
		{
			desc:      "ReferenceGrant status update not required",
			resources: servedResources,
			denied:    []string{"update/referencegrants/status/"},
		},
		{
			desc:        "Access review error",
			resources:   servedResources,
			apiError:    errAccessReview,
			expectedErr: errAccessReview,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			kubeClient := kubefake.NewClientset()

			fakeDiscovery, ok := kubeClient.Discovery().(*fakediscovery.FakeDiscovery)
			require.True(t, ok)
			if test.resources != nil {
				fakeDiscovery.Resources = []*metav1.APIResourceList{{GroupVersion: gatev1.GroupVersion.String(), APIResources: test.resources}}
			}

			kubeClient.PrependReactor("create", "selfsubjectaccessreviews", func(action ktesting.Action) (bool, runtime.Object, error) {
				if test.apiError != nil {
					return true, nil, test.apiError
				}

				review := action.(ktesting.CreateAction).GetObject().(*authorizationv1.SelfSubjectAccessReview)
				attrs := review.Spec.ResourceAttributes
				target := attrs.Resource
				if attrs.Subresource != "" {
					target += "/" + attrs.Subresource
				}
				if attrs.Name != "" {
					target += ":" + attrs.Name
				}
				review.Status.Allowed = !slices.Contains(test.denied, attrs.Verb+"/"+target+"/"+attrs.Namespace)

				return true, review, nil
			})

			client := newClientImpl(kubeClient, nil)
			client.gatewayClass = test.gatewayClass

			err := client.checkGatewayAPIResources(t.Context(), test.namespaces)
			assert.ErrorIs(t, err, test.expectedErr)
		})
	}
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
	staleResolvedRefs := metav1.Condition{Type: string(gatev1.GatewayConditionResolvedRefs), Status: metav1.ConditionFalse}
	staleInsecureFrontend := metav1.Condition{Type: string(gatev1.GatewayConditionInsecureFrontendValidationMode), Status: metav1.ConditionTrue}
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
			desc: "drops owned Gateway condition no longer desired",
			current: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{staleAccepted, staleResolvedRefs},
			},
			expected: desired,
		},
		{
			desc: "drops stale InsecureFrontendValidationMode condition",
			current: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{staleAccepted, staleInsecureFrontend},
			},
			expected: desired,
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
	insecureFrontend := metav1.Condition{Type: string(gatev1.GatewayConditionInsecureFrontendValidationMode), Status: metav1.ConditionTrue}
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
			desc: "keeps InsecureFrontendValidationMode condition",
			status: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{foreign, insecureFrontend, accepted},
			},
			expected: gatev1.GatewayStatus{
				Conditions: []metav1.Condition{insecureFrontend, accepted},
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
