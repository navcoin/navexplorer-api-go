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
	"log"
	"strings"
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
		err := json.Unmarshal(*hit.Source, &transaction)
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
		err := json.Unmarshal(*hit.Source, &transaction)
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

func GetStakingRewardsForAddresses(addresses []string) (rewards []Reward, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", strings.Join(addresses, " ")))
	query = query.Must(elastic.NewMatchQuery("type", "STAKING COLD_STAKING"))

	now := time.Now().UTC().Truncate(time.Second)

	agg := elastic.NewTermsAggregation().Field("address.keyword")
	agg.SubAggregation("last24Hours", dateGroupAgg(now.Add(-(time.Hour * 24)), now))
	agg.SubAggregation("last7Days", dateGroupAgg(now.Add(-(time.Hour * 24 * 7)), now))
	agg.SubAggregation("last30Days", dateGroupAgg(now.Add(-(time.Hour * 24 * 30)), now))
	agg.SubAggregation("lastYear", dateGroupAgg(now.Add(-(time.Hour * 24 * 365)), now))
	agg.SubAggregation("all", dateGroupAgg(now.AddDate(-100, 0, 0), now))

	service := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).Size(0)
	service.Query(query)
	service.Aggregation("groups", agg)

	results, err := service.Do(context.Background())
	if err == nil && results != nil {
		if agg, found := results.Aggregations.Terms("groups"); found {

			for _, bucket := range agg.Buckets {
				reward := Reward{Address: bucket.Key.(string)}
				reward.Periods = append(reward.Periods, stakingPeriodResults(bucket, "last24Hours"))
				reward.Periods = append(reward.Periods, stakingPeriodResults(bucket, "last7Days"))
				reward.Periods = append(reward.Periods, stakingPeriodResults(bucket, "last30Days"))
				reward.Periods = append(reward.Periods, stakingPeriodResults(bucket, "lastYear"))
				reward.Periods = append(reward.Periods, stakingPeriodResults(bucket, "all"))

				rewardJson, _ := json.Marshal(reward)
				log.Println(string(rewardJson))
				rewards = append(rewards, reward)
			}
		}
	}

	return
}

func dateGroupAgg(from time.Time, to time.Time) (aggregation *elastic.RangeAggregation) {
	aggregation = elastic.NewRangeAggregation().Field("time").AddRange(from, to)
	aggregation.SubAggregation("sent", elastic.NewSumAggregation().Field("sent"))
	aggregation.SubAggregation("received", elastic.NewSumAggregation().Field("received"))
	aggregation.SubAggregation("coldStakingSent", elastic.NewSumAggregation().Field("coldStakingSent"))
	aggregation.SubAggregation("coldStakingReceived", elastic.NewSumAggregation().Field("coldStakingReceived"))
	aggregation.SubAggregation("delegateStake", elastic.NewSumAggregation().Field("delegateStake"))

	return
}

func stakingPeriodResults(bucket *elastic.AggregationBucketKeyItem, periodName string) (rewardPeriod RewardPeriod) {
	rewardPeriod = RewardPeriod{Period: periodName}

	if period, found := bucket.Aggregations.Range(rewardPeriod.Period); found {
		aggBucket := period.Buckets[0]

		sent := int64(0)
		received := int64(0)
		if sentValue, found := aggBucket.Aggregations.Sum("sent"); found {
			sent = sent + int64(*sentValue.Value)
		}
		if coldStakingSentValue, found := aggBucket.Aggregations.Sum("coldStakingSent"); found {
			sent = sent + int64(*coldStakingSentValue.Value)
		}
		if receivedValue, found := aggBucket.Aggregations.Sum("received"); found {
			received = received + int64(*receivedValue.Value)
		}
		if coldStakingReceivedValue, found := aggBucket.Aggregations.Sum("coldStakingReceived"); found {
			received = received + int64(*coldStakingReceivedValue.Value)
		}
		if delegateStakeValue, found := aggBucket.Aggregations.Sum("delegateStake"); found {
			received = received + int64(*delegateStakeValue.Value)
		}

		rewardPeriod.Stakes = aggBucket.DocCount
		rewardPeriod.Balance = received - sent
	}

	return rewardPeriod
}