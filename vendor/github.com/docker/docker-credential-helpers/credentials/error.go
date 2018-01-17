package credentials

const (
	// ErrCredentialsNotFound standardizes the not found error, so every helper returns
	// the same message and docker can handle it properly.
	errCredentialsNotFoundMessage = "credentials not found in native keychain"

	// ErrCredentialsMissingServerURL and ErrCredentialsMissingUsername standardize
	// invalid credentials or credentials management operations
	errCredentialsMissingServerURLMessage = "no credentials server URL"
	errCredentialsMissingUsernameMessage  = "no credentials username"
)

// errCredentialsNotFound represents an error
// raised when credentials are not in the store.
type errCredentialsNotFound struct{}

// Error returns the standard error message
// for when the credentials are not in the store.
func (errCredentialsNotFound) Error() string {
	return errCredentialsNotFoundMessage
}

// NewErrCredentialsNotFound creates a new error
// for when the credentials are not in the store.
func NewErrCredentialsNotFound() error {
	return errCredentialsNotFound{}
}

// IsErrCredentialsNotFound returns true if the error
// was caused by not having a set of credentials in a store.
func IsErrCredentialsNotFound(err error) bool {
	_, ok := err.(errCredentialsNotFound)
	return ok
}

// IsErrCredentialsNotFoundMessage returns true if the error
// was caused by not having a set of credentials in a store.
//
// This function helps to check messages returned by an
// external program via its standard output.
func IsErrCredentialsNotFoundMessage(err string) bool {
	return err == errCredentialsNotFoundMessage
}

// errCredentialsMissingServerURL represents an error raised
// when the credentials object has no server URL or when no
// server URL is provided to a credentials operation requiring
// one.
type errCredentialsMissingServerURL struct{}

func (errCredentialsMissingServerURL) Error() string {
	return errCredentialsMissingServerURLMessage
}

// errCredentialsMissingUsername represents an error raised
// when the credentials object has no username or when no
// username is provided to a credentials operation requiring
// one.
type errCredentialsMissingUsername struct{}

func (errCredentialsMissingUsername) Error() string {
	return errCredentialsMissingUsernameMessage
}

// NewErrCredentialsMissingServerURL creates a new error for
// errCredentialsMissingServerURL.
func NewErrCredentialsMissingServerURL() error {
	return errCredentialsMissingServerURL{}
}

// NewErrCredentialsMissingUsername creates a new error for
// errCredentialsMissingUsername.
func NewErrCredentialsMissingUsername() error {
	return errCredentialsMissingUsername{}
}

// IsCredentialsMissingServerURL returns true if the error
// was an errCredentialsMissingServerURL.
func IsCredentialsMissingServerURL(err error) bool {
	_, ok := err.(errCredentialsMissingServerURL)
	return ok
}

// IsCredentialsMissingServerURLMessage checks for an
// errCredentialsMissingServerURL in the error message.
func IsCredentialsMissingServerURLMessage(err string) bool {
	return err == errCredentialsMissingServerURLMessage
}

// IsCredentialsMissingUsername returns true if the error
// was an errCredentialsMissingUsername.
func IsCredentialsMissingUsername(err error) bool {
	_, ok := err.(errCredentialsMissingUsername)
	return ok
}

// IsCredentialsMissingUsernameMessage checks for an
// errCredentialsMissingUsername in the error message.
func IsCredentialsMissingUsernameMessage(err string) bool {
	return err == errCredentialsMissingUsernameMessage
}
