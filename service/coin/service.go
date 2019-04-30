package coin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/service/address"
	"github.com/olivere/elastic"
)

var IndexAddress = config.Get().SelectedNetwork + ".address"

func GetWealthDistribution(groups []int) (distribution []Wealth, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	var totalWealth Wealth
	results, _ := client.Search(config.Get().SelectedNetwork + IndexAddress).
		Aggregation("totalWealth", elastic.NewSumAggregation().Field("balance")).
		Do(context.Background())
	if total, found := results.Aggregations.Sum("totalWealth"); found {
		totalWealth.Balance = *total.Value
		totalWealth.Percentage = 100
	}

	distribution = make([]Wealth, len(groups) + 1)

	for i := 0; i < len(groups); i++ {
		results, _ := client.Search(config.Get().SelectedNetwork + IndexAddress).
			From(0).
			Size(groups[i]).
			Sort("balance", false).
			Do(context.Background())

		var wealth Wealth
		wealth.Group = groups[i]

		for _, element := range results.Hits.Hits {
			var add address.Address
			err = json.Unmarshal(*element.Source, &add)

			wealth.Balance += add.Balance
			wealth.Percentage = int64((wealth.Balance / totalWealth.Balance) * 100)
		}

		distribution[i] = wealth
	}

	distribution[len(groups)] = totalWealth



	return distribution, err
}

func GetTotalSupply() (totalSupply float64, err error) {
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
		totalSupply = *total.Value / 100000000
	}

	return
}
