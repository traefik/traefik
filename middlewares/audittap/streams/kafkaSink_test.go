package streams

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"sync"
	"time"
)

func TestKafkaSink(t *testing.T) {
	tp := &testKafkaAsyncProducer{mu: &sync.Mutex{}}

	NewAsyncProducer = func(endpoints []string, config *sarama.Config) (kafkaAsyncProducer, error) {
		tp.endpoints = endpoints
		return tp, nil
	}

	endpoints := []string{"http://endpoint"}
	kafkaSink, err := NewKafkaSink("hot", endpoints)
	assert.NoError(t, err)
	assert.Exactly(t, endpoints, tp.endpoints)

	err = kafkaSink.Audit(encodedJSONSample)
	assert.NoError(t, err)

	err = kafkaSink.Close()
	assert.NoError(t, err)

	time.Sleep(time.Duration(10 * time.Millisecond))

	assert.True(t, tp.asyncClosed)
	assert.True(t, tp.inputTerminated)
	assert.True(t, tp.errorsTerminated)
}

//-------------------------------------------------------------------------------------------------

type testKafkaAsyncProducer struct {
	endpoints        []string
	input            chan *sarama.ProducerMessage
	messages         []*sarama.ProducerMessage
	errors           []*sarama.ProducerError
	asyncClosed      bool
	inputTerminated  bool
	errorsTerminated bool
	mu               *sync.Mutex
}

func (p *testKafkaAsyncProducer) Input() chan<- *sarama.ProducerMessage {
	p.input = make(chan *sarama.ProducerMessage)
	go func() {
		for m := range p.input {
			p.addSent(m)
		}
		p.inputTerminated = true
	}()
	return p.input
}

func (p *testKafkaAsyncProducer) Errors() <-chan *sarama.ProducerError {
	ch := make(chan *sarama.ProducerError)
	go func() {
		for _, e := range p.errors {
			ch <- e
		}
		close(ch)
		p.errorsTerminated = true
	}()
	return ch
}

func (p *testKafkaAsyncProducer) AsyncClose() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.asyncClosed = true
	close(p.input)
}

func (p *testKafkaAsyncProducer) addSent(m *sarama.ProducerMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, m)
}

func (p *testKafkaAsyncProducer) addError(e *sarama.ProducerError) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errors = append(p.errors, e)
}
