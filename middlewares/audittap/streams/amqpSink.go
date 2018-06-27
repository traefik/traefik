package streams

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/assembla/cony"
	"github.com/beeker1121/goque"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/audittap/configuration"
	"github.com/containous/traefik/middlewares/audittap/encryption"
	atypes "github.com/containous/traefik/middlewares/audittap/types"
	"github.com/streadway/amqp"
)

const undeliveredMessagePrefix = "Message not delivered to MQ because"

type amqpAuditSink struct {
	cli       amqpConyClient
	messages  chan atypes.Encoded
	producers []*amqpProducer
	q         *goque.Queue
	enc       encryption.Encrypter
}

type auditDescription struct {
	EventID     string `json:"eventId"`
	AuditSource string `json:"auditSource"`
	AuditType   string `json:"auditType"`
}

type amqpConyPublisher interface {
	Publish(pub amqp.Publishing) error
	Cancel()
	GetConyPublisher() *cony.Publisher
}

type amqpConyClient interface {
	Declare(d []cony.Declaration)
	Errors() <-chan error
	Blocking() <-chan amqp.Blocking
	Publish(pub amqpConyPublisher)
	Close()
	Loop() bool
}

type conyClientImpl struct {
	cli *cony.Client
}

func (c *conyClientImpl) Declare(d []cony.Declaration) {
	c.cli.Declare(d)
}

func (c *conyClientImpl) Errors() <-chan error {
	return c.cli.Errors()
}

func (c *conyClientImpl) Blocking() <-chan amqp.Blocking {
	return c.cli.Blocking()
}

func (c *conyClientImpl) Publish(pub amqpConyPublisher) {
	c.cli.Publish(pub.GetConyPublisher())
}

func (c *conyClientImpl) Close() {
	c.cli.Close()
}

func (c *conyClientImpl) Loop() bool {
	return c.cli.Loop()
}

type conyPublisherImpl struct {
	publisher *cony.Publisher
}

func (p *conyPublisherImpl) Publish(pub amqp.Publishing) error {
	return p.publisher.Publish(pub)
}

func (p *conyPublisherImpl) Cancel() {
	p.publisher.Cancel()
}

func (p *conyPublisherImpl) GetConyPublisher() *cony.Publisher {
	return p.publisher
}

// NewConyClient is a wrapper for calling cony.NewClient
var NewConyClient = func(endpoint string, clientID string, clientVersion string) amqpConyClient {
	props := amqp.Table{"product": clientID, "version": clientVersion, "connection_name": clientID}
	cfg := amqp.Config{Properties: props}
	return &conyClientImpl{cli: cony.NewClient(cony.URL(endpoint), cony.Config(cfg))}
}

// NewConyPublisher is a wrapper for calling cony.NewPublisher
var NewConyPublisher = func(exchange string) amqpConyPublisher {
	return &conyPublisherImpl{publisher: cony.NewPublisher(exchange, "")}
}

// NewQueue is a wrapper for calling cony.NewPublisher
var NewQueue = func(queueLocation string) (*goque.Queue, error) {
	return goque.OpenQueue(queueLocation)
}

// NewAmqpSink returns an AuditSink for sending messages to an AMQP service.
// A connection is made to the specified endpoint and a number of Producers
// each backed by an AMQP channel are created, ready to send messages.
func NewAmqpSink(config *configuration.AuditSink, messageChan chan atypes.Encoded) (sink AuditSink, err error) {

	clientID := config.ClientID
	if clientID == "" {
		clientID = "hmrc-traefik-" + config.ProxyingFor
	}

	clientVersion := config.ClientVersion
	if clientVersion == "" {
		clientVersion = "not-set"
	}

	cli := NewConyClient(config.Endpoint, clientID, clientVersion)

	exc := cony.Exchange{
		Name:       config.Destination,
		Kind:       "topic",
		AutoDelete: false,
		Durable:    true,
	}

	cli.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})

	producers := make([]*amqpProducer, 0)
	q, err := NewQueue(config.DiskStorePath)
	if err != nil {
		return nil, err
	}

	enc, err := encryption.NewEncrypter(config.EncryptSecret)
	if err != nil {
		return nil, err
	}
	for i := 0; i < config.NumProducers; i++ {
		p, _ := newAmqpProducer(cli, config.Destination, messageChan, q, enc)
		producers = append(producers, p)
	}

	aas := &amqpAuditSink{cli: cli, producers: producers, messages: messageChan, q: q, enc: enc}

	go func() {
		for cli.Loop() {
			select {
			case err := <-cli.Errors():
				log.Errorf("AMQP Client error: %v", err)
			case blocked := <-cli.Blocking():
				log.Warnf("AMQP Client is blocked %v", blocked)
			}
		}
	}()

	return aas, nil
}

func (aas *amqpAuditSink) Audit(encoded atypes.Encoded) error {
	select {
	case aas.messages <- encoded:
	default:
		handleFailedMessage(encoded, "channel full", aas.enc)
	}
	return nil
}

func (aas *amqpAuditSink) Close() error {
	for _, p := range aas.producers {
		p.stop <- true
	}
	aas.cli.Close()
	aas.q.Close()
	return nil
}

type amqpProducer struct {
	cli       amqpConyClient
	exchange  string
	publisher amqpConyPublisher
	messages  chan atypes.Encoded
	q         *goque.Queue
	stop      chan bool
	enc       encryption.Encrypter
}

func newAmqpProducer(cli amqpConyClient, exchange string, messages chan atypes.Encoded, q *goque.Queue, enc encryption.Encrypter) (*amqpProducer, error) {
	publisher := NewConyPublisher(exchange)
	cli.Publish(publisher)

	stop := make(chan bool)
	producer := &amqpProducer{cli: cli, exchange: exchange, messages: messages, publisher: publisher, q: q, stop: stop, enc: enc}
	go producer.audit()
	go producer.publish()
	return producer, nil
}

func (p *amqpProducer) audit() {
	for {
		encoded := <-p.messages
		_, err := p.q.EnqueueObject(encoded)
		if err != nil {
			handleFailedMessage(encoded, "enqueue failed", p.enc)
		}
	}
}

func minimallyDescribeAudit(encoded atypes.Encoded) (auditDescription, error) {
	var data auditDescription
	err := json.Unmarshal(encoded.Bytes, &data)
	return data, err
}

func handleFailedMessage(encoded atypes.Encoded, reason string, crypter encryption.Encrypter) {
	// Assume an indescribable event would be rejected by Datastream
	if desc, err := minimallyDescribeAudit(encoded); err == nil {
		if msgBody, err := crypter.Encrypt(encoded.Bytes); err == nil {
			log.Error(fmt.Sprintf("%s %s eventId=%s auditSource=%s auditType=%s body: {%s}",
				undeliveredMessagePrefix, reason, desc.EventID, desc.AuditSource, desc.AuditType, msgBody))
		} else {
			// Datastream would drop for an encryption failure
			log.Error(fmt.Sprintf("Dropping unencrypted event. eventId=%s auditSource=%s auditType=%s",
				desc.EventID, desc.AuditSource, desc.AuditType))
		}
	} else {
		log.Error(fmt.Sprintf("Dropping invalid audit event. %s", err.Error()))
	}
}

func (p *amqpProducer) Close() error {
	p.publisher.Cancel()
	return nil
}

func (p *amqpProducer) publish() {
	for {
		select {
		case <-p.stop:
			return

		default:
			item, err := p.q.Dequeue()
			if err != nil {
				if err == goque.ErrEmpty {
					time.Sleep(2 * time.Millisecond)
					continue
				}
				// now? nothing to see here ... Should only happen if reference to goque.q is "closed"
				log.Error(err)
				continue
			}
			var encoded atypes.Encoded
			if err = item.ToObject(&encoded); err != nil {
				// well, that didn't work
				log.Error(err)
			}
			// back-off retry
			bo := backoff.NewExponentialBackOff()

		Loop:
			for {
				select {
				case <-p.stop:
					// we've been asked to stop prior to publication: re-enqueue the audit message in disk queue
					p.q.EnqueueObject(encoded)
					return
				default:
					msg := amqp.Publishing{DeliveryMode: amqp.Persistent, Body: encoded.Bytes}
					if err = p.publisher.Publish(msg); err != nil {
						duration := bo.NextBackOff()
						if duration != backoff.Stop {
							time.Sleep(duration)
						} else {
							bo.Reset()
						}
					} else {
						break Loop
					}
				}
			}
		}
	}
}
