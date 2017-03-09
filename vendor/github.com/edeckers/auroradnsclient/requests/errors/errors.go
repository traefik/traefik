package errors

// BadRequest HTTP error wrapper
type BadRequest error

// Unauthorized HTTP error wrapper
type Unauthorized error

// Forbidden HTTP error wrapper
type Forbidden error

// NotFound HTTP error wrapper
type NotFound error

// ServerError HTTP error wrapper
type ServerError error

// InvalidStatusCodeError is used when none of the other types applies
type InvalidStatusCodeError error
