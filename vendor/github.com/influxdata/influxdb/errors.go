package influxdb

import (
	"errors"
	"fmt"
	"strings"
)

// ErrFieldTypeConflict is returned when a new field already exists with a
// different type.
var ErrFieldTypeConflict = errors.New("field type conflict")

// ErrDatabaseNotFound indicates that a database operation failed on the
// specified database because the specified database does not exist.
func ErrDatabaseNotFound(name string) error { return fmt.Errorf("database not found: %s", name) }

// ErrRetentionPolicyNotFound indicates that the named retention policy could
// not be found in the database.
func ErrRetentionPolicyNotFound(name string) error {
	return fmt.Errorf("retention policy not found: %s", name)
}

// IsAuthorizationError indicates whether an error is due to an authorization failure
func IsAuthorizationError(err error) bool {
	e, ok := err.(interface {
		AuthorizationFailed() bool
	})
	return ok && e.AuthorizationFailed()
}

// IsClientError indicates whether an error is a known client error.
func IsClientError(err error) bool {
	if err == nil {
		return false
	}

	if strings.HasPrefix(err.Error(), ErrFieldTypeConflict.Error()) {
		return true
	}

	return false
}
