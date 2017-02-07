package zbase32

import "flag"

// Value implements flag parsing for zbase32 values.
type Value []byte

var _ flag.Value = (*Value)(nil)

// String returns the z-base-32 encoding of the value.
func (v *Value) String() string {
	return EncodeToString(*v)
}

// Set the value to data encoded by string.
func (v *Value) Set(s string) error {
	b, err := DecodeString(s)
	if err != nil {
		return err
	}
	*v = b
	return nil
}

var _ flag.Getter = (*Value)(nil)

// Get the data stored in the value. Returns a value of type []byte.
func (v *Value) Get() interface{} {
	return []byte(*v)
}
