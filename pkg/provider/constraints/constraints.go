package constraints

import (
	"errors"
	"regexp"
	"strings"

	"github.com/vulcand/predicate"
)

// MarathonConstraintPrefix is the prefix for each label's key created from a Marathon application constraint.
// It is used in order to create a specific and unique pattern for these labels.
const MarathonConstraintPrefix = "Traefik-Marathon-505F9E15-BDC7-45E7-828D-C06C7BAB8091"

type constraintFunc func(map[string]string) bool

// Match reports whether the expression matches with the given labels.
// The expression must match any logical boolean combination of:
// - `Label(labelName, labelValue)`
// - `LabelRegex(labelName, regexValue)`
// - `MarathonConstraint(field:operator:value)`
func Match(labels map[string]string, expr string) (bool, error) {
	if expr == "" {
		return true, nil
	}

	p, err := predicate.NewParser(predicate.Def{
		Operators: predicate.Operators{
			AND: andFunc,
			NOT: notFunc,
			OR:  orFunc,
		},
		Functions: map[string]interface{}{
			"Label":              labelFn,
			"LabelRegex":         labelRegexFn,
			"MarathonConstraint": marathonFn,
		},
	})
	if err != nil {
		return false, err
	}

	parse, err := p.Parse(expr)
	if err != nil {
		return false, err
	}

	fn, ok := parse.(constraintFunc)
	if !ok {
		return false, errors.New("not a constraintFunc")
	}
	return fn(labels), nil
}

func labelFn(name, value string) constraintFunc {
	return func(labels map[string]string) bool {
		return labels[name] == value
	}
}

func labelRegexFn(name, expr string) constraintFunc {
	return func(labels map[string]string) bool {
		matched, err := regexp.MatchString(expr, labels[name])
		if err != nil {
			return false
		}
		return matched
	}
}

func marathonFn(value string) constraintFunc {
	return func(labels map[string]string) bool {
		for k, v := range labels {
			if strings.HasPrefix(k, MarathonConstraintPrefix) {
				if v == value {
					return true
				}
			}
		}
		return false
	}
}

func andFunc(a, b constraintFunc) constraintFunc {
	return func(labels map[string]string) bool {
		return a(labels) && b(labels)
	}
}

func orFunc(a, b constraintFunc) constraintFunc {
	return func(labels map[string]string) bool {
		return a(labels) || b(labels)
	}
}

func notFunc(a constraintFunc) constraintFunc {
	return func(labels map[string]string) bool {
		return !a(labels)
	}
}
