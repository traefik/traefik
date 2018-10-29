package egoscale

import (
	"encoding/json"
	"fmt"

	"github.com/satori/go.uuid"
)

// UUID holds a UUID v4
type UUID struct {
	uuid.UUID
}

// Equal returns true if itself is equal to other
func (u UUID) Equal(other UUID) bool {
	return uuid.Equal(u.UUID, other.UUID)
}

// UnmarshalJSON unmarshals the raw JSON into the MAC address
func (u *UUID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	new, err := ParseUUID(s)
	if err == nil {
		u.UUID = new.UUID
	}
	return err
}

// MarshalJSON converts the UUID to a string representation
func (u UUID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", u.String())), nil
}

// ParseUUID parses a string into a UUID
func ParseUUID(s string) (*UUID, error) {
	u, err := uuid.FromString(s)
	if err != nil {
		return nil, err
	}
	return &UUID{u}, nil
}

// MustParseUUID acts like ParseUUID but panic in case of a failure
func MustParseUUID(s string) *UUID {
	u, e := ParseUUID(s)
	if e != nil {
		panic(e)
	}
	return u
}
