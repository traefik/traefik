package decimal

import (
	"fmt"
	"math/big"
)

// Decompose returns the internal decimal state into parts.
// If the provided buf has sufficient capacity, buf may be returned as the coefficient with
// the value set and length set as appropriate.
func (d Decimal) Decompose(buf []byte) (form byte, negative bool, coefficient []byte, exponent int32) {
	negative = d.value.Sign() < 0
	exponent = d.exp
	coefficient = d.value.Bytes()
	return
}

const (
	decomposeFinite   = 0
	decomposeInfinite = 1
	decomposeNaN      = 2
)

// Compose sets the internal decimal value from parts. If the value cannot be
// represented then an error should be returned.
func (d *Decimal) Compose(form byte, negative bool, coefficient []byte, exponent int32) error {
	switch form {
	default:
		return fmt.Errorf("unknown form: %v", form)
	case decomposeFinite:
		// Set rest of finite form below.
	case decomposeInfinite:
		return fmt.Errorf("Infinite form not supported")
	case decomposeNaN:
		return fmt.Errorf("NaN form not supported")
	}
	// Finite form.
	if d.value == nil {
		d.value = &big.Int{}
	}
	d.value.SetBytes(coefficient)
	if negative && d.value.Sign() >= 0 {
		d.value.Neg(d.value)
	}
	d.exp = exponent
	return nil
}
