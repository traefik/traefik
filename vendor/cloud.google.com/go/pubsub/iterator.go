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
	"log"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/support/bundler"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type MessageIterator struct {
	impl interface {
		next() (*Message, error)
		stop()
	}
}

type pollingMessageIterator struct {
	// kaTicker controls how often we send an ack deadline extension request.
	kaTicker *time.Ticker
	// ackTicker controls how often we acknowledge a batch of messages.
	ackTicker *time.Ticker

	ka     *keepAlive
	acker  *acker
	nacker *bundler.Bundler
	puller *puller

	// mu ensures that cleanup only happens once, and concurrent Stop
	// invocations block until cleanup completes.
	mu sync.Mutex

	// closed is used to signal that Stop has been called.
	closed chan struct{}
}

var useStreamingPull = false

// newMessageIterator starts a new MessageIterator.  Stop must be called on the MessageIterator
// when it is no longer needed.
// subName is the full name of the subscription to pull messages from.
// ctx is the context to use for acking messages and extending message deadlines.
func newMessageIterator(ctx context.Context, s service, subName string, po *pullOptions) *MessageIterator {
	if !useStreamingPull {
		return &MessageIterator{
			impl: newPollingMessageIterator(ctx, s, subName, po),
		}
	}
	sp := s.newStreamingPuller(ctx, subName, int32(po.ackDeadline.Seconds()))
	err := sp.open()
	if grpc.Code(err) == codes.Unimplemented {
		log.Println("pubsub: streaming pull unimplemented; falling back to legacy pull")
		return &MessageIterator{
			impl: newPollingMessageIterator(ctx, s, subName, po),
		}
	}
	// TODO(jba): handle other non-nil error?
	log.Println("using streaming pull")
	return &MessageIterator{
		impl: newStreamingMessageIterator(ctx, sp, po),
	}
}

func newPollingMessageIterator(ctx context.Context, s service, subName string, po *pullOptions) *pollingMessageIterator {
	// TODO: make kaTicker frequency more configurable.
	// (ackDeadline - 5s) is a reasonable default for now, because the minimum ack period is 10s.  This gives us 5s grace.
	keepAlivePeriod := po.ackDeadline - 5*time.Second
	kaTicker := time.NewTicker(keepAlivePeriod) // Stopped in it.Stop

	// TODO: make ackTicker more configurable.  Something less than
	// kaTicker is a reasonable default (there's no point extending
	// messages when they could be acked instead).
	ackTicker := time.NewTicker(keepAlivePeriod / 2) // Stopped in it.Stop

	ka := &keepAlive{
		s:             s,
		Ctx:           ctx,
		Sub:           subName,
		ExtensionTick: kaTicker.C,
		Deadline:      po.ackDeadline,
		MaxExtension:  po.maxExtension,
	}

	ack := &acker{
		s:       s,
		Ctx:     ctx,
		Sub:     subName,
		AckTick: ackTicker.C,
		Notify:  ka.Remove,
	}

	nacker := bundler.NewBundler("", func(ackIDs interface{}) {
		// NACK by setting the ack deadline to zero, to make the message
		// immediately available for redelivery.
		//
		// If the RPC fails, nothing we can do about it. In the worst case, the
		// deadline for these messages will expire and they will still get
		// redelivered.
		_ = s.modifyAckDeadline(ctx, subName, 0, ackIDs.([]string))
	})
	nacker.DelayThreshold = keepAlivePeriod / 10 // nack promptly
	nacker.BundleCountThreshold = 10

	pull := newPuller(s, subName, ctx, po.maxPrefetch, ka.Add, ka.Remove)

	ka.Start()
	ack.Start()
	return &pollingMessageIterator{
		kaTicker:  kaTicker,
		ackTicker: ackTicker,
		ka:        ka,
		acker:     ack,
		nacker:    nacker,
		puller:    pull,
		closed:    make(chan struct{}),
	}
}

// Next returns the next Message to be processed.  The caller must call
// Message.Done when finished with it.
// Once Stop has been called, calls to Next will return iterator.Done.
func (it *MessageIterator) Next() (*Message, error) {
	return it.impl.next()
}

func (it *pollingMessageIterator) next() (*Message, error) {
	m, err := it.puller.Next()
	if err == nil {
		m.done = it.done
		return m, nil
	}

	select {
	// If Stop has been called, we return Done regardless the value of err.
	case <-it.closed:
		return nil, iterator.Done
	default:
		return nil, err
	}
}

// Client code must call Stop on a MessageIterator when finished with it.
// Stop will block until Done has been called on all Messages that have been
// returned by Next, or until the context with which the MessageIterator was created
// is cancelled or exceeds its deadline.
// Stop need only be called once, but may be called multiple times from
// multiple goroutines.
func (it *MessageIterator) Stop() {
	it.impl.stop()
}

func (it *pollingMessageIterator) stop() {
	it.mu.Lock()
	defer it.mu.Unlock()

	select {
	case <-it.closed:
		// Cleanup has already been performed.
		return
	default:
	}

	// We close this channel before calling it.puller.Stop to ensure that we
	// reliably return iterator.Done from Next.
	close(it.closed)

	// Stop the puller. Once this completes, no more messages will be added
	// to it.ka.
	it.puller.Stop()

	// Start acking messages as they arrive, ignoring ackTicker.  This will
	// result in it.ka.Stop, below, returning as soon as possible.
	it.acker.FastMode()

	// This will block until
	//   (a) it.ka.Ctx is done, or
	//   (b) all messages have been removed from keepAlive.
	// (b) will happen once all outstanding messages have been either ACKed or NACKed.
	it.ka.Stop()

	// There are no more live messages, so kill off the acker.
	it.acker.Stop()
	it.nacker.Stop()
	it.kaTicker.Stop()
	it.ackTicker.Stop()
}

func (it *pollingMessageIterator) done(ackID string, ack bool) {
	if ack {
		it.acker.Ack(ackID)
		// There's no need to call it.ka.Remove here, as acker will
		// call it via its Notify function.
	} else {
		it.ka.Remove(ackID)
		_ = it.nacker.Add(ackID, len(ackID)) // ignore error; this is just an optimization
	}
}

type streamingMessageIterator struct {
	ctx        context.Context
	po         *pullOptions
	sp         *streamingPuller
	kaTicker   *time.Ticker  // keep-alive (deadline extensions)
	ackTicker  *time.Ticker  // message acks
	nackTicker *time.Ticker  // message nacks (more frequent than acks)
	failed     chan struct{} // closed on stream error
	stopped    chan struct{} // closed when Stop is called
	drained    chan struct{} // closed when stopped && no more pending messages
	msgc       chan *Message
	wg         sync.WaitGroup

	mu                 sync.Mutex
	keepAliveDeadlines map[string]time.Time
	pendingReq         *pb.StreamingPullRequest
	err                error // error from stream failure
}

const messageBufferSize = 1000

func newStreamingMessageIterator(ctx context.Context, sp *streamingPuller, po *pullOptions) *streamingMessageIterator {
	// TODO: make kaTicker frequency more configurable. (ackDeadline - 5s) is a
	// reasonable default for now, because the minimum ack period is 10s. This
	// gives us 5s grace.
	keepAlivePeriod := po.ackDeadline - 5*time.Second
	kaTicker := time.NewTicker(keepAlivePeriod)

	// TODO: make ackTicker more configurable.  Something less than
	// kaTicker is a reasonable default (there's no point extending
	// messages when they could be acked instead).
	ackTicker := time.NewTicker(keepAlivePeriod / 2)
	nackTicker := time.NewTicker(keepAlivePeriod / 10)
	it := &streamingMessageIterator{
		ctx:                ctx,
		sp:                 sp,
		po:                 po,
		kaTicker:           kaTicker,
		ackTicker:          ackTicker,
		nackTicker:         nackTicker,
		failed:             make(chan struct{}),
		stopped:            make(chan struct{}),
		drained:            make(chan struct{}),
		msgc:               make(chan *Message, messageBufferSize),
		keepAliveDeadlines: map[string]time.Time{},
		pendingReq:         &pb.StreamingPullRequest{},
	}
	it.wg.Add(2)
	go it.receiver()
	go it.sender()
	return it
}

func (it *streamingMessageIterator) next() (*Message, error) {
	// If ctx has been cancelled or the iterator is done, return straight
	// away (even if there are buffered messages available).
	select {
	case <-it.ctx.Done():
		return nil, it.ctx.Err()

	case <-it.failed:
		break

	case <-it.stopped:
		break

	default:
		// Wait for a message, but also for one of the above conditions.
		select {
		case msg := <-it.msgc:
			// Since active select cases are chosen at random, this can return
			// nil (from the channel close) even if it.failed or it.stopped is
			// closed.
			if msg == nil {
				break
			}
			msg.done = it.done
			return msg, nil

		case <-it.ctx.Done():
			return nil, it.ctx.Err()

		case <-it.failed:
			break

		case <-it.stopped:
			break
		}
	}
	// Here if the iterator is done.
	it.mu.Lock()
	defer it.mu.Unlock()
	return nil, it.err
}

func (it *streamingMessageIterator) stop() {
	it.mu.Lock()
	select {
	case <-it.stopped:
		it.mu.Unlock()
		it.wg.Wait()
		return
	default:
		close(it.stopped)
	}
	if it.err == nil {
		it.err = iterator.Done
	}
	// Before reading from the channel, see if we're already drained.
	it.checkDrained()
	it.mu.Unlock()
	// Nack all the pending messages.
	// Grab the lock separately for each message to allow the receiver
	// and sender goroutines to make progress.
	// Why this will eventually terminate:
	// - If the receiver is not blocked on a stream Recv, then
	//   it will write all the messages it has received to the channel,
	//   then exit, closing the channel.
	// - If the receiver is blocked, then this loop will eventually
	//   nack all the messages in the channel. Once done is called
	//   on the remaining messages, the iterator will be marked as drained,
	//   which will trigger the sender to terminate. When it does, it
	//   performs a CloseSend on the stream, which will result in the blocked
	//   stream Recv returning.
	for m := range it.msgc {
		it.mu.Lock()
		delete(it.keepAliveDeadlines, m.ackID)
		it.addDeadlineMod(m.ackID, 0)
		it.checkDrained()
		it.mu.Unlock()
	}
	it.wg.Wait()
}

// checkDrained closes the drained channel if the iterator has been stopped and all
// pending messages have either been n/acked or expired.
//
// Called with the lock held.
func (it *streamingMessageIterator) checkDrained() {
	select {
	case <-it.drained:
		return
	default:
	}
	select {
	case <-it.stopped:
		if len(it.keepAliveDeadlines) == 0 {
			close(it.drained)
		}
	default:
	}
}

// Called when a message is acked/nacked.
func (it *streamingMessageIterator) done(ackID string, ack bool) {
	it.mu.Lock()
	defer it.mu.Unlock()
	delete(it.keepAliveDeadlines, ackID)
	if ack {
		it.pendingReq.AckIds = append(it.pendingReq.AckIds, ackID)
	} else {
		it.addDeadlineMod(ackID, 0) // Nack indicated by modifying the deadline to zero.
	}
	it.checkDrained()
}

// addDeadlineMod adds the ack ID to the pending request with the given deadline.
//
// Called with the lock held.
func (it *streamingMessageIterator) addDeadlineMod(ackID string, deadlineSecs int32) {
	pr := it.pendingReq
	pr.ModifyDeadlineAckIds = append(pr.ModifyDeadlineAckIds, ackID)
	pr.ModifyDeadlineSeconds = append(pr.ModifyDeadlineSeconds, deadlineSecs)
}

// fail is called when a stream method returns a permanent error.
func (it *streamingMessageIterator) fail(err error) {
	it.mu.Lock()
	if it.err == nil {
		it.err = err
		close(it.failed)
	}
	it.mu.Unlock()
}

// receiver runs in a goroutine and handles all receives from the stream.
func (it *streamingMessageIterator) receiver() {
	defer it.wg.Done()
	defer close(it.msgc)

	for {
		// Stop retrieving messages if the context is done, the stream
		// failed, or the iterator's Stop method was called.
		select {
		case <-it.ctx.Done():
			return
		case <-it.failed:
			return
		case <-it.stopped:
			return
		default:
		}
		// Receive messages from stream. This may block indefinitely.
		msgs, err := it.sp.fetchMessages()

		// The streamingPuller handles retries, so any error here
		// is fatal to the iterator.
		if err != nil {
			it.fail(err)
			return
		}
		// We received some messages. Remember them so we can
		// keep them alive.
		deadline := time.Now().Add(it.po.maxExtension)
		it.mu.Lock()
		for _, m := range msgs {
			it.keepAliveDeadlines[m.ackID] = deadline
		}
		it.mu.Unlock()
		// Deliver the messages to the channel.
		for _, m := range msgs {
			select {
			case <-it.ctx.Done():
				return
			case <-it.failed:
				return
				// Don't return if stopped. We want to send the remaining
				// messages on the channel, where they will be nacked.
			case it.msgc <- m:
			}
		}
	}
}

// sender runs in a goroutine and handles all sends to the stream.
func (it *streamingMessageIterator) sender() {
	defer it.wg.Done()
	defer it.kaTicker.Stop()
	defer it.ackTicker.Stop()
	defer it.nackTicker.Stop()
	defer it.sp.closeSend()

	done := false
	for !done {
		send := false
		select {
		case <-it.ctx.Done():
			// Context canceled or timed out: stop immediately, without
			// another RPC.
			return

		case <-it.failed:
			// Stream failed: nothing to do, so stop immediately.
			return

		case <-it.drained:
			// All outstanding messages have been marked done:
			// nothing left to do except send the final request.
			it.mu.Lock()
			send = (len(it.pendingReq.AckIds) > 0 || len(it.pendingReq.ModifyDeadlineAckIds) > 0)
			done = true

		case <-it.kaTicker.C:
			it.mu.Lock()
			send = it.handleKeepAlives()

		case <-it.nackTicker.C:
			it.mu.Lock()
			send = (len(it.pendingReq.ModifyDeadlineAckIds) > 0)

		case <-it.ackTicker.C:
			it.mu.Lock()
			send = (len(it.pendingReq.AckIds) > 0)

		}
		// Lock is held here.
		if send {
			req := it.pendingReq
			it.pendingReq = &pb.StreamingPullRequest{}
			it.mu.Unlock()
			err := it.sp.send(req)
			if err != nil {
				// The streamingPuller handles retries, so any error here
				// is fatal to the iterator.
				it.fail(err)
				return
			}
		} else {
			it.mu.Unlock()
		}
	}
}

// handleKeepAlives modifies the pending request to include deadline extensions
// for live messages. It also purges expired messages. It reports whether
// there were any live messages.
//
// Called with the lock held.
func (it *streamingMessageIterator) handleKeepAlives() bool {
	live, expired := getKeepAliveAckIDs(it.keepAliveDeadlines)
	for _, e := range expired {
		delete(it.keepAliveDeadlines, e)
	}
	dl := trunc32(int64(it.po.ackDeadline.Seconds()))
	for _, m := range live {
		it.addDeadlineMod(m, dl)
	}
	it.checkDrained()
	return len(live) > 0
}
