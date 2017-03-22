// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

var consumerSeq uint64

func uniqueConsumerTag() string {
	return fmt.Sprintf("ctag-%s-%d", os.Args[0], atomic.AddUint64(&consumerSeq, 1))
}

type consumerBuffers map[string]chan *Delivery

// Concurrent type that manages the consumerTag ->
// ingress consumerBuffer mapping
type consumers struct {
	sync.Mutex
	chans consumerBuffers
}

func makeConsumers() *consumers {
	return &consumers{chans: make(consumerBuffers)}
}

func bufferDeliveries(in chan *Delivery, out chan Delivery) {
	var queue []*Delivery
	var queueIn = in

	for delivery := range in {
		select {
		case out <- *delivery:
			// delivered immediately while the consumer chan can receive
		default:
			queue = append(queue, delivery)
		}

		for len(queue) > 0 {
			select {
			case out <- *queue[0]:
				queue = queue[1:]
			case delivery, open := <-queueIn:
				if open {
					queue = append(queue, delivery)
				} else {
					// stop receiving to drain the queue
					queueIn = nil
				}
			}
		}
	}

	close(out)
}

// On key conflict, close the previous channel.
func (subs *consumers) add(tag string, consumer chan Delivery) {
	subs.Lock()
	defer subs.Unlock()

	if prev, found := subs.chans[tag]; found {
		close(prev)
	}

	in := make(chan *Delivery)
	go bufferDeliveries(in, consumer)

	subs.chans[tag] = in
}

func (subs *consumers) close(tag string) (found bool) {
	subs.Lock()
	defer subs.Unlock()

	ch, found := subs.chans[tag]

	if found {
		delete(subs.chans, tag)
		close(ch)
	}

	return found
}

func (subs *consumers) closeAll() {
	subs.Lock()
	defer subs.Unlock()

	for _, ch := range subs.chans {
		close(ch)
	}

	subs.chans = make(consumerBuffers)
}

// Sends a delivery to a the consumer identified by `tag`.
// If unbuffered channels are used for Consume this method
// could block all deliveries until the consumer
// receives on the other end of the channel.
func (subs *consumers) send(tag string, msg *Delivery) bool {
	subs.Lock()
	defer subs.Unlock()

	buffer, found := subs.chans[tag]
	if found {
		buffer <- msg
	}

	return found
}
