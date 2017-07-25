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
	"time"

	"github.com/golang/protobuf/ptypes"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
)

// Message represents a Pub/Sub message.
type Message struct {
	// ID identifies this message.
	// This ID is assigned by the server and is populated for Messages obtained from a subscription.
	// This field is read-only.
	ID string

	// Data is the actual data in the message.
	Data []byte

	// Attributes represents the key-value pairs the current message
	// is labelled with.
	Attributes map[string]string

	// ackID is the identifier to acknowledge this message.
	ackID string

	// The time at which the message was published.
	// This is populated by the server for Messages obtained from a subscription.
	// This field is read-only.
	PublishTime time.Time

	calledDone bool

	// The done method of the iterator that created this Message.
	done func(string, bool)
}

func toMessage(resp *pb.ReceivedMessage) (*Message, error) {
	if resp.Message == nil {
		return &Message{ackID: resp.AckId}, nil
	}

	pubTime, err := ptypes.Timestamp(resp.Message.PublishTime)
	if err != nil {
		return nil, err
	}
	return &Message{
		ackID:       resp.AckId,
		Data:        resp.Message.Data,
		Attributes:  resp.Message.Attributes,
		ID:          resp.Message.MessageId,
		PublishTime: pubTime,
	}, nil
}

// Done completes the processing of a Message that was returned from a MessageIterator.
// ack indicates whether the message should be acknowledged.
// Client code must call Done when finished for each Message returned by an iterator.
// Done may only be called on Messages returned by a MessageIterator.
// If message acknowledgement fails, the Message will be redelivered.
// Calls to Done have no effect after the first call.
//
// See MessageIterator.Next for an example.
func (m *Message) Done(ack bool) {
	if m.calledDone {
		return
	}
	m.calledDone = true
	m.done(m.ackID, ack)
}
