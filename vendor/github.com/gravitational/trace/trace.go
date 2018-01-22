/*
Copyright 2015 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package trace implements utility functions for capturing debugging
// information about file and line in error reports and logs.
package trace

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/net/context"
)

var debug int32

// SetDebug turns on/off debugging mode, that causes Fatalf to panic
func SetDebug(enabled bool) {
	if enabled {
		atomic.StoreInt32(&debug, 1)
	} else {
		atomic.StoreInt32(&debug, 0)
	}
}

// IsDebug returns true if debug mode is on, false otherwize
func IsDebug() bool {
	return atomic.LoadInt32(&debug) == 1
}

// Wrap takes the original error and wraps it into the Trace struct
// memorizing the context of the error.
func Wrap(err error, args ...interface{}) Error {
	if len(args) > 0 {
		format := args[0]
		args = args[1:]
		return WrapWithMessage(err, format, args...)
	}
	return wrapWithDepth(err, 2)
}

// Unwrap unwraps error to it's original error
func Unwrap(err error) error {
	if terr, ok := err.(Error); ok {
		return terr.OrigError()
	}
	return err
}

// UserMessage returns user-friendly part of the error
func UserMessage(err error) string {
	if err == nil {
		return ""
	}
	if wrap, ok := err.(Error); ok {
		return wrap.UserMessage()
	}
	return err.Error()
}

// DebugReport returns debug report with all known information
// about the error including stack trace if it was captured
func DebugReport(err error) string {
	if err == nil {
		return ""
	}
	if wrap, ok := err.(Error); ok {
		return wrap.DebugReport()
	}
	return err.Error()
}

// WrapWithMessage wraps the original error into Error and adds user message if any
func WrapWithMessage(err error, message interface{}, args ...interface{}) Error {
	trace := wrapWithDepth(err, 3)
	if trace != nil {
		trace.AddUserMessage(message, args...)
	}
	return trace
}

func wrapWithDepth(err error, depth int) Error {
	if err == nil {
		return nil
	}
	var trace Error
	if wrapped, ok := err.(Error); ok {
		trace = wrapped
	} else {
		trace = newTrace(depth+1, err)
	}

	return trace
}

// Errorf is similar to fmt.Errorf except that it captures
// more information about the origin of error, such as
// callee, line number and function that simplifies debugging
func Errorf(format string, args ...interface{}) (err error) {
	err = fmt.Errorf(format, args...)
	trace := wrapWithDepth(err, 2)
	trace.AddUserMessage(format, args...)
	return trace
}

// Fatalf - If debug is false Fatalf calls Errorf. If debug is
// true Fatalf calls panic
func Fatalf(format string, args ...interface{}) error {
	if IsDebug() {
		panic(fmt.Sprintf(format, args))
	} else {
		return Errorf(format, args)
	}
}

func newTrace(depth int, err error) *TraceErr {
	var pc [32]uintptr
	count := runtime.Callers(depth+1, pc[:])

	traces := make(Traces, count)
	for i := 0; i < count; i++ {
		fn := runtime.FuncForPC(pc[i])
		filePath, line := fn.FileLine(pc[i])
		traces[i] = Trace{
			Func: fn.Name(),
			Path: filePath,
			Line: line,
		}
	}
	return &TraceErr{
		err,
		traces,
		"",
	}
}

// Traces is a list of trace entries
type Traces []Trace

// SetTraces adds new traces to the list
func (s Traces) SetTraces(traces ...Trace) {
	s = append(s, traces...)
}

// Func returns first function in trace list
func (s Traces) Func() string {
	if len(s) == 0 {
		return ""
	}
	return s[0].Func
}

// Func returns just function name
func (s Traces) FuncName() string {
	if len(s) == 0 {
		return ""
	}
	fn := filepath.ToSlash(s[0].Func)
	idx := strings.LastIndex(fn, "/")
	if idx == -1 || idx == len(fn)-1 {
		return fn
	}
	return fn[idx+1:]
}

// Loc points to file/line location in the code
func (s Traces) Loc() string {
	if len(s) == 0 {
		return ""
	}
	return s[0].String()
}

// String returns debug-friendly representaton of trace stack
func (s Traces) String() string {
	if len(s) == 0 {
		return ""
	}
	out := make([]string, len(s))
	for i, t := range s {
		out[i] = fmt.Sprintf("\t%v:%v %v", t.Path, t.Line, t.Func)
	}
	return strings.Join(out, "\n")
}

// Trace stores structured trace entry, including file line and path
type Trace struct {
	// Path is a full file path
	Path string `json:"path"`
	// Func is a function name
	Func string `json:"func"`
	// Line is a code line number
	Line int `json:"line"`
}

// String returns debug-friendly representation of this trace
func (t *Trace) String() string {
	dir, file := filepath.Split(t.Path)
	dirs := strings.Split(filepath.ToSlash(filepath.Clean(dir)), "/")
	if len(dirs) != 0 {
		file = filepath.Join(dirs[len(dirs)-1], file)
	}
	return fmt.Sprintf("%v:%v", file, t.Line)
}

// TraceErr contains error message and some additional
// information about the error origin
type TraceErr struct {
	Err     error `json:"error"`
	Traces  `json:"traces"`
	Message string `json:"message,omitemtpy"`
}

type RawTrace struct {
	Err     json.RawMessage `json:"error"`
	Traces  `json:"traces"`
	Message string `json:"message"`
}

// AddUserMessage adds user-friendly message describing the error nature
func (e *TraceErr) AddUserMessage(formatArg interface{}, rest ...interface{}) {
	newMessage := fmt.Sprintf(fmt.Sprintf("%v", formatArg), rest...)
	if len(e.Message) == 0 {
		e.Message = newMessage
	} else {
		e.Message = strings.Join([]string{e.Message, newMessage}, ", ")
	}
}

// UserMessage returns user-friendly error message
func (e *TraceErr) UserMessage() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// DebugReport returns develeoper-friendly error report
func (e *TraceErr) DebugReport() string {
	return fmt.Sprintf("\nERROR REPORT:\nOriginal Error: %T %v\nStack Trace:\n%v\nUser Message: %v\n", e.Err, e.Err.Error(), e.Traces.String(), e.Message)
}

// Error returns user-friendly error message when not in debug mode
func (e *TraceErr) Error() string {
	if IsDebug() {
		return e.DebugReport()
	}
	return e.UserMessage()
}

// OrigError returns original wrapped error
func (e *TraceErr) OrigError() error {
	err := e.Err
	// this is not an endless loop because I'm being
	// paranoid, this is a safe protection against endless
	// loops
	for i := 0; i < maxHops; i++ {
		newerr, ok := err.(Error)
		if !ok {
			break
		}
		if newerr.OrigError() != err {
			err = newerr.OrigError()
		}
	}
	return err
}

// maxHops is a max supported nested depth for errors
const maxHops = 50

// Error is an interface that helps to adapt usage of trace in the code
// When applications define new error types, they can implement the interface
// So error handlers can use OrigError() to retrieve error from the wrapper
type Error interface {
	error
	// OrigError returns original error wrapped in this error
	OrigError() error
	// AddMessage adds formatted user-facing message
	// to the error, depends on the implementation,
	// usually works as fmt.Sprintf(formatArg, rest...)
	// but implementations can choose another way, e.g. treat
	// arguments as structured args
	AddUserMessage(formatArg interface{}, rest ...interface{})

	// UserMessage returns user-friendly error message
	UserMessage() string

	// DebugReport returns develeoper-friendly error report
	DebugReport() string
}

// NewAggregate creates a new aggregate instance from the specified
// list of errors
func NewAggregate(errs ...error) error {
	// filter out possible nil values
	var nonNils []error
	for _, err := range errs {
		if err != nil {
			nonNils = append(nonNils, err)
		}
	}
	if len(nonNils) == 0 {
		return nil
	}
	return wrapWithDepth(aggregate(nonNils), 2)
}

// NewAggregateFromChannel creates a new aggregate instance from the provided
// errors channel.
//
// A context.Context can be passed in so the caller has the ability to cancel
// the operation. If this is not desired, simply pass context.Background().
func NewAggregateFromChannel(errCh chan error, ctx context.Context) error {
	var errs []error

Loop:
	for {
		select {
		case err, ok := <-errCh:
			if !ok { // the channel is closed, time to exit
				break Loop
			}
			errs = append(errs, err)
		case <-ctx.Done():
			break Loop
		}
	}

	return NewAggregate(errs...)
}

// Aggregate interface combines several errors into one error
type Aggregate interface {
	error
	// Errors obtains the list of errors this aggregate combines
	Errors() []error
}

// aggregate implements Aggregate
type aggregate []error

// Error implements the error interface
func (r aggregate) Error() string {
	if len(r) == 0 {
		return ""
	}
	output := r[0].Error()
	for i := 1; i < len(r); i++ {
		output = fmt.Sprintf("%v, %v", output, r[i])
	}
	return output
}

// Errors obtains the list of errors this aggregate combines
func (r aggregate) Errors() []error {
	return []error(r)
}

// IsAggregate returns whether this error of Aggregate error type
func IsAggregate(err error) bool {
	_, ok := Unwrap(err).(Aggregate)
	return ok
}
