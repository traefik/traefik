package kingpin

import (
	"fmt"
	"os"
	"strings"
)

type flagGroup struct {
	short     map[string]*FlagClause
	long      map[string]*FlagClause
	flagOrder []*FlagClause
}

func newFlagGroup() *flagGroup {
	return &flagGroup{
		short: make(map[string]*FlagClause),
		long:  make(map[string]*FlagClause),
	}
}

func (f *flagGroup) merge(o *flagGroup) {
	for _, flag := range o.flagOrder {
		if flag.shorthand != 0 {
			f.short[string(flag.shorthand)] = flag
		}
		f.long[flag.name] = flag
		f.flagOrder = append(f.flagOrder, flag)
	}
}

// Flag defines a new flag with the given long name and help.
func (f *flagGroup) Flag(name, help string) *FlagClause {
	flag := newFlag(name, help)
	f.long[name] = flag
	f.flagOrder = append(f.flagOrder, flag)
	return flag
}

func (f *flagGroup) init() error {
	for _, flag := range f.long {
		if err := flag.init(); err != nil {
			return err
		}
		if flag.shorthand != 0 {
			f.short[string(flag.shorthand)] = flag
		}
	}
	return nil
}

func (f *flagGroup) parse(context *ParseContext) error {
	var token *Token

loop:
	for {
		token = context.Peek()
		switch token.Type {
		case TokenEOL:
			break loop

		case TokenLong, TokenShort:
			flagToken := token
			defaultValue := ""
			var flag *FlagClause
			var ok bool
			invert := false

			name := token.Value
			if token.Type == TokenLong {
				if strings.HasPrefix(name, "no-") {
					name = name[3:]
					invert = true
				}
				flag, ok = f.long[name]
				if !ok {
					return fmt.Errorf("unknown long flag '%s'", flagToken)
				}
			} else {
				flag, ok = f.short[name]
				if !ok {
					return fmt.Errorf("unknown short flag '%s'", flagToken)
				}
			}

			context.Next()

			fb, ok := flag.value.(boolFlag)
			if ok && fb.IsBoolFlag() {
				if invert {
					defaultValue = "false"
				} else {
					defaultValue = "true"
				}
			} else {
				if invert {
					return fmt.Errorf("unknown long flag '%s'", flagToken)
				}
				token = context.Peek()
				if token.Type != TokenArg {
					return fmt.Errorf("expected argument for flag '%s'", flagToken)
				}
				context.Next()
				defaultValue = token.Value
			}

			context.matchedFlag(flag, defaultValue)

		default:
			break loop
		}
	}
	return nil
}

func (f *flagGroup) visibleFlags() int {
	count := 0
	for _, flag := range f.long {
		if !flag.hidden {
			count++
		}
	}
	return count
}

// FlagClause is a fluid interface used to build flags.
type FlagClause struct {
	parserMixin
	name         string
	shorthand    byte
	help         string
	envar        string
	defaultValue string
	placeholder  string
	action       Action
	preAction    Action
	hidden       bool
}

func newFlag(name, help string) *FlagClause {
	f := &FlagClause{
		name: name,
		help: help,
	}
	return f
}

func (f *FlagClause) needsValue() bool {
	return f.required && f.defaultValue == ""
}

func (f *FlagClause) formatPlaceHolder() string {
	if f.placeholder != "" {
		return f.placeholder
	}
	if f.defaultValue != "" {
		if _, ok := f.value.(*stringValue); ok {
			return fmt.Sprintf("%q", f.defaultValue)
		}
		return f.defaultValue
	}
	return strings.ToUpper(f.name)
}

func (f *FlagClause) init() error {
	if f.required && f.defaultValue != "" {
		return fmt.Errorf("required flag '--%s' with default value that will never be used", f.name)
	}
	if f.value == nil {
		return fmt.Errorf("no type defined for --%s (eg. .String())", f.name)
	}
	if f.envar != "" {
		if v := os.Getenv(f.envar); v != "" {
			f.defaultValue = v
		}
	}
	return nil
}

// Dispatch to the given function after the flag is parsed and validated.
func (f *FlagClause) Action(action Action) *FlagClause {
	f.action = action
	return f
}

func (f *FlagClause) PreAction(action Action) *FlagClause {
	f.preAction = action
	return f
}

// Default value for this flag. It *must* be parseable by the value of the flag.
func (f *FlagClause) Default(value string) *FlagClause {
	f.defaultValue = value
	return f
}

// OverrideDefaultFromEnvar overrides the default value for a flag from an
// environment variable, if available.
func (f *FlagClause) OverrideDefaultFromEnvar(envar string) *FlagClause {
	f.envar = envar
	return f
}

// PlaceHolder sets the place-holder string used for flag values in the help. The
// default behaviour is to use the value provided by Default() if provided,
// then fall back on the capitalized flag name.
func (f *FlagClause) PlaceHolder(placeholder string) *FlagClause {
	f.placeholder = placeholder
	return f
}

// Hidden hides a flag from usage but still allows it to be used.
func (f *FlagClause) Hidden() *FlagClause {
	f.hidden = true
	return f
}

// Required makes the flag required. You can not provide a Default() value to a Required() flag.
func (f *FlagClause) Required() *FlagClause {
	f.required = true
	return f
}

// Short sets the short flag name.
func (f *FlagClause) Short(name byte) *FlagClause {
	f.shorthand = name
	return f
}

// Bool makes this flag a boolean flag.
func (f *FlagClause) Bool() (target *bool) {
	target = new(bool)
	f.SetValue(newBoolValue(target))
	return
}
