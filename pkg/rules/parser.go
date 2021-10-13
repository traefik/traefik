package rules

import (
	"strings"

	"github.com/vulcand/predicate"
)

const (
	and = "and"
	or  = "or"
)

type TreeBuilder func() *Tree

type Tree struct {
	Matcher   string
	Not       bool
	Value     []string
	RuleLeft  *Tree
	RuleRight *Tree
}

func NewParser(matchers []string) (predicate.Parser, error) {
	parserFuncs := make(map[string]interface{})

	for _, matcherName := range matchers {
		matcherName := matcherName
		fn := func(value ...string) TreeBuilder {
			return func() *Tree {
				return &Tree{
					Matcher: matcherName,
					Value:   value,
				}
			}
		}
		parserFuncs[matcherName] = fn
		parserFuncs[strings.ToLower(matcherName)] = fn
		parserFuncs[strings.ToUpper(matcherName)] = fn
		parserFuncs[strings.Title(strings.ToLower(matcherName))] = fn
	}

	return predicate.NewParser(predicate.Def{
		Operators: predicate.Operators{
			AND: andFunc,
			OR:  orFunc,
			NOT: notFunc,
		},
		Functions: parserFuncs,
	})
}

func andFunc(left, right TreeBuilder) TreeBuilder {
	return func() *Tree {
		return &Tree{
			Matcher:   and,
			RuleLeft:  left(),
			RuleRight: right(),
		}
	}
}

func orFunc(left, right TreeBuilder) TreeBuilder {
	return func() *Tree {
		return &Tree{
			Matcher:   or,
			RuleLeft:  left(),
			RuleRight: right(),
		}
	}
}

func invert(t *Tree) *Tree {
	switch t.Matcher {
	case or:
		t.Matcher = and
		t.RuleLeft = invert(t.RuleLeft)
		t.RuleRight = invert(t.RuleRight)
	case and:
		t.Matcher = or
		t.RuleLeft = invert(t.RuleLeft)
		t.RuleRight = invert(t.RuleRight)
	default:
		t.Not = !t.Not
	}

	return t
}

func notFunc(elem TreeBuilder) TreeBuilder {
	return func() *Tree {
		return invert(elem())
	}
}

func (tree *Tree) ParseMatchers(matchers []string) []string {
	switch tree.Matcher {
	case and, or:
		return append(tree.RuleLeft.ParseMatchers(matchers), tree.RuleRight.ParseMatchers(matchers)...)
	default:
		for _, matcher := range matchers {
			if tree.Matcher == matcher {
				return lower(tree.Value)
			}
		}
		return nil
	}
}

func lower(slice []string) []string {
	var lowerStrings []string
	for _, value := range slice {
		lowerStrings = append(lowerStrings, strings.ToLower(value))
	}
	return lowerStrings
}
