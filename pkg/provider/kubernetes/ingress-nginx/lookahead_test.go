package ingressnginx

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	netv1 "k8s.io/api/networking/v1"
)

func Test_splitNegativeLookahead(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc            string
		path            string
		expectedKeep    string
		expectedExclude string
		expectedOK      bool
	}{
		{
			desc:            "single excluded prefix",
			path:            `/api/licensing/((?!_internal).*)`,
			expectedKeep:    `/api/licensing/(.*)`,
			expectedExclude: `/api/licensing/(?:_internal)`,
			expectedOK:      true,
		},
		{
			desc:            "several excluded prefixes",
			path:            `/api/example/v1/((?!invitations|session|self-registration).*)`,
			expectedKeep:    `/api/example/v1/(.*)`,
			expectedExclude: `/api/example/v1/(?:invitations|session|self-registration)`,
			expectedOK:      true,
		},
		{
			desc:            "excluded prefix spanning several segments",
			path:            `/api/configuration/v1/((?!administration/manage).*)`,
			expectedKeep:    `/api/configuration/v1/(.*)`,
			expectedExclude: `/api/configuration/v1/(?:administration/manage)`,
			expectedOK:      true,
		},
		{
			desc:       "no lookahead to translate",
			path:       `/api/example/(.*)`,
			expectedOK: false,
		},
		{
			desc:       "R1: several negative lookaheads",
			path:       `/a/((?!b).*)/((?!c).*)`,
			expectedOK: false,
		},
		{
			desc:       "R2: positive lookahead is not the negation of a match",
			path:       `/a/((?=b).*)`,
			expectedOK: false,
		},
		{
			desc:       "R2: negative lookbehind would anchor from the right",
			path:       `/a/((?<!b).*)`,
			expectedOK: false,
		},
		{
			desc:       "R3: lookahead does not open the capture group",
			path:       `/a/(x(?!y).*)`,
			expectedOK: false,
		},
		{
			desc:       "R4: quantifier makes the prefix variable-length, so the exclusion offset is not determinate",
			path:       `/a/.*((?!b).*)`,
			expectedOK: false,
		},
		{
			desc:       "R4: alternation in the prefix is sound but narrowed out of this translation",
			path:       `/a(b|c)/((?!d).*)`,
			expectedOK: false,
		},
		{
			desc:       "R5: capture group in the alternation",
			path:       `/a/((?!(b|c)).*)`,
			expectedOK: false,
		},
		{
			desc:       "R5: capture group in the alternation would renumber the tail group",
			path:       `/a/((?!(b|c))(.*))`,
			expectedOK: false,
		},
		{
			desc:       "R6: empty alternation excludes every path",
			path:       `/a/((?!).*)`,
			expectedOK: false,
		},
		{
			desc:       "R6: empty alternative excludes every path",
			path:       `/a/((?!b|).*)`,
			expectedOK: false,
		},
		{
			desc:       "R7: top-level alternation in the tail bypasses the assertion",
			path:       `/a/((?!b)c|b)`,
			expectedOK: false,
		},
		{
			desc:       "R7: character class tail is outside the supported shape",
			path:       `/a/((?!b)[^/]*)`,
			expectedOK: false,
		},
		{
			desc:       "R7: one-or-more tail is outside the supported shape",
			path:       `/a/((?!b).+)`,
			expectedOK: false,
		},
		{
			desc:       "R7: trailing literal inside the outer capture",
			path:       `/a/((?!b)c)`,
			expectedOK: false,
		},
		{
			desc:       "R7: nested group in the tail",
			path:       `/a/((?!b)(.*))`,
			expectedOK: false,
		},
		{
			desc:       "R7: quantifier after the outer capture makes the guarded group optional",
			path:       `/a/((?!b).*)?b$`,
			expectedOK: false,
		},
		{
			desc:       "R7: trailing literal after the outer capture",
			path:       `/a/((?!b).*)/c`,
			expectedOK: false,
		},
		{
			desc:       "unbalanced lookahead",
			path:       `/a/((?!b`,
			expectedOK: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			keep, exclude, ok := splitNegativeLookahead(test.path)

			assert.Equal(t, test.expectedOK, ok)
			assert.Equal(t, test.expectedKeep, keep)
			assert.Equal(t, test.expectedExclude, exclude)
		})
	}
}

// Test_splitNegativeLookahead_matchesNGINXSemantics checks the translated halves
// against the routing decision ingress-nginx would make, including the cases
// where a segment-aware or case-sensitive path matcher would diverge from
// nginx's character-level, case-insensitive assertion.
func Test_splitNegativeLookahead_matchesNGINXSemantics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		path          string
		requestPath   string
		expectedMatch bool
		expectedGroup string
	}{
		{
			desc:          "path below the general route is served",
			path:          `/api/licensing/((?!_internal).*)`,
			requestPath:   "/api/licensing/public",
			expectedMatch: true,
			expectedGroup: "public",
		},
		{
			desc:        "excluded prefix is not served",
			path:        `/api/licensing/((?!_internal).*)`,
			requestPath: "/api/licensing/_internal",
		},
		{
			desc:        "path below the excluded prefix is not served",
			path:        `/api/licensing/((?!_internal).*)`,
			requestPath: "/api/licensing/_internal/keys",
		},
		{
			desc:        "exclusion is a character prefix, not a path segment",
			path:        `/api/licensing/((?!_internal).*)`,
			requestPath: "/api/licensing/_internalfoo",
		},
		{
			desc:          "exclusion is anchored at the prefix, not searched for",
			path:          `/api/licensing/((?!_internal).*)`,
			requestPath:   "/api/licensing/x/_internal",
			expectedMatch: true,
			expectedGroup: "x/_internal",
		},
		{
			desc:          "unexcluded path is served",
			path:          `/api/example/v1/((?!invitations|session|self-registration).*)`,
			requestPath:   "/api/example/v1/customers",
			expectedMatch: true,
			expectedGroup: "customers",
		},
		{
			desc:        "excluded alternative is not served",
			path:        `/api/example/v1/((?!invitations|session|self-registration).*)`,
			requestPath: "/api/example/v1/session/current",
		},
		{
			desc:        "path merely starting with an excluded alternative is not served",
			path:        `/api/example/v1/((?!invitations|session|self-registration).*)`,
			requestPath: "/api/example/v1/sessions-list",
		},
		{
			desc:        "exclusion is case-insensitive, as ingress-nginx regex locations are",
			path:        `/api/example/v1/((?!invitations|session|self-registration).*)`,
			requestPath: "/api/example/v1/SESSION/current",
		},
		{
			desc:          "sibling of a multi-segment exclusion is served",
			path:          `/api/configuration/v1/((?!administration/manage).*)`,
			requestPath:   "/api/configuration/v1/administration",
			expectedMatch: true,
			expectedGroup: "administration",
		},
		{
			desc:        "path merely starting with a multi-segment exclusion is not served",
			path:        `/api/configuration/v1/((?!administration/manage).*)`,
			requestPath: "/api/configuration/v1/administration/manageXYZ",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			keep, exclude, ok := splitNegativeLookahead(test.path)
			require.True(t, ok)

			keepRegexp := regexp.MustCompile(nginxRegexPrefix + keep)
			excludeRegexp := regexp.MustCompile(nginxRegexPrefix + exclude)

			// Mirrors the rule a caller builds: PathRegexp(keep) && !PathRegexp(exclude).
			match := keepRegexp.MatchString(test.requestPath) && !excludeRegexp.MatchString(test.requestPath)
			assert.Equal(t, test.expectedMatch, match)

			if !test.expectedMatch {
				return
			}

			// The capture group rewrite-target replacements refer to must survive
			// the removal of the assertion.
			groups := keepRegexp.FindStringSubmatch(test.requestPath)
			require.Len(t, groups, 2)
			assert.Equal(t, test.expectedGroup, groups[1])
		})
	}
}

func Test_matchingParen(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		expr     string
		start    int
		expected int
	}{
		{
			desc:     "adjacent parentheses",
			expr:     `(ab)`,
			expected: 3,
		},
		{
			desc:     "nested parentheses are counted",
			expr:     `(a(b)c)`,
			expected: 6,
		},
		{
			desc:     "escaped parenthesis does not change the depth",
			expr:     `(a\)b)`,
			expected: 5,
		},
		{
			desc:     "parenthesis inside a character class is literal",
			expr:     `(a[)]b)`,
			expected: 6,
		},
		{
			desc:     "unbalanced",
			expr:     `(ab`,
			expected: -1,
		},
		{
			desc:     "scan starts at the given index",
			expr:     `xx(ab)`,
			start:    2,
			expected: 5,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, matchingParen(test.expr, test.start))
		})
	}
}

func Test_buildRule_negativeLookahead(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc                string
		path                string
		expectedRule        string
		expectedPreNegation string
	}{
		{
			desc:                "path without a lookahead is left verbatim",
			path:                `/api/example/(.*)`,
			expectedRule:        `Host("example.localhost") && PathRegexp("(?i)^/api/example/(.*)")`,
			expectedPreNegation: `Host("example.localhost") && PathRegexp("(?i)^/api/example/(.*)")`,
		},
		{
			desc:                "lookahead becomes a keep arm and a negated exclude arm",
			path:                `/api/licensing/((?!_internal).*)`,
			expectedRule:        `Host("example.localhost") && PathRegexp("(?i)^/api/licensing/(.*)") && !PathRegexp("(?i)^/api/licensing/(?:_internal)")`,
			expectedPreNegation: `Host("example.localhost") && PathRegexp("(?i)^/api/licensing/((?!_internal).*)")`,
		},
		{
			desc:                "several excluded prefixes stay in a single negated arm",
			path:                `/api/example/v1/((?!invitations|session|self-registration).*)`,
			expectedRule:        `Host("example.localhost") && PathRegexp("(?i)^/api/example/v1/(.*)") && !PathRegexp("(?i)^/api/example/v1/(?:invitations|session|self-registration)")`,
			expectedPreNegation: `Host("example.localhost") && PathRegexp("(?i)^/api/example/v1/((?!invitations|session|self-registration).*)")`,
		},
		{
			desc:                "unsupported shape is left verbatim",
			path:                `/a/((?!b).*)/((?!c).*)`,
			expectedRule:        `Host("example.localhost") && PathRegexp("(?i)^/a/((?!b).*)/((?!c).*)")`,
			expectedPreNegation: `Host("example.localhost") && PathRegexp("(?i)^/a/((?!b).*)/((?!c).*)")`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loc := &location{
				Path:     test.path,
				UseRegex: true,
			}
			resolveNegativeLookahead(loc)

			rule, preNegation := buildRule("example.localhost", loc)

			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPreNegation, preNegation)
		})
	}
}

// Test_buildRule_priorityPinning covers why buildRule returns a pre-negation rule
// at all. Traefik derives router priority from rule length, so a translated
// router must not be left to score its own two-armed rule.
func Test_buildRule_priorityPinning(t *testing.T) {
	t.Parallel()

	loc := &location{
		Path:     `/api/licensing/((?!_internal).*)`,
		UseRegex: true,
	}
	resolveNegativeLookahead(loc)

	rule, preNegation := buildRule("example.localhost", loc)

	// The negated arm lengthens the rule, which is the inflation being avoided.
	require.Greater(t, httpmuxer.GetRulePriority(rule), httpmuxer.GetRulePriority(preNegation))

	// Deriving each router's priority from its own pre-negation rule keeps a canary
	// router ranked above the base router it was built from.
	canary := &canaryConfig{Header: "X-Canary"}
	assert.Greater(t,
		httpmuxer.GetRulePriority(appendCanaryRule(preNegation, canary)),
		httpmuxer.GetRulePriority(preNegation))
}

// Test_buildRedirect_negativeLookahead covers the third compile site. The
// redirect regex is built from the raw path, so a lookahead location combined
// with a redirect annotation would produce a router whose RedirectRegex fails
// to compile at runtime.
func Test_buildRedirect_negativeLookahead(t *testing.T) {
	t.Parallel()

	loc := &location{
		Path:     `/api/licensing/((?!_internal).*)`,
		UseRegex: true,
		Config: IngressConfig{
			PermanentRedirect: new("https://example.localhost/moved"),
		},
	}

	resolveNegativeLookahead(loc)

	(&Provider{}).buildRedirect(loc)

	require.NotNil(t, loc.Redirect)
	assert.Equal(t, `^https?://[^/]+/api/licensing/(.*)`, loc.Redirect.Regex)

	_, err := regexp.Compile(loc.Redirect.Regex)
	require.NoError(t, err)
}

// Test_resolveNegativeLookahead_absoluteRewriteTarget pins the one shape that carries
// a translatable lookahead and is still deliberately left alone. An absolute
// rewrite-target on an ImplementationSpecific path widens the path before the
// translation sees it, which replaces the literal prefix the exclusion arm would have
// anchored to, so the translation declines. Go's regexp then cannot compile the path
// and the location is dropped, which is the safe outcome: the alternative would be
// serving the prefixes the assertion excluded.
func Test_resolveNegativeLookahead_absoluteRewriteTarget(t *testing.T) {
	t.Parallel()

	loc := &location{
		Path:     `/api/licensing/((?!_internal).*)`,
		PathType: new(netv1.PathTypeImplementationSpecific),
		UseRegex: true,
		Config: IngressConfig{
			RewriteTarget: new("https://backend.localhost/$1"),
		},
	}

	// Widening runs first, so what the translation is offered is no longer the path
	// the ingress declared, and its prefix is no longer a literal.
	require.Equal(t, `/api/licensing(?:/((?!_internal).*))?`, pathRegexp(loc))

	resolveNegativeLookahead(loc)

	assert.Empty(t, loc.PathKeep)
	assert.Empty(t, loc.PathExclude)

	// The router rule keeps the assertion, so it is scored as its own pre-negation
	// rule and has nothing to negate.
	rule, preNegation := buildRule("example.localhost", loc)

	assert.Equal(t, preNegation, rule)
	assert.Contains(t, rule, `(?!_internal)`)

	// So does the rewrite-target regex, which is the other half of the combination.
	(&Provider{}).buildRewriteTarget(loc)

	require.NotNil(t, loc.RewriteTarget)
	assert.Contains(t, loc.RewriteTarget.Regex, `(?!_internal)`)

	// Neither compiles under RE2, which is what makes the combination unsupported.
	_, err := regexp.Compile(nginxRegexPrefix + pathRegexp(loc))
	require.Error(t, err)
}
