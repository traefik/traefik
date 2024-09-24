package gateway

import "sigs.k8s.io/gateway-api/pkg/features"

func SupportedFeatures() []features.SupportedFeature {
	return []features.SupportedFeature{
		features.SupportGateway,
		features.SupportGatewayPort8080,
		features.SupportGRPCRoute,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteQueryParamMatching,
		features.SupportHTTPRouteMethodMatching,
		features.SupportHTTPRoutePortRedirect,
		features.SupportHTTPRouteSchemeRedirect,
		features.SupportHTTPRouteHostRewrite,
		features.SupportHTTPRoutePathRewrite,
		features.SupportHTTPRoutePathRedirect,
		features.SupportHTTPRouteResponseHeaderModification,
		features.SupportTLSRoute,
		features.SupportHTTPRouteBackendProtocolH2C,
		features.SupportHTTPRouteBackendProtocolWebSocket,
	}
}
