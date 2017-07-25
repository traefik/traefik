package secret

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestSecret(t *testing.T) { TestingT(t) }

type SecretSuite struct {
}

var _ = Suite(&SecretSuite{})

func (s *SecretSuite) TestEncryptDecryptCylce(c *C) {
	keyS, err := NewKeyString()
	c.Assert(err, IsNil)

	key, err := KeyFromString(keyS)
	c.Assert(err, IsNil)

	b, err := NewBox(key)
	c.Assert(err, IsNil)

	message := []byte("hello, box!")
	sealed, err := b.Seal(message)
	c.Assert(err, IsNil)

	out, err := b.Open(sealed)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, message)
}

func (s *SecretSuite) TestEncryptDecryptJSON(c *C) {
	keyS, err := NewKeyString()
	c.Assert(err, IsNil)

	key, err := KeyFromString(keyS)
	c.Assert(err, IsNil)

	b, err := NewBox(key)
	c.Assert(err, IsNil)

	message := []byte("hello, box!")
	sealed, err := b.Seal(message)
	c.Assert(err, IsNil)

	data, err := SealedValueToJSON(sealed)
	c.Assert(err, IsNil)
	c.Assert(data, NotNil)

	bytes, err := SealedValueFromJSON(data)
	c.Assert(err, IsNil)
	c.Assert(bytes, NotNil)

	out, err := b.Open(sealed)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, message)
}
