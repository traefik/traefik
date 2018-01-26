package integration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	"github.com/streadway/amqp"
	checker "github.com/vdemeester/shakers"
)

// Rabbitmq test suites (using libcompose)
type RabbitMQSuite struct {
	BaseSuite
	deliveryChannel <-chan amqp.Delivery
}

func (s *RabbitMQSuite) setupRabbitMQ(c *check.C) {
	s.createComposeProject(c, "rabbitmq")
	s.composeProject.Start(c)

	rabbitMQHost := s.composeProject.Container(c, "rabbitmq").NetworkSettings.IPAddress

	var conn *amqp.Connection

	// Wait for Rabbit to start
	err := try.Do(20*time.Second, func() error {
		var e2 error
		conn, e2 = amqp.Dial(fmt.Sprintf("amqp://%s:5672/", rabbitMQHost))

		if e2 != nil {
			return e2
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	ch, err := conn.Channel()
	c.Assert(err, checker.IsNil)

	err = ch.ExchangeDeclare("audit", "topic", true, false, false, false, nil)
	c.Assert(err, checker.IsNil)

	queue, err := ch.QueueDeclare("splunk", true, false, false, false, nil)
	c.Assert(err, checker.IsNil)

	err = ch.QueueBind(queue.Name, "", "audit", false, nil)
	c.Assert(err, checker.IsNil)

	s.deliveryChannel, err = ch.Consume(queue.Name, "c", false, false, false, false, nil)
	c.Assert(err, checker.IsNil)
}

func (s *RabbitMQSuite) TearDownTest(c *check.C) {
	// shutdown and delete compose project
	if s.composeProject != nil {
		s.composeProject.Stop(c)
	}
}

func (s *RabbitMQSuite) TearDownSuite(c *check.C) {}

func (s *RabbitMQSuite) TestSimpleConfiguration(c *check.C) {
	s.setupRabbitMQ(c)

	rabbitMQHost := s.composeProject.Container(c, "rabbitmq").NetworkSettings.IPAddress
	file := s.adaptFile(c, "fixtures/rabbitmq/simple.toml", struct{ RabbitMQHost string }{rabbitMQHost})
	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)

	var b bytes.Buffer
	foo := bufio.NewWriter(&b)

	cmd.Stdout = foo

	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	time.Sleep(1000 * time.Millisecond)
	resp, err := http.Get("http://127.0.0.1:8000/test?foo=bar")
	c.Assert(err, checker.IsNil)

	// Expected a 502 as Nginx isn't configured
	c.Assert(resp.StatusCode, checker.Equals, 502)

	msg := <-s.deliveryChannel
	msg.Ack(false)

	var dat map[string]interface{}
	json.Unmarshal(msg.Body, &dat)

	// Simple check to make sure it's the message we sent
	c.Assert(dat["auditSource"].(string), checker.Equals, "testAuditSource")
	c.Assert(dat["auditType"].(string), checker.Equals, "testAuditType")
}
