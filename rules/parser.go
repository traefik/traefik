package rules

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/vulcand/predicate"
)

type treeBuilder func() *tree

// ParseDomains extract domains from rule
func ParseDomains(rule string) ([]string, error) {
	parser, err := newParser()
	if err != nil {
		return nil, err
	}

	parse, err := parser.Parse(rule)
	if err != nil {
		return nil, err
	}

	buildTree, ok := parse.(treeBuilder)
	if !ok {
		return nil, errors.New("cannot parse")
	}

	return lower(parseDomain(buildTree())), nil
}

func lower(slice []string) []string {
	var lowerStrings []string
	for _, value := range slice {
		lowerStrings = append(lowerStrings, strings.ToLower(value))
	}
	return lowerStrings
}

func parseDomain(tree *tree) []string {
	switch tree.matcher {
	case "and", "or":
		return append(parseDomain(tree.ruleLeft), parseDomain(tree.ruleRight)...)
	case "Host":
		return tree.value
	default:
		return nil
	}
}

func andFunc(left, right treeBuilder) treeBuilder {
	return func() *tree {
		return &tree{
			matcher:   "and",
			ruleLeft:  left(),
			ruleRight: right(),
		}
	}
}

func orFunc(left, right treeBuilder) treeBuilder {
	return func() *tree {
		return &tree{
			matcher:   "or",
			ruleLeft:  left(),
			ruleRight: right(),
		}
	}
}

func newParser() (predicate.Parser, error) {
	parserFuncs := make(map[string]interface{})

	for matcherName := range funcs {
		matcherName := matcherName
		fn := func(value ...string) treeBuilder {
			return func() *tree {
				return &tree{
					matcher: matcherName,
					value:   value,
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
		},
		Functions: parserFuncs,
	})
}
