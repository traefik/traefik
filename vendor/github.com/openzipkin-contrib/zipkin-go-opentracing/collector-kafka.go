package zipkintracer

import (
	"github.com/Shopify/sarama"
	"github.com/apache/thrift/lib/go/thrift"

	"github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

// defaultKafkaTopic sets the standard Kafka topic our Collector will publish
// on. The default topic for zipkin-receiver-kafka is "zipkin", see:
// https://github.com/openzipkin/zipkin/tree/master/zipkin-receiver-kafka
const defaultKafkaTopic = "zipkin"

// KafkaCollector implements Collector by publishing spans to a Kafka
// broker.
type KafkaCollector struct {
	producer sarama.AsyncProducer
	logger   Logger
	topic    string
}

// KafkaOption sets a parameter for the KafkaCollector
type KafkaOption func(c *KafkaCollector)

// KafkaLogger sets the logger used to report errors in the collection
// process. By default, a no-op logger is used, i.e. no errors are logged
// anywhere. It's important to set this option.
func KafkaLogger(logger Logger) KafkaOption {
	return func(c *KafkaCollector) { c.logger = logger }
}

// KafkaProducer sets the producer used to produce to Kafka.
func KafkaProducer(p sarama.AsyncProducer) KafkaOption {
	return func(c *KafkaCollector) { c.producer = p }
}

// KafkaTopic sets the kafka topic to attach the collector producer on.
func KafkaTopic(t string) KafkaOption {
	return func(c *KafkaCollector) { c.topic = t }
}

// NewKafkaCollector returns a new Kafka-backed Collector. addrs should be a
// slice of TCP endpoints of the form "host:port".
func NewKafkaCollector(addrs []string, options ...KafkaOption) (Collector, error) {
	c := &KafkaCollector{
		logger: NewNopLogger(),
		topic:  defaultKafkaTopic,
	}

	for _, option := range options {
		option(c)
	}
	if c.producer == nil {
		p, err := sarama.NewAsyncProducer(addrs, nil)
		if err != nil {
			return nil, err
		}
		c.producer = p
	}

	go c.logErrors()

	return c, nil
}

func (c *KafkaCollector) logErrors() {
	for pe := range c.producer.Errors() {
		_ = c.logger.Log("msg", pe.Msg, "err", pe.Err, "result", "failed to produce msg")
	}
}

// Collect implements Collector.
func (c *KafkaCollector) Collect(s *zipkincore.Span) error {
	c.producer.Input() <- &sarama.ProducerMessage{
		Topic: c.topic,
		Key:   nil,
		Value: sarama.ByteEncoder(kafkaSerialize(s)),
	}
	return nil
}

// Close implements Collector.
func (c *KafkaCollector) Close() error {
	return c.producer.Close()
}

func kafkaSerialize(s *zipkincore.Span) []byte {
	t := thrift.NewTMemoryBuffer()
	p := thrift.NewTBinaryProtocolTransport(t)
	if err := s.Write(p); err != nil {
		panic(err)
	}
	return t.Buffer.Bytes()
}
