package types

import "strings"

// Traefik labels
const (
	LabelPrefix                                  = "traefik."
	LabelDomain                                  = LabelPrefix + "domain"
	LabelEnable                                  = LabelPrefix + "enable"
	LabelPort                                    = LabelPrefix + "port"
	LabelPortIndex                               = LabelPrefix + "portIndex"
	LabelProtocol                                = LabelPrefix + "protocol"
	LabelTags                                    = LabelPrefix + "tags"
	LabelWeight                                  = LabelPrefix + "weight"
	LabelFrontendAuthBasic                       = LabelPrefix + "frontend.auth.basic"
	LabelFrontendEntryPoints                     = LabelPrefix + "frontend.entryPoints"
	LabelFrontendRequestHeader                   = LabelPrefix + "frontend.headers.customrequestheaders"
	LabelFrontendResponseHeader                  = LabelPrefix + "frontend.headers.customresoponseheaders"
	LabelFrontendPassHostHeader                  = LabelPrefix + "frontend.passHostHeader"
	LabelFrontendPriority                        = LabelPrefix + "frontend.priority"
	LabelFrontendRule                            = LabelPrefix + "frontend.rule"
	LabelFrontendRuleType                        = LabelPrefix + "frontend.rule.type"
	LabelTraefikFrontendValue                    = LabelPrefix + "frontend.value"
	LabelTraefikFrontendWhitelistSourceRange     = LabelPrefix + "frontend.whitelistSourceRange"
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
