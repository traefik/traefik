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
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/api/iterator"
)

type subListService struct {
	service
	subs []string
	err  error

	t *testing.T // for error logging.
}

func (s *subListService) newNextStringFunc() nextStringFunc {
	return func() (string, error) {
		if len(s.subs) == 0 {
			return "", iterator.Done
		}
		sn := s.subs[0]
		s.subs = s.subs[1:]
		return sn, s.err
	}
}

func (s *subListService) listProjectSubscriptions(ctx context.Context, projName string) nextStringFunc {
	if projName != "projects/projid" {
		s.t.Fatalf("unexpected call: projName: %q", projName)
		return nil
	}
	return s.newNextStringFunc()
}

func (s *subListService) listTopicSubscriptions(ctx context.Context, topicName string) nextStringFunc {
	if topicName != "projects/projid/topics/topic" {
		s.t.Fatalf("unexpected call: topicName: %q", topicName)
		return nil
	}
	return s.newNextStringFunc()
}

// All returns the remaining subscriptions from this iterator.
func slurpSubs(it *SubscriptionIterator) ([]*Subscription, error) {
	var subs []*Subscription
	for {
		switch sub, err := it.Next(); err {
		case nil:
			subs = append(subs, sub)
		case iterator.Done:
			return subs, nil
		default:
			return nil, err
		}
	}
}

func TestSubscriptionID(t *testing.T) {
	const id = "id"
	serv := &subListService{
		subs: []string{"projects/projid/subscriptions/s1", "projects/projid/subscriptions/s2"},
		t:    t,
	}
	c := &Client{projectID: "projid", s: serv}
	s := c.Subscription(id)
	if got, want := s.ID(), id; got != want {
		t.Errorf("Subscription.ID() = %q; want %q", got, want)
	}
	want := []string{"s1", "s2"}
	subs, err := slurpSubs(c.Subscriptions(context.Background()))
	if err != nil {
		t.Errorf("error listing subscriptions: %v", err)
	}
	for i, s := range subs {
		if got, want := s.ID(), want[i]; got != want {
			t.Errorf("Subscription.ID() = %q; want %q", got, want)
		}
	}
}

func TestListProjectSubscriptions(t *testing.T) {
	snames := []string{"projects/projid/subscriptions/s1", "projects/projid/subscriptions/s2",
		"projects/projid/subscriptions/s3"}
	s := &subListService{subs: snames, t: t}
	c := &Client{projectID: "projid", s: s}
	subs, err := slurpSubs(c.Subscriptions(context.Background()))
	if err != nil {
		t.Errorf("error listing subscriptions: %v", err)
	}
	got := subNames(subs)
	want := []string{
		"projects/projid/subscriptions/s1",
		"projects/projid/subscriptions/s2",
		"projects/projid/subscriptions/s3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sub list: got: %v, want: %v", got, want)
	}
	if len(s.subs) != 0 {
		t.Errorf("outstanding subs: %v", s.subs)
	}
}

func TestListTopicSubscriptions(t *testing.T) {
	snames := []string{"projects/projid/subscriptions/s1", "projects/projid/subscriptions/s2",
		"projects/projid/subscriptions/s3"}
	s := &subListService{subs: snames, t: t}
	c := &Client{projectID: "projid", s: s}
	subs, err := slurpSubs(c.Topic("topic").Subscriptions(context.Background()))
	if err != nil {
		t.Errorf("error listing subscriptions: %v", err)
	}
	got := subNames(subs)
	want := []string{
		"projects/projid/subscriptions/s1",
		"projects/projid/subscriptions/s2",
		"projects/projid/subscriptions/s3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sub list: got: %v, want: %v", got, want)
	}
	if len(s.subs) != 0 {
		t.Errorf("outstanding subs: %v", s.subs)
	}
}

func subNames(subs []*Subscription) []string {
	var names []string

	for _, sub := range subs {
		names = append(names, sub.name)

	}
	return names
}
