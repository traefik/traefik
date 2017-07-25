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
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"
)

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var _ = io.EOF
var _ = ptypes.MarshalAny
var _ status.Status

type mockPublisherServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	pubsubpb.PublisherServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockPublisherServer) CreateTopic(_ context.Context, req *pubsubpb.Topic) (*pubsubpb.Topic, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.Topic), nil
}

func (s *mockPublisherServer) Publish(_ context.Context, req *pubsubpb.PublishRequest) (*pubsubpb.PublishResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.PublishResponse), nil
}

func (s *mockPublisherServer) GetTopic(_ context.Context, req *pubsubpb.GetTopicRequest) (*pubsubpb.Topic, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.Topic), nil
}

func (s *mockPublisherServer) ListTopics(_ context.Context, req *pubsubpb.ListTopicsRequest) (*pubsubpb.ListTopicsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.ListTopicsResponse), nil
}

func (s *mockPublisherServer) ListTopicSubscriptions(_ context.Context, req *pubsubpb.ListTopicSubscriptionsRequest) (*pubsubpb.ListTopicSubscriptionsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.ListTopicSubscriptionsResponse), nil
}

func (s *mockPublisherServer) DeleteTopic(_ context.Context, req *pubsubpb.DeleteTopicRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

type mockIamPolicyServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	iampb.IAMPolicyServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockIamPolicyServer) SetIamPolicy(_ context.Context, req *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.Policy), nil
}

func (s *mockIamPolicyServer) GetIamPolicy(_ context.Context, req *iampb.GetIamPolicyRequest) (*iampb.Policy, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.Policy), nil
}

func (s *mockIamPolicyServer) TestIamPermissions(_ context.Context, req *iampb.TestIamPermissionsRequest) (*iampb.TestIamPermissionsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.TestIamPermissionsResponse), nil
}

type mockSubscriberServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	pubsubpb.SubscriberServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockSubscriberServer) CreateSubscription(_ context.Context, req *pubsubpb.Subscription) (*pubsubpb.Subscription, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.Subscription), nil
}

func (s *mockSubscriberServer) GetSubscription(_ context.Context, req *pubsubpb.GetSubscriptionRequest) (*pubsubpb.Subscription, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.Subscription), nil
}

func (s *mockSubscriberServer) ListSubscriptions(_ context.Context, req *pubsubpb.ListSubscriptionsRequest) (*pubsubpb.ListSubscriptionsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.ListSubscriptionsResponse), nil
}

func (s *mockSubscriberServer) DeleteSubscription(_ context.Context, req *pubsubpb.DeleteSubscriptionRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockSubscriberServer) ModifyAckDeadline(_ context.Context, req *pubsubpb.ModifyAckDeadlineRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockSubscriberServer) Acknowledge(_ context.Context, req *pubsubpb.AcknowledgeRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockSubscriberServer) Pull(_ context.Context, req *pubsubpb.PullRequest) (*pubsubpb.PullResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*pubsubpb.PullResponse), nil
}

func (s *mockSubscriberServer) StreamingPull(stream pubsubpb.Subscriber_StreamingPullServer) error {
	for {
		if req, err := stream.Recv(); err == io.EOF {
			break
		} else if err != nil {
			return err
		} else {
			s.reqs = append(s.reqs, req)
		}
	}
	if s.err != nil {
		return s.err
	}
	for _, v := range s.resps {
		if err := stream.Send(v.(*pubsubpb.StreamingPullResponse)); err != nil {
			return err
		}
	}
	return nil
}

func (s *mockSubscriberServer) ModifyPushConfig(_ context.Context, req *pubsubpb.ModifyPushConfigRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var (
	mockPublisher  mockPublisherServer
	mockIamPolicy  mockIamPolicyServer
	mockSubscriber mockSubscriberServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	pubsubpb.RegisterPublisherServer(serv, &mockPublisher)
	iampb.RegisterIAMPolicyServer(serv, &mockIamPolicy)
	pubsubpb.RegisterSubscriberServer(serv, &mockSubscriber)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	go serv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	clientOpt = option.WithGRPCConn(conn)

	os.Exit(m.Run())
}

func TestPublisherCreateTopic(t *testing.T) {
	var name2 string = "name2-1052831874"
	var expectedResponse = &pubsubpb.Topic{
		Name: name2,
	}

	mockPublisher.err = nil
	mockPublisher.reqs = nil

	mockPublisher.resps = append(mockPublisher.resps[:0], expectedResponse)

	var formattedName string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.Topic{
		Name: formattedName,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateTopic(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockPublisher.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestPublisherCreateTopicError(t *testing.T) {
	errCode := codes.Internal
	mockPublisher.err = grpc.Errorf(errCode, "test error")

	var formattedName string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.Topic{
		Name: formattedName,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateTopic(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestPublisherPublish(t *testing.T) {
	var messageIdsElement string = "messageIdsElement-744837059"
	var messageIds = []string{messageIdsElement}
	var expectedResponse = &pubsubpb.PublishResponse{
		MessageIds: messageIds,
	}

	mockPublisher.err = nil
	mockPublisher.reqs = nil

	mockPublisher.resps = append(mockPublisher.resps[:0], expectedResponse)

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var data []byte = []byte("-86")
	var messagesElement = &pubsubpb.PubsubMessage{
		Data: data,
	}
	var messages = []*pubsubpb.PubsubMessage{messagesElement}
	var request = &pubsubpb.PublishRequest{
		Topic:    formattedTopic,
		Messages: messages,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Publish(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockPublisher.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestPublisherPublishError(t *testing.T) {
	errCode := codes.Internal
	mockPublisher.err = grpc.Errorf(errCode, "test error")

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var data []byte = []byte("-86")
	var messagesElement = &pubsubpb.PubsubMessage{
		Data: data,
	}
	var messages = []*pubsubpb.PubsubMessage{messagesElement}
	var request = &pubsubpb.PublishRequest{
		Topic:    formattedTopic,
		Messages: messages,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Publish(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestPublisherGetTopic(t *testing.T) {
	var name string = "name3373707"
	var expectedResponse = &pubsubpb.Topic{
		Name: name,
	}

	mockPublisher.err = nil
	mockPublisher.reqs = nil

	mockPublisher.resps = append(mockPublisher.resps[:0], expectedResponse)

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.GetTopicRequest{
		Topic: formattedTopic,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetTopic(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockPublisher.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestPublisherGetTopicError(t *testing.T) {
	errCode := codes.Internal
	mockPublisher.err = grpc.Errorf(errCode, "test error")

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.GetTopicRequest{
		Topic: formattedTopic,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetTopic(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestPublisherListTopics(t *testing.T) {
	var nextPageToken string = ""
	var topicsElement *pubsubpb.Topic = &pubsubpb.Topic{}
	var topics = []*pubsubpb.Topic{topicsElement}
	var expectedResponse = &pubsubpb.ListTopicsResponse{
		NextPageToken: nextPageToken,
		Topics:        topics,
	}

	mockPublisher.err = nil
	mockPublisher.reqs = nil

	mockPublisher.resps = append(mockPublisher.resps[:0], expectedResponse)

	var formattedProject string = PublisherProjectPath("[PROJECT]")
	var request = &pubsubpb.ListTopicsRequest{
		Project: formattedProject,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTopics(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockPublisher.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Topics[0])
	got := (interface{})(resp)
	var ok bool

	switch want := (want).(type) {
	case proto.Message:
		ok = proto.Equal(want, got.(proto.Message))
	default:
		ok = want == got
	}
	if !ok {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestPublisherListTopicsError(t *testing.T) {
	errCode := codes.Internal
	mockPublisher.err = grpc.Errorf(errCode, "test error")

	var formattedProject string = PublisherProjectPath("[PROJECT]")
	var request = &pubsubpb.ListTopicsRequest{
		Project: formattedProject,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTopics(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestPublisherListTopicSubscriptions(t *testing.T) {
	var nextPageToken string = ""
	var subscriptionsElement string = "subscriptionsElement1698708147"
	var subscriptions = []string{subscriptionsElement}
	var expectedResponse = &pubsubpb.ListTopicSubscriptionsResponse{
		NextPageToken: nextPageToken,
		Subscriptions: subscriptions,
	}

	mockPublisher.err = nil
	mockPublisher.reqs = nil

	mockPublisher.resps = append(mockPublisher.resps[:0], expectedResponse)

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.ListTopicSubscriptionsRequest{
		Topic: formattedTopic,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTopicSubscriptions(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockPublisher.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Subscriptions[0])
	got := (interface{})(resp)
	var ok bool

	switch want := (want).(type) {
	case proto.Message:
		ok = proto.Equal(want, got.(proto.Message))
	default:
		ok = want == got
	}
	if !ok {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestPublisherListTopicSubscriptionsError(t *testing.T) {
	errCode := codes.Internal
	mockPublisher.err = grpc.Errorf(errCode, "test error")

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.ListTopicSubscriptionsRequest{
		Topic: formattedTopic,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTopicSubscriptions(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestPublisherDeleteTopic(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockPublisher.err = nil
	mockPublisher.reqs = nil

	mockPublisher.resps = append(mockPublisher.resps[:0], expectedResponse)

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.DeleteTopicRequest{
		Topic: formattedTopic,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteTopic(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockPublisher.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestPublisherDeleteTopicError(t *testing.T) {
	errCode := codes.Internal
	mockPublisher.err = grpc.Errorf(errCode, "test error")

	var formattedTopic string = PublisherTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.DeleteTopicRequest{
		Topic: formattedTopic,
	}

	c, err := NewPublisherClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteTopic(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestSubscriberCreateSubscription(t *testing.T) {
	var name2 string = "name2-1052831874"
	var topic2 string = "topic2-1139259102"
	var ackDeadlineSeconds int32 = 2135351438
	var expectedResponse = &pubsubpb.Subscription{
		Name:               name2,
		Topic:              topic2,
		AckDeadlineSeconds: ackDeadlineSeconds,
	}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedName string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var formattedTopic string = SubscriberTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.Subscription{
		Name:  formattedName,
		Topic: formattedTopic,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateSubscription(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestSubscriberCreateSubscriptionError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedName string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var formattedTopic string = SubscriberTopicPath("[PROJECT]", "[TOPIC]")
	var request = &pubsubpb.Subscription{
		Name:  formattedName,
		Topic: formattedTopic,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateSubscription(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestSubscriberGetSubscription(t *testing.T) {
	var name string = "name3373707"
	var topic string = "topic110546223"
	var ackDeadlineSeconds int32 = 2135351438
	var expectedResponse = &pubsubpb.Subscription{
		Name:               name,
		Topic:              topic,
		AckDeadlineSeconds: ackDeadlineSeconds,
	}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var request = &pubsubpb.GetSubscriptionRequest{
		Subscription: formattedSubscription,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetSubscription(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestSubscriberGetSubscriptionError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var request = &pubsubpb.GetSubscriptionRequest{
		Subscription: formattedSubscription,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetSubscription(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestSubscriberListSubscriptions(t *testing.T) {
	var nextPageToken string = ""
	var subscriptionsElement *pubsubpb.Subscription = &pubsubpb.Subscription{}
	var subscriptions = []*pubsubpb.Subscription{subscriptionsElement}
	var expectedResponse = &pubsubpb.ListSubscriptionsResponse{
		NextPageToken: nextPageToken,
		Subscriptions: subscriptions,
	}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedProject string = SubscriberProjectPath("[PROJECT]")
	var request = &pubsubpb.ListSubscriptionsRequest{
		Project: formattedProject,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListSubscriptions(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Subscriptions[0])
	got := (interface{})(resp)
	var ok bool

	switch want := (want).(type) {
	case proto.Message:
		ok = proto.Equal(want, got.(proto.Message))
	default:
		ok = want == got
	}
	if !ok {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestSubscriberListSubscriptionsError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedProject string = SubscriberProjectPath("[PROJECT]")
	var request = &pubsubpb.ListSubscriptionsRequest{
		Project: formattedProject,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListSubscriptions(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestSubscriberDeleteSubscription(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var request = &pubsubpb.DeleteSubscriptionRequest{
		Subscription: formattedSubscription,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteSubscription(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestSubscriberDeleteSubscriptionError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var request = &pubsubpb.DeleteSubscriptionRequest{
		Subscription: formattedSubscription,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteSubscription(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestSubscriberModifyAckDeadline(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var ackIds []string = nil
	var ackDeadlineSeconds int32 = 2135351438
	var request = &pubsubpb.ModifyAckDeadlineRequest{
		Subscription:       formattedSubscription,
		AckIds:             ackIds,
		AckDeadlineSeconds: ackDeadlineSeconds,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.ModifyAckDeadline(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestSubscriberModifyAckDeadlineError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var ackIds []string = nil
	var ackDeadlineSeconds int32 = 2135351438
	var request = &pubsubpb.ModifyAckDeadlineRequest{
		Subscription:       formattedSubscription,
		AckIds:             ackIds,
		AckDeadlineSeconds: ackDeadlineSeconds,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.ModifyAckDeadline(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestSubscriberAcknowledge(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var ackIds []string = nil
	var request = &pubsubpb.AcknowledgeRequest{
		Subscription: formattedSubscription,
		AckIds:       ackIds,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.Acknowledge(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestSubscriberAcknowledgeError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var ackIds []string = nil
	var request = &pubsubpb.AcknowledgeRequest{
		Subscription: formattedSubscription,
		AckIds:       ackIds,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.Acknowledge(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestSubscriberPull(t *testing.T) {
	var expectedResponse *pubsubpb.PullResponse = &pubsubpb.PullResponse{}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var maxMessages int32 = 496131527
	var request = &pubsubpb.PullRequest{
		Subscription: formattedSubscription,
		MaxMessages:  maxMessages,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Pull(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestSubscriberPullError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var maxMessages int32 = 496131527
	var request = &pubsubpb.PullRequest{
		Subscription: formattedSubscription,
		MaxMessages:  maxMessages,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Pull(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestSubscriberStreamingPull(t *testing.T) {
	var expectedResponse *pubsubpb.StreamingPullResponse = &pubsubpb.StreamingPullResponse{}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var streamAckDeadlineSeconds int32 = 1875467245
	var request = &pubsubpb.StreamingPullRequest{
		Subscription:             formattedSubscription,
		StreamAckDeadlineSeconds: streamAckDeadlineSeconds,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	stream, err := c.StreamingPull(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Send(request); err != nil {
		t.Fatal(err)
	}
	if err := stream.CloseSend(); err != nil {
		t.Fatal(err)
	}
	resp, err := stream.Recv()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestSubscriberStreamingPullError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var streamAckDeadlineSeconds int32 = 1875467245
	var request = &pubsubpb.StreamingPullRequest{
		Subscription:             formattedSubscription,
		StreamAckDeadlineSeconds: streamAckDeadlineSeconds,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	stream, err := c.StreamingPull(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Send(request); err != nil {
		t.Fatal(err)
	}
	if err := stream.CloseSend(); err != nil {
		t.Fatal(err)
	}
	resp, err := stream.Recv()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestSubscriberModifyPushConfig(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockSubscriber.err = nil
	mockSubscriber.reqs = nil

	mockSubscriber.resps = append(mockSubscriber.resps[:0], expectedResponse)

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var pushConfig *pubsubpb.PushConfig = &pubsubpb.PushConfig{}
	var request = &pubsubpb.ModifyPushConfigRequest{
		Subscription: formattedSubscription,
		PushConfig:   pushConfig,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.ModifyPushConfig(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockSubscriber.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestSubscriberModifyPushConfigError(t *testing.T) {
	errCode := codes.Internal
	mockSubscriber.err = grpc.Errorf(errCode, "test error")

	var formattedSubscription string = SubscriberSubscriptionPath("[PROJECT]", "[SUBSCRIPTION]")
	var pushConfig *pubsubpb.PushConfig = &pubsubpb.PushConfig{}
	var request = &pubsubpb.ModifyPushConfigRequest{
		Subscription: formattedSubscription,
		PushConfig:   pushConfig,
	}

	c, err := NewSubscriberClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.ModifyPushConfig(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
