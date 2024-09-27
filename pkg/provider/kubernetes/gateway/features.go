package gateway

import "sigs.k8s.io/gateway-api/pkg/features"

func SupportedFeatures() []features.FeatureName {
	return []features.FeatureName{
		features.GatewayFeature.Name,
		features.GatewayPort8080Feature.Name,
		features.GRPCRouteFeature.Name,
		features.HTTPRouteFeature.Name,
		features.HTTPRouteQueryParamMatchingFeature.Name,
		features.HTTPRouteMethodMatchingFeature.Name,
		features.HTTPRoutePortRedirectFeature.Name,
		features.HTTPRouteSchemeRedirectFeature.Name,
		features.HTTPRouteHostRewriteFeature.Name,
		features.HTTPRoutePathRewriteFeature.Name,
		features.HTTPRoutePathRedirectFeature.Name,
		features.HTTPRouteResponseHeaderModificationFeature.Name,
		features.HTTPRouteBackendProtocolH2CFeature.Name,
		features.HTTPRouteBackendProtocolWebSocketFeature.Name,
		features.HTTPRouteDestinationPortMatchingFeature.Name,
		features.TLSRouteFeature.Name,
	}
}
