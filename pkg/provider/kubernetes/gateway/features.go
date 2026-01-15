package gateway

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/pkg/features"
)

var SupportedFeatures = sync.OnceValue(func() []features.FeatureName {
	featureSet := sets.New[features.Feature]().
		Insert(features.GatewayCoreFeatures.UnsortedList()...).
		Insert(features.GatewayExtendedFeatures.Intersection(extendedGatewayFeatures()).UnsortedList()...).
		Insert(features.HTTPRouteCoreFeatures.UnsortedList()...).
		Insert(features.HTTPRouteExtendedFeatures.Intersection(extendedHTTPRouteFeatures()).UnsortedList()...).
		Insert(features.ReferenceGrantCoreFeatures.UnsortedList()...).
		Insert(features.BackendTLSPolicyCoreFeatures.UnsortedList()...).
		Insert(features.GRPCRouteCoreFeatures.UnsortedList()...).
		Insert(features.TLSRouteCoreFeatures.UnsortedList()...)

	featureNames := make([]features.FeatureName, 0, featureSet.Len())
	for f := range featureSet {
		featureNames = append(featureNames, f.Name)
	}
	return featureNames
})

// extendedGatewayFeatures returns the supported extended Gateway features.
func extendedGatewayFeatures() sets.Set[features.Feature] {
	return sets.New(features.GatewayPort8080Feature)
}

// extendedHTTPRouteFeatures returns the supported extended HTTP Route features.
func extendedHTTPRouteFeatures() sets.Set[features.Feature] {
	return sets.New(
		features.HTTPRouteQueryParamMatchingFeature,
		features.HTTPRouteMethodMatchingFeature,
		features.HTTPRoutePortRedirectFeature,
		features.HTTPRouteSchemeRedirectFeature,
		features.HTTPRouteHostRewriteFeature,
		features.HTTPRoutePathRewriteFeature,
		features.HTTPRoutePathRedirectFeature,
		features.HTTPRouteResponseHeaderModificationFeature,
		features.HTTPRouteBackendProtocolH2CFeature,
		features.HTTPRouteBackendProtocolWebSocketFeature,
		features.HTTPRouteDestinationPortMatchingFeature,
	)
}
