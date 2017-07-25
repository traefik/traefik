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

package pubsub

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/internal/testutil"
	"google.golang.org/api/option"
)

// messageData is used to hold the contents of a message so that it can be compared against the contents
// of another message without regard to irrelevant fields.
type messageData struct {
	ID         string
	Data       []byte
	Attributes map[string]string
}

func extractMessageData(m *Message) *messageData {
	return &messageData{
		ID:         m.ID,
		Data:       m.Data,
		Attributes: m.Attributes,
	}
}

func TestAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	ctx := context.Background()
	ts := testutil.TokenSource(ctx, ScopePubSub, ScopeCloudPlatform)
	if ts == nil {
		t.Skip("Integration tests skipped. See CONTRIBUTING.md for details")
	}

	now := time.Now()
	topicName := fmt.Sprintf("topic-%d", now.Unix())
	subName := fmt.Sprintf("subscription-%d", now.Unix())

	client, err := NewClient(ctx, testutil.ProjID(), option.WithTokenSource(ts))
	if err != nil {
		t.Fatalf("Creating client error: %v", err)
	}
	defer client.Close()

	var topic *Topic
	if topic, err = client.CreateTopic(ctx, topicName); err != nil {
		t.Errorf("CreateTopic error: %v", err)
	}

	var sub *Subscription
	if sub, err = client.CreateSubscription(ctx, subName, topic, 0, nil); err != nil {
		t.Errorf("CreateSub error: %v", err)
	}

	exists, err := topic.Exists(ctx)
	if err != nil {
		t.Fatalf("TopicExists error: %v", err)
	}
	if !exists {
		t.Errorf("topic %s should exist, but it doesn't", topic)
	}

	exists, err = sub.Exists(ctx)
	if err != nil {
		t.Fatalf("SubExists error: %v", err)
	}
	if !exists {
		t.Errorf("subscription %s should exist, but it doesn't", subName)
	}

	msgs := []*Message{}
	for i := 0; i < 10; i++ {
		text := fmt.Sprintf("a message with an index %d", i)
		attrs := make(map[string]string)
		attrs["foo"] = "bar"
		msgs = append(msgs, &Message{
			Data:       []byte(text),
			Attributes: attrs,
		})
	}

	ids, err := topic.Publish(ctx, msgs...)
	if err != nil {
		t.Fatalf("Publish (1) error: %v", err)
	}

	if len(ids) != len(msgs) {
		t.Errorf("unexpected number of message IDs received; %d, want %d", len(ids), len(msgs))
	}

	want := make(map[string]*messageData)
	for i, m := range msgs {
		md := extractMessageData(m)
		md.ID = ids[i]
		want[md.ID] = md
	}

	// Use a timeout to ensure that Pull does not block indefinitely if there are unexpectedly few messages available.
	timeoutCtx, _ := context.WithTimeout(ctx, time.Minute)
	it, err := sub.Pull(timeoutCtx)
	if err != nil {
		t.Fatalf("error constructing iterator: %v", err)
	}
	defer it.Stop()
	got := make(map[string]*messageData)
	for i := 0; i < len(want); i++ {
		m, err := it.Next()
		if err != nil {
			t.Fatalf("error getting next message: %v", err)
		}
		md := extractMessageData(m)
		got[md.ID] = md
		m.Done(true)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("messages: got: %v ; want: %v", got, want)
	}

	// base64 test
	data := "=@~"
	_, err = topic.Publish(ctx, &Message{Data: []byte(data)})
	if err != nil {
		t.Fatalf("Publish error: %v", err)
	}

	m, err := it.Next()
	if err != nil {
		t.Fatalf("Pull error: %v", err)
	}

	if string(m.Data) != data {
		t.Errorf("unexpected message received; %s, want %s", string(m.Data), data)
	}
	m.Done(true)

	if msg, ok := testIAM(ctx, topic.IAM(), "pubsub.topics.get"); !ok {
		t.Errorf("topic IAM: %s", msg)
	}
	if msg, ok := testIAM(ctx, sub.IAM(), "pubsub.subscriptions.get"); !ok {
		t.Errorf("sub IAM: %s", msg)
	}

	err = sub.Delete(ctx)
	if err != nil {
		t.Errorf("DeleteSub error: %v", err)
	}

	err = topic.Delete(ctx)
	if err != nil {
		t.Errorf("DeleteTopic error: %v", err)
	}
}

// IAM tests.
// NOTE: for these to succeed, the test runner identity must have the Pub/Sub Admin or Owner roles.
// To set, visit https://console.developers.google.com, select "IAM & Admin" from the top-left
// menu, choose the account, click the Roles dropdown, and select "Pub/Sub > Pub/Sub Admin".
// TODO(jba): move this to a testing package within cloud.google.com/iam, so we can re-use it.
func testIAM(ctx context.Context, h *iam.Handle, permission string) (msg string, ok bool) {
	// Attempting to add an non-existent identity  (e.g. "alice@example.com") causes the service
	// to return an internal error, so use a real identity.
	const member = "domain:google.com"

	var policy *iam.Policy
	var err error

	if policy, err = h.Policy(ctx); err != nil {
		return fmt.Sprintf("Policy: %v", err), false
	}
	// The resource is new, so the policy should be empty.
	if got := policy.Roles(); len(got) > 0 {
		return fmt.Sprintf("initially: got roles %v, want none", got), false
	}
	// Add a member, set the policy, then check that the member is present.
	policy.Add(member, iam.Viewer)
	if err := h.SetPolicy(ctx, policy); err != nil {
		return fmt.Sprintf("SetPolicy: %v", err), false
	}
	if policy, err = h.Policy(ctx); err != nil {
		return fmt.Sprintf("Policy: %v", err), false
	}
	if got, want := policy.Members(iam.Viewer), []string{member}; !reflect.DeepEqual(got, want) {
		return fmt.Sprintf("after Add: got %v, want %v", got, want), false
	}
	// Now remove that member, set the policy, and check that it's empty again.
	policy.Remove(member, iam.Viewer)
	if err := h.SetPolicy(ctx, policy); err != nil {
		return fmt.Sprintf("SetPolicy: %v", err), false
	}
	if policy, err = h.Policy(ctx); err != nil {
		return fmt.Sprintf("Policy: %v", err), false
	}
	if got := policy.Roles(); len(got) > 0 {
		return fmt.Sprintf("after Remove: got roles %v, want none", got), false
	}
	// Call TestPermissions.
	// Because this user is an admin, it has all the permissions on the
	// resource type. Note: the service fails if we ask for inapplicable
	// permissions (e.g. a subscription permission on a topic, or a topic
	// create permission on a topic rather than its parent).
	wantPerms := []string{permission}
	gotPerms, err := h.TestPermissions(ctx, wantPerms)
	if err != nil {
		return fmt.Sprintf("TestPermissions: %v", err), false
	}
	if !reflect.DeepEqual(gotPerms, wantPerms) {
		return fmt.Sprintf("TestPermissions: got %v, want %v", gotPerms, wantPerms), false
	}
	return "", true
}
