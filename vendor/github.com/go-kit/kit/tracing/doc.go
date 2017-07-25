// Package tracing provides helpers and bindings for distributed tracing.
//
// As your infrastructure grows, it becomes important to be able to trace a
// request, as it travels through multiple services and back to the user.
// Package tracing provides endpoints and transport helpers and middlewares to
// capture and emit request-scoped information. We use the excellent OpenTracing
// project to bind to concrete tracing systems.
package tracing
