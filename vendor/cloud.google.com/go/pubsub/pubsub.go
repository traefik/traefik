// Copyright 2014 Google Inc. All Rights Reserved.
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

package pubsub // import "cloud.google.com/go/pubsub"

import (
	"fmt"
	"os"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"golang.org/x/net/context"
)

const (
	// ScopePubSub grants permissions to view and manage Pub/Sub
	// topics and subscriptions.
	ScopePubSub = "https://www.googleapis.com/auth/pubsub"

	// ScopeCloudPlatform grants permissions to view and manage your data
	// across Google Cloud Platform services.
	ScopeCloudPlatform = "https://www.googleapis.com/auth/cloud-platform"
)

const prodAddr = "https://pubsub.googleapis.com/"
const userAgent = "gcloud-golang-pubsub/20160927"

// Client is a Google Pub/Sub client scoped to a single project.
//
// Clients should be reused rather than being created as needed.
// A Client may be shared by multiple goroutines.
type Client struct {
	projectID string
	s         service
}

// NewClient creates a new PubSub client.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	var o []option.ClientOption
	// Environment variables for gcloud emulator:
	// https://cloud.google.com/sdk/gcloud/reference/beta/emulators/pubsub/
	if addr := os.Getenv("PUBSUB_EMULATOR_HOST"); addr != "" {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("grpc.Dial: %v", err)
		}
		o = []option.ClientOption{option.WithGRPCConn(conn)}
	} else {
		o = []option.ClientOption{option.WithUserAgent(userAgent)}
	}
	o = append(o, opts...)
	s, err := newPubSubService(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("constructing pubsub client: %v", err)
	}

	c := &Client{
		projectID: projectID,
		s:         s,
	}

	return c, nil
}

// Close closes any resources held by the client.
//
// Close need not be called at program exit.
func (c *Client) Close() error {
	return c.s.close()
}

func (c *Client) fullyQualifiedProjectName() string {
	return fmt.Sprintf("projects/%s", c.projectID)
}

// pageToken stores the next page token for a server response which is split over multiple pages.
type pageToken struct {
	tok      string
	explicit bool
}

func (pt *pageToken) set(tok string) {
	pt.tok = tok
	pt.explicit = true
}

func (pt *pageToken) get() string {
	return pt.tok
}

// more returns whether further pages should be fetched from the server.
func (pt *pageToken) more() bool {
	return pt.tok != "" || !pt.explicit
}

// stringsIterator provides an iterator API for a sequence of API page fetches that return lists of strings.
type stringsIterator struct {
	ctx     context.Context
	strings []string
	token   pageToken
	fetch   func(ctx context.Context, tok string) (*stringsPage, error)
}

// Next returns the next string. If there are no more strings, iterator.Done will be returned.
func (si *stringsIterator) Next() (string, error) {
	for len(si.strings) == 0 && si.token.more() {
		page, err := si.fetch(si.ctx, si.token.get())
		if err != nil {
			return "", err
		}
		si.token.set(page.tok)
		si.strings = page.strings
	}

	if len(si.strings) == 0 {
		return "", iterator.Done
	}

	s := si.strings[0]
	si.strings = si.strings[1:]

	return s, nil
}
