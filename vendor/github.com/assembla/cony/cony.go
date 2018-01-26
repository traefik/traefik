// Package cony is a high-level wrapper around http://github.com/streadway/amqp library,
// for working declaratively with AMQP. Cony will manage AMQP
// connect/reconnect to AMQP broker, along with recovery of consumers.
package cony

import (
	"sync"

	"github.com/streadway/amqp"
)

// Queue hold definition of AMQP queue
type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	Args       amqp.Table

	l sync.Mutex
}

// Exchange hold definition of AMQP exchange
type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Args       amqp.Table
}

// Binding used to declare binding between AMQP Queue and AMQP Exchange
type Binding struct {
	Queue    *Queue
	Exchange Exchange
	Key      string
	Args     amqp.Table
}

type mqDeleter interface {
	deletePublisher(*Publisher)
	deleteConsumer(*Consumer)
}

type mqChannel interface {
	Close() error
	Consume(string, string, bool, bool, bool, bool, amqp.Table) (<-chan amqp.Delivery, error)
	NotifyClose(chan *amqp.Error) chan *amqp.Error
	Publish(string, string, bool, bool, amqp.Publishing) error
	Qos(int, int, bool) error
}
