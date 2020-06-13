package event

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Consumer struct {
	address string
	prefix  string
}

func NewConsumer(user string, password string, host string, port int, prefix string) *Consumer {
	return &Consumer{
		address: fmt.Sprintf("amqp://%s:%s@%s:%d", user, password, host, port),
		prefix:  prefix,
	}
}

func (c *Consumer) Consume(network string, name string, callback func(msg string)) {
	xname := fmt.Sprintf("%s.%s", network, name)
	qname := fmt.Sprintf("%s.%s", xname, c.prefix)

	conn, err := amqp.Dial(c.address)
	c.handleError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	c.handleError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(xname, "fanout", true, false, false, false, nil)
	c.handleError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(qname, false, false, false, false, nil)
	c.handleError(err, "Failed to declare a queue")

	err = ch.QueueBind(q.Name, "", xname, false, nil)
	c.handleError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Debugf("[Event] Received message: %s", d.Body)
			callback(string(d.Body))
		}
	}()

	log.Debugf("[Event] Waiting for messages")
	<-forever
}

func (p *Consumer) handleError(err error, msg string) {
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", msg)
	}
}
