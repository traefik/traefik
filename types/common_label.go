package types

import "strings"

// Traefik labels
const (
	LabelPrefix                                  = "traefik."
	SuffixPort                                   = "port"
	SuffixProtocol                               = "protocol"
	SuffixWeight                                 = "weight"
	SuffixFrontendAuthBasic                      = "frontend.auth.basic"
	SuffixFrontendBackend                        = "frontend.backend"
	SuffixFrontendEntryPoints                    = "frontend.entryPoints"
	SuffixFrontendPassHostHeader                 = "frontend.passHostHeader"
	SuffixFrontendPriority                       = "frontend.priority"
	SuffixFrontendRedirect                       = "frontend.redirect"
	SuffixFrontendRule                           = "frontend.rule"
	LabelDomain                                  = LabelPrefix + "domain"
	LabelEnable                                  = LabelPrefix + "enable"
	LabelPort                                    = LabelPrefix + SuffixPort
	LabelPortIndex                               = LabelPrefix + "portIndex"
	LabelProtocol                                = LabelPrefix + SuffixProtocol
	LabelTags                                    = LabelPrefix + "tags"
	LabelWeight                                  = LabelPrefix + SuffixWeight
	LabelFrontendAuthBasic                       = LabelPrefix + SuffixFrontendAuthBasic
	LabelFrontendEntryPoints                     = LabelPrefix + SuffixFrontendEntryPoints
	LabelFrontendPassHostHeader                  = LabelPrefix + SuffixFrontendPassHostHeader
	LabelFrontendPassTLSCert                     = LabelPrefix + "frontend.passTLSCert"
	LabelFrontendPriority                        = LabelPrefix + SuffixFrontendPriority
	LabelFrontendRule                            = LabelPrefix + SuffixFrontendRule
	LabelFrontendRuleType                        = LabelPrefix + "frontend.rule.type"
	LabelFrontendRedirect                        = LabelPrefix + SuffixFrontendRedirect
	LabelTraefikFrontendValue                    = LabelPrefix + "frontend.value"
	LabelTraefikFrontendWhitelistSourceRange     = LabelPrefix + "frontend.whitelistSourceRange"
	LabelFrontendRequestHeader                   = LabelPrefix + "frontend.headers.customrequestheaders"
	LabelFrontendResponseHeader                  = LabelPrefix + "frontend.headers.customresponseheaders"
	LabelFrontendAllowedHosts                    = LabelPrefix + "frontend.headers.allowedHosts"
	LabelFrontendHostsProxyHeaders               = LabelPrefix + "frontend.headers.hostsProxyHeaders"
	LabelFrontendSSLRedirect                     = LabelPrefix + "frontend.headers.SSLRedirect"
	LabelFrontendSSLTemporaryRedirect            = LabelPrefix + "frontend.headers.SSLTemporaryRedirect"
	LabelFrontendSSLHost                         = LabelPrefix + "frontend.headers.SSLHost"
	LabelFrontendSSLProxyHeaders                 = LabelPrefix + "frontend.headers.SSLProxyHeaders"
	LabelFrontendSTSSeconds                      = LabelPrefix + "frontend.headers.STSSeconds"
	LabelFrontendSTSIncludeSubdomains            = LabelPrefix + "frontend.headers.STSIncludeSubdomains"
	LabelFrontendSTSPreload                      = LabelPrefix + "frontend.headers.STSPreload"
	LabelFrontendForceSTSHeader                  = LabelPrefix + "frontend.headers.forceSTSHeader"
	LabelFrontendFrameDeny                       = LabelPrefix + "frontend.headers.frameDeny"
	LabelFrontendCustomFrameOptionsValue         = LabelPrefix + "frontend.headers.customFrameOptionsValue"
	LabelFrontendContentTypeNosniff              = LabelPrefix + "frontend.headers.contentTypeNosniff"
	LabelFrontendBrowserXSSFilter                = LabelPrefix + "frontend.headers.browserXSSFilter"
	LabelFrontendContentSecurityPolicy           = LabelPrefix + "frontend.headers.contentSecurityPolicy"
	LabelFrontendPublicKey                       = LabelPrefix + "frontend.headers.publicKey"
	LabelFrontendReferrerPolicy                  = LabelPrefix + "frontend.headers.referrerPolicy"
	LabelFrontendIsDevelopment                   = LabelPrefix + "frontend.headers.isDevelopment"
	LabelBackend                                 = LabelPrefix + "backend"
	LabelBackendID                               = LabelPrefix + "backend.id"
	LabelTraefikBackendCircuitbreaker            = LabelPrefix + "backend.circuitbreaker"
	LabelBackendCircuitbreakerExpression         = LabelPrefix + "backend.circuitbreaker.expression"
	LabelBackendHealthcheckPath                  = LabelPrefix + "backend.healthcheck.path"
	LabelBackendHealthcheckInterval              = LabelPrefix + "backend.healthcheck.interval"
	LabelBackendLoadbalancerMethod               = LabelPrefix + "backend.loadbalancer.method"
	LabelBackendLoadbalancerSticky               = LabelPrefix + "backend.loadbalancer.sticky"
	LabelBackendLoadbalancerStickiness           = LabelPrefix + "backend.loadbalancer.stickiness"
	LabelBackendLoadbalancerStickinessCookieName = LabelPrefix + "backend.loadbalancer.stickiness.cookieName"
	LabelBackendMaxconnAmount                    = LabelPrefix + "backend.maxconn.amount"
	LabelBackendMaxconnExtractorfunc             = LabelPrefix + "backend.maxconn.extractorfunc"
)

//ServiceLabel converts a key value of Label*, given a serviceName, into a pattern <LabelPrefix>.<serviceName>.<property>
//    i.e. For LabelFrontendRule and serviceName=app it will return "traefik.app.frontend.rule"
func ServiceLabel(key, serviceName string) string {
	if len(serviceName) > 0 {
		property := strings.TrimPrefix(key, LabelPrefix)
		return LabelPrefix + serviceName + "." + property
	}
	return key
}
