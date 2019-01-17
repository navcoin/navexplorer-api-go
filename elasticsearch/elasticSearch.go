package elasticsearch

import (
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/olivere/elastic"
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
		log.Print("Error: ", err)
		err = ErrDatabaseConnection
	}

	return client, err
}

var (
	ErrDatabaseConnection = errors.New("could not connect to the database")
)
