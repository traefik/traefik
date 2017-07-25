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

package speech

import (
	"time"

	"cloud.google.com/go/internal/version"
	"cloud.google.com/go/longrunning"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1beta1"
	longrunningpb "google.golang.org/genproto/googleapis/longrunning"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// CallOptions contains the retry settings for each method of Client.
type CallOptions struct {
	SyncRecognize      []gax.CallOption
	AsyncRecognize     []gax.CallOption
	StreamingRecognize []gax.CallOption
}

func defaultClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("speech.googleapis.com:443"),
		option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
		),
	}
}

func defaultCallOptions() *CallOptions {
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
	return &CallOptions{
		SyncRecognize:      retry[[2]string{"default", "idempotent"}],
		AsyncRecognize:     retry[[2]string{"default", "idempotent"}],
		StreamingRecognize: retry[[2]string{"default", "non_idempotent"}],
	}
}

// Client is a client for interacting with Google Cloud Speech API.
type Client struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client speechpb.SpeechClient

	// The call options for this service.
	CallOptions *CallOptions

	// The metadata to be sent with each request.
	xGoogHeader string
}

// NewClient creates a new speech client.
//
// Service that implements Google Cloud Speech API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &Client{
		conn:        conn,
		CallOptions: defaultCallOptions(),

		client: speechpb.NewSpeechClient(conn),
	}
	c.SetGoogleClientInfo()
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *Client) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *Client) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *Client) SetGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", version.Go()}, keyval...)
	kv = append(kv, "gapic", version.Repo, "gax", gax.Version, "grpc", "")
	c.xGoogHeader = gax.XGoogHeader(kv...)
}

// SyncRecognize perform synchronous speech-recognition: receive results after all audio
// has been sent and processed.
func (c *Client) SyncRecognize(ctx context.Context, req *speechpb.SyncRecognizeRequest) (*speechpb.SyncRecognizeResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *speechpb.SyncRecognizeResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.SyncRecognize(ctx, req)
		return err
	}, c.CallOptions.SyncRecognize...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AsyncRecognize perform asynchronous speech-recognition: receive results via the
// google.longrunning.Operations interface. Returns either an
// `Operation.error` or an `Operation.response` which contains
// an `AsyncRecognizeResponse` message.
func (c *Client) AsyncRecognize(ctx context.Context, req *speechpb.AsyncRecognizeRequest) (*AsyncRecognizeResponseOperation, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *longrunningpb.Operation
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.AsyncRecognize(ctx, req)
		return err
	}, c.CallOptions.AsyncRecognize...)
	if err != nil {
		return nil, err
	}
	return &AsyncRecognizeResponseOperation{
		lro: longrunning.InternalNewOperation(c.Connection(), resp),
	}, nil
}

// StreamingRecognize perform bidirectional streaming speech-recognition: receive results while
// sending audio. This method is only available via the gRPC API (not REST).
func (c *Client) StreamingRecognize(ctx context.Context) (speechpb.Speech_StreamingRecognizeClient, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp speechpb.Speech_StreamingRecognizeClient
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.StreamingRecognize(ctx)
		return err
	}, c.CallOptions.StreamingRecognize...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AsyncRecognizeResponseOperation manages a long-running operation yielding speechpb.AsyncRecognizeResponse.
type AsyncRecognizeResponseOperation struct {
	lro *longrunning.Operation
}

// AsyncRecognizeResponseOperation returns a new AsyncRecognizeResponseOperation from a given name.
// The name must be that of a previously created AsyncRecognizeResponseOperation, possibly from a different process.
func (c *Client) AsyncRecognizeResponseOperation(name string) *AsyncRecognizeResponseOperation {
	return &AsyncRecognizeResponseOperation{
		lro: longrunning.InternalNewOperation(c.Connection(), &longrunningpb.Operation{Name: name}),
	}
}

// Wait blocks until the long-running operation is completed, returning the response and any errors encountered.
//
// See documentation of Poll for error-handling information.
func (op *AsyncRecognizeResponseOperation) Wait(ctx context.Context) (*speechpb.AsyncRecognizeResponse, error) {
	var resp speechpb.AsyncRecognizeResponse
	if err := op.lro.Wait(ctx, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Poll fetches the latest state of the long-running operation.
//
// Poll also fetches the latest metadata, which can be retrieved by Metadata.
//
// If Poll fails, the error is returned and op is unmodified. If Poll succeeds and
// the operation has completed with failure, the error is returned and op.Done will return true.
// If Poll succeeds and the operation has completed successfully,
// op.Done will return true, and the response of the operation is returned.
// If Poll succeeds and the operation has not completed, the returned response and error are both nil.
func (op *AsyncRecognizeResponseOperation) Poll(ctx context.Context) (*speechpb.AsyncRecognizeResponse, error) {
	var resp speechpb.AsyncRecognizeResponse
	if err := op.lro.Poll(ctx, &resp); err != nil {
		return nil, err
	}
	if !op.Done() {
		return nil, nil
	}
	return &resp, nil
}

// Metadata returns metadata associated with the long-running operation.
// Metadata itself does not contact the server, but Poll does.
// To get the latest metadata, call this method after a successful call to Poll.
// If the metadata is not available, the returned metadata and error are both nil.
func (op *AsyncRecognizeResponseOperation) Metadata() (*speechpb.AsyncRecognizeMetadata, error) {
	var meta speechpb.AsyncRecognizeMetadata
	if err := op.lro.Metadata(&meta); err == longrunning.ErrNoMetadata {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &meta, nil
}

// Done reports whether the long-running operation has completed.
func (op *AsyncRecognizeResponseOperation) Done() bool {
	return op.lro.Done()
}

// Name returns the name of the long-running operation.
// The name is assigned by the server and is unique within the service from which the operation is created.
func (op *AsyncRecognizeResponseOperation) Name() string {
	return op.lro.Name()
}
