package audittap

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"github.com/containous/traefik/types"
	"github.com/satori/go.uuid"
)

// MaximumEntityLength sets the upper limit for request and response entities. This will
// probably be removed in future versions.
const MaximumEntityLength = 32 * 1024

// AuditTap writes an event to the audit streams for every request.
type AuditTap struct {
	AuditSource     string
	AuditType       string
	AuditStreams    []audittypes.AuditStream
	Backend         string
	MaxEntityLength int
	next            http.Handler
}

// NewAuditTap returns a new AuditTap handler.
func NewAuditTap(config *types.AuditSink, streams []audittypes.AuditStream, backend string, next http.Handler) (*AuditTap, error) {
	var th int64 = MaximumEntityLength
	var err error
	if config.MaxEntityLength != "" {
		th, _, err = asSI(config.MaxEntityLength)
		if err != nil {
			return nil, err
		}
	}

	if config.AuditSource == "" {
		return nil, fmt.Errorf("AuditSource not set in configuration")
	}

	if config.AuditType == "" {
		return nil, fmt.Errorf("AuditType not set in configuration")
	}

	return &AuditTap{config.AuditSource, config.AuditType, streams, backend, int(th), next}, nil
}

func (s *AuditTap) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	hdr := NewHeaders(r.Header).SimplifyCookies()

	// Need to create a URL from the RequestURI, because the URL in the request is overwritten
	// by oxy's RoundRobin and loses Path and RawQuery
	u, _ := url.ParseRequestURI(r.RequestURI)

	clientHeaders, requestHeaders := hdr.ClientAndRequestHeaders()
	flatHdr := hdr.Flatten()

	requestPayload := audittypes.DataMap{
		"type": flatHdr.GetString("content-type"),
	}

	// Create the summary and populate with all the request info we need
	summary := audittypes.Summary{
		AuditSource:        s.AuditSource,
		AuditType:          s.AuditType,
		EventID:            uuid.NewV4().String(),
		GeneratedAt:        audittypes.TheClock.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		Version:            "1",
		RequestID:          flatHdr.GetString("x-request-id"),
		Method:             r.Method,
		Path:               u.Path,
		QueryString:        u.RawQuery,
		ClientIP:           flatHdr.GetString("true-client-ip"),
		ClientPort:         flatHdr.GetString("true-client-port"),
		ReceivingIP:        flatHdr.GetString("x-source"),
		AuthorisationToken: flatHdr.GetString("authorization"),
		ClientHeaders:      clientHeaders,
		RequestHeaders:     requestHeaders,
		RequestPayload:     requestPayload,
		ResponseHeaders:    nil,
		ResponseStatus:     "",
		ResponsePayload:    nil,
	}

	ww := NewAuditResponseWriter(rw, s.MaxEntityLength)
	s.next.ServeHTTP(ww, r)

	// Now add in the response info
	ww.SummariseResponse(&summary)

	for _, sink := range s.AuditStreams {
		sink.Audit(summary)
	}
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
