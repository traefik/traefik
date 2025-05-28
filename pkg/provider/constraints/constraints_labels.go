package constraints

import (
	"errors"
	"regexp"

	"github.com/vulcand/predicate"
)

type constraintLabelFunc func(map[string]string) bool

// MatchLabels reports whether the expression matches with the given labels.
// The expression must match any logical boolean combination of:
// - `Label(labelName, labelValue)`
// - `LabelRegex(labelName, regexValue)`.
func MatchLabels(labels map[string]string, expr string) (bool, error) {
	if expr == "" {
		return true, nil
	}

	p, err := predicate.NewParser(predicate.Def{
		Operators: predicate.Operators{
			AND: andLabelFunc,
			NOT: notLabelFunc,
			OR:  orLabelFunc,
		},
		Functions: map[string]interface{}{
			"Label":      labelFn,
			"LabelRegex": labelRegexFn,
		},
	})
	if err != nil {
		return false, err
	}

	parse, err := p.Parse(expr)
	if err != nil {
		return false, err
	}

	fn, ok := parse.(constraintLabelFunc)
	if !ok {
		return false, errors.New("not a constraintLabelFunc")
	}
	return fn(labels), nil
}

func labelFn(name, value string) constraintLabelFunc {
	return func(labels map[string]string) bool {
		return labels[name] == value
	}
}

func labelRegexFn(name, expr string) constraintLabelFunc {
	return func(labels map[string]string) bool {
		matched, err := regexp.MatchString(expr, labels[name])
		if err != nil {
			return false
		}
		return matched
	}
}

func andLabelFunc(a, b constraintLabelFunc) constraintLabelFunc {
	return func(labels map[string]string) bool {
		return a(labels) && b(labels)
	}
}

func orLabelFunc(a, b constraintLabelFunc) constraintLabelFunc {
	return func(labels map[string]string) bool {
		return a(labels) || b(labels)
	}
}

func notLabelFunc(a constraintLabelFunc) constraintLabelFunc {
	return func(labels map[string]string) bool {
		return !a(labels)
	}
}
