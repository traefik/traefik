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

// This file provides a fake/mock in-memory pubsub server.
// (Really just a mock at the moment, but we hope to turn it into
// more of a fake.)

import (
	"io"
	"sync"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/internal/testutil"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
)

type fakeServer struct {
	pb.PublisherServer
	pb.SubscriberServer

	Addr          string
	Acked         map[string]bool  // acked message IDs
	Deadlines     map[string]int32 // deadlines by message ID
	pullResponses []*pullResponse
	wg            sync.WaitGroup
}

type pullResponse struct {
	msgs []*pb.ReceivedMessage
	err  error
}

func newFakeServer() (*fakeServer, error) {
	srv, err := testutil.NewServer()
	if err != nil {
		return nil, err
	}
	fake := &fakeServer{
		Addr:      srv.Addr,
		Acked:     map[string]bool{},
		Deadlines: map[string]int32{},
	}
	pb.RegisterPublisherServer(srv.Gsrv, fake)
	pb.RegisterSubscriberServer(srv.Gsrv, fake)
	srv.Start()
	return fake, nil
}

// Each call to addStreamingPullMessages results in one StreamingPullResponse.
func (s *fakeServer) addStreamingPullMessages(msgs []*pb.ReceivedMessage) {
	s.pullResponses = append(s.pullResponses, &pullResponse{msgs, nil})
}

func (s *fakeServer) addStreamingPullError(err error) {
	s.pullResponses = append(s.pullResponses, &pullResponse{nil, err})
}

func (s *fakeServer) wait() {
	s.wg.Wait()
}

func (s *fakeServer) StreamingPull(stream pb.Subscriber_StreamingPullServer) error {
	// Receive initial request.
	_, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	// Consume and ignore subsequent requests.
	errc := make(chan error, 1)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			req, err := stream.Recv()
			if err != nil {
				errc <- err
				return
			}
			for _, id := range req.AckIds {
				s.Acked[id] = true
			}
			for i, id := range req.ModifyDeadlineAckIds {
				s.Deadlines[id] = req.ModifyDeadlineSeconds[i]
			}
		}
	}()
	// Send responses.
	for {
		if len(s.pullResponses) == 0 {
			// Nothing to send, so wait for the client to shut down the stream.
			err := <-errc // a real error, or at least EOF
			if err == io.EOF {
				return nil
			}
			return err
		}
		pr := s.pullResponses[0]
		s.pullResponses = s.pullResponses[1:]
		if pr.err != nil {
			// Add a slight delay to ensure the server receives any
			// messages en route from the client before shutting down the stream.
			// This reduces flakiness of tests involving retry.
			time.Sleep(100 * time.Millisecond)
		}
		if pr.err == io.EOF {
			return nil
		}
		if pr.err != nil {
			return pr.err
		}
		// Return any error from Recv.
		select {
		case err := <-errc:
			return err
		default:
		}
		res := &pb.StreamingPullResponse{ReceivedMessages: pr.msgs}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *fakeServer) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.Subscription, error) {
	return &pb.Subscription{
		Name:               req.Subscription,
		AckDeadlineSeconds: 10,
		PushConfig:         &pb.PushConfig{},
	}, nil
}
