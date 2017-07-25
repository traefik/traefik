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

package pubsub

import (
	"math"
	"time"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/internal/version"
	gax "github.com/googleapis/gax-go"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	subscriberProjectPathTemplate      = gax.MustCompilePathTemplate("projects/{project}")
	subscriberSubscriptionPathTemplate = gax.MustCompilePathTemplate("projects/{project}/subscriptions/{subscription}")
	subscriberTopicPathTemplate        = gax.MustCompilePathTemplate("projects/{project}/topics/{topic}")
)

// SubscriberCallOptions contains the retry settings for each method of SubscriberClient.
type SubscriberCallOptions struct {
	CreateSubscription []gax.CallOption
	GetSubscription    []gax.CallOption
	ListSubscriptions  []gax.CallOption
	DeleteSubscription []gax.CallOption
	ModifyAckDeadline  []gax.CallOption
	Acknowledge        []gax.CallOption
	Pull               []gax.CallOption
	StreamingPull      []gax.CallOption
	ModifyPushConfig   []gax.CallOption
}

func defaultSubscriberClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("pubsub.googleapis.com:443"),
		option.WithScopes(
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/pubsub",
		),
	}
}

func defaultSubscriberCallOptions() *SubscriberCallOptions {
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
	return &SubscriberCallOptions{
		CreateSubscription: retry[[2]string{"default", "idempotent"}],
		GetSubscription:    retry[[2]string{"default", "idempotent"}],
		ListSubscriptions:  retry[[2]string{"default", "idempotent"}],
		DeleteSubscription: retry[[2]string{"default", "idempotent"}],
		ModifyAckDeadline:  retry[[2]string{"default", "non_idempotent"}],
		Acknowledge:        retry[[2]string{"messaging", "non_idempotent"}],
		Pull:               retry[[2]string{"messaging", "non_idempotent"}],
		StreamingPull:      retry[[2]string{"messaging", "non_idempotent"}],
		ModifyPushConfig:   retry[[2]string{"default", "non_idempotent"}],
	}
}

// SubscriberClient is a client for interacting with Google Cloud Pub/Sub API.
type SubscriberClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	subscriberClient pubsubpb.SubscriberClient

	// The call options for this service.
	CallOptions *SubscriberCallOptions

	// The metadata to be sent with each request.
	xGoogHeader string
}

// NewSubscriberClient creates a new subscriber client.
//
// The service that an application uses to manipulate subscriptions and to
// consume messages from a subscription via the `Pull` method.
func NewSubscriberClient(ctx context.Context, opts ...option.ClientOption) (*SubscriberClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultSubscriberClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &SubscriberClient{
		conn:        conn,
		CallOptions: defaultSubscriberCallOptions(),

		subscriberClient: pubsubpb.NewSubscriberClient(conn),
	}
	c.SetGoogleClientInfo()
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *SubscriberClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *SubscriberClient) Close() error {
	return c.conn.Close()
}

// SetGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *SubscriberClient) SetGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", version.Go()}, keyval...)
	kv = append(kv, "gapic", version.Repo, "gax", gax.Version, "grpc", "")
	c.xGoogHeader = gax.XGoogHeader(kv...)
}

// SubscriberProjectPath returns the path for the project resource.
func SubscriberProjectPath(project string) string {
	path, err := subscriberProjectPathTemplate.Render(map[string]string{
		"project": project,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// SubscriberSubscriptionPath returns the path for the subscription resource.
func SubscriberSubscriptionPath(project, subscription string) string {
	path, err := subscriberSubscriptionPathTemplate.Render(map[string]string{
		"project":      project,
		"subscription": subscription,
	})
	if err != nil {
		panic(err)
	}
	return path
}

// SubscriberTopicPath returns the path for the topic resource.
func SubscriberTopicPath(project, topic string) string {
	path, err := subscriberTopicPathTemplate.Render(map[string]string{
		"project": project,
		"topic":   topic,
	})
	if err != nil {
		panic(err)
	}
	return path
}

func (c *SubscriberClient) SubscriptionIAM(subscription *pubsubpb.Subscription) *iam.Handle {
	return iam.InternalNewHandle(c.Connection(), subscription.Name)
}

func (c *SubscriberClient) TopicIAM(topic *pubsubpb.Topic) *iam.Handle {
	return iam.InternalNewHandle(c.Connection(), topic.Name)
}

// CreateSubscription creates a subscription to a given topic.
// If the subscription already exists, returns `ALREADY_EXISTS`.
// If the corresponding topic doesn't exist, returns `NOT_FOUND`.
//
// If the name is not provided in the request, the server will assign a random
// name for this subscription on the same project as the topic, conforming
// to the
// [resource name format](https://cloud.google.com/pubsub/docs/overview#names).
// The generated name is populated in the returned Subscription object.
// Note that for REST API requests, you must specify a name in the request.
func (c *SubscriberClient) CreateSubscription(ctx context.Context, req *pubsubpb.Subscription) (*pubsubpb.Subscription, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *pubsubpb.Subscription
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.subscriberClient.CreateSubscription(ctx, req)
		return err
	}, c.CallOptions.CreateSubscription...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSubscription gets the configuration details of a subscription.
func (c *SubscriberClient) GetSubscription(ctx context.Context, req *pubsubpb.GetSubscriptionRequest) (*pubsubpb.Subscription, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *pubsubpb.Subscription
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.subscriberClient.GetSubscription(ctx, req)
		return err
	}, c.CallOptions.GetSubscription...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListSubscriptions lists matching subscriptions.
func (c *SubscriberClient) ListSubscriptions(ctx context.Context, req *pubsubpb.ListSubscriptionsRequest) *SubscriptionIterator {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	it := &SubscriptionIterator{}
	it.InternalFetch = func(pageSize int, pageToken string) ([]*pubsubpb.Subscription, string, error) {
		var resp *pubsubpb.ListSubscriptionsResponse
		req.PageToken = pageToken
		if pageSize > math.MaxInt32 {
			req.PageSize = math.MaxInt32
		} else {
			req.PageSize = int32(pageSize)
		}
		err := gax.Invoke(ctx, func(ctx context.Context) error {
			var err error
			resp, err = c.subscriberClient.ListSubscriptions(ctx, req)
			return err
		}, c.CallOptions.ListSubscriptions...)
		if err != nil {
			return nil, "", err
		}
		return resp.Subscriptions, resp.NextPageToken, nil
	}
	fetch := func(pageSize int, pageToken string) (string, error) {
		items, nextPageToken, err := it.InternalFetch(pageSize, pageToken)
		if err != nil {
			return "", err
		}
		it.items = append(it.items, items...)
		return nextPageToken, nil
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(fetch, it.bufLen, it.takeBuf)
	return it
}

// DeleteSubscription deletes an existing subscription. All messages retained in the subscription
// are immediately dropped. Calls to `Pull` after deletion will return
// `NOT_FOUND`. After a subscription is deleted, a new one may be created with
// the same name, but the new one has no association with the old
// subscription or its topic unless the same topic is specified.
func (c *SubscriberClient) DeleteSubscription(ctx context.Context, req *pubsubpb.DeleteSubscriptionRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.subscriberClient.DeleteSubscription(ctx, req)
		return err
	}, c.CallOptions.DeleteSubscription...)
	return err
}

// ModifyAckDeadline modifies the ack deadline for a specific message. This method is useful
// to indicate that more time is needed to process a message by the
// subscriber, or to make the message available for redelivery if the
// processing was interrupted. Note that this does not modify the
// subscription-level `ackDeadlineSeconds` used for subsequent messages.
func (c *SubscriberClient) ModifyAckDeadline(ctx context.Context, req *pubsubpb.ModifyAckDeadlineRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.subscriberClient.ModifyAckDeadline(ctx, req)
		return err
	}, c.CallOptions.ModifyAckDeadline...)
	return err
}

// Acknowledge acknowledges the messages associated with the `ack_ids` in the
// `AcknowledgeRequest`. The Pub/Sub system can remove the relevant messages
// from the subscription.
//
// Acknowledging a message whose ack deadline has expired may succeed,
// but such a message may be redelivered later. Acknowledging a message more
// than once will not result in an error.
func (c *SubscriberClient) Acknowledge(ctx context.Context, req *pubsubpb.AcknowledgeRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.subscriberClient.Acknowledge(ctx, req)
		return err
	}, c.CallOptions.Acknowledge...)
	return err
}

// Pull pulls messages from the server. Returns an empty list if there are no
// messages available in the backlog. The server may return `UNAVAILABLE` if
// there are too many concurrent pull requests pending for the given
// subscription.
func (c *SubscriberClient) Pull(ctx context.Context, req *pubsubpb.PullRequest) (*pubsubpb.PullResponse, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp *pubsubpb.PullResponse
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.subscriberClient.Pull(ctx, req)
		return err
	}, c.CallOptions.Pull...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// StreamingPull (EXPERIMENTAL) StreamingPull is an experimental feature. This RPC will
// respond with UNIMPLEMENTED errors unless you have been invited to test
// this feature. Contact cloud-pubsub@google.com with any questions.
//
// Establishes a stream with the server, which sends messages down to the
// client. The client streams acknowledgements and ack deadline modifications
// back to the server. The server will close the stream and return the status
// on any error. The server may close the stream with status `OK` to reassign
// server-side resources, in which case, the client should re-establish the
// stream. `UNAVAILABLE` may also be returned in the case of a transient error
// (e.g., a server restart). These should also be retried by the client. Flow
// control can be achieved by configuring the underlying RPC channel.
func (c *SubscriberClient) StreamingPull(ctx context.Context) (pubsubpb.Subscriber_StreamingPullClient, error) {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	var resp pubsubpb.Subscriber_StreamingPullClient
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.subscriberClient.StreamingPull(ctx)
		return err
	}, c.CallOptions.StreamingPull...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ModifyPushConfig modifies the `PushConfig` for a specified subscription.
//
// This may be used to change a push subscription to a pull one (signified by
// an empty `PushConfig`) or vice versa, or change the endpoint URL and other
// attributes of a push subscription. Messages will accumulate for delivery
// continuously through the call regardless of changes to the `PushConfig`.
func (c *SubscriberClient) ModifyPushConfig(ctx context.Context, req *pubsubpb.ModifyPushConfigRequest) error {
	ctx = insertXGoog(ctx, c.xGoogHeader)
	err := gax.Invoke(ctx, func(ctx context.Context) error {
		var err error
		_, err = c.subscriberClient.ModifyPushConfig(ctx, req)
		return err
	}, c.CallOptions.ModifyPushConfig...)
	return err
}

// SubscriptionIterator manages a stream of *pubsubpb.Subscription.
type SubscriptionIterator struct {
	items    []*pubsubpb.Subscription
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*pubsubpb.Subscription, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *SubscriptionIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *SubscriptionIterator) Next() (*pubsubpb.Subscription, error) {
	var item *pubsubpb.Subscription
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *SubscriptionIterator) bufLen() int {
	return len(it.items)
}

func (it *SubscriptionIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}
