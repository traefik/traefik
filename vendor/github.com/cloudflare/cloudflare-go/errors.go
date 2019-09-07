package cloudflare

// Error messages
const (
	errEmptyCredentials     = "invalid credentials: key & email must not be empty"
	errEmptyAPIToken        = "invalid credentials: API Token must not be empty"
	errMakeRequestError     = "error from makeRequest"
	errUnmarshalError       = "error unmarshalling the JSON response"
	errRequestNotSuccessful = "error reported by API"
	errMissingAccountID     = "account ID is empty and must be provided"
)

var _ Error = &UserError{}

// Error represents an error returned from this library.
type Error interface {
	error
	// Raised when user credentials or configuration is invalid.
	User() bool
	// Raised when a parsing error (e.g. JSON) occurs.
	Parse() bool
	// Raised when a network error occurs.
	Network() bool
	// Contains the most recent error.
}

// UserError represents a user-generated error.
type UserError struct {
	Err error
}

// User is a user-caused error.
func (e *UserError) User() bool {
	return true
}

// Network error.
func (e *UserError) Network() bool {
	return false
}

// Parse error.
func (e *UserError) Parse() bool {
	return true
}

// Error wraps the underlying error.
func (e *UserError) Error() string {
	return e.Err.Error()
}
