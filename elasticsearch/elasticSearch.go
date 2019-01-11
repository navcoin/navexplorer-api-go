package elasticsearch

import (
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/olivere/elastic"
)

func NewClient() (client *elastic.Client) {
	client, err := elastic.NewClient(
		elastic.SetURL(config.Get().ElasticSearch.Urls),
		elastic.SetSniff(config.Get().ElasticSearch.Sniff),
		elastic.SetHealthcheck(config.Get().ElasticSearch.HealthCheck))

	if err != nil {
		// Handle error
		panic(err)
	}

	return client
}