package crd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kschema "k8s.io/apimachinery/pkg/runtime/schema"
)

// assertStatusInvariants checks invariants that must hold between the produced
// configuration and the status tracker state. It is called from the existing
// table-driven tests in kubernetes_test.go.
func assertStatusInvariants(t *testing.T, conf *dynamic.Configuration, statuses configStatuses) {
	t.Helper()
	// If routers were produced, the corresponding resource type must have been visited.
	if len(conf.HTTP.Routers) > 0 {
		assert.NotEmpty(t, statuses.ingressRoutes.seen, "conf has HTTP routers but no IngressRoutes were visited")
	}
	if conf.TCP != nil && len(conf.TCP.Routers) > 0 {
		assert.NotEmpty(t, statuses.ingressRoutesTCP.seen, "conf has TCP routers but no IngressRouteTCPs were visited")
	}
	if conf.UDP != nil && len(conf.UDP.Routers) > 0 {
		assert.NotEmpty(t, statuses.ingressRoutesUDP.seen, "conf has UDP routers but no IngressRouteUDPs were visited")
	}
	// Every resource that has recorded errors must also be in the seen map.
	for key := range statuses.ingressRoutes.errors {
		assert.Truef(t, statuses.ingressRoutes.seen[key], "ingressroute %q has errors but was not visited", key)
	}
	for key := range statuses.ingressRoutesTCP.errors {
		assert.Truef(t, statuses.ingressRoutesTCP.seen[key], "ingressroutetcp %q has errors but was not visited", key)
	}
	for key := range statuses.ingressRoutesUDP.errors {
		assert.Truef(t, statuses.ingressRoutesUDP.seen[key], "ingressrouteudp %q has errors but was not visited", key)
	}
}

func TestBuildResourceCondition(t *testing.T) {
	testCases := []struct {
		desc            string
		errs            []error
		generation      int64
		expectedStatus  metav1.ConditionStatus
		expectedReason  string
		expectedMessage string
	}{
		{
			desc:            "no errors gives Valid=True",
			generation:      3,
			expectedStatus:  metav1.ConditionTrue,
			expectedReason:  "Processed",
			expectedMessage: "Resource processed successfully.",
		},
		{
			desc:            "single error gives Valid=False",
			errs:            []error{errors.New("service not found")},
			generation:      1,
			expectedStatus:  metav1.ConditionFalse,
			expectedReason:  "ProcessingError",
			expectedMessage: "service not found",
		},
		{
			desc:            "multiple errors are joined with semicolon",
			errs:            []error{errors.New("err one"), errors.New("err two")},
			generation:      2,
			expectedStatus:  metav1.ConditionFalse,
			expectedReason:  "ProcessingError",
			expectedMessage: "err one; err two",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			tracker := newStatusTracker()
			tracker.visit("default", "my-route")
			for _, err := range test.errs {
				tracker.addError("default", "my-route", err)
			}

			conds := buildResourceCondition(tracker, "default", "my-route", test.generation)

			require.Len(t, conds, 1)
			assert.Equal(t, "Valid", conds[0].Type)
			assert.Equal(t, test.expectedStatus, conds[0].Status)
			assert.Equal(t, test.expectedReason, conds[0].Reason)
			assert.Equal(t, test.expectedMessage, conds[0].Message)
			assert.Equal(t, test.generation, conds[0].ObservedGeneration)
		})
	}
}

// statusMockClient records UpdateXxx calls and returns configurable errors.
// All Get methods return whatever is set on the struct fields.
// By default all UpdateXxx methods succeed.
type statusMockClient struct {
	ingressRoutes        []*traefikv1alpha1.IngressRoute
	ingressRoutesTCP     []*traefikv1alpha1.IngressRouteTCP
	ingressRoutesUDP     []*traefikv1alpha1.IngressRouteUDP
	middlewares          []*traefikv1alpha1.Middleware
	middlewaresTCP       []*traefikv1alpha1.MiddlewareTCP
	serversTransports    []*traefikv1alpha1.ServersTransport
	serversTransportsTCP []*traefikv1alpha1.ServersTransportTCP
	tlsOptions           []*traefikv1alpha1.TLSOption
	tlsStores            []*traefikv1alpha1.TLSStore
	traefikServices      []*traefikv1alpha1.TraefikService

	// written tracks conditions passed to UpdateXxxStatus by "resource/namespace/name".
	written map[string][]metav1.Condition

	// updateIngressRouteErr, if non-nil, is returned by UpdateIngressRouteStatus.
	updateIngressRouteErr error
}

func newStatusMockClient() *statusMockClient {
	return &statusMockClient{written: make(map[string][]metav1.Condition)}
}

func (m *statusMockClient) WatchAll(_ []string, _ <-chan struct{}) (<-chan any, error) {
	return nil, nil
}

func (m *statusMockClient) GetIngressRoutes() []*traefikv1alpha1.IngressRoute {
	return m.ingressRoutes
}

func (m *statusMockClient) GetIngressRouteTCPs() []*traefikv1alpha1.IngressRouteTCP {
	return m.ingressRoutesTCP
}

func (m *statusMockClient) GetIngressRouteUDPs() []*traefikv1alpha1.IngressRouteUDP {
	return m.ingressRoutesUDP
}
func (m *statusMockClient) GetMiddlewares() []*traefikv1alpha1.Middleware { return m.middlewares }
func (m *statusMockClient) GetMiddlewareTCPs() []*traefikv1alpha1.MiddlewareTCP {
	return m.middlewaresTCP
}

func (m *statusMockClient) GetTraefikService(_, _ string) (*traefikv1alpha1.TraefikService, bool, error) {
	return nil, false, nil
}

func (m *statusMockClient) GetTraefikServices() []*traefikv1alpha1.TraefikService {
	return m.traefikServices
}
func (m *statusMockClient) GetTLSOptions() []*traefikv1alpha1.TLSOption { return m.tlsOptions }
func (m *statusMockClient) GetServersTransports() []*traefikv1alpha1.ServersTransport {
	return m.serversTransports
}

func (m *statusMockClient) GetServersTransportTCPs() []*traefikv1alpha1.ServersTransportTCP {
	return m.serversTransportsTCP
}
func (m *statusMockClient) GetTLSStores() []*traefikv1alpha1.TLSStore { return m.tlsStores }
func (m *statusMockClient) GetService(_, _ string) (*corev1.Service, bool, error) {
	return nil, false, nil
}

func (m *statusMockClient) GetSecret(_, _ string) (*corev1.Secret, bool, error) {
	return nil, false, nil
}

func (m *statusMockClient) GetEndpointSlicesForService(_, _ string) ([]*discoveryv1.EndpointSlice, error) {
	return nil, nil
}
func (m *statusMockClient) GetNodes() ([]*corev1.Node, bool, error) { return nil, false, nil }
func (m *statusMockClient) GetConfigMap(_, _ string) (*corev1.ConfigMap, bool, error) {
	return nil, false, nil
}

func (m *statusMockClient) UpdateIngressRouteStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	if m.updateIngressRouteErr != nil {
		return m.updateIngressRouteErr
	}
	m.record("ingressroutes", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateIngressRouteTCPStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("ingressroutetcps", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateIngressRouteUDPStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("ingressrouteudps", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateMiddlewareStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("middlewares", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateMiddlewareTCPStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("middlewaretcps", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateServersTransportStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("serverstransports", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateServersTransportTCPStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("serverstransporttcps", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateTLSOptionStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("tlsoptions", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateTLSStoreStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("tlsstores", namespace, name, conds)
	return nil
}

func (m *statusMockClient) UpdateTraefikServiceStatus(_ context.Context, namespace, name string, conds []metav1.Condition) error {
	m.record("traefikservices", namespace, name, conds)
	return nil
}

func (m *statusMockClient) record(resource, namespace, name string, conds []metav1.Condition) {
	m.written[resource+"/"+namespace+"/"+name] = conds
}

func TestUpdateCRDStatuses(t *testing.T) {
	testCases := []struct {
		desc                  string
		ingressRoutes         []*traefikv1alpha1.IngressRoute
		ingressRoutesTCP      []*traefikv1alpha1.IngressRouteTCP
		buildStatuses         func() configStatuses
		updateIngressRouteErr error
		expectedWritten       map[string]metav1.ConditionStatus // "resource/namespace/name" -> expected Status
		expectTCPNotWritten   bool
	}{
		{
			desc: "visited IngressRoute with no errors gets Valid=True",
			ingressRoutes: []*traefikv1alpha1.IngressRoute{
				{ObjectMeta: metav1.ObjectMeta{Name: "ok-route", Namespace: "default", Generation: 5}},
			},
			buildStatuses: func() configStatuses {
				s := newConfigStatuses()
				s.ingressRoutes.visit("default", "ok-route")
				return s
			},
			expectedWritten: map[string]metav1.ConditionStatus{
				"ingressroutes/default/ok-route": metav1.ConditionTrue,
			},
		},
		{
			desc: "visited IngressRoute with errors gets Valid=False",
			ingressRoutes: []*traefikv1alpha1.IngressRoute{
				{ObjectMeta: metav1.ObjectMeta{Name: "bad-route", Namespace: "default"}},
			},
			buildStatuses: func() configStatuses {
				s := newConfigStatuses()
				s.ingressRoutes.addError("default", "bad-route", errors.New("service not found"))
				return s
			},
			expectedWritten: map[string]metav1.ConditionStatus{
				"ingressroutes/default/bad-route": metav1.ConditionFalse,
			},
		},
		{
			// Nothing is visited — simulates an IngressRoute filtered out by ingressClass.
			desc: "unseen IngressRoute is skipped",
			ingressRoutes: []*traefikv1alpha1.IngressRoute{
				{ObjectMeta: metav1.ObjectMeta{Name: "other-class-route", Namespace: "default"}},
			},
			buildStatuses:   newConfigStatuses,
			expectedWritten: map[string]metav1.ConditionStatus{},
		},
		{
			desc: "Forbidden error on IngressRoute aborts before TCP update",
			ingressRoutes: []*traefikv1alpha1.IngressRoute{
				{ObjectMeta: metav1.ObjectMeta{Name: "route", Namespace: "default"}},
			},
			ingressRoutesTCP: []*traefikv1alpha1.IngressRouteTCP{
				{ObjectMeta: metav1.ObjectMeta{Name: "tcp-route", Namespace: "default"}},
			},
			buildStatuses: func() configStatuses {
				s := newConfigStatuses()
				s.ingressRoutes.visit("default", "route")
				s.ingressRoutesTCP.visit("default", "tcp-route")
				return s
			},
			updateIngressRouteErr: kerror.NewForbidden(kschema.GroupResource{Group: "traefik.io", Resource: "ingressroutes"}, "route", errors.New("forbidden")),
			expectedWritten:       map[string]metav1.ConditionStatus{},
			expectTCPNotWritten:   true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mockClient := newStatusMockClient()
			mockClient.ingressRoutes = test.ingressRoutes
			mockClient.ingressRoutesTCP = test.ingressRoutesTCP
			mockClient.updateIngressRouteErr = test.updateIngressRouteErr

			statuses := test.buildStatuses()
			updateCRDStatuses(t.Context(), mockClient, statuses)

			for key, expectedStatus := range test.expectedWritten {
				conds, ok := mockClient.written[key]
				require.True(t, ok, "expected status update for %q to be written", key)
				require.Len(t, conds, 1)
				assert.Equal(t, expectedStatus, conds[0].Status, "wrong condition status for %q", key)
			}
			assert.Len(t, mockClient.written, len(test.expectedWritten), "unexpected extra status updates written")

			if test.expectTCPNotWritten {
				_, tcpWritten := mockClient.written["ingressroutetcps/default/tcp-route"]
				assert.False(t, tcpWritten, "TCP route status should not be written after 403 abort")
			}
		})
	}
}
