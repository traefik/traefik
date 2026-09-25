package ingressnginx

import (
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// nginxRegexPrefix is applied to every ingress-nginx path compiled as a regular
// expression: ingress-nginx matches regex locations case-insensitively (the `~*`
// operator) and anchors them at the start of the path.
const nginxRegexPrefix = "(?i)^"

// pathLiteralClass contains the supported literal characters; regex metacharacters and escapes are excluded.
const pathLiteralClass = `A-Za-z0-9_~/-`

var (
	// literalPathPrefix matches a supported literal prefix starting with a slash.
	literalPathPrefix = regexp.MustCompile(`^/[` + pathLiteralClass + `]*$`)

	// literalAlternation matches non-empty literals separated by "|".
	literalAlternation = regexp.MustCompile(`^[` + pathLiteralClass + `]+(?:\|[` + pathLiteralClass + `]+)*$`)
)

// splitNegativeLookahead translates only X((?!A|B|C).*) with a literal path
// prefix and non-empty literal alternatives. The caller combines keep && !exclude.
// Unsupported patterns return ok=false and must be left untranslated.
func splitNegativeLookahead(path string) (keep, exclude string, ok bool) {
	prefix, rest, found := strings.Cut(path, "((?!")
	if !found || !literalPathPrefix.MatchString(prefix) {
		return "", "", false
	}

	alternation, tail, found := strings.Cut(rest, ")")
	if !found || !literalAlternation.MatchString(alternation) || tail != ".*)" {
		return "", "", false
	}

	// Retain the outer capture for rewrite-target references.
	keep = prefix + "(.*)"
	exclude = prefix + "(?:" + alternation + ")"

	for _, expr := range []string{nginxRegexPrefix + keep, nginxRegexPrefix + exclude} {
		if _, err := regexp.Compile(expr); err != nil {
			return "", "", false
		}
	}

	return keep, exclude, true
}

// resolveNegativeLookahead stores one translation for all consumers without changing Path.
// Paths widened by makeTrailingGroupOptional fall outside the supported grammar.
func resolveNegativeLookahead(loc *location) {
	if !loc.UseRegex || loc.Path == "" {
		return
	}

	keep, exclude, ok := splitNegativeLookahead(pathRegexp(loc))
	if !ok {
		if strings.Contains(loc.Path, "(?!") {
			log.Warn().Msgf("Unsupported negative lookahead in path %q of ingress %s/%s, the path is used as-is and Go's regexp cannot compile it.", loc.Path, loc.Namespace, loc.IngressName)
		}

		return
	}

	loc.PathKeep, loc.PathExclude = keep, exclude
}
