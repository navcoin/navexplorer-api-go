package event

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
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

func (c *Consumer) Consume(network network.Network, name string, callback func(msg string)) {
	xname := fmt.Sprintf("%s.%s", network.ToString(), name)
	qname := fmt.Sprintf("%s.%s", xname, c.prefix)

	conn, err := amqp.Dial(c.address)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to connect to RabbitMQ")
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to open a channel")
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(xname, "fanout", true, false, false, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to declare an exchange")
		return
	}

	q, err := ch.QueueDeclare(qname, false, false, false, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to declare a queue")
		return
	}

	err = ch.QueueBind(q.Name, "", xname, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to bind a queue")
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to consume the queue")
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Debugf("[Event] Received message: %s", d.Body)
			callback(string(d.Body))
		}
	}()

	log.WithFields(log.Fields{"network": network.ToString(), "exchange": xname, "queue": qname}).Debugf("[Event] Waiting for messages")
	<-forever
}

func (p *Consumer) handleError(err error, msg string) {
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", msg)
	}
}
