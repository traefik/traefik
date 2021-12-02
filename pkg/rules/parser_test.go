package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleMatch(t *testing.T) {
	matchers := []string{"m"}
	testCases := []struct {
		desc           string
		rule           string
		tree           Tree
		expectParseErr bool
		expectLeafErr  bool
	}{
		{
			desc:           "No rule",
			rule:           "",
			expectParseErr: true,
		},
		{
			desc:           "No matcher in rule",
			rule:           "m",
			expectParseErr: true,
		},
		{
			desc: "No value in rule",
			rule: "m()",
			tree: Tree{
				Matcher: "m",
				Value:   []string{},
			},
			expectLeafErr: true,
		},
		{
			desc:           "One value in rule with and",
			rule:           "m(`1`) &&",
			expectParseErr: true,
		},
		{
			desc:           "One value in rule with or",
			rule:           "m(`1`) ||",
			expectParseErr: true,
		},
		{
			desc:           "One value in rule with missing back tick",
			rule:           "m(`1)",
			expectParseErr: true,
		},
		{
			desc:           "One value in rule with missing opening parenthesis",
			rule:           "m(`1`))",
			expectParseErr: true,
		},
		{
			desc:           "One value in rule with missing closing parenthesis",
			rule:           "(m(`1`)",
			expectParseErr: true,
		},
		{
			desc: "One value in rule",
			rule: "m(`1`)",
			tree: Tree{
				Matcher: "m",
				Value:   []string{"1"},
			},
		},
		{
			desc: "One value in rule with superfluous parenthesis",
			rule: "(m(`1`))",
			tree: Tree{
				Matcher: "m",
				Value:   []string{"1"},
			},
		},
		{
			desc: "Rule with CAPS matcher",
			rule: "M(`1`)",
			tree: Tree{
				Matcher: "m",
				Value:   []string{"1"},
			},
		},
		{
			desc:           "Invalid matcher in rule",
			rule:           "w(`1`)",
			expectParseErr: true,
		},
		{
			desc: "Two value in rule",
			rule: "m(`1`, `2`)",
			tree: Tree{
				Matcher: "m",
				Value:   []string{"1", "2"},
			},
		},
		{
			desc: "Not one value in rule",
			rule: "!m(`1`)",
			tree: Tree{
				Matcher: "m",
				Not:     true,
				Value:   []string{"1"},
			},
		},
		{
			desc: "Two value in rule with and",
			rule: "m(`1`) && m(`2`)",
			tree: Tree{
				Matcher: "and",
				RuleLeft: &Tree{
					Matcher: "m",
					Value:   []string{"1"},
				},
				RuleRight: &Tree{
					Matcher: "m",
					Value:   []string{"2"},
				},
			},
		},
		{
			desc: "Two value in rule with or",
			rule: "m(`1`) || m(`2`)",
			tree: Tree{
				Matcher: "or",
				RuleLeft: &Tree{
					Matcher: "m",
					Value:   []string{"1"},
				},
				RuleRight: &Tree{
					Matcher: "m",
					Value:   []string{"2"},
				},
			},
		},
		{
			desc: "Two value in rule with and negated",
			rule: "!(m(`1`) && m(`2`))",
			tree: Tree{
				Matcher: "or",
				RuleLeft: &Tree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"1"},
				},
				RuleRight: &Tree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"2"},
				},
			},
		},
		{
			desc: "Two value in rule with or negated",
			rule: "!(m(`1`) || m(`2`))",
			tree: Tree{
				Matcher: "and",
				RuleLeft: &Tree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"1"},
				},
				RuleRight: &Tree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"2"},
				},
			},
		},
	}

	parser, err := NewParser(matchers)
	require.NoError(t, err)

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			parse, err := parser.Parse(test.rule)
			if test.expectParseErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			treeBuilder, ok := parse.(TreeBuilder)
			require.True(t, ok)

			tree := treeBuilder()

			assert.Equal(t, &test.tree, tree)

			checkRule(t, tree, test.expectLeafErr)
		})
	}
}

func checkRule(t *testing.T, actual *Tree, expectLeafErr bool) {
	t.Helper()

	if actual == nil {
		return
	}

	if actual.RuleLeft != nil {
		checkRule(t, actual.RuleLeft, expectLeafErr)
	}

	if actual.RuleRight != nil {
		checkRule(t, actual.RuleRight, expectLeafErr)
	}

	if actual.RuleRight == nil && actual.RuleLeft == nil && !expectLeafErr {
		assert.NoError(t, CheckRule(actual))
	} else {
		assert.Error(t, CheckRule(actual))
	}
}
