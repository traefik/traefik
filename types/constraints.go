package types

import (
	"encoding"
	"errors"
	"fmt"
	"strings"

	"github.com/ryanuber/go-glob"
)

// Constraint holds a parsed constraint expression.
type Constraint struct {
	Key string `export:"true"`
	// MustMatch is true if operator is "==" or false if operator is "!="
	MustMatch bool `export:"true"`
	// TODO: support regex
	Regex string `export:"true"`
}

// NewConstraint receives a string and return a *Constraint, after checking syntax and parsing the constraint expression.
func NewConstraint(exp string) (*Constraint, error) {
	sep := ""
	constraint := &Constraint{}

	switch {
	case strings.Contains(exp, "=="):
		sep = "=="
		constraint.MustMatch = true
	case strings.Contains(exp, "!="):
		sep = "!="
		constraint.MustMatch = false
	default:
		return nil, errors.New("constraint expression missing valid operator: '==' or '!='")
	}

	kv := strings.SplitN(exp, sep, 2)
	if len(kv) == 2 {
		// At the moment, it only supports tags
		if kv[0] != "tag" {
			return nil, errors.New("constraint must be tag-based. Syntax: tag==us-*")
		}

		constraint.Key = kv[0]
		constraint.Regex = kv[1]
		return constraint, nil
	}

	return nil, fmt.Errorf("incorrect constraint expression: %s", exp)
}

func (c *Constraint) String() string {
	if c.MustMatch {
		return c.Key + "==" + c.Regex
	}
	return c.Key + "!=" + c.Regex
}

var _ encoding.TextUnmarshaler = (*Constraint)(nil)

// UnmarshalText defines how unmarshal in TOML parsing
func (c *Constraint) UnmarshalText(text []byte) error {
	constraint, err := NewConstraint(string(text))
	if err != nil {
		return err
	}
	c.Key = constraint.Key
	c.MustMatch = constraint.MustMatch
	c.Regex = constraint.Regex
	return nil
}

var _ encoding.TextMarshaler = (*Constraint)(nil)

// MarshalText encodes the receiver into UTF-8-encoded text and returns the result.
func (c *Constraint) MarshalText() (text []byte, err error) {
	return []byte(c.String()), nil
}

// MatchConstraintWithAtLeastOneTag tests a constraint for one single service.
func (c *Constraint) MatchConstraintWithAtLeastOneTag(tags []string) bool {
	for _, tag := range tags {
		if glob.Glob(c.Regex, tag) {
			return true
		}
	}
	return false
}

// Set []*Constraint.
func (cs *Constraints) Set(str string) error {
	exps := strings.Split(str, ",")
	if len(exps) == 0 {
		return fmt.Errorf("bad Constraint format: %s", str)
	}
	for _, exp := range exps {
		constraint, err := NewConstraint(exp)
		if err != nil {
			return err
		}
		*cs = append(*cs, constraint)
	}
	return nil
}

// Constraints holds a Constraint parser.
type Constraints []*Constraint

// Get []*Constraint
func (cs *Constraints) Get() interface{} { return []*Constraint(*cs) }

// String returns []*Constraint in string.
func (cs *Constraints) String() string { return fmt.Sprintf("%+v", *cs) }

// SetValue sets []*Constraint into the parser.
func (cs *Constraints) SetValue(val interface{}) {
	*cs = val.(Constraints)
}

// Type exports the Constraints type as a string.
func (cs *Constraints) Type() string {
	return "constraint"
}
