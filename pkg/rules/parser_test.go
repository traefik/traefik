package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTree = Tree + CheckErr
type testTree struct {
	Matcher   string
	Not       bool
	Value     []string
	RuleLeft  *testTree
	RuleRight *testTree

	// CheckErr allow knowing if a Tree has a rule error.
	CheckErr bool
}

func TestRuleMatch(t *testing.T) {
	matchers := []string{"m"}
	testCases := []struct {
		desc           string
		rule           string
		tree           testTree
		matchers       []string
		values         []string
		expectParseErr bool
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
			tree: testTree{
				Matcher:  "m",
				Value:    []string{},
				CheckErr: true,
			},
		},
		{
			desc: "Empty value in rule",
			rule: "m(``)",
			tree: testTree{
				Matcher:  "m",
				Value:    []string{""},
				CheckErr: true,
			},
			matchers: []string{"m"},
			values:   []string{""},
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
			tree: testTree{
				Matcher: "m",
				Value:   []string{"1"},
			},
			matchers: []string{"m"},
			values:   []string{"1"},
		},
		{
			desc: "One value in rule with superfluous parenthesis",
			rule: "(m(`1`))",
			tree: testTree{
				Matcher: "m",
				Value:   []string{"1"},
			},
			matchers: []string{"m"},
			values:   []string{"1"},
		},
		{
			desc: "Rule with CAPS matcher",
			rule: "M(`1`)",
			tree: testTree{
				Matcher: "m",
				Value:   []string{"1"},
			},
			matchers: []string{"m"},
			values:   []string{"1"},
		},
		{
			desc:           "Invalid matcher in rule",
			rule:           "w(`1`)",
			expectParseErr: true,
		},
		{
			desc: "Invalid matchers",
			rule: "m(`1`)",
			tree: testTree{
				Matcher: "m",
				Value:   []string{"1"},
			},
			matchers: []string{"not-m"},
		},
		{
			desc: "Two value in rule",
			rule: "m(`1`, `2`)",
			tree: testTree{
				Matcher: "m",
				Value:   []string{"1", "2"},
			},
			matchers: []string{"m"},
			values:   []string{"1", "2"},
		},
		{
			desc: "Not one value in rule",
			rule: "!m(`1`)",
			tree: testTree{
				Matcher: "m",
				Not:     true,
				Value:   []string{"1"},
			},
			matchers: []string{"m"},
			values:   []string{"1"},
		},
		{
			desc: "Two value in rule with and",
			rule: "m(`1`) && m(`2`)",
			tree: testTree{
				Matcher:  "and",
				CheckErr: true,
				RuleLeft: &testTree{
					Matcher: "m",
					Value:   []string{"1"},
				},
				RuleRight: &testTree{
					Matcher: "m",
					Value:   []string{"2"},
				},
			},
			matchers: []string{"m"},
			values:   []string{"1", "2"},
		},
		{
			desc: "Two value in rule with or",
			rule: "m(`1`) || m(`2`)",
			tree: testTree{
				Matcher:  "or",
				CheckErr: true,
				RuleLeft: &testTree{
					Matcher: "m",
					Value:   []string{"1"},
				},
				RuleRight: &testTree{
					Matcher: "m",
					Value:   []string{"2"},
				},
			},
			matchers: []string{"m"},
			values:   []string{"1", "2"},
		},
		{
			desc: "Two value in rule with and negated",
			rule: "!(m(`1`) && m(`2`))",
			tree: testTree{
				Matcher:  "or",
				CheckErr: true,
				RuleLeft: &testTree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"1"},
				},
				RuleRight: &testTree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"2"},
				},
			},
			matchers: []string{"m"},
			values:   []string{"1", "2"},
		},
		{
			desc: "Two value in rule with or negated",
			rule: "!(m(`1`) || m(`2`))",
			tree: testTree{
				Matcher:  "and",
				CheckErr: true,
				RuleLeft: &testTree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"1"},
				},
				RuleRight: &testTree{
					Matcher: "m",
					Not:     true,
					Value:   []string{"2"},
				},
			},
			matchers: []string{"m"},
			values:   []string{"1", "2"},
		},
		{
			desc: "No value in rule",
			rule: "m(`1`) && m()",
			tree: testTree{
				Matcher:  "and",
				CheckErr: true,
				RuleLeft: &testTree{
					Matcher: "m",
					Value:   []string{"1"},
				},
				RuleRight: &testTree{
					Matcher:  "m",
					Value:    []string{},
					CheckErr: true,
				},
			},
			matchers: []string{"m"},
			values:   []string{"1"},
		},
	}

	parser, err := NewParser(matchers)
	require.NoError(t, err)

	for _, test := range testCases {
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
			checkEquivalence(t, &test.tree, tree)

			assert.Equal(t, test.values, tree.ParseMatchers(test.matchers))
		})
	}
}

func checkEquivalence(t *testing.T, expected *testTree, actual *Tree) {
	t.Helper()

	if actual == nil {
		return
	}

	if actual.RuleLeft != nil {
		checkEquivalence(t, expected.RuleLeft, actual.RuleLeft)
	}

	if actual.RuleRight != nil {
		checkEquivalence(t, expected.RuleRight, actual.RuleRight)
	}

	assert.Equal(t, expected.Matcher, actual.Matcher)
	assert.Equal(t, expected.Not, actual.Not)
	assert.Equal(t, expected.Value, actual.Value)

	t.Logf("%+v -> %v", actual, CheckRule(actual))
	if expected.CheckErr {
		assert.Error(t, CheckRule(actual))
	} else {
		assert.NoError(t, CheckRule(actual))
	}
}
