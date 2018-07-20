package audittap

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/containous/traefik/middlewares/audittap/configuration"
)

// Possible ProxyingFor types
const (
	RATE = "rate"
	API  = "api"
	MDTP = "mdtp"
)

// AuditConfig specifies audit construction characteristics
type AuditConfig struct {
	AuditSource string
	AuditType   string
	ProxyingFor string
	audittypes.AuditSpecification
}

// AuditTap writes an event to the audit streams for every request.
type AuditTap struct {
	AuditConfig
	AuditStreams    []audittypes.AuditStream
	Backend         string
	MaxEntityLength int
	next            http.Handler
}

// NewAuditTap returns a new AuditTap handler.
func NewAuditTap(config *configuration.AuditSink, streams []audittypes.AuditStream, backend string, next http.Handler) (*AuditTap, error) {

	var err error

	var maxAudit int64
	if config.MaxAuditLength != "" {
		if maxAudit, _, err = asSI(config.MaxAuditLength); err != nil {
			return nil, err
		}
	} else {
		maxAudit = 100000
	}

	var maxPayload int64
	if config.MaxPayloadContentsLength != "" {
		if maxPayload, _, err = asSI(config.MaxPayloadContentsLength); err != nil {
			return nil, err
		}
	} else {
		maxPayload = 96000
	}

	var th = maxAudit // Default the max recorded response length to the max audit size
	if config.MaxEntityLength != "" {
		th, _, err = asSI(config.MaxEntityLength)
		if err != nil {
			return nil, err
		}
	}

	pf := strings.ToLower(config.ProxyingFor)
	if pf != MDTP && pf != API && pf != RATE {
		return nil, fmt.Errorf(fmt.Sprintf("ProxyingFor value '%s' is invalid", config.ProxyingFor))
	}

	// RATE values are either constant or chosen dynamically
	if pf != RATE && pf != MDTP {
		if config.AuditSource == "" {
			return nil, fmt.Errorf("AuditSource not set in configuration")
		}

		if config.AuditType == "" {
			return nil, fmt.Errorf("AuditType not set in configuration")
		}
	}

	exclusions, err := optionsToFilters(config.Exclusions)
	if err != nil {
		return nil, fmt.Errorf("Failed to construct audit exclusion filter %v", err)
	}

	obfuscate := audittypes.AuditObfuscation{}
	obfuscate.MaskFields = config.MaskFields
	if len(obfuscate.MaskFields) > 0 {
		if config.MaskValue != "" {
			obfuscate.MaskValue = config.MaskValue
		} else {
			obfuscate.MaskValue = "#########"
		}
	}

	dynamicFields := make(audittypes.HeaderMappings)
	for section, mappings := range config.HeaderMappings {
		dynamicFields[section] = audittypes.HeaderMapping(mappings)
	}

	constraints := audittypes.AuditConstraints{MaxAuditLength: maxAudit, MaxPayloadContentsLength: maxPayload}

	auditSpec := audittypes.AuditSpecification{
		AuditConstraints: constraints,
		AuditObfuscation: obfuscate,
		HeaderMappings:   dynamicFields,
		Exclusions:       exclusions,
	}

	ac := AuditConfig{
		AuditSource:        config.AuditSource,
		AuditType:          config.AuditType,
		ProxyingFor:        config.ProxyingFor,
		AuditSpecification: auditSpec,
	}
	return &AuditTap{ac, streams, backend, int(th), next}, nil
}

func (tap *AuditTap) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var auditer audittypes.Auditer
	ctx := audittypes.NewRequestContext(req)
	shouldAudit := audittypes.ShouldAudit(ctx, &tap.AuditSpecification)

	log.Debugf("Should audit is %t for Host:%s URI:%s Headers:%v", shouldAudit, req.Host, req.RequestURI, req.Header)

	if shouldAudit {
		switch strings.ToLower(tap.ProxyingFor) {
		case "api":
			auditer = audittypes.NewAPIAuditEvent(tap.AuditSource, tap.AuditType)
		case "rate":
			auditer = audittypes.NewRATEAuditEvent()
		case "mdtp":
			auditer = audittypes.NewMdtpAuditEvent()
		}
		auditer.AppendRequest(ctx, &tap.AuditSpecification)
	}

	ww := NewAuditResponseWriter(rw, tap.MaxEntityLength)
	tap.next.ServeHTTP(ww, req)

	if shouldAudit {
		auditer.AppendResponse(ww.Header(), ww.GetResponseInfo(), &tap.AuditSpecification)
		if auditer.EnforceConstraints(tap.AuditSpecification.AuditConstraints) {
			tap.submitAudit(auditer)
		}
	}
}

func (tap *AuditTap) submitAudit(auditer audittypes.Auditer) error {
	enc := auditer.ToEncoded()
	if enc.Err != nil {
		return enc.Err
	}
	if int64(enc.Length()) <= tap.AuditConstraints.MaxAuditLength {
		for _, sink := range tap.AuditStreams {
			sink.Audit(enc)
		}
	} else {
		log.Errorf("Dropping audit event. Length %d exceeds limit %d", enc.Length(), tap.AuditConstraints.MaxAuditLength)
	}
	return nil
}

func optionsToFilters(opts map[string]*configuration.FilterOption) ([]*audittypes.Filter, error) {
	exclusions := []*audittypes.Filter{}
	for _, exc := range opts {
		if exc.Enabled() {
			filter, err := makeFilter(exc)
			if err != nil {
				return nil, err
			}
			exclusions = append(exclusions, filter)
		}
	}
	return exclusions, nil
}

func makeFilter(opt *configuration.FilterOption) (*audittypes.Filter, error) {
	expressions := []*regexp.Regexp{}
	for _, pattern := range opt.Matches {
		exp, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, exp)
	}
	f := &audittypes.Filter{
		Source:     opt.HeaderName,
		StartsWith: opt.StartsWith,
		EndsWith:   opt.EndsWith,
		Contains:   opt.Contains,
		Matches:    expressions,
	}
	return f, nil
}

// asSI parses a string for its number. Suffixes are allowed that loosely follow SI rules: K, Ki, M, Mi.
// 'k' and 'K' are equivalent.
// Example: "2 KiB" returns 2048, "B", nil
func asSI(value string) (int64, string, error) {
	if value == "" {
		return 0, "", fmt.Errorf("Blank value")
	}

	numEnd := len(value)
	for i, r := range value {
		if unicode.IsDigit(r) {
			numEnd = i + 1
		}
	}

	number := value[:numEnd]
	unit := strings.TrimSpace(value[numEnd:])

	if strings.HasPrefix(unit, "Ki") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1024, unit[2:], e
	}

	if strings.HasPrefix(strings.ToUpper(unit), "K") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1000, unit[1:], e
	}

	if strings.HasPrefix(unit, "Mi") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1024 * 1024, unit[2:], e
	}

	if strings.HasPrefix(unit, "M") {
		i, e := strconv.ParseInt(number, 10, 64)
		return i * 1000000, unit[1:], e
	}

	i, e := strconv.ParseInt(number, 10, 64)
	return i, "", e
}
