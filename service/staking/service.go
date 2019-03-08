package staking

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/service/address"
	"github.com/olivere/elastic"
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

	to := time.Now().UTC().Truncate(time.Second)
	from := to.AddDate(0,0, -1)

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("time").Gte(from))
	query = query.Must(elastic.NewTermsQuery("type.keyword", "COLD_STAKING", "STAKING"))
	query = query.Must(elastic.NewTermQuery("standard", true))

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).
		Query(query).
		Size(10000).
		Sort("height", false).
		Collapse(elastic.NewCollapseBuilder("address.keyword")).
		Do(context.Background())

	for _, hit := range results.Hits.Hits {
		var transaction address.Transaction
		err := json.Unmarshal(*hit.Source, &transaction)
		if err == nil {
			var reporter Reporter
			reporter.Address = transaction.Address
			reporter.Balance = transaction.Balance / 100000000
			report.Addresses = append(report.Addresses, reporter)

			report.Staking += reporter.Balance
		}
	}

	report.To = to
	report.From = from

	return report, err
}

var (
	ErrAddressesNotAvailable = errors.New("addresses not available")
)