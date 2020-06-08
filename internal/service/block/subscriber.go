package block

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/event"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	once     sync.Once
	instance *Subscriber
)

type Subscriber struct {
	consumer *event.Consumer
	networks []string
}

func NewBlockSubscriber(networks []string, consumer *event.Consumer) *Subscriber {
	once.Do(func() {
		instance = &Subscriber{
			consumer: consumer,
			networks: networks,
		}
	})

	return instance
}

func (s *Subscriber) Subscribe() {
	for _, n := range s.networks {
		go s.consumer.Consume(n, "indexed.block", react(n))
	}
}

func react(network string) func(string) {
	return func(msg string) {
		log.Infof("Block %s indexed for %s", msg, network)
		return
	}
}
