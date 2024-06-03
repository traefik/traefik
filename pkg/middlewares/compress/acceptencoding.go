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
	identityName  = "identity"
	wildcardName  = "*"
	notAcceptable = "not_acceptable"
)

type Encoding struct {
	Type   string
	Weight *float64
}

func getCompressionType(acceptEncoding []string, defaultType string) string {
	accepts, hasWeight := parseAcceptsEncoding(acceptEncoding)

	if hasWeight {
		for _, a := range accepts {
			if a.Type == identityName && a.Weight != nil && *a.Weight == 0 {
				return notAcceptable
			}

			if a.Type == wildcardName && a.Weight != nil && *a.Weight == 0 {
				return notAcceptable
			}

			if a.Type == wildcardName {
				continue
			}

			return a.Type
		}

		if defaultType != "" {
			return defaultType
		}

		// Follows the pre-existing default order inside Traefik: br > gzip.
		return brotliName
	}

	// fallback on pre-existing order inside Traefik
	defaultTypes := []string{defaultType, brotliName, gzipName}

	for _, dt := range defaultTypes {
		if dt == "" {
			continue
		}

		if slices.ContainsFunc(accepts, func(e Encoding) bool {
			return e.Type == dt || e.Type == wildcardName
		}) {
			return dt
		}
	}

	return defaultType
}

func parseAcceptsEncoding(acceptEncoding []string) ([]Encoding, bool) {
	var values []Encoding
	var hasWeight bool

	for _, ae := range acceptEncoding {
		for _, e := range strings.Split(strings.ReplaceAll(ae, " ", ""), ",") {
			parsed := strings.SplitN(strings.TrimSpace(e), ";", 2)
			if len(parsed) == 0 {
				continue
			}

			switch parsed[0] {
			case "br", "gzip", "identity", "*":
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

			values = append(values, Encoding{
				Type:   parsed[0],
				Weight: weight,
			})
		}
	}

	slices.SortFunc(values, func(a, b Encoding) int {
		return floatCompare(a.Weight, b.Weight)
	})

	return values, hasWeight
}

func floatCompare(lhs, rhs *float64) int {
	if lhs == nil && rhs == nil {
		return 0
	}

	if lhs == nil {
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
