package flag

// Flags provides the ability to create and communicate feature flags.
type Flags uint64

// Flags is a bitset
const (
	Debug Flags = 1 << 0

	// All flags below deal with binaryPropagators. They will be discarded in the
	// textMapPropagator (not read and not set)

	// SamplingSet and Sampled handle Sampled tribool logic for interop with
	// instrumenting libraries / propagation channels not using a separate Sampled
	// header and potentially encoding this in flags.
	//
	// When we receive a flag we do this:
	// 1. Sampled bit is set => true
	// 2. Sampled bit is not set => inspect SamplingSet bit.
	// 2a. SamplingSet bit is set => false
	// 2b. SamplingSet bit is not set => null
	// Note on 2b.: depending on the propagator having a separate Sampled header
	// we either assume Sampling is false or unknown. In the latter case we will
	// run our sampler even though we are not the root of the trace.
	//
	// When propagating to a downstream service we will always be explicit and
	// will provide a set SamplingSet bit in case of our binary propagator either
	SamplingSet Flags = 1 << 1
	Sampled     Flags = 1 << 2
	// When set, we can ignore the value of the parentId. This is used for binary
	// fixed width transports or transports like proto3 that return a default
	// value if a value has not been set (thus not enabling you to distinguish
	// between the value being set to the default or not set at all).
	//
	// While many zipkin systems re-use a trace id as the root span id, we know
	// that some don't. With this flag, we can tell for sure if the span is root
	// as opposed to the convention trace id == span id == parent id.
	IsRoot Flags = 1 << 3
)
