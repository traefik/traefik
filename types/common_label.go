package types

const (
	// LabelDomain Traefik label
	LabelDomain = "traefik.domain"
	// LabelEnable Traefik label
	LabelEnable = "traefik.enable"
	// LabelPort Traefik label
	LabelPort = "traefik.port"
	// LabelPortIndex Traefik label
	LabelPortIndex = "traefik.portIndex"
	// LabelProtocol Traefik label
	LabelProtocol = "traefik.protocol"
	// LabelTags Traefik label
	LabelTags = "traefik.tags"
	// LabelWeight Traefik label
	LabelWeight = "traefik.weight"
	// LabelFrontendAuthBasic Traefik label
	LabelFrontendAuthBasic = "traefik.frontend.auth.basic"
	// LabelFrontendEntryPoints Traefik label
	LabelFrontendEntryPoints = "traefik.frontend.entryPoints"
	// LabelFrontendPassHostHeader Traefik label
	LabelFrontendPassHostHeader = "traefik.frontend.passHostHeader"
	// LabelFrontendPriority Traefik label
	LabelFrontendPriority = "traefik.frontend.priority"
	// LabelFrontendRule Traefik label
	LabelFrontendRule = "traefik.frontend.rule"
	// LabelFrontendRuleType Traefik label
	LabelFrontendRuleType = "traefik.frontend.rule.type"
	// LabelTraefikFrontendValue Traefik label
	LabelTraefikFrontendValue = "traefik.frontend.value"
	// LabelTraefikFrontendWhitelistSourceRange Traefik label
	LabelTraefikFrontendWhitelistSourceRange = "traefik.frontend.whitelistSourceRange"
	// LabelBackend Traefik label
	LabelBackend = "traefik.backend"
	// LabelBackendID Traefik label
	LabelBackendID = "traefik.backend.id"
	// LabelTraefikBackendCircuitbreaker Traefik label
	LabelTraefikBackendCircuitbreaker = "traefik.backend.circuitbreaker"
	// LabelBackendCircuitbreakerExpression Traefik label
	LabelBackendCircuitbreakerExpression = "traefik.backend.circuitbreaker.expression"
	// LabelBackendHealthcheckPath Traefik label
	LabelBackendHealthcheckPath = "traefik.backend.healthcheck.path"
	// LabelBackendHealthcheckInterval Traefik label
	LabelBackendHealthcheckInterval = "traefik.backend.healthcheck.interval"
	// LabelBackendLoadbalancerMethod Traefik label
	LabelBackendLoadbalancerMethod = "traefik.backend.loadbalancer.method"
	// LabelBackendLoadbalancerSticky Traefik label
	LabelBackendLoadbalancerSticky = "traefik.backend.loadbalancer.sticky"
	// LabelBackendMaxconnAmount Traefik label
	LabelBackendMaxconnAmount = "traefik.backend.maxconn.amount"
	// LabelBackendMaxconnExtractorfunc Traefik label
	LabelBackendMaxconnExtractorfunc = "traefik.backend.maxconn.extractorfunc"
)
