// Copyright 2017, Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AUTO-GENERATED CODE. DO NOT EDIT.

package debugger

import (
	"time"

	"cloud.google.com/go/internal/version"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	clouddebuggerpb "google.golang.org/genproto/googleapis/devtools/clouddebugger/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Debugger2CallOptions contains the retry settings for each method of Debugger2Client.
type Debugger2CallOptions struct {
	SetBreakpoint    []gax.CallOption
	GetBreakpoint    []gax.CallOption
	DeleteBreakpoint []gax.CallOption
	ListBreakpoints  []gax.CallOption
	ListDebuggees    []gax.CallOption
}

func defaultDebugger2ClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("clouddebugger.googleapis.com:443"),
		option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/cloud_debugger",
		),
	}
}

func defaultDebugger2CallOptions() *Debugger2CallOptions {
	retry := map[[2]string][]gax.CallOption{
		{"default", "idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
		{"default", "non_idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
	}
	return &Debugger2CallOptions{
		SetBreakpoint:    retry[[2]string{"default", "non_idempotent"}],
		GetBreakpoint:    retry[[2]string{"default", "idempotent"}],
		DeleteBreakpoint: retry[[2]string{"default", "idempotent"}],
		ListBreakpoints:  retry[[2]string{"default", "idempotent"}],
		ListDebuggees:    retry[[2]string{"default", "idempotent"}],
	}
}

// Debugger2Client is a client for interacting with Stackdriver Debugger API.
type Debugger2Client struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	debugger2Client clouddebuggerpb.Debugger2Client

	// The call options for this service.
	CallOptions *Debugger2CallOptions

	// The metadata to be sent with each request.
	xGoogHeader string
}

// NewDebugger2Client creates a new debugger2 client.
//
// The Debugger service provides the API that allows users to collect run-time
// information from a running application, without stopping or slowing it down
// and without modifying its state.  An application may include one or
// more replicated processes performing the same work.
//
// The application is represented using the Debuggee concept. The Debugger
// service provides a way to query for available Debuggees, but does not
// provide a way to create one.  A debuggee is created using the Controller
// service, usually by running a debugger agent with the application.
//
// The Debugger service enables the client to set one or more Breakpoints on a
// Debuggee and collect the results of the set Breakpoints.
func NewDebugger2Client(ctx context.Context, opts ...option.ClientOption) (*Debugger2Client, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultDebugger2ClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &Debugger2Client{
		conn:        conn,
		CallOptions: defaultDebugger2CallOptions(),

		debugger2Client: clouddebuggerpb.NewDebugger2Client(conn),
	}
	c.SetGoogleClientInfo()
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *Debugger2Client) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *Debugger2Client) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *Debugger2Client) SetGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", version.Go()}, keyval...)
	kv = append(kv, "gapic", version.Repo, "gax", gax.Version, "grpc", "")
	c.xGoogHeader = gax.XGoogHeader(kv...)
}

// SetBreakpoint sets the breakpoint to the debuggee.
func (c *Debugger2Client) SetBreakpoint(ctx context.Context, req *clouddebuggerpb.SetBreakpointRequest) (*clouddebuggerpb.SetBreakpointResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *clouddebuggerpb.SetBreakpointResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.debugger2Client.SetBreakpoint(ctx, req)
		return err
	}, c.CallOptions.SetBreakpoint...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetBreakpoint gets breakpoint information.
func (c *Debugger2Client) GetBreakpoint(ctx context.Context, req *clouddebuggerpb.GetBreakpointRequest) (*clouddebuggerpb.GetBreakpointResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *clouddebuggerpb.GetBreakpointResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.debugger2Client.GetBreakpoint(ctx, req)
		return err
	}, c.CallOptions.GetBreakpoint...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteBreakpoint deletes the breakpoint from the debuggee.
func (c *Debugger2Client) DeleteBreakpoint(ctx context.Context, req *clouddebuggerpb.DeleteBreakpointRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.debugger2Client.DeleteBreakpoint(ctx, req)
		return err
	}, c.CallOptions.DeleteBreakpoint...)
	return err
}

// ListBreakpoints lists all breakpoints for the debuggee.
func (c *Debugger2Client) ListBreakpoints(ctx context.Context, req *clouddebuggerpb.ListBreakpointsRequest) (*clouddebuggerpb.ListBreakpointsResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *clouddebuggerpb.ListBreakpointsResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.debugger2Client.ListBreakpoints(ctx, req)
		return err
	}, c.CallOptions.ListBreakpoints...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListDebuggees lists all the debuggees that the user can set breakpoints to.
func (c *Debugger2Client) ListDebuggees(ctx context.Context, req *clouddebuggerpb.ListDebuggeesRequest) (*clouddebuggerpb.ListDebuggeesResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *clouddebuggerpb.ListDebuggeesResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.debugger2Client.ListDebuggees(ctx, req)
		return err
	}, c.CallOptions.ListDebuggees...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
