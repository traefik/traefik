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

type topicListService struct {
	service
	topics []string
	err    error
	t      *testing.T // for error logging.
}

func (s *topicListService) newNextStringFunc() nextStringFunc {
	return func() (string, error) {
		if len(s.topics) == 0 {
			return "", iterator.Done
		}
		tn := s.topics[0]
		s.topics = s.topics[1:]
		return tn, s.err
	}
}

func (s *topicListService) listProjectTopics(ctx context.Context, projName string) nextStringFunc {
	if projName != "projects/projid" {
		s.t.Fatalf("unexpected call: projName: %q", projName)
		return nil
	}
	return s.newNextStringFunc()
}

func checkTopicListing(t *testing.T, want []string) {
	s := &topicListService{topics: want, t: t}
	c := &Client{projectID: "projid", s: s}
	topics, err := slurpTopics(c.Topics(context.Background()))
	if err != nil {
		t.Errorf("error listing topics: %v", err)
	}
	got := topicNames(topics)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("topic list: got: %v, want: %v", got, want)
	}
	if len(s.topics) != 0 {
		t.Errorf("outstanding topics: %v", s.topics)
	}
}

// All returns the remaining topics from this iterator.
func slurpTopics(it *TopicIterator) ([]*Topic, error) {
	var topics []*Topic
	for {
		switch topic, err := it.Next(); err {
		case nil:
			topics = append(topics, topic)
		case iterator.Done:
			return topics, nil
		default:
			return nil, err
		}
	}
}

func TestTopicID(t *testing.T) {
	const id = "id"
	serv := &topicListService{
		topics: []string{"projects/projid/topics/t1", "projects/projid/topics/t2"},
		t:      t,
	}
	c := &Client{projectID: "projid", s: serv}
	s := c.Topic(id)
	if got, want := s.ID(), id; got != want {
		t.Errorf("Token.ID() = %q; want %q", got, want)
	}
	want := []string{"t1", "t2"}
	topics, err := slurpTopics(c.Topics(context.Background()))
	if err != nil {
		t.Errorf("error listing topics: %v", err)
	}
	for i, topic := range topics {
		if got, want := topic.ID(), want[i]; got != want {
			t.Errorf("Token.ID() = %q; want %q", got, want)
		}
	}
}

func TestListTopics(t *testing.T) {
	checkTopicListing(t, []string{
		"projects/projid/topics/t1",
		"projects/projid/topics/t2",
		"projects/projid/topics/t3",
		"projects/projid/topics/t4"})
}

func TestListCompletelyEmptyTopics(t *testing.T) {
	var want []string
	checkTopicListing(t, want)
}

func topicNames(topics []*Topic) []string {
	var names []string

	for _, topic := range topics {
		names = append(names, topic.name)

	}
	return names
}
