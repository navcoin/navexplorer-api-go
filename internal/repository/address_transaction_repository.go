package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"time"
)

type AddressTransactionRepository struct {
	elastic *elastic_cache.Index
}

func NewAddressTransactionRepository(elastic *elastic_cache.Index) *AddressTransactionRepository {
	return &AddressTransactionRepository{elastic}
}

func (r *AddressTransactionRepository) TransactionsByHash(hash string, types string, cold bool, dir bool, size int, page int) ([]*explorer.AddressTransaction, int64, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))
	query = query.Must(elastic.NewTermQuery("cold", cold))

	if len(types) != 0 {
		if strings.Contains(types, string(explorer.TransferStake)) {
			types += fmt.Sprintf(" %s", explorer.TransferColdStake)
		}
		if strings.Contains(types, string(explorer.TransferReceive)) {
			types += fmt.Sprintf(" %s", explorer.TransferCommunityFundPayout)
		}
		query = query.Must(elastic.NewMatchQuery("type", types))
	}

	results, err := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).
		Query(query).
		Sort("height", dir).
		Sort("index", dir).
		From((page * size) - size).
		Size(size).
		TrackTotalHits(true).
		Do(context.Background())

	return r.findMany(results, err)
}

func (r *AddressTransactionRepository) BalanceChart(address string) (chart entity.Chart, err error) {
	now := time.Now().UTC().Truncate(time.Second)
	from := time.Date(now.Year(), now.Month(), now.Day()-30, 0, 0, 0, 0, now.Location())

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", address))
	query = query.Must(elastic.NewRangeQuery("time").Gte(from))

	results, err := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).
		Query(query).
		Sort("height", false).
		Size(10000).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	for _, hit := range results.Hits.Hits {
		var tx explorer.AddressTransaction
		err := json.Unmarshal(hit.Source, &tx)
		if err == nil {
			chartPoint := &entity.ChartPoint{
				Time:  tx.Time,
				Value: float64(tx.Balance) / 100000000,
			}
			chart.Points = append(chart.Points, chartPoint)
		}
	}

	return chart, err
}

func (r *AddressTransactionRepository) StakingChart(period string, hash string) (groups []*entity.StakingGroup, err error) {
	count := 12
	now := time.Now().UTC().Truncate(time.Second)

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("hash", hash))
	query = query.Must(elastic.NewMatchQuery("type", "stake cold_stake"))

	agg := elastic.NewFilterAggregation().Filter(query)

	for i := 0; i < count; i++ {
		group := &entity.StakingGroup{End: now}

		switch period {
		case "hourly":
			{
				if i == 0 {
					group.Start = now.Truncate(time.Hour)
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.Add(-time.Hour)
				}
				break
			}
		case "daily":
			{
				if i == 0 {
					group.Start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.AddDate(0, 0, -1)
				}
				break
			}
		case "monthly":
			{
				if i == 0 {
					group.Start = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location())
					group.Start = group.Start.AddDate(0, 0, 1)
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.AddDate(0, -1, 0)
				}
				break
			}
		}

		timeAgg := elastic.NewRangeAggregation().Field("time").AddRange(group.Start, group.End)
		timeAgg.SubAggregation("total", elastic.NewSumAggregation().Field("total"))

		agg.SubAggregation(fmt.Sprintf("group-%d", i), timeAgg)

		groups = append(groups, group)
	}

	results, err := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).
		Aggregation("groups", agg).
		Size(0).
		Do(context.Background())
	if results != nil {
		i := 0
		for {
			if agg, found := results.Aggregations.Filter("groups"); found {
				if groupAgg, found := agg.Aggregations.Range(fmt.Sprintf("group-%d", i)); found {
					bucket := groupAgg.Buckets[0]
					groups[i].Stakes = bucket.DocCount

					if totalValue, found := bucket.Aggregations.Sum("total"); found {
						groups[i].Amount = int64(*totalValue.Value)
					}
					i++
				} else {
					break
				}
			}
		}
	}

	return groups, err
}

func (r *AddressTransactionRepository) TransactionsForAddresses(addresses []string, txType string, start *time.Time, end *time.Time) ([]*explorer.AddressTransaction, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", strings.Join(addresses, " ")))
	query = query.Must(elastic.NewMatchQuery("type", txType))
	query = query.Must(elastic.NewRangeQuery("time").Gt(&start).Lte(&end))

	results, err := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).
		Query(query).
		Size(5000).
		Sort("time", false).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	txs := make([]*explorer.AddressTransaction, 0)
	for _, hit := range results.Hits.Hits {
		tx := new(explorer.AddressTransaction)
		err := json.Unmarshal(hit.Source, &tx)
		if err == nil {
			txs = append(txs, tx)
		}
	}

	return txs, nil
}

func (r *AddressTransactionRepository) GetStakingReport(report *entity.StakingReport) error {

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("time").Gte(report.From))
	query = query.Must(elastic.NewTermsQuery("type.keyword", "COLD_STAKING", "STAKING"))
	query = query.Must(elastic.NewTermQuery("standard", true))

	results, err := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).
		Query(query).
		Size(10000).
		Sort("height", false).
		Collapse(elastic.NewCollapseBuilder("address.keyword")).
		Do(context.Background())
	if err != nil {
		return err
	}

	for _, hit := range results.Hits.Hits {
		transaction := &explorer.AddressTransaction{}
		err := json.Unmarshal(hit.Source, &transaction)
		if err != nil {
			return err
		}
		var reporter entity.Reporter
		reporter.Address = transaction.Hash
		reporter.Balance = float64(transaction.Balance) / 100000000
		report.Addresses = append(report.Addresses, reporter)

		report.Staking += reporter.Balance
	}

	return nil
}

func (r *AddressTransactionRepository) GetStakingHigherThan(height uint64, count int) (*entity.StakingBlocks, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("height").Gt(height))

	hashAgg := elastic.NewTermsAggregation().Field("hash.keyword")
	hashAgg.SubAggregation("balance", elastic.NewMaxAggregation().Field("balance"))

	stakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("type.keyword", explorer.TransferStake))
	stakeAgg.SubAggregation("hash", hashAgg)

	coldStakeAgg := elastic.NewFilterAggregation().Filter(
		elastic.NewBoolQuery().
			Must(elastic.NewTermQuery("type.keyword", explorer.TransferColdStake)).
			Must(elastic.NewTermQuery("cold", true)),
	).SubAggregation("hash", hashAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).
		Query(query).
		Aggregation("stake", stakeAgg).
		Aggregation("coldStake", coldStakeAgg).
		Size(0).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	stakeBalance := float64(0)
	if stake, found := results.Aggregations.Filter("stake"); found {
		if hash, found := stake.Aggregations.Terms("hash"); found {
			for _, bucket := range hash.Buckets {
				if balance, found := bucket.Aggregations.Max("balance"); found {
					stakeBalance += *balance.Value
				}
			}
		}
	}

	coldStakeBalance := float64(0)
	if stake, found := results.Aggregations.Filter("coldStake"); found {
		if hash, found := stake.Aggregations.Terms("hash"); found {
			for _, bucket := range hash.Buckets {
				if balance, found := bucket.Aggregations.Max("balance"); found {
					coldStakeBalance += *balance.Value
				}
			}
		}
	}

	return &entity.StakingBlocks{
		Staking:     stakeBalance / 100000000,
		ColdStaking: coldStakeBalance / 100000000,
		BlockCount:  count,
	}, nil
}

func (r *AddressTransactionRepository) StakingRewardsForAddresses(addresses []string) ([]*entity.StakingReward, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", strings.Join(addresses, " ")))
	query = query.Must(elastic.NewMatchQuery("type", "STAKING COLD_STAKING"))

	now := time.Now().UTC().Truncate(time.Second)

	agg := elastic.NewTermsAggregation().Field("address.keyword")
	agg.SubAggregation("last24Hours", dateGroupAgg(now.Add(-(time.Hour*24)), now))
	agg.SubAggregation("last7Days", dateGroupAgg(now.Add(-(time.Hour*24*7)), now))
	agg.SubAggregation("last30Days", dateGroupAgg(now.Add(-(time.Hour*24*30)), now))
	agg.SubAggregation("lastYear", dateGroupAgg(now.Add(-(time.Hour*24*365)), now))
	agg.SubAggregation("all", dateGroupAgg(now.AddDate(-100, 0, 0), now))

	service := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get())
	service.Query(query)
	service.Aggregation("groups", agg)

	results, err := service.Do(context.Background())
	if err != nil {
		return nil, err
	}

	rewards := make([]*entity.StakingReward, 0)
	if agg, found := results.Aggregations.Terms("groups"); found {

		for _, bucket := range agg.Buckets {
			reward := &entity.StakingReward{Address: bucket.Key.(string)}
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

	return rewards, nil
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

func stakingPeriodResults(bucket *elastic.AggregationBucketKeyItem, periodName string) *entity.StakingRewardPeriod {
	rewardPeriod := &entity.StakingRewardPeriod{Period: periodName}

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

func (r *AddressTransactionRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.AddressTransaction, int64, error) {
	if err != nil {
		return nil, 0, err
	}

	txs := make([]*explorer.AddressTransaction, 0)
	for _, hit := range results.Hits.Hits {
		var tx *explorer.AddressTransaction
		if err := json.Unmarshal(hit.Source, &tx); err == nil {
			txs = append(txs, tx)
		} else {
			logrus.WithError(err).Info("Failed to get transaction")
		}
	}

	return txs, results.TotalHits(), err
}
