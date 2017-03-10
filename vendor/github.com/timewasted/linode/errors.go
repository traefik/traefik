package linode

const (
	ERR_OK                        = 0
	ERR_BAD_REQUEST               = 1
	ERR_NO_ACTION_REQUESTED       = 2
	ERR_CLASS_DOES_NOT_EXIST      = 3
	ERR_AUTHENTICATION_FAILED     = 4
	ERR_OBJECT_NOT_FOUND          = 5
	ERR_REQUIRED_PROPERTY_MISSING = 6
	ERR_INVALID_PROPERTY          = 7
	ERR_DATA_VALIDATION_FAILED    = 8
	ERR_METHOD_NOT_IMPLEMENTED    = 9
	ERR_TOO_MANY_BATCHED_REQUESTS = 10
	ERR_REQUEST_INVALID_JSON      = 11
	ERR_BATCH_TIMED_OUT           = 12
	ERR_PERMISSION_DENIED         = 13
	ERR_API_RATE_LIMIT_EXCEEDED   = 14

	ERR_CREDIT_CARD_CHARGE_FAILED = 30
	ERR_CREDIT_CARD_EXPIRED       = 31

	ERR_LINODES_PER_HOUR_LIMIT_EXCEEDED = 40
	ERR_LINODE_STILL_HAS_DISKS          = 41
)

type (
	// apiError represents an error as returned by the API.
	apiError struct {
		code int
		msg  string
	}
	// Error represents an error, either API related or not.
	Error struct {
		apiError *apiError
		Err      error
	}
)

// NewError returns an instance of Error which represents a non-API error.
func NewError(err error) *Error {
	return &Error{
		apiError: nil,
		Err:      err,
	}
}

// NewApiError returns an instance of Error which represents an API error.
func NewApiError(code int, msg string) *Error {
	return &Error{
		apiError: &apiError{
			code: code,
			msg:  msg,
		},
		Err: nil,
	}
}

// Error implements the Error() method of the error interface.
func (e *Error) Error() string {
	if e.apiError != nil {
		return e.apiError.msg
	}
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// IsApiError returns true if the error is an API error, or false otherwise.
func (e *Error) IsApiError() bool {
	return e.apiError != nil
}
