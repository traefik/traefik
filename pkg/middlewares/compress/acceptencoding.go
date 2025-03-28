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

func (c *compress) getCompressionEncoding(acceptEncoding []string) string {
	// RFC says: An Accept-Encoding header field with a field value that is empty implies that the user agent does not want any content coding in response.
	// https://datatracker.ietf.org/doc/html/rfc9110#name-accept-encoding
	if len(acceptEncoding) == 1 && acceptEncoding[0] == "" {
		return identityName
	}

	acceptableEncodings := parseAcceptableEncodings(acceptEncoding, c.supportedEncodings)

	// An empty Accept-Encoding header field would have been handled earlier.
	// If empty, it means no encoding is supported, we do not encode.
	if len(acceptableEncodings) == 0 {
		// TODO: return 415 status code instead of deactivating the compression, if the backend was not to compress as well.
		return notAcceptable
	}

	slices.SortFunc(acceptableEncodings, func(a, b Encoding) int {
		if a.Weight == b.Weight {
			// At same weight, we want to prioritize based on the encoding priority.
			// the lower the index, the higher the priority.
			return cmp.Compare(c.supportedEncodings[a.Type], c.supportedEncodings[b.Type])
		}
		return cmp.Compare(b.Weight, a.Weight)
	})

	if acceptableEncodings[0].Type == wildcardName {
		if c.defaultEncoding == "" {
			return c.encodings[0]
		}

		return c.defaultEncoding
	}

	return acceptableEncodings[0].Type
}

func parseAcceptableEncodings(acceptEncoding []string, supportedEncodings map[string]int) []Encoding {
	var encodings []Encoding

	for _, line := range acceptEncoding {
		for _, item := range strings.Split(strings.ReplaceAll(line, " ", ""), ",") {
			parsed := strings.SplitN(item, ";", 2)
			if len(parsed) == 0 {
				continue
			}

			if _, ok := supportedEncodings[parsed[0]]; !ok {
				continue
			}

			// If no "q" parameter is present, the default weight is 1.
			// https://www.rfc-editor.org/rfc/rfc9110.html#name-quality-values
			weight := 1.0
			if len(parsed) > 1 && strings.HasPrefix(parsed[1], "q=") {
				w, _ := strconv.ParseFloat(strings.TrimPrefix(parsed[1], "q="), 64)

				// If the weight is 0, the encoding is not acceptable.
				// We can skip the encoding.
				if w == 0 {
					continue
				}

				weight = w
			}

			encodings = append(encodings, Encoding{
				Type:   parsed[0],
				Weight: weight,
			})
		}
	}

	return encodings
}
