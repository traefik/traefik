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

package pubsub

import (
	"fmt"
	"strings"

	"cloud.google.com/go/iam"
	"golang.org/x/net/context"
)

const MaxPublishBatchSize = 1000

// Topic is a reference to a PubSub topic.
type Topic struct {
	s service

	// The fully qualified identifier for the topic, in the format "projects/<projid>/topics/<name>"
	name string
}

// CreateTopic creates a new topic.
// The specified topic ID must start with a letter, and contain only letters
// ([A-Za-z]), numbers ([0-9]), dashes (-), underscores (_), periods (.),
// tildes (~), plus (+) or percent signs (%). It must be between 3 and 255
// characters in length, and must not start with "goog".
// If the topic already exists an error will be returned.
func (c *Client) CreateTopic(ctx context.Context, id string) (*Topic, error) {
	t := c.Topic(id)
	err := c.s.createTopic(ctx, t.name)
	return t, err
}

// Topic creates a reference to a topic.
func (c *Client) Topic(id string) *Topic {
	return &Topic{
		s:    c.s,
		name: fmt.Sprintf("projects/%s/topics/%s", c.projectID, id),
	}
}

// Topics returns an iterator which returns all of the topics for the client's project.
func (c *Client) Topics(ctx context.Context) *TopicIterator {
	return &TopicIterator{
		s:    c.s,
		next: c.s.listProjectTopics(ctx, c.fullyQualifiedProjectName()),
	}
}

// TopicIterator is an iterator that returns a series of topics.
type TopicIterator struct {
	s    service
	next nextStringFunc
}

// Next returns the next topic. If there are no more topics, iterator.Done will be returned.
func (tps *TopicIterator) Next() (*Topic, error) {
	topicName, err := tps.next()
	if err != nil {
		return nil, err
	}
	return &Topic{s: tps.s, name: topicName}, nil
}

// ID returns the unique idenfier of the topic within its project.
func (t *Topic) ID() string {
	slash := strings.LastIndex(t.name, "/")
	if slash == -1 {
		// name is not a fully-qualified name.
		panic("bad topic name")
	}
	return t.name[slash+1:]
}

// String returns the printable globally unique name for the topic.
func (t *Topic) String() string {
	return t.name
}

// Delete deletes the topic.
func (t *Topic) Delete(ctx context.Context) error {
	return t.s.deleteTopic(ctx, t.name)
}

// Exists reports whether the topic exists on the server.
func (t *Topic) Exists(ctx context.Context) (bool, error) {
	if t.name == "_deleted-topic_" {
		return false, nil
	}

	return t.s.topicExists(ctx, t.name)
}

// Subscriptions returns an iterator which returns the subscriptions for this topic.
func (t *Topic) Subscriptions(ctx context.Context) *SubscriptionIterator {
	// NOTE: zero or more Subscriptions that are ultimately returned by this
	// Subscriptions iterator may belong to a different project to t.
	return &SubscriptionIterator{
		s:    t.s,
		next: t.s.listTopicSubscriptions(ctx, t.name),
	}
}

// Publish publishes the supplied Messages to the topic.
// If successful, the server-assigned message IDs are returned in the same order as the supplied Messages.
// At most MaxPublishBatchSize messages may be supplied.
func (t *Topic) Publish(ctx context.Context, msgs ...*Message) ([]string, error) {
	if len(msgs) == 0 {
		return nil, nil
	}
	if len(msgs) > MaxPublishBatchSize {
		return nil, fmt.Errorf("pubsub: got %d messages, but maximum batch size is %d", len(msgs), MaxPublishBatchSize)
	}
	return t.s.publishMessages(ctx, t.name, msgs)
}

func (t *Topic) IAM() *iam.Handle {
	return t.s.iamHandle(t.name)
}
