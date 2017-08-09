package types

const (
	// LabelPrefix Traefik label
	LabelPrefix = "traefik."
	// LabelDomain Traefik label
	LabelDomain = LabelPrefix + "domain"
	// LabelEnable Traefik label
	LabelEnable = LabelPrefix + "enable"
	// LabelPort Traefik label
	LabelPort = LabelPrefix + "port"
	// LabelPortIndex Traefik label
	LabelPortIndex = LabelPrefix + "portIndex"
	// LabelProtocol Traefik label
	LabelProtocol = LabelPrefix + "protocol"
	// LabelTags Traefik label
	LabelTags = LabelPrefix + "tags"
	// LabelWeight Traefik label
	LabelWeight = LabelPrefix + "weight"
	// LabelFrontendAuthBasic Traefik label
	LabelFrontendAuthBasic = LabelPrefix + "frontend.auth.basic"
	// LabelFrontendEntryPoints Traefik label
	LabelFrontendEntryPoints = LabelPrefix + "frontend.entryPoints"
	// LabelFrontendPassHostHeader Traefik label
	LabelFrontendPassHostHeader = LabelPrefix + "frontend.passHostHeader"
	// LabelFrontendPriority Traefik label
	LabelFrontendPriority = LabelPrefix + "frontend.priority"
	// LabelFrontendRule Traefik label
	LabelFrontendRule = LabelPrefix + "frontend.rule"
	// LabelFrontendRuleType Traefik label
	LabelFrontendRuleType = LabelPrefix + "frontend.rule.type"
	// LabelTraefikFrontendValue Traefik label
	LabelTraefikFrontendValue = LabelPrefix + "frontend.value"
	// LabelTraefikFrontendWhitelistSourceRange Traefik label
	LabelTraefikFrontendWhitelistSourceRange = LabelPrefix + "frontend.whitelistSourceRange"
	// LabelBackend Traefik label
	LabelBackend = LabelPrefix + "backend"
	// LabelBackendID Traefik label
	LabelBackendID = LabelPrefix + "backend.id"
	// LabelTraefikBackendCircuitbreaker Traefik label
	LabelTraefikBackendCircuitbreaker = LabelPrefix + "backend.circuitbreaker"
	// LabelBackendCircuitbreakerExpression Traefik label
	LabelBackendCircuitbreakerExpression = LabelPrefix + "backend.circuitbreaker.expression"
	// LabelBackendHealthcheckPath Traefik label
	LabelBackendHealthcheckPath = LabelPrefix + "backend.healthcheck.path"
	// LabelBackendHealthcheckInterval Traefik label
	LabelBackendHealthcheckInterval = LabelPrefix + "backend.healthcheck.interval"
	// LabelBackendLoadbalancerMethod Traefik label
	LabelBackendLoadbalancerMethod = LabelPrefix + "backend.loadbalancer.method"
	// LabelBackendLoadbalancerSticky Traefik label
	LabelBackendLoadbalancerSticky = LabelPrefix + "backend.loadbalancer.sticky"
	// LabelBackendMaxconnAmount Traefik label
	LabelBackendMaxconnAmount = LabelPrefix + "backend.maxconn.amount"
	// LabelBackendMaxconnExtractorfunc Traefik label
	LabelBackendMaxconnExtractorfunc = LabelPrefix + "backend.maxconn.extractorfunc"
)
