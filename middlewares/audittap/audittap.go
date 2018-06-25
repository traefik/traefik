package audittap

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"net/url"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/containous/traefik/types"
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
	Exclusions  []*types.Exclusion
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
func NewAuditTap(config *types.AuditSink, streams []audittypes.AuditStream, backend string, next http.Handler) (*AuditTap, error) {

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

	exclusions := []*types.Exclusion{}
	for _, exc := range config.Exclusions {
		if exc.Enabled() {
			exclusions = append(exclusions, exc)
		}
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
	// for section, mappings := range config.HeaderMappings {
	// 	dynamicFields[section] = audittypes.HeaderMapping(mappings)
	// }

	constraints := audittypes.AuditConstraints{MaxAuditLength: maxAudit, MaxPayloadContentsLength: maxPayload}

	auditSpec := audittypes.AuditSpecification{
		AuditConstraints: constraints,
		AuditObfuscation: obfuscate,
		HeaderMappings:   dynamicFields,
	}

	ac := AuditConfig{
		AuditSource:        config.AuditSource,
		AuditType:          config.AuditType,
		ProxyingFor:        config.ProxyingFor,
		Exclusions:         exclusions,
		AuditSpecification: auditSpec,
	}
	return &AuditTap{ac, streams, backend, int(th), next}, nil
}

func (tap *AuditTap) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var auditer audittypes.Auditer
	excludeAudit := isExcluded(tap.Exclusions, req)

	log.Debugf("Exclude audit is %t for Host:%s URI:%s Headers:%v", excludeAudit, req.Host, req.RequestURI, req.Header)

	if !excludeAudit {
		switch strings.ToLower(tap.ProxyingFor) {
		case "api":
			auditer = audittypes.NewAPIAuditEvent(tap.AuditSource, tap.AuditType)
		case "rate":
			auditer = audittypes.NewRATEAuditEvent()
		case "mdtp":
			auditer = audittypes.NewMdtpAuditEvent()
		}
		auditer.AppendRequest(req, &tap.AuditSpecification)
	}

	ww := NewAuditResponseWriter(rw, tap.MaxEntityLength)
	tap.next.ServeHTTP(ww, req)

	if !excludeAudit {
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

// isExcluded asserts if request metadata matches specified exclusions from config
func isExcluded(exclusions []*types.Exclusion, req *http.Request) bool {

	for _, exc := range exclusions {
		lcHdr := strings.ToLower(exc.HeaderName)
		// Get host or path direct from request
		if (lcHdr == "host" || lcHdr == "requesthost") && shouldExclude(req.Host, exc) {
			return true
		} else if lcHdr == "path" || lcHdr == "requestpath" {
			if url, err := url.ParseRequestURI(req.RequestURI); err == nil && shouldExclude(url.Path, exc) {
				return true
			}
		} else if shouldExclude(req.Header.Get(exc.HeaderName), exc) {
			return true
		}
	}

	return false
}

func shouldExclude(v string, exc *types.Exclusion) bool {
	return excludeValue(v, exc.StartsWith, strings.HasPrefix) ||
		excludeValue(v, exc.EndsWith, strings.HasSuffix) ||
		excludeValue(v, exc.Contains, strings.Contains) ||
		excludeMatch(v, exc.Matches)
}

func excludeValue(v string, exclusions []string, fn func(string, string) bool) bool {
	if v != "" {
		for _, x := range exclusions {
			if fn(v, x) {
				return true
			}
		}

	}
	return false
}

// TODO: Do paths need to be precompiled
func excludeMatch(v string, expressions []string) bool {
	if v != "" {
		for _, pattern := range expressions {
			if matched, err := regexp.Match(pattern, []byte(v)); err == nil && matched {
				return true
			}
		}

	}
	return false
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
