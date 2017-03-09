package errorutil

// ErrorFunction A function definition that returns an error
// to be passed to the Ignore or Panic error handler
type ErrorFunction func() error

// Ignore Calls an ErrorFunction, and ignores the result.
// This allows us to be more explicit when there is no error
// handling to be done, for example in defers
func Ignore(f ErrorFunction) {
	_ = f()
}
