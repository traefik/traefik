// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package errors is a Google Stackdriver Error Reporting library.
//
// This package is still experimental and subject to change.
//
// See https://cloud.google.com/error-reporting/ for more information.
//
// To initialize a client, use the NewClient function.  Generally you will want
// to do this on program initialization.  The NewClient function takes as
// arguments a context, the project name, a service name, and a version string.
// The service name and version string identify the running program, and are
// included in error reports.  The version string can be left empty.  NewClient
// also takes a bool that indicates whether to report errors using Stackdriver
// Logging, which will result in errors appearing in both the logs and the error
// dashboard.  This is useful if you are already a user of Stackdriver Logging.
//
//   import "cloud.google.com/go/errors"
//   ...
//   errorsClient, err = errors.NewClient(ctx, projectID, "myservice", "v1.0", true)
//
// The client can recover panics in your program and report them as errors.
// To use this functionality, defer its Catch method, as you would any other
// function for recovering panics.
//
//   func foo(ctx context.Context, ...) {
//     defer errorsClient.Catch(ctx)
//     ...
//   }
//
// Catch writes an error report containing the recovered value and a stack trace
// to Stackdriver Error Reporting.
//
// There are various options you can add to the call to Catch that modify how
// panics are handled.
//
// WithMessage and WithMessagef add a custom message after the recovered value,
// using fmt.Sprint and fmt.Sprintf respectively.
//
//   defer errorsClient.Catch(ctx, errors.WithMessagef("x=%d", x))
//
// WithRequest fills in various fields in the error report with information
// about an http.Request that's being handled.
//
//   defer errorsClient.Catch(ctx, errors.WithRequest(httpReq))
//
// By default, after recovering a panic, Catch will panic again with the
// recovered value.  You can turn off this behavior with the Repanic option.
//
//   defer errorsClient.Catch(ctx, errors.Repanic(false))
//
// You can also change the default behavior for the client by changing the
// RepanicDefault field.
//
//   errorsClient.RepanicDefault = false
//
// It is also possible to write an error report directly without recovering a
// panic, using Report or Reportf.
//
//   if err != nil {
//     errorsClient.Reportf(ctx, r, "unexpected error %v", err)
//   }
//
// If you try to write an error report with a nil client, or if the client
// fails to write the report to the server, the error report is logged using
// log.Println.
package errors // import "cloud.google.com/go/errors"

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	api "cloud.google.com/go/errorreporting/apiv1beta1"
	"cloud.google.com/go/internal/version"
	"cloud.google.com/go/logging"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	erpb "google.golang.org/genproto/googleapis/devtools/clouderrorreporting/v1beta1"
)

const (
	userAgent = `gcloud-golang-errorreporting/20160701`
)

type apiInterface interface {
	ReportErrorEvent(ctx context.Context, req *erpb.ReportErrorEventRequest) (*erpb.ReportErrorEventResponse, error)
}

var newApiInterface = func(ctx context.Context, opts ...option.ClientOption) (apiInterface, error) {
	client, err := api.NewReportErrorsClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	client.SetGoogleClientInfo("gccl", version.Repo)
	return client, nil
}

type loggerInterface interface {
	LogSync(ctx context.Context, e logging.Entry) error
}

var newLoggerInterface = func(ctx context.Context, projectID string, opts ...option.ClientOption) (loggerInterface, error) {
	lc, err := logging.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating Logging client: %v", err)
	}
	l := lc.Logger("errorreports")
	return l, nil
}

type sender interface {
	send(ctx context.Context, r *http.Request, message string)
}

// errorApiSender sends error reports using the Stackdriver Error Reporting API.
type errorApiSender struct {
	apiClient      apiInterface
	projectID      string
	serviceContext erpb.ServiceContext
}

// loggingSender sends error reports using the Stackdriver Logging API.
type loggingSender struct {
	logger         loggerInterface
	projectID      string
	serviceContext map[string]string
}

type Client struct {
	sender
	// RepanicDefault determines whether Catch will re-panic after recovering a
	// panic.  This behavior can be overridden for an individual call to Catch using
	// the Repanic option.
	RepanicDefault bool
}

func NewClient(ctx context.Context, projectID, serviceName, serviceVersion string, useLogging bool, opts ...option.ClientOption) (*Client, error) {
	if useLogging {
		l, err := newLoggerInterface(ctx, projectID, opts...)
		if err != nil {
			return nil, fmt.Errorf("creating Logging client: %v", err)
		}
		sender := &loggingSender{
			logger:    l,
			projectID: projectID,
			serviceContext: map[string]string{
				"service": serviceName,
			},
		}
		if serviceVersion != "" {
			sender.serviceContext["version"] = serviceVersion
		}
		c := &Client{
			sender:         sender,
			RepanicDefault: true,
		}
		return c, nil
	} else {
		a, err := newApiInterface(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("creating Error Reporting client: %v", err)
		}
		c := &Client{
			sender: &errorApiSender{
				apiClient: a,
				projectID: "projects/" + projectID,
				serviceContext: erpb.ServiceContext{
					Service: serviceName,
					Version: serviceVersion,
				},
			},
			RepanicDefault: true,
		}
		return c, nil
	}
}

// An Option is an optional argument to Catch.
type Option interface {
	isOption()
}

// PanicFlag returns an Option that can inform Catch that a panic has occurred.
// If *p is true when Catch is called, an error report is made even if recover
// returns nil.  This allows Catch to report an error for panic(nil).
// If p is nil, the option is ignored.
//
// Here is an example of how to use PanicFlag:
//
//   func foo(ctx context.Context, ...) {
//     hasPanicked := true
//     defer errorsClient.Catch(ctx, errors.PanicFlag(&hasPanicked))
//     ...
//     ...
//     // We have reached the end of the function, so we're not panicking.
//     hasPanicked = false
//   }
func PanicFlag(p *bool) Option { return panicFlag{p} }

type panicFlag struct {
	*bool
}

func (h panicFlag) isOption() {}

// Repanic returns an Option that determines whether Catch will re-panic after
// it reports an error.  This overrides the default in the client.
func Repanic(r bool) Option { return repanic(r) }

type repanic bool

func (r repanic) isOption() {}

// WithRequest returns an Option that informs Catch or Report of an http.Request
// that is being handled.  Information from the Request is included in the error
// report, if one is made.
func WithRequest(r *http.Request) Option { return withRequest{r} }

type withRequest struct {
	*http.Request
}

func (w withRequest) isOption() {}

// WithMessage returns an Option that sets a message to be included in the error
// report, if one is made.  v is converted to a string with fmt.Sprint.
func WithMessage(v ...interface{}) Option { return message(v) }

type message []interface{}

func (m message) isOption() {}

// WithMessagef returns an Option that sets a message to be included in the error
// report, if one is made.  format and v are converted to a string with fmt.Sprintf.
func WithMessagef(format string, v ...interface{}) Option { return messagef{format, v} }

type messagef struct {
	format string
	v      []interface{}
}

func (m messagef) isOption() {}

// Catch tries to recover a panic; if it succeeds, it writes an error report.
// It should be called by deferring it, like any other function for recovering
// panics.
//
// Catch can be called concurrently with other calls to Catch, Report or Reportf.
func (c *Client) Catch(ctx context.Context, opt ...Option) {
	panicked := false
	for _, o := range opt {
		switch o := o.(type) {
		case panicFlag:
			panicked = panicked || o.bool != nil && *o.bool
		}
	}
	x := recover()
	if x == nil && !panicked {
		return
	}
	var (
		r             *http.Request
		shouldRepanic = true
		messages      = []string{fmt.Sprint(x)}
	)
	if c != nil {
		shouldRepanic = c.RepanicDefault
	}
	for _, o := range opt {
		switch o := o.(type) {
		case repanic:
			shouldRepanic = bool(o)
		case withRequest:
			r = o.Request
		case message:
			messages = append(messages, fmt.Sprint(o...))
		case messagef:
			messages = append(messages, fmt.Sprintf(o.format, o.v...))
		}
	}
	c.logInternal(ctx, r, true, strings.Join(messages, " "))
	if shouldRepanic {
		panic(x)
	}
}

// Report writes an error report unconditionally, instead of only when a panic
// occurs.
// If r is non-nil, information from the Request is included in the error report.
//
// Report can be called concurrently with other calls to Catch, Report or Reportf.
func (c *Client) Report(ctx context.Context, r *http.Request, v ...interface{}) {
	c.logInternal(ctx, r, false, fmt.Sprint(v...))
}

// Reportf writes an error report unconditionally, instead of only when a panic
// occurs.
// If r is non-nil, information from the Request is included in the error report.
//
// Reportf can be called concurrently with other calls to Catch, Report or Reportf.
func (c *Client) Reportf(ctx context.Context, r *http.Request, format string, v ...interface{}) {
	c.logInternal(ctx, r, false, fmt.Sprintf(format, v...))
}

func (c *Client) logInternal(ctx context.Context, r *http.Request, isPanic bool, msg string) {
	// limit the stack trace to 16k.
	var buf [16384]byte
	stack := buf[0:runtime.Stack(buf[:], false)]
	message := msg + "\n" + chopStack(stack, isPanic)
	if c == nil {
		log.Println("Error report used nil client:", message)
		return
	}
	c.send(ctx, r, message)
}

func (s *loggingSender) send(ctx context.Context, r *http.Request, message string) {
	payload := map[string]interface{}{
		"eventTime":      time.Now().In(time.UTC).Format(time.RFC3339Nano),
		"message":        message,
		"serviceContext": s.serviceContext,
	}
	if r != nil {
		payload["context"] = map[string]interface{}{
			"httpRequest": map[string]interface{}{
				"method":    r.Method,
				"url":       r.Host + r.RequestURI,
				"userAgent": r.UserAgent(),
				"referrer":  r.Referer(),
				"remoteIp":  r.RemoteAddr,
			},
		}
	}
	e := logging.Entry{
		Severity: logging.Error,
		Payload:  payload,
	}
	err := s.logger.LogSync(ctx, e)
	if err != nil {
		log.Println("Error writing error report:", err, "report:", payload)
	}
}

func (s *errorApiSender) send(ctx context.Context, r *http.Request, message string) {
	time := time.Now()
	var errorContext *erpb.ErrorContext
	if r != nil {
		errorContext = &erpb.ErrorContext{
			HttpRequest: &erpb.HttpRequestContext{
				Method:    r.Method,
				Url:       r.Host + r.RequestURI,
				UserAgent: r.UserAgent(),
				Referrer:  r.Referer(),
				RemoteIp:  r.RemoteAddr,
			},
		}
	}
	req := erpb.ReportErrorEventRequest{
		ProjectName: s.projectID,
		Event: &erpb.ReportedErrorEvent{
			EventTime: &timestamp.Timestamp{
				Seconds: time.Unix(),
				Nanos:   int32(time.Nanosecond()),
			},
			ServiceContext: &s.serviceContext,
			Message:        message,
			Context:        errorContext,
		},
	}
	_, err := s.apiClient.ReportErrorEvent(ctx, &req)
	if err != nil {
		log.Println("Error writing error report:", err, "report:", message)
	}
}

// chopStack trims a stack trace so that the function which panics or calls
// Report is first.
func chopStack(s []byte, isPanic bool) string {
	var f []byte
	if isPanic {
		f = []byte("panic(")
	} else {
		f = []byte("cloud.google.com/go/errors.(*Client).Report")
	}

	lfFirst := bytes.IndexByte(s, '\n')
	if lfFirst == -1 {
		return string(s)
	}
	stack := s[lfFirst:]
	panicLine := bytes.Index(stack, f)
	if panicLine == -1 {
		return string(s)
	}
	stack = stack[panicLine+1:]
	for i := 0; i < 2; i++ {
		nextLine := bytes.IndexByte(stack, '\n')
		if nextLine == -1 {
			return string(s)
		}
		stack = stack[nextLine+1:]
	}
	return string(s[:lfFirst+1]) + string(stack)
}
