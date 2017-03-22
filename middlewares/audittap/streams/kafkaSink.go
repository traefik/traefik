package streams

import (
	"github.com/Shopify/sarama"
	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"log"
	"sync"
)

type kafkaAsyncProducer interface {
	Input() chan<- *sarama.ProducerMessage
	Errors() <-chan *sarama.ProducerError
	AsyncClose()
}

type kafkaAuditSink struct {
	topic    string
	producer kafkaAsyncProducer
	join     *sync.WaitGroup
}

// NewAsyncProducer is a seam for testing
var NewAsyncProducer = func(endpoints []string, config *sarama.Config) (kafkaAsyncProducer, error) {
	return sarama.NewAsyncProducer(endpoints, config)
}

// NewKafkaSink returns an AuditSink for sending messages to Kafka.
func NewKafkaSink(topic string, endpoints []string) (sink AuditSink, err error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = false
	producer, err := NewAsyncProducer(endpoints, config)
	if err != nil {
		return nil, err
	}

	kas := &kafkaAuditSink{topic, producer, &sync.WaitGroup{}}
	kas.logErrors()
	return kas, nil
}

func (kas *kafkaAuditSink) logErrors() {
	kas.join.Add(1)

	go func() {
		// read errors and log them, until the producer is closed
		for err := range kas.producer.Errors() {
			log.Println(err.Error())
		}
		kas.join.Done()
	}()
}

func (kas *kafkaAuditSink) Audit(encoded audittypes.Encoded) error {
	message := &sarama.ProducerMessage{Topic: kas.topic, Value: encoded}
	kas.producer.Input() <- message
	return nil
}

func (kas *kafkaAuditSink) Close() error {
	kas.producer.AsyncClose() // closes all the producer's goroutines and channels
	kas.join.Wait()
	return nil
}
