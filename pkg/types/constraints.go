package types

import (
	"encoding"
	"errors"
	"fmt"
	"strings"

	"github.com/ryanuber/go-glob"
)

// Constraint holds a parsed constraint expression.
// FIXME replace by a string.
type Constraint struct {
	Key string `description:"The provider label that will be matched against. In practice, it is always 'tag'." export:"true"`
	// MustMatch is true if operator is "==" or false if operator is "!="
	MustMatch bool   `description:"Whether the matching operator is equals or not equals." export:"true"`
	Value     string `description:"The value that will be matched against." export:"true"` // TODO: support regex
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
		constraint.Value = kv[1]
		return constraint, nil
	}

	return nil, fmt.Errorf("incorrect constraint expression: %s", exp)
}

func (c *Constraint) String() string {
	if c.MustMatch {
		return c.Key + "==" + c.Value
	}
	return c.Key + "!=" + c.Value
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
	c.Value = constraint.Value
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
		if glob.Glob(c.Value, tag) {
			return true
		}
	}
	return false
}
