package rules

import (
	"errors"
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

// ParseDomains extract domains from rule.
func ParseDomains(rule string) ([]string, error) {
	parser, err := newParser()
	if err != nil {
		return nil, err
	}

	parse, err := parser.Parse(rule)
	if err != nil {
		return nil, err
	}

	buildTree, ok := parse.(TreeBuilder)
	if !ok {
		return nil, errors.New("cannot parse")
	}

	return lower(parseDomain(buildTree())), nil
}

// ParseHostSNI extracts the HostSNIs declared in a rule.
// This is a first naive implementation used in TCP routing.
func ParseHostSNI(rule string) ([]string, error) {
	parser, err := NewTCPParser()
	if err != nil {
		return nil, err
	}

	parse, err := parser.Parse(rule)
	if err != nil {
		return nil, err
	}

	buildTree, ok := parse.(TreeBuilder)
	if !ok {
		return nil, errors.New("cannot parse")
	}

	return lower(parseDomain(buildTree())), nil
}

//// ParseClientIP extracts the ClientIPs declared in a rule.
//// This is a first naive implementation used in TCP routing.
//func ParseClientIP(rule string) ([]string, error) {
//	parser, err := newParser()
//	if err != nil {
//		return nil, err
//	}
//
//	parse, err := parser.Parse(rule)
//	if err != nil {
//		return nil, err
//	}
//
//	buildTree, ok := parse.(TreeBuilder)
//	if !ok {
//		return nil, errors.New("cannot parse")
//	}
//
//	return lower(parseClientIP(buildTree())), nil
//}

func lower(slice []string) []string {
	var lowerStrings []string
	for _, value := range slice {
		lowerStrings = append(lowerStrings, strings.ToLower(value))
	}
	return lowerStrings
}

func parseDomain(tree *Tree) []string {
	switch tree.Matcher {
	case and, or:
		return append(parseDomain(tree.RuleLeft), parseDomain(tree.RuleRight)...)
	case "Host", "HostSNI":
		return tree.Value
	default:
		return nil
	}
}

//func parseClientIP(tree *Tree) []string {
//	switch tree.Matcher {
//	case and, or:
//		return append(parseClientIP(tree.RuleLeft), parseClientIP(tree.RuleRight)...)
//	case "ClientIP":
//		return tree.Value
//	default:
//		return nil
//	}
//}

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

func newParser() (predicate.Parser, error) {
	parserFuncs := make(map[string]interface{})

	for matcherName := range funcs {
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

func NewTCPParser() (predicate.Parser, error) {
	parserFuncs := make(map[string]interface{})

	for _, matcherName := range []string{"HostSNI", "ClientIP"} {
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
