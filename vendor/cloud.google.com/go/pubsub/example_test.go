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

package pubsub_test

import (
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func ExampleNewClient() {
	ctx := context.Background()
	_, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	// See the other examples to learn how to use the Client.
}

func ExampleClient_CreateTopic() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	// Create a new topic with the given name.
	topic, err := client.CreateTopic(ctx, "topicName")
	if err != nil {
		// TODO: Handle error.
	}

	_ = topic // TODO: use the topic.
}

func ExampleClient_CreateSubscription() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	// Create a new topic with the given name.
	topic, err := client.CreateTopic(ctx, "topicName")
	if err != nil {
		// TODO: Handle error.
	}

	// Create a new subscription to the previously created topic
	// with the given name.
	sub, err := client.CreateSubscription(ctx, "subName", topic, 10*time.Second, nil)
	if err != nil {
		// TODO: Handle error.
	}

	_ = sub // TODO: use the subscription.
}

func ExampleTopic_Delete() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	topic := client.Topic("topicName")
	if err := topic.Delete(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleTopic_Exists() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	topic := client.Topic("topicName")
	ok, err := topic.Exists(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	if !ok {
		// Topic doesn't exist.
	}
}

func ExampleTopic_Publish() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	topic := client.Topic("topicName")
	msgIDs, err := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("hello world"),
	})
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Printf("Published a message with a message ID: %s\n", msgIDs[0])
}

func ExampleTopic_Subscriptions() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	topic := client.Topic("topic-name")
	// List all subscriptions of the topic (maybe of multiple projects).
	for subs := topic.Subscriptions(ctx); ; {
		sub, err := subs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		_ = sub // TODO: use the subscription.
	}
}

func ExampleSubscription_Delete() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	sub := client.Subscription("subName")
	if err := sub.Delete(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleSubscription_Exists() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	sub := client.Subscription("subName")
	ok, err := sub.Exists(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	if !ok {
		// Subscription doesn't exist.
	}
}

func ExampleSubscription_Config() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	sub := client.Subscription("subName")
	config, err := sub.Config(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(config)
}

func ExampleSubscription_Pull() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it, err := client.Subscription("subName").Pull(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// Ensure that the iterator is closed down cleanly.
	defer it.Stop()
}

func ExampleSubscription_Pull_options() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	sub := client.Subscription("subName")
	// This program is expected to process and acknowledge messages
	// in 5 seconds. If not, Pub/Sub API will assume the message is not
	// acknowledged.
	it, err := sub.Pull(ctx, pubsub.MaxExtension(5*time.Second))
	if err != nil {
		// TODO: Handle error.
	}
	// Ensure that the iterator is closed down cleanly.
	defer it.Stop()
}

func ExampleSubscription_ModifyPushConfig() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	sub := client.Subscription("subName")
	if err := sub.ModifyPushConfig(ctx, &pubsub.PushConfig{Endpoint: "https://example.com/push"}); err != nil {
		// TODO: Handle error.
	}
}

func ExampleMessageIterator_Next() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it, err := client.Subscription("subName").Pull(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// Ensure that the iterator is closed down cleanly.
	defer it.Stop()
	// Consume 10 messages.
	for i := 0; i < 10; i++ {
		m, err := it.Next()
		if err == iterator.Done {
			// There are no more messages.  This will happen if it.Stop is called.
			break
		}
		if err != nil {
			// TODO: Handle error.
			break
		}
		fmt.Printf("message %d: %s\n", i, m.Data)

		// Acknowledge the message.
		m.Done(true)
	}
}

func ExampleMessageIterator_Stop_defer() {
	// If all uses of the iterator occur within the lifetime of a single
	// function, stop it with defer.
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it, err := client.Subscription("subName").Pull(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	// Ensure that the iterator is closed down cleanly.
	defer it.Stop()

	// TODO: Use the iterator (see the example for MessageIterator.Next).
}

func ExampleMessageIterator_Stop_goroutine() *pubsub.MessageIterator {
	// If you use the iterator outside the lifetime of a single function, you
	// must still stop it.
	// This (contrived) example returns an iterator that will yield messages
	// for ten seconds, and then stop.
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it, err := client.Subscription("subName").Pull(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// Stop the iterator after receiving messages for ten seconds.
	go func() {
		time.Sleep(10 * time.Second)
		it.Stop()
	}()
	return it
}
