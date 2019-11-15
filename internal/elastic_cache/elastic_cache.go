package elastic_cache

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/config"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strings"
)

type Index struct {
	Client *elastic.Client
}

func New() (*Index, error) {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(strings.Join(config.Get().ElasticSearch.Hosts, ",")),
		elastic.SetSniff(config.Get().ElasticSearch.Sniff),
		elastic.SetHealthcheck(config.Get().ElasticSearch.HealthCheck),
	}

	if config.Get().ElasticSearch.Username != "" {
		logrus.Info("Using ES Authentication")
		opts = append(opts, elastic.SetBasicAuth(
			config.Get().ElasticSearch.Username,
			config.Get().ElasticSearch.Password,
		))
	}

	if config.Get().ElasticSearch.Debug {
		opts = append(opts, elastic.SetTraceLog(log.New(os.Stdout, "", 0)))
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		log.Println("Error: ", err)
	}

	return &Index{Client: client}, err
}
