package gateway

import (
	"context"
	"reflect"

	"github.com/rs/zerolog/log"
	ktypes "k8s.io/apimachinery/pkg/types"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// statusReport collects the status writes produced by a single rebuild so they
// can be flushed to the apiserver after the dynamic configuration has been published.
type statusReport struct {
	gatewayClasses     map[string]gatev1.GatewayClassStatus
	gateways           map[ktypes.NamespacedName]gatev1.GatewayStatus
	listenerSets       map[ktypes.NamespacedName]gatev1.ListenerSetStatus
	httpRoutes         map[ktypes.NamespacedName]gatev1.RouteStatus
	grpcRoutes         map[ktypes.NamespacedName]gatev1.RouteStatus
	tcpRoutes          map[ktypes.NamespacedName]gatev1.RouteStatus
	tlsRoutes          map[ktypes.NamespacedName]gatev1.RouteStatus
	backendTLSPolicies map[ktypes.NamespacedName]gatev1.PolicyStatus
}

func newStatusReport() *statusReport {
	return &statusReport{
		gatewayClasses:     map[string]gatev1.GatewayClassStatus{},
		gateways:           map[ktypes.NamespacedName]gatev1.GatewayStatus{},
		listenerSets:       map[ktypes.NamespacedName]gatev1.ListenerSetStatus{},
		httpRoutes:         map[ktypes.NamespacedName]gatev1.RouteStatus{},
		grpcRoutes:         map[ktypes.NamespacedName]gatev1.RouteStatus{},
		tcpRoutes:          map[ktypes.NamespacedName]gatev1.RouteStatus{},
		tlsRoutes:          map[ktypes.NamespacedName]gatev1.RouteStatus{},
		backendTLSPolicies: map[ktypes.NamespacedName]gatev1.PolicyStatus{},
	}
}

// Flush sends every status write collected during the
// routing configuration build to the Kubernetes API server.
func (r *statusReport) Flush(ctx context.Context, client *clientWrapper) {
	logger := log.Ctx(ctx)

	for name, status := range r.gatewayClasses {
		if err := client.UpdateGatewayClassStatus(ctx, name, status); err != nil {
			logger.Warn().Err(err).Str("gateway_class", name).Msg("Unable to update GatewayClass status")
		}
	}

	for name, status := range r.gateways {
		if err := client.UpdateGatewayStatus(ctx, name, status); err != nil {
			logger.Warn().Err(err).Str("gateway", name.Name).Str("namespace", name.Namespace).Msg("Unable to update Gateway status")
		}
	}

	for name, status := range r.listenerSets {
		if err := client.UpdateListenerSetStatus(ctx, name, status); err != nil {
			logger.Warn().Err(err).Str("listener_set", name.Name).Str("namespace", name.Namespace).Msg("Unable to update ListenerSet status")
		}
	}

	for name, routeStatus := range r.httpRoutes {
		status := gatev1.HTTPRouteStatus{RouteStatus: routeStatus}
		if err := client.UpdateHTTPRouteStatus(ctx, name, status); err != nil {
			logger.Warn().Err(err).Str("http_route", name.Name).Str("namespace", name.Namespace).Msg("Unable to update HTTPRoute status")
		}
	}

	for name, routeStatus := range r.grpcRoutes {
		status := gatev1.GRPCRouteStatus{RouteStatus: routeStatus}
		if err := client.UpdateGRPCRouteStatus(ctx, name, status); err != nil {
			logger.Warn().Err(err).Str("grpc_route", name.Name).Str("namespace", name.Namespace).Msg("Unable to update GRPCRoute status")
		}
	}

	for name, routeStatus := range r.tcpRoutes {
		status := gatev1alpha2.TCPRouteStatus{RouteStatus: routeStatus}
		if err := client.UpdateTCPRouteStatus(ctx, name, status); err != nil {
			logger.Warn().Err(err).Str("tcp_route", name.Name).Str("namespace", name.Namespace).Msg("Unable to update TCPRoute status")
		}
	}

	for name, routeStatus := range r.tlsRoutes {
		status := gatev1.TLSRouteStatus{RouteStatus: routeStatus}
		if err := client.UpdateTLSRouteStatus(ctx, name, status); err != nil {
			logger.Warn().Err(err).Str("tls_route", name.Name).Str("namespace", name.Namespace).Msg("Unable to update TLSRoute status")
		}
	}

	for name, policyStatus := range r.backendTLSPolicies {
		if err := client.UpdateBackendTLSPolicyStatus(ctx, name, policyStatus); err != nil {
			logger.Warn().Err(err).Str("backend_tls_policy", name.Name).Str("namespace", name.Namespace).Msg("Unable to update BackendTLSPolicy status")
		}
	}
}

func (r *statusReport) RecordGatewayClassStatus(gatewayClassName string, status gatev1.GatewayClassStatus) {
	r.gatewayClasses[gatewayClassName] = status
}

func (r *statusReport) RecordGatewayStatus(gateway ktypes.NamespacedName, status gatev1.GatewayStatus) {
	r.gateways[gateway] = status
}

func (r *statusReport) RecordListenerSetStatus(listenerSet ktypes.NamespacedName, status gatev1.ListenerSetStatus) {
	r.listenerSets[listenerSet] = status
}

func (r *statusReport) RecordHTTPRouteStatus(route ktypes.NamespacedName, status gatev1.RouteParentStatus) {
	r.httpRoutes[route] = gatev1.RouteStatus{
		Parents: append(r.httpRoutes[route].Parents, status),
	}
}

func (r *statusReport) RecordGRPCRouteStatus(route ktypes.NamespacedName, status gatev1.RouteParentStatus) {
	r.grpcRoutes[route] = gatev1.RouteStatus{
		Parents: append(r.grpcRoutes[route].Parents, status),
	}
}

func (r *statusReport) RecordTCPRouteStatus(route ktypes.NamespacedName, status gatev1.RouteParentStatus) {
	r.tcpRoutes[route] = gatev1.RouteStatus{
		Parents: append(r.tcpRoutes[route].Parents, status),
	}
}

func (r *statusReport) RecordTLSRouteStatus(route ktypes.NamespacedName, status gatev1.RouteParentStatus) {
	r.tlsRoutes[route] = gatev1.RouteStatus{
		Parents: append(r.tlsRoutes[route].Parents, status),
	}
}

func (r *statusReport) RecordBackendTLSPolicyStatus(policy ktypes.NamespacedName, status gatev1.PolicyAncestorStatus) {
	var ancestors []gatev1.PolicyAncestorStatus

	// Keep existing ancestor statuses, except if it matches the status to merge.
	for _, existing := range r.backendTLSPolicies[policy].Ancestors {
		if reflect.DeepEqual(existing.AncestorRef, status.AncestorRef) {
			continue
		}

		ancestors = append(ancestors, existing)
	}

	r.backendTLSPolicies[policy] = gatev1.PolicyStatus{
		Ancestors: append(ancestors, status), // Add the new status to the existing ancestors statuses.
	}
}
