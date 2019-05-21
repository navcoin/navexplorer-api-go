package staking

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/service/address"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"github.com/NavExplorer/navexplorer-api-go/service/coin"
	"github.com/olivere/elastic"
	"time"
)

var IndexAddressTransaction = ".addresstransaction"
var IndexBlock = ".block"


func GetStakingReport() (report Report, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	totalSupply, err := coin.GetTotalSupply()
	if err == nil {
		report.TotalSupply = totalSupply
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
		err := json.Unmarshal(*&hit.Source, &transaction)
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

func GetStakingByBlockCount(blockCount int) (stakingBlocks StakingBlocks, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	bestBlock, err := block.GetBestBlock()
	if err != nil {
		return
	}

	if blockCount > bestBlock.Height {
		blockCount = bestBlock.Height
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("height").Gt(bestBlock.Height - blockCount))
	query = query.Must(elastic.NewTermQuery("type.keyword", "STAKING"))
	query = query.Must(elastic.NewTermQuery("standard", true))

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).
		Query(query).
		Size(blockCount).
		Sort("height", false).
		Collapse(elastic.NewCollapseBuilder("address.keyword")).
		Do(context.Background())

	for _, hit := range results.Hits.Hits {
		var transaction address.Transaction
		err := json.Unmarshal(*&hit.Source, &transaction)
		if err == nil {
			stakingBlocks.Staking += transaction.Balance / 100000000
		}
	}

	query = elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("height").Gt(bestBlock.Height - blockCount))
	query = query.Must(elastic.NewTermQuery("type.keyword", "COLD_STAKING"))
	query = query.Must(elastic.NewTermQuery("standard", true))

	results, err = client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).
		Query(query).
		Size(blockCount).
		Sort("height", false).
		Collapse(elastic.NewCollapseBuilder("address.keyword")).
		Do(context.Background())

	for _, hit := range results.Hits.Hits {
		var transaction address.Transaction
		err := json.Unmarshal(*&hit.Source, &transaction)
		if err == nil {
			stakingBlocks.ColdStaking += transaction.Balance / 100000000
		}
	}

	fees, err := block.GetFeesForLastBlocks(blockCount)
	if err == nil {
		stakingBlocks.Fees = fees
	}

	stakingBlocks.BlockCount = blockCount

	return
}