package stats

// Options are stats options.
type Options struct {
	statusCode *int
	size       int
	recorder   ResponseWriter
}

// StatusCode returns the response status code.
func (o Options) StatusCode() int {
	if o.recorder != nil {
		return o.recorder.Status()
	}

	return *o.statusCode
}

// Size returns the response size.
func (o Options) Size() int {
	if o.recorder != nil {
		return o.recorder.Size()
	}

	return o.size
}

// Option represents a stats option.
type Option func(*Options)

// WithStatusCode sets the status code to use in stats.
func WithStatusCode(statusCode int) Option {
	return func(o *Options) {
		o.statusCode = &statusCode
	}
}

// WithSize sets the size to use in stats.
func WithSize(size int) Option {
	return func(o *Options) {
		o.size = size
	}
}

// WithRecorder sets the recorder to use in stats.
func WithRecorder(recorder ResponseWriter) Option {
	return func(o *Options) {
		o.recorder = recorder
	}
}

// newOptions takes functional options and returns options.
func newOptions(options ...Option) *Options {
	opts := &Options{}
	for _, o := range options {
		o(opts)
	}
	return opts
}
