// Copyright 2015 Google Inc. All Rights Reserved.
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

package bigquery

// TODO(mcgreevy): support dry-run mode when creating jobs.

import (
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"

	"golang.org/x/net/context"
	bq "google.golang.org/api/bigquery/v2"
)

const prodAddr = "https://www.googleapis.com/bigquery/v2/"

// ExternalData is a table which is stored outside of BigQuery.  It is implemented by GCSReference.
type ExternalData interface {
	externalDataConfig() bq.ExternalDataConfiguration
}

const Scope = "https://www.googleapis.com/auth/bigquery"
const userAgent = "gcloud-golang-bigquery/20160429"

// Client may be used to perform BigQuery operations.
type Client struct {
	service   service
	projectID string
}

// NewClient constructs a new Client which can perform BigQuery operations.
// Operations performed via the client are billed to the specified GCP project.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	o := []option.ClientOption{
		option.WithEndpoint(prodAddr),
		option.WithScopes(Scope),
		option.WithUserAgent(userAgent),
	}
	o = append(o, opts...)
	httpClient, endpoint, err := transport.NewHTTPClient(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("dialing: %v", err)
	}

	s, err := newBigqueryService(httpClient, endpoint)
	if err != nil {
		return nil, fmt.Errorf("constructing bigquery client: %v", err)
	}

	c := &Client{
		service:   s,
		projectID: projectID,
	}
	return c, nil
}

// Close closes any resources held by the client.
// Close should be called when the client is no longer needed.
// It need not be called at program exit.
func (c *Client) Close() error {
	return nil
}
