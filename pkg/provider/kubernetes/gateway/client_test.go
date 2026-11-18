package gateway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authorizationv1 "k8s.io/api/authorization/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestWatchAll_clusterTrustBundleChecksError(t *testing.T) {
	forbiddenErr := kerror.NewForbidden(authorizationv1.Resource(""), "", errors.New("RBAC denied"))
	internalErr := kerror.NewInternalError(errors.New("internal error"))

	testCases := []struct {
		desc        string
		verb        string
		resource    string
		err         error
		expectedErr bool
	}{
		{
			desc:     "Discovery forbidden",
			verb:     "get",
			resource: "resource",
			err:      forbiddenErr,
		},
		{
			desc:        "Discovery error",
			verb:        "get",
			resource:    "resource",
			err:         internalErr,
			expectedErr: true,
		},
		{
			desc:     "Access review forbidden",
			verb:     "create",
			resource: "selfsubjectaccessreviews",
			err:      forbiddenErr,
		},
		{
			desc:        "Access review error",
			verb:        "create",
			resource:    "selfsubjectaccessreviews",
			err:         internalErr,
			expectedErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			csKube := kubefake.NewClientset()
			csKube.Resources = []*metav1.APIResourceList{{
				GroupVersion: certificatesv1.SchemeGroupVersion.String(),
				APIResources: []metav1.APIResource{{Name: "clustertrustbundles"}},
			}}
			// The reactors are evaluated from the last prepended one, so the tested error comes first.
			csKube.PrependReactor("create", "selfsubjectaccessreviews", func(ktesting.Action) (bool, runtime.Object, error) {
				return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
			})
			csKube.PrependReactor(test.verb, test.resource, func(ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.err
			})

			client := newClientImpl(csKube, newGatewaySimpleClientSet(t))

			stopCh := make(chan struct{})
			t.Cleanup(func() { close(stopCh) })

			_, err := client.WatchAll(nil, stopCh)
			if test.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.True(t, client.clusterTrustBundleAPIServed)
			assert.True(t, client.clusterTrustBundleAllowed)
			assert.Contains(t, client.factoryNamespace.WaitForCacheSync(stopCh), reflect.TypeFor[*certificatesv1.ClusterTrustBundle]())
		})
	}
}

func Test_newWatchErrorHandler(t *testing.T) {
	testCases := []struct {
		desc        string
		err         error
		expectedLog string
	}{
		{
			desc:        "Resource expired",
			err:         kerror.NewResourceExpired("too old resource version"),
			expectedLog: `"level":"debug","error":"too old resource version","message":"Watch of *v1.TCPRoute closed"`,
		},
		{
			desc:        "Resource gone",
			err:         kerror.NewGone("gone"),
			expectedLog: `"level":"debug","error":"gone","message":"Watch of *v1.TCPRoute closed"`,
		},
		{
			desc: "Watch closed",
			err:  io.EOF,
		},
		{
			desc:        "Watch closed with unexpected EOF",
			err:         io.ErrUnexpectedEOF,
			expectedLog: `"level":"debug","error":"unexpected EOF","message":"Watch of *v1.TCPRoute closed with unexpected EOF"`,
		},
		{
			desc:        "Resource not served",
			err:         fmt.Errorf("failed to list *v1.TCPRoute: %w", kerror.NewNotFound(gatev1.Resource("tcproutes"), "")),
			expectedLog: `"level":"error",` + `"error":"failed to list *v1.TCPRoute: tcproutes.gateway.networking.k8s.io \"\" not found","message":"Cannot list or watch *v1.TCPRoute, the resource is not served: install the Gateway API CRDs v1.7.0 or later"`,
		},
		{
			desc:        "Access forbidden",
			err:         kerror.NewForbidden(gatev1.Resource("tcproutes"), "", errors.New("RBAC denied")),
			expectedLog: `"level":"error",` + `"error":"tcproutes.gateway.networking.k8s.io is forbidden: RBAC denied","message":"Cannot list or watch *v1.TCPRoute, the access is forbidden: update the Traefik RBAC for the Kubernetes Gateway API provider"`,
		},
		{
			desc:        "Other error",
			err:         errors.New("connection refused"),
			expectedLog: `"level":"error","error":"connection refused","message":"Cannot list or watch *v1.TCPRoute"`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			reflector := cache.NewReflector(&cache.ListWatch{}, &gatev1.TCPRoute{}, cache.NewStore(cache.MetaNamespaceKeyFunc), 0)

			newWatchErrorHandler(zerolog.New(&buf))(context.Background(), reflector, test.err)

			if test.expectedLog == "" {
				assert.Empty(t, buf.String())
				return
			}

			assert.Contains(t, buf.String(), test.expectedLog)
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
