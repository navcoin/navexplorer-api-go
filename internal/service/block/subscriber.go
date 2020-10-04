package block

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/cache"
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
	cache    *cache.Cache
}

func NewBlockSubscriber(networks []string, consumer *event.Consumer, cache *cache.Cache) *Subscriber {
	once.Do(func() {
		instance = &Subscriber{
			consumer: consumer,
			networks: networks,
			cache:    cache,
		}
	})

	return instance
}

func (s *Subscriber) Subscribe() {
	log.Info("API is subscribing to events")

	for _, n := range s.networks {
		go s.consumer.Consume(n, "indexed.block", react(s, n))
	}
}

func react(s *Subscriber, network string) func(string) {
	return func(msg string) {
		log.Infof("Block %s indexed for %s", msg, network)
		s.cache.Refresh(network)

		return
	}
}
