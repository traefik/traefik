package compress

import (
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
	Weight *float64
}

func getCompressionType(acceptEncoding []string, defaultType string) string {
	if defaultType == "" {
		// Keeps the pre-existing default inside Traefik.
		defaultType = brotliName
	}

	encodings, hasWeight := parseAcceptEncoding(acceptEncoding)

	if hasWeight {
		if len(encodings) == 0 {
			return identityName
		}

		encoding := encodings[0]

		if encoding.Type == identityName && encoding.Weight != nil && *encoding.Weight == 0 {
			return notAcceptable
		}

		if encoding.Type == wildcardName && encoding.Weight != nil && *encoding.Weight == 0 {
			return notAcceptable
		}

		if encoding.Type == wildcardName {
			return defaultType
		}

		return encoding.Type
	}

	for _, dt := range []string{zstdName, brotliName, gzipName} {
		if slices.ContainsFunc(encodings, func(e Encoding) bool { return e.Type == dt }) {
			return dt
		}
	}

	if slices.ContainsFunc(encodings, func(e Encoding) bool { return e.Type == wildcardName }) {
		return defaultType
	}

	return identityName
}

func parseAcceptEncoding(acceptEncoding []string) ([]Encoding, bool) {
	var encodings []Encoding
	var hasWeight bool

	for _, line := range acceptEncoding {
		for _, item := range strings.Split(strings.ReplaceAll(line, " ", ""), ",") {
			parsed := strings.SplitN(item, ";", 2)
			if len(parsed) == 0 {
				continue
			}

			switch parsed[0] {
			case zstdName, brotliName, gzipName, identityName, wildcardName:
				// supported encoding
			default:
				continue
			}

			var weight *float64
			if len(parsed) > 1 && strings.HasPrefix(parsed[1], "q=") {
				w, _ := strconv.ParseFloat(strings.TrimPrefix(parsed[1], "q="), 64)

				weight = &w
				hasWeight = true
			}

			encodings = append(encodings, Encoding{
				Type:   parsed[0],
				Weight: weight,
			})
		}
	}

	slices.SortFunc(encodings, compareEncoding)

	return encodings, hasWeight
}

func compareEncoding(a, b Encoding) int {
	lhs, rhs := a.Weight, b.Weight

	if lhs == nil && rhs == nil {
		return 0
	}

	if lhs == nil && *rhs == 0 {
		return -1
	}

	if lhs == nil {
		return 1
	}

	if rhs == nil && *lhs == 0 {
		return 1
	}

	if rhs == nil {
		return -1
	}

	if *lhs < *rhs {
		return 1
	}

	if *lhs > *rhs {
		return -1
	}

	return 0
}
