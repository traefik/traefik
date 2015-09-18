package kingpin

import "fmt"

type argGroup struct {
	args []*ArgClause
}

func newArgGroup() *argGroup {
	return &argGroup{}
}

func (a *argGroup) have() bool {
	return len(a.args) > 0
}

func (a *argGroup) Arg(name, help string) *ArgClause {
	arg := newArg(name, help)
	a.args = append(a.args, arg)
	return arg
}

func (a *argGroup) init() error {
	required := 0
	seen := map[string]struct{}{}
	previousArgMustBeLast := false
	for i, arg := range a.args {
		if previousArgMustBeLast {
			return fmt.Errorf("Args() can't be followed by another argument '%s'", arg.name)
		}
		if arg.consumesRemainder() {
			previousArgMustBeLast = true
		}
		if _, ok := seen[arg.name]; ok {
			return fmt.Errorf("duplicate argument '%s'", arg.name)
		}
		seen[arg.name] = struct{}{}
		if arg.required && required != i {
			return fmt.Errorf("required arguments found after non-required")
		}
		if arg.required {
			required++
		}
		if err := arg.init(); err != nil {
			return err
		}
	}
	return nil
}

type ArgClause struct {
	parserMixin
	name         string
	help         string
	defaultValue string
	required     bool
	action       Action
	preAction    Action
}

func newArg(name, help string) *ArgClause {
	a := &ArgClause{
		name: name,
		help: help,
	}
	return a
}

func (a *ArgClause) consumesRemainder() bool {
	if r, ok := a.value.(remainderArg); ok {
		return r.IsCumulative()
	}
	return false
}

// Required arguments must be input by the user. They can not have a Default() value provided.
func (a *ArgClause) Required() *ArgClause {
	a.required = true
	return a
}

// Default value for this argument. It *must* be parseable by the value of the argument.
func (a *ArgClause) Default(value string) *ArgClause {
	a.defaultValue = value
	return a
}

func (a *ArgClause) Action(action Action) *ArgClause {
	a.action = action
	return a
}

func (a *ArgClause) PreAction(action Action) *ArgClause {
	a.preAction = action
	return a
}

func (a *ArgClause) init() error {
	if a.required && a.defaultValue != "" {
		return fmt.Errorf("required argument '%s' with unusable default value", a.name)
	}
	if a.value == nil {
		return fmt.Errorf("no parser defined for arg '%s'", a.name)
	}
	return nil
}
