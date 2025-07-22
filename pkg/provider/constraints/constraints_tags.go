package constraints

import (
	"errors"
	"regexp"
	"slices"

	"github.com/vulcand/predicate"
)

type constraintTagFunc func([]string) bool

// MatchTags reports whether the expression matches with the given tags.
// The expression must match any logical boolean combination of:
// - `Tag(tagValue)`
// - `TagRegex(regexValue)`.
func MatchTags(tags []string, expr string) (bool, error) {
	if expr == "" {
		return true, nil
	}

	p, err := predicate.NewParser(predicate.Def{
		Operators: predicate.Operators{
			AND: andTagFunc,
			NOT: notTagFunc,
			OR:  orTagFunc,
		},
		Functions: map[string]interface{}{
			"Tag":      tagFn,
			"TagRegex": tagRegexFn,
		},
	})
	if err != nil {
		return false, err
	}

	parse, err := p.Parse(expr)
	if err != nil {
		return false, err
	}

	fn, ok := parse.(constraintTagFunc)
	if !ok {
		return false, errors.New("not a constraintTagFunc")
	}
	return fn(tags), nil
}

func tagFn(name string) constraintTagFunc {
	return func(tags []string) bool {
		return slices.Contains(tags, name)
	}
}

func tagRegexFn(expr string) constraintTagFunc {
	return func(tags []string) bool {
		exp, err := regexp.Compile(expr)
		if err != nil {
			return false
		}

		return slices.ContainsFunc(tags, func(tag string) bool {
			return exp.MatchString(tag)
		})
	}
}

func andTagFunc(a, b constraintTagFunc) constraintTagFunc {
	return func(tags []string) bool {
		return a(tags) && b(tags)
	}
}

func orTagFunc(a, b constraintTagFunc) constraintTagFunc {
	return func(tags []string) bool {
		return a(tags) || b(tags)
	}
}

func notTagFunc(a constraintTagFunc) constraintTagFunc {
	return func(tags []string) bool {
		return !a(tags)
	}
}
