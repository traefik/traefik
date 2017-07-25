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

package errorreporting

import (
	"time"

	"cloud.google.com/go/internal/version"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	clouderrorreportingpb "google.golang.org/genproto/googleapis/devtools/clouderrorreporting/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	errorGroupGroupPathTemplate = gax.MustCompilePathTemplate("projects/{project}/groups/{group}")
)

// ErrorGroupCallOptions contains the retry settings for each method of ErrorGroupClient.
type ErrorGroupCallOptions struct {
	GetGroup    []gax.CallOption
	UpdateGroup []gax.CallOption
}

func defaultErrorGroupClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("clouderrorreporting.googleapis.com:443"),
		option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
		),
	}
}

func defaultErrorGroupCallOptions() *ErrorGroupCallOptions {
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
	}
	return &ErrorGroupCallOptions{
		GetGroup:    retry[[2]string{"default", "idempotent"}],
		UpdateGroup: retry[[2]string{"default", "idempotent"}],
	}
}

// ErrorGroupClient is a client for interacting with Stackdriver Error Reporting API.
type ErrorGroupClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	errorGroupClient clouderrorreportingpb.ErrorGroupServiceClient

	// The call options for this service.
	CallOptions *ErrorGroupCallOptions

	// The metadata to be sent with each request.
	xGoogHeader string
}

// NewErrorGroupClient creates a new error group service client.
//
// Service for retrieving and updating individual error groups.
func NewErrorGroupClient(ctx context.Context, opts ...option.ClientOption) (*ErrorGroupClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultErrorGroupClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &ErrorGroupClient{
		conn:        conn,
		CallOptions: defaultErrorGroupCallOptions(),

		errorGroupClient: clouderrorreportingpb.NewErrorGroupServiceClient(conn),
	}
	c.SetGoogleClientInfo()
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *ErrorGroupClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *ErrorGroupClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *ErrorGroupClient) SetGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", version.Go()}, keyval...)
	kv = append(kv, "gapic", version.Repo, "gax", gax.Version, "grpc", "")
	c.xGoogHeader = gax.XGoogHeader(kv...)
}

// ErrorGroupGroupPath returns the path for the group resource.
func ErrorGroupGroupPath(project, group string) string {
	path, err := errorGroupGroupPathTemplate.Render(map[string]string{
		"project": project,
		"group":   group,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// GetGroup get the specified group.
func (c *ErrorGroupClient) GetGroup(ctx context.Context, req *clouderrorreportingpb.GetGroupRequest) (*clouderrorreportingpb.ErrorGroup, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *clouderrorreportingpb.ErrorGroup
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.errorGroupClient.GetGroup(ctx, req)
		return err
	}, c.CallOptions.GetGroup...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateGroup replace the data for the specified group.
// Fails if the group does not exist.
func (c *ErrorGroupClient) UpdateGroup(ctx context.Context, req *clouderrorreportingpb.UpdateGroupRequest) (*clouderrorreportingpb.ErrorGroup, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *clouderrorreportingpb.ErrorGroup
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.errorGroupClient.UpdateGroup(ctx, req)
		return err
	}, c.CallOptions.UpdateGroup...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
