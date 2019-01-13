package elasticsearch

import (
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/olivere/elastic"
	"log"
	"os"
)

func NewClient() (client *elastic.Client) {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(config.Get().ElasticSearch.Urls),
		elastic.SetSniff(config.Get().ElasticSearch.Sniff),
		elastic.SetHealthcheck(config.Get().ElasticSearch.HealthCheck),
	}

	if config.Get().Debug {
		opts = append(opts, elastic.SetTraceLog(log.New(os.Stdout, "ELASTIC ", 0)))
	}

	client, err := elastic.NewClient(opts...)

	if err != nil {
		// Handle error
		panic(err)
	}

	return client
}