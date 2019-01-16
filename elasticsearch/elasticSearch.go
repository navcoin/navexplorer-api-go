package elasticsearch

import (
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
	"log"
	"os"
)

func NewClient() (client *elastic.Client, err error) {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(config.Get().ElasticSearch.Urls),
		elastic.SetSniff(config.Get().ElasticSearch.Sniff),
		elastic.SetHealthcheck(config.Get().ElasticSearch.HealthCheck),
	}

	if config.Get().Debug {
		opts = append(opts, elastic.SetTraceLog(log.New(os.Stdout, "", 0)))
	}

	client, err = elastic.NewClient(opts...)
	if err != nil {
		err = errors.New("Unable to connect to elastic search")
		log.Print(err)
	}

	return client, err
}