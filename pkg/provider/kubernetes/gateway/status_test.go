package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestStatusReport_RecordGatewayClassStatus(t *testing.T) {
	report := newStatusReport()

	accepted := gatev1.GatewayClassStatus{
		Conditions: []metav1.Condition{{Type: string(gatev1.GatewayClassConditionStatusAccepted)}},
	}
	report.RecordGatewayClassStatus("traefik", accepted)
	assert.Equal(t, accepted, report.gatewayClasses["traefik"])

	// A later record for the same GatewayClass overwrites the previous one.
	unsupported := gatev1.GatewayClassStatus{
		Conditions: []metav1.Condition{{Type: string(gatev1.GatewayClassReasonUnsupportedVersion)}},
	}
	report.RecordGatewayClassStatus("traefik", unsupported)
	assert.Equal(t, unsupported, report.gatewayClasses["traefik"])
}

func TestStatusReport_RecordGatewayStatus(t *testing.T) {
	report := newStatusReport()
	gateway := ktypes.NamespacedName{Namespace: "default", Name: "my-gateway"}

	accepted := gatev1.GatewayStatus{
		Conditions: []metav1.Condition{{Type: string(gatev1.GatewayConditionAccepted)}},
	}
	report.RecordGatewayStatus(gateway, accepted)
	assert.Equal(t, accepted, report.gateways[gateway])

	// A later record for the same Gateway overwrites the previous one.
	programmed := gatev1.GatewayStatus{
		Conditions: []metav1.Condition{{Type: string(gatev1.GatewayConditionProgrammed)}},
	}
	report.RecordGatewayStatus(gateway, programmed)
	assert.Equal(t, programmed, report.gateways[gateway])
}

func TestStatusReport_RecordHTTPRouteStatus(t *testing.T) {
	report := newStatusReport()
	route := ktypes.NamespacedName{Namespace: "default", Name: "my-route"}

	gatewayParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "gateway"}}
	otherParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "other-gateway"}}

	report.RecordHTTPRouteStatus(route, gatewayParent)
	report.RecordHTTPRouteStatus(route, otherParent)

	// Each parentRef accumulates as a distinct parent status.
	assert.Equal(t, []gatev1.RouteParentStatus{gatewayParent, otherParent}, report.httpRoutes[route].Parents)
}

func TestStatusReport_RecordGRPCRouteStatus(t *testing.T) {
	report := newStatusReport()
	route := ktypes.NamespacedName{Namespace: "default", Name: "my-route"}

	gatewayParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "gateway"}}
	otherParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "other-gateway"}}

	report.RecordGRPCRouteStatus(route, gatewayParent)
	report.RecordGRPCRouteStatus(route, otherParent)

	assert.Equal(t, []gatev1.RouteParentStatus{gatewayParent, otherParent}, report.grpcRoutes[route].Parents)
}

func TestStatusReport_RecordTCPRouteStatus(t *testing.T) {
	report := newStatusReport()
	route := ktypes.NamespacedName{Namespace: "default", Name: "my-route"}

	gatewayParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "gateway"}}
	otherParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "other-gateway"}}

	report.RecordTCPRouteStatus(route, gatewayParent)
	report.RecordTCPRouteStatus(route, otherParent)

	assert.Equal(t, []gatev1.RouteParentStatus{gatewayParent, otherParent}, report.tcpRoutes[route].Parents)
}

func TestStatusReport_RecordTLSRouteStatus(t *testing.T) {
	report := newStatusReport()
	route := ktypes.NamespacedName{Namespace: "default", Name: "my-route"}

	gatewayParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "gateway"}}
	otherParent := gatev1.RouteParentStatus{ParentRef: gatev1.ParentReference{Name: "other-gateway"}}

	report.RecordTLSRouteStatus(route, gatewayParent)
	report.RecordTLSRouteStatus(route, otherParent)

	assert.Equal(t, []gatev1.RouteParentStatus{gatewayParent, otherParent}, report.tlsRoutes[route].Parents)
}

func TestStatusReport_RecordBackendTLSPolicyStatus(t *testing.T) {
	gatewayAncestor := gatev1.PolicyAncestorStatus{
		AncestorRef:    gatev1.ParentReference{Name: "gateway"},
		ControllerName: controllerName,
	}
	otherAncestor := gatev1.PolicyAncestorStatus{
		AncestorRef:    gatev1.ParentReference{Name: "other-gateway"},
		ControllerName: controllerName,
	}

	testCases := []struct {
		desc     string
		records  []gatev1.PolicyAncestorStatus
		expected []gatev1.PolicyAncestorStatus
	}{
		{
			desc:     "distinct ancestor refs accumulate",
			records:  []gatev1.PolicyAncestorStatus{gatewayAncestor, otherAncestor},
			expected: []gatev1.PolicyAncestorStatus{gatewayAncestor, otherAncestor},
		},
		{
			desc: "same ancestor ref is replaced, not duplicated",
			records: []gatev1.PolicyAncestorStatus{
				gatewayAncestor,
				{
					AncestorRef:    gatev1.ParentReference{Name: "gateway"},
					ControllerName: "another.io/controller",
				},
			},
			expected: []gatev1.PolicyAncestorStatus{
				{
					AncestorRef:    gatev1.ParentReference{Name: "gateway"},
					ControllerName: "another.io/controller",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			report := newStatusReport()
			policy := ktypes.NamespacedName{Namespace: "default", Name: "my-policy"}

			for _, record := range test.records {
				report.RecordBackendTLSPolicyStatus(policy, record)
			}

			assert.Equal(t, test.expected, report.backendTLSPolicies[policy].Ancestors)
		})
	}
}
