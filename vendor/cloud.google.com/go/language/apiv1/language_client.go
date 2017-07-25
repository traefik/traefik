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

package language

import (
	"time"

	"cloud.google.com/go/internal/version"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// CallOptions contains the retry settings for each method of Client.
type CallOptions struct {
	AnalyzeSentiment []gax.CallOption
	AnalyzeEntities  []gax.CallOption
	AnalyzeSyntax    []gax.CallOption
	AnnotateText     []gax.CallOption
}

func defaultClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("language.googleapis.com:443"),
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
	}
	return &CallOptions{
		AnalyzeSentiment: retry[[2]string{"default", "idempotent"}],
		AnalyzeEntities:  retry[[2]string{"default", "idempotent"}],
		AnalyzeSyntax:    retry[[2]string{"default", "idempotent"}],
		AnnotateText:     retry[[2]string{"default", "idempotent"}],
	}
}

// Client is a client for interacting with Google Cloud Natural Language API.
type Client struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	client languagepb.LanguageServiceClient

	// The call options for this service.
	CallOptions *CallOptions

	// The metadata to be sent with each request.
	xGoogHeader string
}

// NewClient creates a new language service client.
//
// Provides text analysis operations such as sentiment analysis and entity
// recognition.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &Client{
		conn:        conn,
		CallOptions: defaultCallOptions(),

		client: languagepb.NewLanguageServiceClient(conn),
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

// AnalyzeSentiment analyzes the sentiment of the provided text.
func (c *Client) AnalyzeSentiment(ctx context.Context, req *languagepb.AnalyzeSentimentRequest) (*languagepb.AnalyzeSentimentResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *languagepb.AnalyzeSentimentResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.AnalyzeSentiment(ctx, req)
		return err
	}, c.CallOptions.AnalyzeSentiment...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AnalyzeEntities finds named entities (currently finds proper names) in the text,
// entity types, salience, mentions for each entity, and other properties.
func (c *Client) AnalyzeEntities(ctx context.Context, req *languagepb.AnalyzeEntitiesRequest) (*languagepb.AnalyzeEntitiesResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *languagepb.AnalyzeEntitiesResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.AnalyzeEntities(ctx, req)
		return err
	}, c.CallOptions.AnalyzeEntities...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AnalyzeSyntax analyzes the syntax of the text and provides sentence boundaries and
// tokenization along with part of speech tags, dependency trees, and other
// properties.
func (c *Client) AnalyzeSyntax(ctx context.Context, req *languagepb.AnalyzeSyntaxRequest) (*languagepb.AnalyzeSyntaxResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *languagepb.AnalyzeSyntaxResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.AnalyzeSyntax(ctx, req)
		return err
	}, c.CallOptions.AnalyzeSyntax...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// AnnotateText a convenience method that provides all the features that analyzeSentiment,
// analyzeEntities, and analyzeSyntax provide in one call.
func (c *Client) AnnotateText(ctx context.Context, req *languagepb.AnnotateTextRequest) (*languagepb.AnnotateTextResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *languagepb.AnnotateTextResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.client.AnnotateText(ctx, req)
		return err
	}, c.CallOptions.AnnotateText...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
