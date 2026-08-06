package gateway

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestLoadUDPRouteWeightedInvalidBackend(t *testing.T) {
	client := newUDPRouteTestClient(t, []runtime.Object{
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "udp-backend", Namespace: "default"},
			Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{
				Name:     "udp",
				Protocol: corev1.ProtocolUDP,
				Port:     5300,
			}}},
		},
		&discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "udp-backend-abc",
				Namespace: "default",
				Labels:    map[string]string{discoveryv1.LabelServiceName: "udp-backend"},
			},
			AddressType: discoveryv1.AddressTypeIPv4,
			Ports: []discoveryv1.EndpointPort{{
				Name:     new("udp"),
				Protocol: ptr.To(corev1.ProtocolUDP),
				Port:     new(int32(5300)),
			}},
			Endpoints: []discoveryv1.Endpoint{{
				Addresses:  []string{"10.0.0.1"},
				Conditions: discoveryv1.EndpointConditions{Ready: new(true)},
			}},
		},
	}, nil)

	route := &gatev1.UDPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: "weighted", Namespace: "default", Generation: 2},
		Spec: gatev1.UDPRouteSpec{
			CommonRouteSpec: gatev1.CommonRouteSpec{},
			Rules: []gatev1.UDPRouteRule{{BackendRefs: []gatev1.BackendRef{
				{
					BackendObjectReference: gatev1.BackendObjectReference{
						Name: "udp-backend",
						Port: ptr.To(gatev1.PortNumber(5300)),
					},
					Weight: new(int32(7)),
				},
				{
					BackendObjectReference: gatev1.BackendObjectReference{
						Name: "missing-backend",
						Port: ptr.To(gatev1.PortNumber(5300)),
					},
					Weight: new(int32(3)),
				},
			}}},
		},
	}

	p := Provider{client: client}
	conf, condition := p.loadUDPRoute("gateway", "default", gatewayListener{EPName: "udp"}, route)

	assert.Equal(t, metav1.ConditionFalse, condition.Status)
	assert.Equal(t, string(gatev1.RouteReasonBackendNotFound), condition.Reason)

	routerName := makeRouterName("udproute", "", "default", "weighted", "default", "gateway", "udp", 0)
	require.Equal(t, &dynamic.UDPRouter{
		EntryPoints: []string{"udp"},
		Service:     routerName + "-wrr",
	}, conf.UDP.Routers[routerName])

	require.Equal(t, []dynamic.UDPWRRService{
		{Name: routerName + "-svc-default-udp-backend-0", Weight: new(7)},
		{Name: routerName + "-err-lb", Weight: new(3)},
	}, conf.UDP.Services[routerName+"-wrr"].Weighted.Services)
	assert.Equal(t, []dynamic.UDPServer{{Address: "10.0.0.1:5300"}}, conf.UDP.Services[routerName+"-svc-default-udp-backend-0"].LoadBalancer.Servers)
	assert.Empty(t, conf.UDP.Services[routerName+"-err-lb"].LoadBalancer.Servers)
}

func TestLoadUDPService(t *testing.T) {
	coreGroup := gatev1.Group(groupCore)
	emptyGroup := gatev1.Group("")
	unsupportedGroup := gatev1.Group("example.com")
	unsupportedKind := gatev1.Kind("ExampleBackend")
	serviceKind := gatev1.Kind(kindService)
	traefikServiceKind := gatev1.Kind(kindTraefikService)
	backendNamespace := gatev1.Namespace("backend")
	emptyEndpointObjects := newUDPBackendObjects("default", corev1.ProtocolUDP, true)
	emptyEndpointObjects[1].(*discoveryv1.EndpointSlice).Endpoints = nil

	testCases := []struct {
		desc        string
		kubeObjects []runtime.Object
		backendRef  gatev1.BackendRef
		wantReason  string
		wantServers []dynamic.UDPServer
	}{
		{
			desc:        "UDP Service",
			kubeObjects: newUDPBackendObjects("default", corev1.ProtocolUDP, true),
			backendRef:  newUDPBackendRef(nil, 5300),
			wantServers: []dynamic.UDPServer{{Address: "10.0.0.1:5300"}},
		},
		{
			desc:        "TCP Service",
			kubeObjects: newUDPBackendObjects("default", corev1.ProtocolTCP, true),
			backendRef:  newUDPBackendRef(nil, 5300),
			wantReason:  string(gatev1.RouteReasonUnsupportedProtocol),
		},
		{
			desc:        "TCP Service without EndpointSlice",
			kubeObjects: newUDPBackendObjects("default", corev1.ProtocolTCP, false),
			backendRef:  newUDPBackendRef(nil, 5300),
			wantReason:  string(gatev1.RouteReasonUnsupportedProtocol),
		},
		{
			desc:       "missing Service",
			backendRef: newUDPBackendRef(nil, 5300),
			wantReason: string(gatev1.RouteReasonBackendNotFound),
		},
		{
			desc: "unsupported Group",
			backendRef: gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
				Group: &unsupportedGroup,
				Kind:  &serviceKind,
				Name:  "backend",
				Port:  ptr.To(gatev1.PortNumber(5300)),
			}},
			wantReason: string(gatev1.RouteReasonInvalidKind),
		},
		{
			desc: "explicit core Group",
			backendRef: gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
				Group: &coreGroup,
				Kind:  &serviceKind,
				Name:  "backend",
				Port:  ptr.To(gatev1.PortNumber(5300)),
			}},
			wantReason: string(gatev1.RouteReasonInvalidKind),
		},
		{
			desc: "explicit core Group in another namespace",
			backendRef: gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
				Group:     &coreGroup,
				Kind:      &serviceKind,
				Name:      "backend",
				Namespace: &backendNamespace,
				Port:      ptr.To(gatev1.PortNumber(5300)),
			}},
			wantReason: string(gatev1.RouteReasonInvalidKind),
		},
		{
			desc: "unsupported Kind",
			backendRef: gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
				Kind: &unsupportedKind,
				Name: "backend",
				Port: ptr.To(gatev1.PortNumber(5300)),
			}},
			wantReason: string(gatev1.RouteReasonInvalidKind),
		},
		{
			desc: "TraefikService without Group",
			backendRef: gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
				Kind: &traefikServiceKind,
				Name: "api@internal",
			}},
			wantReason: string(gatev1.RouteReasonInvalidKind),
		},
		{
			desc: "TraefikService with empty Group",
			backendRef: gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
				Group: &emptyGroup,
				Kind:  &traefikServiceKind,
				Name:  "api@internal",
			}},
			wantReason: string(gatev1.RouteReasonInvalidKind),
		},
		{
			desc: "TraefikService with unsupported Group",
			backendRef: gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
				Group: &unsupportedGroup,
				Kind:  &traefikServiceKind,
				Name:  "api@internal",
			}},
			wantReason: string(gatev1.RouteReasonInvalidKind),
		},
		{
			desc:        "wrong Service port",
			kubeObjects: newUDPBackendObjects("default", corev1.ProtocolUDP, true),
			backendRef:  newUDPBackendRef(nil, 5400),
			wantReason:  string(gatev1.RouteReasonBackendNotFound),
		},
		{
			desc:        "Service with empty EndpointSlice",
			kubeObjects: emptyEndpointObjects,
			backendRef:  newUDPBackendRef(nil, 5300),
		},
		{
			desc:        "Service without EndpointSlice",
			kubeObjects: newUDPBackendObjects("default", corev1.ProtocolUDP, false),
			backendRef:  newUDPBackendRef(nil, 5300),
		},
		{
			desc:        "cross-namespace Service without ReferenceGrant",
			kubeObjects: newUDPBackendObjects("backend", corev1.ProtocolUDP, true),
			backendRef:  newUDPBackendRef(&backendNamespace, 5300),
			wantReason:  string(gatev1.RouteReasonRefNotPermitted),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := newUDPRouteTestClient(t, test.kubeObjects, nil)
			p := Provider{client: client}
			route := &gatev1.UDPRoute{ObjectMeta: metav1.ObjectMeta{
				Name:       "route",
				Namespace:  "default",
				Generation: 4,
			}}

			serviceName, service, condition := p.loadUDPService("router", route, 0, test.backendRef)
			assert.NotEmpty(t, serviceName)

			if test.wantReason == "" {
				require.Nil(t, condition)
				require.NotNil(t, service)
				require.NotNil(t, service.LoadBalancer)
				assert.Equal(t, test.wantServers, service.LoadBalancer.Servers)
				return
			}

			require.NotNil(t, condition)
			assert.Equal(t, metav1.ConditionFalse, condition.Status)
			assert.Equal(t, test.wantReason, condition.Reason)
			assert.Equal(t, route.Generation, condition.ObservedGeneration)
			assert.Nil(t, service)
		})
	}
}

func TestLoadUDPServiceReferenceGrantLifecycle(t *testing.T) {
	backendNamespace := gatev1.Namespace("backend")
	backendRef := newUDPBackendRef(&backendNamespace, 5300)
	client := newUDPRouteTestClient(t, newUDPBackendObjects("backend", corev1.ProtocolUDP, true), nil)
	p := Provider{client: client}
	route := &gatev1.UDPRoute{ObjectMeta: metav1.ObjectMeta{Name: "route", Namespace: "default"}}

	load := func() (*dynamic.UDPService, *metav1.Condition) {
		t.Helper()
		_, service, condition := p.loadUDPService("router", route, 0, backendRef)
		return service, condition
	}

	service, condition := load()
	assert.Nil(t, service)
	require.NotNil(t, condition)
	assert.Equal(t, string(gatev1.RouteReasonRefNotPermitted), condition.Reason)

	referenceGrant := &gatev1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{Name: "allow-udp-route", Namespace: "backend"},
		Spec: gatev1.ReferenceGrantSpec{
			From: []gatev1.ReferenceGrantFrom{{
				Group:     gatev1.Group(gatev1.GroupName),
				Kind:      gatev1.Kind(kindUDPRoute),
				Namespace: "default",
			}},
			To: []gatev1.ReferenceGrantTo{{
				Kind: gatev1.Kind(kindService),
			}},
		},
	}
	_, err := client.csGateway.GatewayV1().ReferenceGrants("backend").Create(t.Context(), referenceGrant, metav1.CreateOptions{})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		grants, listErr := client.ListReferenceGrants("backend")
		return listErr == nil && len(grants) == 1
	}, 5*time.Second, 10*time.Millisecond)

	service, condition = load()
	require.Nil(t, condition)
	require.NotNil(t, service)
	require.NotNil(t, service.LoadBalancer)
	assert.Equal(t, []dynamic.UDPServer{{Address: "10.0.0.1:5300"}}, service.LoadBalancer.Servers)

	err = client.csGateway.GatewayV1().ReferenceGrants("backend").Delete(t.Context(), referenceGrant.Name, metav1.DeleteOptions{})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		grants, listErr := client.ListReferenceGrants("backend")
		return listErr == nil && len(grants) == 0
	}, 5*time.Second, 10*time.Millisecond)

	service, condition = load()
	assert.Nil(t, service)
	require.NotNil(t, condition)
	assert.Equal(t, string(gatev1.RouteReasonRefNotPermitted), condition.Reason)
}

func TestLoadUDPRouteV1Fixture(t *testing.T) {
	kubeObjects, gatewayObjects := readResources(t, []string{"udproute/simple.yml"})
	client := newUDPRouteTestClient(t, kubeObjects, gatewayObjects)
	p := Provider{
		EntryPoints: map[string]Entrypoint{
			"udp": {Address: ":5300", Protocol: entryPointProtocolUDP},
		},
		client: client,
	}

	conf, report, err := p.loadConfigurationFromGateways(t.Context())
	require.NoError(t, err)

	routerName := makeRouterName("udproute", "", "default", "udp-route", "default", "udp-gateway", "udp", 0)
	require.Equal(t, &dynamic.UDPRouter{
		EntryPoints: []string{"udp"},
		Service:     routerName + "-wrr",
	}, conf.UDP.Routers[routerName])
	assert.Equal(t, []dynamic.UDPServer{{Address: "10.0.0.1:5300"}}, conf.UDP.Services[routerName+"-svc-default-udp-backend-0"].LoadBalancer.Servers)

	routeStatus := report.udpRoutes[ktypes.NamespacedName{Namespace: "default", Name: "udp-route"}]
	require.Len(t, routeStatus.Parents, 1)
	accepted := meta.FindStatusCondition(routeStatus.Parents[0].Conditions, string(gatev1.RouteConditionAccepted))
	require.NotNil(t, accepted)
	assert.Equal(t, metav1.ConditionTrue, accepted.Status)
}

func TestLoadUDPRoutesConflictPrecedence(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)
	parentRef := gatev1.ParentReference{Name: "gateway"}

	newer := newInternalServiceUDPRoute("newer", "newer@internal", metav1.NewTime(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)), parentRef, group, kind)
	older := newInternalServiceUDPRoute("older", "older@internal", metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), parentRef, group, kind)
	client := newUDPRouteTestClient(t, nil, []runtime.Object{newer, older})

	listenerStatus := &gatev1.ListenerStatus{Name: "udp"}
	gateways := []gatewayWithListeners{{
		Name:      "gateway",
		Namespace: "default",
		listeners: []gatewayListener{{
			Name:              "udp",
			Port:              5300,
			Protocol:          gatev1.UDPProtocolType,
			Status:            listenerStatus,
			AllowedNamespaces: []string{"default"},
			AllowedRouteKinds: []string{kindUDPRoute},
			Attached:          true,
			EPName:            "udp",
		}},
	}}
	conf := &dynamic.Configuration{UDP: &dynamic.UDPConfiguration{
		Routers:  map[string]*dynamic.UDPRouter{},
		Services: map[string]*dynamic.UDPService{},
	}}
	report := newStatusReport()

	p := Provider{client: client}
	p.loadUDPRoutes(t.Context(), gateways, conf, report)

	require.Len(t, conf.UDP.Routers, 1)
	for _, router := range conf.UDP.Routers {
		assert.Equal(t, "older@internal", router.Service)
	}
	assert.Equal(t, int32(2), listenerStatus.AttachedRoutes)

	for _, routeName := range []string{"older", "newer"} {
		status := report.udpRoutes[ktypes.NamespacedName{Namespace: "default", Name: routeName}]
		require.Len(t, status.Parents, 1)
		accepted := meta.FindStatusCondition(status.Parents[0].Conditions, string(gatev1.RouteConditionAccepted))
		require.NotNil(t, accepted)
		assert.Equal(t, metav1.ConditionTrue, accepted.Status)
		assert.Equal(t, string(gatev1.RouteReasonAccepted), accepted.Reason)
	}
}

func TestLoadUDPRoutesConflictFailover(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)
	parentRef := gatev1.ParentReference{Name: "gateway"}

	newer := newInternalServiceUDPRoute("newer", "newer@internal", metav1.NewTime(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)), parentRef, group, kind)
	older := newInternalServiceUDPRoute("older", "older@internal", metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), parentRef, group, kind)
	client := newUDPRouteTestClient(t, nil, []runtime.Object{newer, older})
	p := Provider{client: client}

	load := func() (*dynamic.Configuration, *statusReport, *gatev1.ListenerStatus) {
		t.Helper()

		listenerStatus := &gatev1.ListenerStatus{Name: "udp"}
		gateways := []gatewayWithListeners{newUDPRouteGatewayWithListener(listenerStatus)}
		conf := &dynamic.Configuration{UDP: &dynamic.UDPConfiguration{
			Routers:  map[string]*dynamic.UDPRouter{},
			Services: map[string]*dynamic.UDPService{},
		}}
		report := newStatusReport()

		p.loadUDPRoutes(t.Context(), gateways, conf, report)
		return conf, report, listenerStatus
	}

	conf, _, listenerStatus := load()
	require.Len(t, conf.UDP.Routers, 1)
	for _, router := range conf.UDP.Routers {
		assert.Equal(t, "older@internal", router.Service)
	}
	assert.Equal(t, int32(2), listenerStatus.AttachedRoutes)

	err := client.csGateway.GatewayV1().UDPRoutes("default").Delete(t.Context(), "older", metav1.DeleteOptions{})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		routes, listErr := client.ListUDPRoutes()
		return listErr == nil && len(routes) == 1 && routes[0].Name == "newer"
	}, 5*time.Second, 10*time.Millisecond)

	conf, report, listenerStatus := load()
	require.Len(t, conf.UDP.Routers, 1)
	for _, router := range conf.UDP.Routers {
		assert.Equal(t, "newer@internal", router.Service)
	}
	assert.Equal(t, int32(1), listenerStatus.AttachedRoutes)
	assert.NotContains(t, report.udpRoutes, ktypes.NamespacedName{Namespace: "default", Name: "older"})

	status := report.udpRoutes[ktypes.NamespacedName{Namespace: "default", Name: "newer"}]
	require.Len(t, status.Parents, 1)
	accepted := meta.FindStatusCondition(status.Parents[0].Conditions, string(gatev1.RouteConditionAccepted))
	require.NotNil(t, accepted)
	assert.Equal(t, metav1.ConditionTrue, accepted.Status)
}

func TestLoadUDPRoutesGatewayListenerConflictLifecycle(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)

	gatewayClass := &gatev1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{Name: "traefik"},
		Spec:       gatev1.GatewayClassSpec{ControllerName: controllerName},
	}
	newGateway := func(name, listenerName string, port gatev1.PortNumber) *gatev1.Gateway {
		return &gatev1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default", Generation: 1},
			Spec: gatev1.GatewaySpec{
				GatewayClassName: "traefik",
				Listeners: []gatev1.Listener{{
					Name:     gatev1.SectionName(listenerName),
					Protocol: gatev1.UDPProtocolType,
					Port:     port,
				}},
			},
		}
	}

	gatewayA := newGateway("gateway-a", "udp-a", 5300)
	gatewayB := newGateway("gateway-b", "udp-b", 5300)
	routeA := newInternalServiceUDPRoute("route-a", "api@internal", metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), gatev1.ParentReference{Name: "gateway-a"}, group, kind)
	routeB := newInternalServiceUDPRoute("route-b", "dashboard@internal", metav1.NewTime(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)), gatev1.ParentReference{Name: "gateway-b"}, group, kind)

	client := newUDPRouteTestClient(t, nil, []runtime.Object{gatewayClass, gatewayA, gatewayB, routeA, routeB})
	p := Provider{
		EntryPoints: map[string]Entrypoint{
			"udp":     {Address: ":5300", Protocol: entryPointProtocolUDP},
			"udp-alt": {Address: ":5400", Protocol: entryPointProtocolUDP},
		},
		client: client,
	}

	load := func() (*dynamic.Configuration, *statusReport) {
		t.Helper()

		conf, report, err := p.loadConfigurationFromGateways(t.Context())
		require.NoError(t, err)
		return conf, report
	}
	assertState := func(conf *dynamic.Configuration, report *statusReport, conflicted bool) {
		t.Helper()

		if conflicted {
			assert.Empty(t, conf.UDP.Routers)
		} else {
			assert.Len(t, conf.UDP.Routers, 2)
		}

		for _, gatewayName := range []string{"gateway-a", "gateway-b"} {
			status := report.gateways[ktypes.NamespacedName{Namespace: "default", Name: gatewayName}]
			require.Len(t, status.Listeners, 1)
			assert.Equal(t, int32(1), status.Listeners[0].AttachedRoutes)

			accepted := meta.FindStatusCondition(status.Listeners[0].Conditions, string(gatev1.ListenerConditionAccepted))
			require.NotNil(t, accepted)
			assert.Equal(t, metav1.ConditionTrue, accepted.Status)
			resolvedRefs := meta.FindStatusCondition(status.Listeners[0].Conditions, string(gatev1.ListenerConditionResolvedRefs))
			require.NotNil(t, resolvedRefs)
			assert.Equal(t, metav1.ConditionTrue, resolvedRefs.Status)
			programmed := meta.FindStatusCondition(status.Listeners[0].Conditions, string(gatev1.ListenerConditionProgrammed))
			require.NotNil(t, programmed)

			condition := meta.FindStatusCondition(status.Listeners[0].Conditions, string(gatev1.ListenerConditionConflicted))
			if conflicted {
				require.NotNil(t, condition)
				assert.Equal(t, metav1.ConditionTrue, condition.Status)
				assert.Equal(t, string(gatev1.ListenerReasonProtocolConflict), condition.Reason)
				assert.Equal(t, metav1.ConditionFalse, programmed.Status)
				assert.Equal(t, string(gatev1.ListenerReasonInvalid), programmed.Reason)

				gatewayAccepted := meta.FindStatusCondition(status.Conditions, string(gatev1.GatewayConditionAccepted))
				require.NotNil(t, gatewayAccepted)
				assert.Equal(t, metav1.ConditionFalse, gatewayAccepted.Status)
				assert.Equal(t, string(gatev1.GatewayReasonListenersNotValid), gatewayAccepted.Reason)
				gatewayProgrammed := meta.FindStatusCondition(status.Conditions, string(gatev1.GatewayConditionProgrammed))
				require.NotNil(t, gatewayProgrammed)
				assert.Equal(t, metav1.ConditionFalse, gatewayProgrammed.Status)
				assert.Equal(t, string(gatev1.GatewayReasonInvalid), gatewayProgrammed.Reason)
				continue
			}

			assert.Nil(t, condition)
			assert.Equal(t, metav1.ConditionTrue, programmed.Status)
			assert.Equal(t, string(gatev1.ListenerReasonProgrammed), programmed.Reason)

			gatewayAccepted := meta.FindStatusCondition(status.Conditions, string(gatev1.GatewayConditionAccepted))
			require.NotNil(t, gatewayAccepted)
			assert.Equal(t, metav1.ConditionTrue, gatewayAccepted.Status)
			gatewayProgrammed := meta.FindStatusCondition(status.Conditions, string(gatev1.GatewayConditionProgrammed))
			require.NotNil(t, gatewayProgrammed)
			assert.Equal(t, metav1.ConditionTrue, gatewayProgrammed.Status)
		}

		for _, routeName := range []string{"route-a", "route-b"} {
			status := report.udpRoutes[ktypes.NamespacedName{Namespace: "default", Name: routeName}]
			require.Len(t, status.Parents, 1)
			accepted := meta.FindStatusCondition(status.Parents[0].Conditions, string(gatev1.RouteConditionAccepted))
			require.NotNil(t, accepted)
			assert.Equal(t, metav1.ConditionTrue, accepted.Status)
		}
	}
	updateGatewayBPort := func(port gatev1.PortNumber) {
		t.Helper()

		gateway, err := client.csGateway.GatewayV1().Gateways("default").Get(t.Context(), "gateway-b", metav1.GetOptions{})
		require.NoError(t, err)
		gateway.Spec.Listeners[0].Port = port
		gateway.Generation++
		_, err = client.csGateway.GatewayV1().Gateways("default").Update(t.Context(), gateway, metav1.UpdateOptions{})
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			for _, current := range client.ListGateways() {
				if current.Name == "gateway-b" {
					return current.Spec.Listeners[0].Port == port
				}
			}
			return false
		}, 5*time.Second, 10*time.Millisecond)
	}

	conf, report := load()
	assertState(conf, report, true)

	updateGatewayBPort(5400)
	conf, report = load()
	assertState(conf, report, false)

	updateGatewayBPort(5300)
	conf, report = load()
	assertState(conf, report, true)
}

func TestLoadUDPRoutesPortUnavailableAttachment(t *testing.T) {
	testLoadUDPRoutesUnavailableEntryPointAttachment(t, nil)
}

func TestLoadUDPRoutesAmbiguousEntryPointAttachment(t *testing.T) {
	testLoadUDPRoutesUnavailableEntryPointAttachment(t, map[string]Entrypoint{
		"udp-a": {Address: ":5300", Protocol: entryPointProtocolUDP},
		"udp-b": {Address: ":5300", Protocol: entryPointProtocolUDP},
	})
}

func testLoadUDPRoutesUnavailableEntryPointAttachment(t *testing.T, entryPoints map[string]Entrypoint) {
	t.Helper()

	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)

	gatewayClass := &gatev1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{Name: "traefik"},
		Spec:       gatev1.GatewayClassSpec{ControllerName: controllerName},
	}
	gateway := &gatev1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gateway", Namespace: "default", Generation: 2},
		Spec: gatev1.GatewaySpec{
			GatewayClassName: "traefik",
			Listeners: []gatev1.Listener{{
				Name:     "udp",
				Protocol: gatev1.UDPProtocolType,
				Port:     5300,
			}},
		},
	}
	route := newInternalServiceUDPRoute("route", "api@internal", metav1.Time{}, gatev1.ParentReference{Name: "gateway"}, group, kind)

	client := newUDPRouteTestClient(t, nil, []runtime.Object{gatewayClass, gateway, route})
	p := Provider{EntryPoints: entryPoints, client: client}

	conf, report, err := p.loadConfigurationFromGateways(t.Context())
	require.NoError(t, err)
	assert.Empty(t, conf.UDP.Routers)

	gatewayStatus := report.gateways[ktypes.NamespacedName{Namespace: "default", Name: "gateway"}]
	require.Len(t, gatewayStatus.Listeners, 1)
	listenerStatus := gatewayStatus.Listeners[0]
	assert.Equal(t, int32(1), listenerStatus.AttachedRoutes)
	assert.Equal(t, []gatev1.RouteGroupKind{{
		Group: new(gatev1.Group(gatev1.GroupName)),
		Kind:  kindUDPRoute,
	}}, listenerStatus.SupportedKinds)

	accepted := meta.FindStatusCondition(listenerStatus.Conditions, string(gatev1.ListenerConditionAccepted))
	require.NotNil(t, accepted)
	assert.Equal(t, metav1.ConditionFalse, accepted.Status)
	assert.Equal(t, string(gatev1.ListenerReasonPortUnavailable), accepted.Reason)
	assert.Nil(t, meta.FindStatusCondition(listenerStatus.Conditions, string(gatev1.ListenerConditionConflicted)))
	resolvedRefs := meta.FindStatusCondition(listenerStatus.Conditions, string(gatev1.ListenerConditionResolvedRefs))
	require.NotNil(t, resolvedRefs)
	assert.Equal(t, metav1.ConditionTrue, resolvedRefs.Status)
	programmed := meta.FindStatusCondition(listenerStatus.Conditions, string(gatev1.ListenerConditionProgrammed))
	require.NotNil(t, programmed)
	assert.Equal(t, metav1.ConditionFalse, programmed.Status)
	assert.Equal(t, string(gatev1.ListenerReasonInvalid), programmed.Reason)

	gatewayAccepted := meta.FindStatusCondition(gatewayStatus.Conditions, string(gatev1.GatewayConditionAccepted))
	require.NotNil(t, gatewayAccepted)
	assert.Equal(t, metav1.ConditionFalse, gatewayAccepted.Status)
	gatewayProgrammed := meta.FindStatusCondition(gatewayStatus.Conditions, string(gatev1.GatewayConditionProgrammed))
	require.NotNil(t, gatewayProgrammed)
	assert.Equal(t, metav1.ConditionFalse, gatewayProgrammed.Status)

	routeStatus := report.udpRoutes[ktypes.NamespacedName{Namespace: "default", Name: "route"}]
	require.Len(t, routeStatus.Parents, 1)
	routeAccepted := meta.FindStatusCondition(routeStatus.Parents[0].Conditions, string(gatev1.RouteConditionAccepted))
	require.NotNil(t, routeAccepted)
	assert.Equal(t, metav1.ConditionTrue, routeAccepted.Status)
}

func TestRejectConflictingUDPListeners(t *testing.T) {
	newGateway := func(name string, entryPoints ...string) gatewayWithListeners {
		gateway := gatewayWithListeners{Name: name, Namespace: "default", Generation: 3}
		for _, entryPoint := range entryPoints {
			port := gatev1.PortNumber(5300)
			if entryPoint == "udp-alt" {
				port = 5400
			}

			listenerName := gatev1.SectionName(name + "-" + entryPoint)
			gateway.listeners = append(gateway.listeners, gatewayListener{
				Name:     string(listenerName),
				Port:     port,
				Protocol: gatev1.UDPProtocolType,
				Status:   &gatev1.ListenerStatus{Name: listenerName},
				Attached: true,
				EPName:   entryPoint,
			})
		}
		return gateway
	}
	invalidGateway := newGateway("invalid-gateway", "udp")
	invalidGateway.listeners[0].Attached = false
	invalidGateway.listeners[0].Status.Conditions = []metav1.Condition{{
		Type:   string(gatev1.ListenerConditionAccepted),
		Status: metav1.ConditionFalse,
		Reason: "InvalidTLSConfiguration",
	}}
	sameGateway := newGateway("gateway", "udp", "udp-other")
	sameGateway.listeners[1].Attached = false
	sameGateway.listeners[1].Status.Conditions = []metav1.Condition{{
		Type:   string(gatev1.ListenerConditionConflicted),
		Status: metav1.ConditionTrue,
		Reason: "DuplicateListener",
	}}
	mixedProtocolGateway := newGateway("gateway", "udp", "udp-other")
	tcpListenerName := gatev1.SectionName("gateway-tcp")
	mixedProtocolGateway.listeners = append(mixedProtocolGateway.listeners, gatewayListener{
		Name:     string(tcpListenerName),
		Port:     5300,
		Protocol: gatev1.TCPProtocolType,
		Status:   &gatev1.ListenerStatus{Name: tcpListenerName},
		Attached: true,
		EPName:   "tcp",
	})

	for _, test := range []struct {
		desc                    string
		gateways                []gatewayWithListeners
		wantConflictedCount     int
		wantAttachedCount       int
		wantConflictedListeners []string
		wantAttachedListeners   []string
	}{
		{
			desc:              "single listener",
			gateways:          []gatewayWithListeners{newGateway("gateway", "udp")},
			wantAttachedCount: 1,
		},
		{
			desc:                "same Gateway",
			gateways:            []gatewayWithListeners{sameGateway},
			wantConflictedCount: 2,
		},
		{
			desc:                "different Gateways",
			gateways:            []gatewayWithListeners{newGateway("gateway-a", "udp"), newGateway("gateway-b", "udp")},
			wantConflictedCount: 2,
		},
		{
			desc:                "same port on different entryPoints",
			gateways:            []gatewayWithListeners{newGateway("gateway-a", "udp"), newGateway("gateway-b", "udp-other")},
			wantConflictedCount: 2,
		},
		{
			desc:                    "same port with another protocol",
			gateways:                []gatewayWithListeners{mixedProtocolGateway},
			wantConflictedCount:     2,
			wantAttachedCount:       1,
			wantConflictedListeners: []string{"gateway/gateway-udp", "gateway/gateway-udp-other"},
			wantAttachedListeners:   []string{"gateway/gateway-tcp"},
		},
		{
			desc:              "different ports",
			gateways:          []gatewayWithListeners{newGateway("gateway-a", "udp"), newGateway("gateway-b", "udp-alt")},
			wantAttachedCount: 2,
		},
		{
			desc:                "already rejected listener",
			gateways:            []gatewayWithListeners{newGateway("gateway", "udp"), invalidGateway},
			wantConflictedCount: 2,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			rejectConflictingUDPListeners(test.gateways)

			var conflictedCount int
			var attachedCount int
			var conflictedListeners []string
			var attachedListeners []string
			for _, gateway := range test.gateways {
				for _, listener := range gateway.listeners {
					if listener.Attached {
						attachedCount++
						attachedListeners = append(attachedListeners, gateway.Name+"/"+listener.Name)
					}
					condition := meta.FindStatusCondition(listener.Status.Conditions, string(gatev1.ListenerConditionConflicted))
					if condition == nil {
						continue
					}

					conflictedCount++
					conflictedListeners = append(conflictedListeners, gateway.Name+"/"+listener.Name)
					assert.False(t, listener.Attached)
					assert.Equal(t, metav1.ConditionTrue, condition.Status)
					assert.Equal(t, string(gatev1.ListenerReasonProtocolConflict), condition.Reason)
					assert.Equal(t, gateway.Generation, condition.ObservedGeneration)

					var conflictConditionCount int
					for _, current := range listener.Status.Conditions {
						if current.Type == string(gatev1.ListenerConditionConflicted) {
							conflictConditionCount++
						}
					}
					assert.Equal(t, 1, conflictConditionCount)
				}
			}
			assert.Equal(t, test.wantConflictedCount, conflictedCount)
			assert.Equal(t, test.wantAttachedCount, attachedCount)
			if test.wantConflictedListeners != nil {
				assert.ElementsMatch(t, test.wantConflictedListeners, conflictedListeners)
			}
			if test.wantAttachedListeners != nil {
				assert.ElementsMatch(t, test.wantAttachedListeners, attachedListeners)
			}
		})
	}
}

func TestLoadUDPRoutesRemovesStatusAfterParentRefRemoval(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)
	sectionName := gatev1.SectionName("udp")
	parentRef := gatev1.ParentReference{Name: "gateway", SectionName: &sectionName}
	route := newInternalServiceUDPRoute("route", "api@internal", metav1.Now(), parentRef, group, kind)

	gatewayClass := &gatev1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{Name: "traefik"},
		Spec:       gatev1.GatewayClassSpec{ControllerName: controllerName},
	}
	gateway := &gatev1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gateway", Namespace: "default"},
		Spec: gatev1.GatewaySpec{
			GatewayClassName: "traefik",
			Listeners: []gatev1.Listener{{
				Name:     sectionName,
				Protocol: gatev1.UDPProtocolType,
				Port:     5300,
			}},
		},
	}

	client := newUDPRouteTestClient(t, nil, []runtime.Object{gatewayClass, gateway, route})
	p := Provider{
		EntryPoints: map[string]Entrypoint{
			"udp": {Address: ":5300", Protocol: entryPointProtocolUDP},
		},
		client: client,
	}

	conf, report, err := p.loadConfigurationFromGateways(t.Context())
	require.NoError(t, err)
	require.Len(t, conf.UDP.Routers, 1)
	report.Flush(t.Context(), client)

	currentRoute, err := client.csGateway.GatewayV1().UDPRoutes("default").Get(t.Context(), "route", metav1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, currentRoute.Status.Parents, 1)
	assert.Equal(t, gatev1.GatewayController(controllerName), currentRoute.Status.Parents[0].ControllerName)

	currentRoute.Spec.ParentRefs = nil
	currentRoute.Generation++
	_, err = client.csGateway.GatewayV1().UDPRoutes("default").Update(t.Context(), currentRoute, metav1.UpdateOptions{})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		routes, listErr := client.ListUDPRoutes()
		return listErr == nil && len(routes) == 1 && len(routes[0].Spec.ParentRefs) == 0 && len(routes[0].Status.Parents) == 1
	}, 5*time.Second, 10*time.Millisecond)

	conf, report, err = p.loadConfigurationFromGateways(t.Context())
	require.NoError(t, err)
	assert.Empty(t, conf.UDP.Routers)
	report.Flush(t.Context(), client)

	require.Eventually(t, func() bool {
		updatedRoute, getErr := client.csGateway.GatewayV1().UDPRoutes("default").Get(t.Context(), "route", metav1.GetOptions{})
		return getErr == nil && len(updatedRoute.Status.Parents) == 0
	}, 5*time.Second, 10*time.Millisecond)
}

func TestLoadUDPRouteCrossProviderNamespaces(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)
	route := newInternalServiceUDPRoute("route", "api@internal", metav1.Time{}, gatev1.ParentReference{}, group, kind)

	for _, test := range []struct {
		desc                    string
		crossProviderNamespaces []string
		wantAllowed             bool
	}{
		{desc: "unset allow-list", crossProviderNamespaces: nil, wantAllowed: true},
		{desc: "empty allow-list", crossProviderNamespaces: []string{}, wantAllowed: false},
		{desc: "route namespace allowed", crossProviderNamespaces: []string{"default"}, wantAllowed: true},
		{desc: "route namespace not allowed", crossProviderNamespaces: []string{"other"}, wantAllowed: false},
	} {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{CrossProviderNamespaces: test.crossProviderNamespaces}
			conf, condition := p.loadUDPRoute("gateway", "default", gatewayListener{EPName: "udp"}, route)

			assert.Equal(t, test.wantAllowed, len(conf.UDP.Routers) == 1)
			if test.wantAllowed {
				assert.Equal(t, metav1.ConditionTrue, condition.Status)
				return
			}

			assert.Equal(t, metav1.ConditionFalse, condition.Status)
			assert.Equal(t, string(gatev1.RouteReasonRefNotPermitted), condition.Reason)
		})
	}
}

func TestLoadUDPRouteCrossProviderZeroWeight(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)
	route := newInternalServiceUDPRoute("route", "api@internal", metav1.Time{}, gatev1.ParentReference{}, group, kind)
	route.Spec.Rules[0].BackendRefs[0].Weight = new(int32(0))

	p := Provider{}
	conf, condition := p.loadUDPRoute("gateway", "default", gatewayListener{EPName: "udp"}, route)

	assert.Equal(t, metav1.ConditionTrue, condition.Status)

	routerName := makeRouterName("udproute", "", "default", "route", "default", "gateway", "udp", 0)
	require.Equal(t, &dynamic.UDPRouter{
		EntryPoints: []string{"udp"},
		Service:     routerName + "-wrr",
	}, conf.UDP.Routers[routerName])
	require.Equal(t, []dynamic.UDPWRRService{{
		Name:   "api@internal",
		Weight: new(0),
	}}, conf.UDP.Services[routerName+"-wrr"].Weighted.Services)
}

func TestLoadUDPRouteCrossProviderReferenceWithNamespace(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)
	route := newInternalServiceUDPRoute("route", "api@internal", metav1.Time{}, gatev1.ParentReference{}, group, kind)
	route.Spec.Rules[0].BackendRefs[0].Namespace = new(gatev1.Namespace("other"))

	p := Provider{}
	conf, condition := p.loadUDPRoute("gateway", "default", gatewayListener{EPName: "udp"}, route)

	assert.Equal(t, metav1.ConditionFalse, condition.Status)
	assert.Equal(t, string(gatev1.RouteReasonRefNotPermitted), condition.Reason)

	routerName := makeRouterName("udproute", "", "default", "route", "default", "gateway", "udp", 0)
	router, ok := conf.UDP.Routers[routerName]
	require.True(t, ok)

	service, ok := conf.UDP.Services[router.Service]
	require.True(t, ok)
	require.NotNil(t, service.Weighted)
	require.Len(t, service.Weighted.Services, 1)

	invalidService, ok := conf.UDP.Services[service.Weighted.Services[0].Name]
	require.True(t, ok)
	require.NotNil(t, invalidService.LoadBalancer)
	assert.Empty(t, invalidService.LoadBalancer.Servers)
}

func TestLoadUDPRouteCrossProviderBackendRefs(t *testing.T) {
	group := gatev1.Group(traefikv1alpha1.GroupName)
	kind := gatev1.Kind(kindTraefikService)

	for _, test := range []struct {
		desc                    string
		crossProviderNamespaces []string
		backendNamespace        *gatev1.Namespace
		wantError               bool
	}{
		{desc: "unset allow-list", crossProviderNamespaces: nil, wantError: false},
		{desc: "empty allow-list", crossProviderNamespaces: []string{}, wantError: true},
		{desc: "route namespace allowed", crossProviderNamespaces: []string{"default"}, wantError: false},
		{desc: "route namespace not allowed", crossProviderNamespaces: []string{"other"}, wantError: true},
		{desc: "explicit namespace with cross-provider reference", backendNamespace: ptr.To(gatev1.Namespace("other")), wantError: true},
	} {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			route := newInternalServiceUDPRoute("route", "api@internal", metav1.Time{}, gatev1.ParentReference{}, group, kind)
			route.Spec.Rules[0].BackendRefs[0].Namespace = test.backendNamespace
			route.Spec.Rules[0].BackendRefs = append(route.Spec.Rules[0].BackendRefs, gatev1.BackendRef{
				BackendObjectReference: gatev1.BackendObjectReference{
					Group:     &group,
					Kind:      &kind,
					Name:      "dashboard@internal",
					Namespace: test.backendNamespace,
				},
			})

			p := Provider{CrossProviderNamespaces: test.crossProviderNamespaces}
			conf, condition := p.loadUDPRoute("gateway", "default", gatewayListener{EPName: "udp"}, route)

			routerName := makeRouterName("udproute", "", "default", "route", "default", "gateway", "udp", 0)
			router, ok := conf.UDP.Routers[routerName]
			require.True(t, ok)

			service, ok := conf.UDP.Services[router.Service]
			require.True(t, ok)
			require.NotNil(t, service.Weighted)
			require.Len(t, service.Weighted.Services, 2)

			var hasError bool
			for _, wrrService := range service.Weighted.Services {
				if strings.Contains(wrrService.Name, "@") {
					continue
				}

				lbService, ok := conf.UDP.Services[wrrService.Name]
				require.True(t, ok)
				require.NotNil(t, lbService.LoadBalancer)
				if len(lbService.LoadBalancer.Servers) == 0 {
					hasError = true
				}
			}

			assert.Equal(t, test.wantError, hasError)
			assert.Equal(t, test.wantError, condition.Status == metav1.ConditionFalse)
		})
	}
}

func TestLoadGatewayListenersRejectsTLSOnUDP(t *testing.T) {
	p := Provider{EntryPoints: map[string]Entrypoint{
		"udp": {Address: ":5300", Protocol: entryPointProtocolUDP},
	}}
	gateway := &gatev1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: "gateway", Namespace: "default", Generation: 3},
		Spec: gatev1.GatewaySpec{
			GatewayClassName: "traefik",
			Listeners: []gatev1.Listener{{
				Name:     "udp",
				Protocol: gatev1.UDPProtocolType,
				Port:     5300,
				TLS:      &gatev1.ListenerTLSConfig{},
			}},
		},
	}

	listeners := p.loadGatewayListeners(t.Context(), gateway, &dynamic.Configuration{TLS: &dynamic.TLSConfiguration{}})

	require.Len(t, listeners, 1)
	accepted := meta.FindStatusCondition(listeners[0].Status.Conditions, string(gatev1.ListenerConditionAccepted))
	require.NotNil(t, accepted)
	assert.Equal(t, metav1.ConditionFalse, accepted.Status)
	assert.Equal(t, "InvalidTLSConfiguration", accepted.Reason)
	assert.False(t, listeners[0].Attached)
}

func newUDPRouteTestClient(t *testing.T, kubeObjects, gatewayObjects []runtime.Object) *clientWrapper {
	t.Helper()

	client := newClientImpl(kubefake.NewClientset(kubeObjects...), newGatewaySimpleClientSet(t, gatewayObjects...))
	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })

	_, err := client.WatchAll(nil, stopCh)
	require.NoError(t, err)

	return client
}

func newInternalServiceUDPRoute(name, serviceName string, creationTimestamp metav1.Time, parentRef gatev1.ParentReference, group gatev1.Group, kind gatev1.Kind) *gatev1.UDPRoute {
	return &gatev1.UDPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: creationTimestamp,
		},
		Spec: gatev1.UDPRouteSpec{
			CommonRouteSpec: gatev1.CommonRouteSpec{ParentRefs: []gatev1.ParentReference{parentRef}},
			Rules: []gatev1.UDPRouteRule{{BackendRefs: []gatev1.BackendRef{{
				BackendObjectReference: gatev1.BackendObjectReference{
					Group: &group,
					Kind:  &kind,
					Name:  gatev1.ObjectName(serviceName),
				},
			}}}},
		},
	}
}

func newUDPBackendObjects(namespace string, protocol corev1.Protocol, withEndpoints bool) []runtime.Object {
	objects := []runtime.Object{&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "backend", Namespace: namespace},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{
			Name:     "backend",
			Protocol: protocol,
			Port:     5300,
		}}},
	}}
	if !withEndpoints {
		return objects
	}

	return append(objects, &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-abc",
			Namespace: namespace,
			Labels:    map[string]string{discoveryv1.LabelServiceName: "backend"},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Ports: []discoveryv1.EndpointPort{{
			Name:     new("backend"),
			Protocol: &protocol,
			Port:     new(int32(5300)),
		}},
		Endpoints: []discoveryv1.Endpoint{{
			Addresses:  []string{"10.0.0.1"},
			Conditions: discoveryv1.EndpointConditions{Ready: new(true)},
		}},
	})
}

func newUDPBackendRef(namespace *gatev1.Namespace, port gatev1.PortNumber) gatev1.BackendRef {
	return gatev1.BackendRef{BackendObjectReference: gatev1.BackendObjectReference{
		Name:      "backend",
		Namespace: namespace,
		Port:      &port,
	}}
}

func newUDPRouteGatewayWithListener(status *gatev1.ListenerStatus) gatewayWithListeners {
	return gatewayWithListeners{
		Name:      "gateway",
		Namespace: "default",
		listeners: []gatewayListener{{
			Name:              "udp",
			Port:              5300,
			Protocol:          gatev1.UDPProtocolType,
			Status:            status,
			AllowedNamespaces: []string{"default"},
			AllowedRouteKinds: []string{kindUDPRoute},
			Attached:          true,
			EPName:            "udp",
		}},
	}
}
