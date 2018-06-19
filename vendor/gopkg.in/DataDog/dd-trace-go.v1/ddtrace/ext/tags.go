// Package ext contains a set of Datadog-specific constants. Most of them are used
// for setting span metadata.
package ext

const (
	// TargetHost sets the target host address.
	TargetHost = "out.host"

	// TargetPort sets the target host port.
	TargetPort = "out.port"

	// SamplingPriority is the tag that marks the sampling priority of a span.
	SamplingPriority = "sampling.priority"

	// SQLType sets the sql type tag.
	SQLType = "sql"

	// SQLQuery sets the sql query tag on a span.
	SQLQuery = "sql.query"

	// HTTPMethod specifies the HTTP method used in a span.
	HTTPMethod = "http.method"

	// HTTPCode sets the HTTP status code as a tag.
	HTTPCode = "http.status_code"

	// HTTPURL sets the HTTP URL for a span.
	HTTPURL = "http.url"

	// SpanType defines the Span type (web, db, cache).
	SpanType = "span.type"

	// ServiceName defines the Service name for this Span.
	ServiceName = "service.name"

	// ResourceName defines the Resource name for the Span.
	ResourceName = "resource.name"

	// Error specifies the error tag. It's value is usually of type "error".
	Error = "error"

	// ErrorMsg specifies the error message.
	ErrorMsg = "error.msg"

	// ErrorType specifies the error type.
	ErrorType = "error.type"

	// ErrorStack specifies the stack dump.
	ErrorStack = "error.stack"
)
