package staking

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/service/address"
	"github.com/olivere/elastic"
	"log"
	"time"
)

var IndexAddress = ".address"
var IndexAddressTransaction = ".addresstransaction"

func GetStakingReport() (report Report, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	supplyResult, err := client.Search(config.Get().SelectedNetwork + IndexAddress).
		Aggregation("totalWealth", elastic.NewSumAggregation().Field("balance")).
		Size(0).
		Do(context.Background())
	if err != nil {
		return
	}

	if total, found := supplyResult.Aggregations.Sum("totalWealth"); found {
		report.TotalSupply = *total.Value / 100000000
	} else {
		err = ErrAddressesNotAvailable
		return
	}

	from := time.Now().UTC().Truncate(time.Second).AddDate(0,0, -1)

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("type", "STAKING"))

	heightResult, err := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).
		Query(elastic.NewRangeQuery("time").Gte(from)).
		Size(1).
		Sort("height", false).
		Collapse(elastic.NewCollapseBuilder("address")).
		Do(context.Background())

	var transaction address.Transaction
	err = json.Unmarshal(*heightResult.Hits.Hits[0].Source, &transaction)
	if err != nil {
		err = ErrAddressesNotAvailable
		return
	}

	log.Printf("Height greater than %d\n", transaction.Height)


	return report, err
}

var (
	ErrAddressesNotAvailable = errors.New("addresses not available")
)