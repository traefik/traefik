package compress

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
)

const acceptEncodingHeader = "Accept-Encoding"

const (
	brotliName    = "br"
	gzipName      = "gzip"
	zstdName      = "zstd"
	identityName  = "identity"
	wildcardName  = "*"
	notAcceptable = "not_acceptable"
)

type Encoding struct {
	Type   string
	Weight float64
}

func getCompressionEncoding(acceptEncoding []string, defaultEncoding string, supportedEncodings []string) string {
	if defaultEncoding == "" {
		if slices.Contains(supportedEncodings, brotliName) {
			// Keeps the pre-existing default inside Traefik if brotli is a supported encoding.
			defaultEncoding = brotliName
		} else if len(supportedEncodings) > 0 {
			// Otherwise use the first supported encoding.
			defaultEncoding = supportedEncodings[0]
		}
	}

	encodings, hasWeight := parseAcceptEncoding(acceptEncoding, supportedEncodings)

	if hasWeight {
		if len(encodings) == 0 {
			return identityName
		}

		encoding := encodings[0]

		if encoding.Type == identityName && encoding.Weight == 0 {
			return notAcceptable
		}

		if encoding.Type == wildcardName && encoding.Weight == 0 {
			return notAcceptable
		}

		if encoding.Type == wildcardName {
			return defaultEncoding
		}

		return encoding.Type
	}

	for _, dt := range supportedEncodings {
		if slices.ContainsFunc(encodings, func(e Encoding) bool { return e.Type == dt }) {
			return dt
		}
	}

	if slices.ContainsFunc(encodings, func(e Encoding) bool { return e.Type == wildcardName }) {
		return defaultEncoding
	}

	return identityName
}

func parseAcceptEncoding(acceptEncoding, supportedEncodings []string) ([]Encoding, bool) {
	var encodings []Encoding
	var hasWeight bool

	for _, line := range acceptEncoding {
		for _, item := range strings.Split(strings.ReplaceAll(line, " ", ""), ",") {
			parsed := strings.SplitN(item, ";", 2)
			if len(parsed) == 0 {
				continue
			}

			if !slices.Contains(supportedEncodings, parsed[0]) &&
				parsed[0] != identityName &&
				parsed[0] != wildcardName {
				continue
			}

			// If no "q" parameter is present, the default weight is 1.
			// https://www.rfc-editor.org/rfc/rfc9110.html#name-quality-values
			weight := 1.0
			if len(parsed) > 1 && strings.HasPrefix(parsed[1], "q=") {
				w, _ := strconv.ParseFloat(strings.TrimPrefix(parsed[1], "q="), 64)

				weight = w
				hasWeight = true
			}

			encodings = append(encodings, Encoding{
				Type:   parsed[0],
				Weight: weight,
			})
		}
	}

	slices.SortFunc(encodings, func(a, b Encoding) int {
		return cmp.Compare(b.Weight, a.Weight)
	})

	return encodings, hasWeight
}
