package cony

import (
	"fmt"
	"os"
	"sync"

	"github.com/streadway/amqp"
)

// ConsumerOpt is a consumer's functional option type
type ConsumerOpt func(*Consumer)

// Consumer holds definition for AMQP consumer
type Consumer struct {
	q          *Queue
	deliveries chan amqp.Delivery
	errs       chan error
	qos        int
	tag        string
	autoAck    bool
	exclusive  bool
	noLocal    bool
	stop       chan struct{}
	dead       bool
	m          sync.Mutex
}

// Deliveries return deliveries shipped to this consumer
// this channel never closed, even on disconnects
func (c *Consumer) Deliveries() <-chan amqp.Delivery {
	return c.deliveries
}

// Errors returns channel with AMQP channel level errors
func (c *Consumer) Errors() <-chan error {
	return c.errs
}

// Cancel this consumer.
//
// This will CLOSE Deliveries() channel
func (c *Consumer) Cancel() {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.dead {
		close(c.deliveries)
		close(c.stop)
		c.dead = true
	}
}

func (c *Consumer) reportErr(err error) bool {
	if err != nil {
		select {
		case c.errs <- err:
		default:
		}
		return true
	}
	return false
}

func (c *Consumer) serve(client mqDeleter, ch mqChannel) {
	if c.reportErr(ch.Qos(c.qos, 0, false)) {
		return
	}

	deliveries, err2 := ch.Consume(c.q.Name,
		c.tag,       // consumer tag
		c.autoAck,   // autoAck,
		c.exclusive, // exclusive,
		c.noLocal,   // noLocal,
		false,       // noWait,
		nil,         // args Table
	)
	if c.reportErr(err2) {
		return
	}

	for {
		select {
		case <-c.stop:
			client.deleteConsumer(c)
			ch.Close()
			return
		case d, ok := <-deliveries: // deliveries will be closed once channel is closed (disconnected from network)
			if !ok {
				return
			}
			c.deliveries <- d
		}
	}
}

// NewConsumer Consumer's constructor
func NewConsumer(q *Queue, opts ...ConsumerOpt) *Consumer {
	c := &Consumer{
		q:          q,
		deliveries: make(chan amqp.Delivery),
		errs:       make(chan error, 100),
		stop:       make(chan struct{}),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Qos on channel
func Qos(count int) ConsumerOpt {
	return func(c *Consumer) {
		c.qos = count
	}
}

// Tag the consumer
func Tag(tag string) ConsumerOpt {
	return func(c *Consumer) {
		c.tag = tag
	}
}

// AutoTag set automatically generated tag like this
//	fmt.Sprintf(QueueName+"-pid-%d@%s", os.Getpid(), os.Hostname())
func AutoTag() ConsumerOpt {
	return func(c *Consumer) {
		host, _ := os.Hostname()
		tag := fmt.Sprintf(c.q.Name+"-pid-%d@%s", os.Getpid(), host)
		Tag(tag)(c)
	}
}

// AutoAck set this consumer in AutoAck mode
func AutoAck() ConsumerOpt {
	return func(c *Consumer) {
		c.autoAck = true
	}
}

// Exclusive set this consumer in exclusive mode
func Exclusive() ConsumerOpt {
	return func(c *Consumer) {
		c.exclusive = true
	}
}

// NoLocal set this consumer in NoLocal mode.
func NoLocal() ConsumerOpt {
	return func(c *Consumer) {
		c.noLocal = true
	}
}
