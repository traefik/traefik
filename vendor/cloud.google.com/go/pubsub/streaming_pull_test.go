// Copyright 2017 Google Inc. All Rights Reserved.
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

// TODO(jba): test keepalive
// TODO(jba): test that expired messages are not kept alive
// TODO(jba): test that when all messages expire, Stop returns.

import (
	"io"
	"reflect"
	"strconv"
	"testing"
	"time"

	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	timestamp    = &tspb.Timestamp{}
	testMessages = []*pb.ReceivedMessage{
		{AckId: "1", Message: &pb.PubsubMessage{Data: []byte{1}, PublishTime: timestamp}},
		{AckId: "2", Message: &pb.PubsubMessage{Data: []byte{2}, PublishTime: timestamp}},
		{AckId: "3", Message: &pb.PubsubMessage{Data: []byte{3}, PublishTime: timestamp}},
	}
)

func TestStreamingPullBasic(t *testing.T) {
	client, server := newFake(t)
	server.addStreamingPullMessages(testMessages)
	testStreamingPullIteration(t, client, server, testMessages)
}

func TestStreamingPullMultipleFetches(t *testing.T) {
	client, server := newFake(t)
	server.addStreamingPullMessages(testMessages[:1])
	server.addStreamingPullMessages(testMessages[1:])
	testStreamingPullIteration(t, client, server, testMessages)
}

func testStreamingPullIteration(t *testing.T, client *Client, server *fakeServer, msgs []*pb.ReceivedMessage) {
	if !useStreamingPull {
		t.SkipNow()
	}
	sub := client.Subscription("s")
	iter, err := sub.Pull(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(msgs); i++ {
		got, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		got.Done(i%2 == 0) // ack evens, nack odds
		want, err := toMessage(msgs[i])
		if err != nil {
			t.Fatal(err)
		}
		want.calledDone = true
		// Don't compare done; it's a function.
		got.done = nil
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got\n%#v\nwant\n%#v", i, got, want)
		}

	}
	iter.Stop()
	server.wait()
	for i := 0; i < len(msgs); i++ {
		id := msgs[i].AckId
		if i%2 == 0 {
			if !server.Acked[id] {
				t.Errorf("msg %q should have been acked but wasn't", id)
			}
		} else {
			if dl, ok := server.Deadlines[id]; !ok || dl != 0 {
				t.Errorf("msg %q should have been nacked but wasn't", id)
			}
		}
	}
}

func TestStreamingPullStop(t *testing.T) {
	if !useStreamingPull {
		t.SkipNow()
	}
	// After Stop is called, Next returns iterator.Done.
	client, server := newFake(t)
	server.addStreamingPullMessages(testMessages)
	sub := client.Subscription("s")
	iter, err := sub.Pull(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	msg, err := iter.Next()
	if err != nil {
		t.Fatal(err)
	}
	msg.Done(true)
	iter.Stop()
	// Next should always return the same error.
	for i := 0; i < 3; i++ {
		_, err = iter.Next()
		if want := iterator.Done; err != want {
			t.Fatalf("got <%v> %p, want <%v> %p", err, err, want, want)
		}
	}
}

func TestStreamingPullError(t *testing.T) {
	if !useStreamingPull {
		t.SkipNow()
	}
	client, server := newFake(t)
	server.addStreamingPullError(grpc.Errorf(codes.Internal, ""))
	sub := client.Subscription("s")
	iter, err := sub.Pull(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// Next should always return the same error.
	for i := 0; i < 3; i++ {
		_, err = iter.Next()
		if want := codes.Internal; grpc.Code(err) != want {
			t.Fatalf("got <%v>, want code %v", err, want)
		}
	}
}

func TestStreamingPullCancel(t *testing.T) {
	if !useStreamingPull {
		t.SkipNow()
	}
	// Test that canceling the iterator's context behaves correctly.
	client, server := newFake(t)
	server.addStreamingPullMessages(testMessages)
	sub := client.Subscription("s")
	ctx, cancel := context.WithCancel(context.Background())
	iter, err := sub.Pull(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = iter.Next()
	if err != nil {
		t.Fatal(err)
	}
	// Here we have one message read (but not acked), and two
	// in the iterator's buffer.
	cancel()
	// Further calls to Next will return Canceled.
	_, err = iter.Next()
	if got, want := err, context.Canceled; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	// Despite the unacked message, Stop will still return promptly.
	done := make(chan struct{})
	go func() {
		iter.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("iter.Stop timed out")
	}
}

func TestStreamingPullRetry(t *testing.T) {
	if !useStreamingPull {
		t.SkipNow()
	}
	// Check that we retry on io.EOF or Unavailable.
	client, server := newFake(t)
	server.addStreamingPullMessages(testMessages[:1])
	server.addStreamingPullError(io.EOF)
	server.addStreamingPullError(io.EOF)
	server.addStreamingPullMessages(testMessages[1:2])
	server.addStreamingPullError(grpc.Errorf(codes.Unavailable, ""))
	server.addStreamingPullError(grpc.Errorf(codes.Unavailable, ""))
	server.addStreamingPullMessages(testMessages[2:])
	testStreamingPullIteration(t, client, server, testMessages)
}

func TestStreamingPullConcurrent(t *testing.T) {
	if !useStreamingPull {
		t.SkipNow()
	}
	newMsg := func(i int) *pb.ReceivedMessage {
		return &pb.ReceivedMessage{
			AckId:   strconv.Itoa(i),
			Message: &pb.PubsubMessage{Data: []byte{byte(i)}, PublishTime: timestamp},
		}
	}

	// Multiple goroutines should be able to read from the same iterator.
	client, server := newFake(t)
	// Add a lot of messages, a few at a time, to make sure both threads get a chance.
	nMessages := 100
	for i := 0; i < nMessages; i += 2 {
		server.addStreamingPullMessages([]*pb.ReceivedMessage{newMsg(i), newMsg(i + 1)})
	}
	sub := client.Subscription("s")
	iter, err := sub.Pull(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	seenc := make(chan string)
	errc := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			for {
				msg, err := iter.Next()
				if err == iterator.Done {
					return
				}
				if err != nil {
					errc <- err
					return
				}
				// Must ack before sending to channel, or Stop may hang.
				msg.Done(true)
				seenc <- msg.ackID
			}
		}()
	}
	seen := map[string]bool{}
	for i := 0; i < nMessages; i++ {
		select {
		case err := <-errc:
			t.Fatal(err)
		case id := <-seenc:
			if seen[id] {
				t.Fatalf("duplicate ID %q", id)
			}
			seen[id] = true
		}
	}
	iter.Stop()
	if len(seen) != nMessages {
		t.Fatalf("got %d messages, want %d", len(seen), nMessages)
	}
}

func newFake(t *testing.T) (*Client, *fakeServer) {
	srv, err := newFakeServer()
	if err != nil {
		t.Fatal(err)
	}
	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(context.Background(), "projectID", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}
	return client, srv
}
