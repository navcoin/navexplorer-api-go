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

func (r *AddressTransactionRepository) BalanceChart(address string) (chart entity.Chart, err error) {
	now := time.Now().UTC().Truncate(time.Second)
	from := time.Date(now.Year(), now.Month(), now.Day()-30, 0, 0, 0, 0, now.Location())

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("hash", address))
	query = query.Must(elastic.NewRangeQuery("time").Gte(from))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Sort("height", false).
		Size(10000).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	for _, hit := range results.Hits.Hits {
		var history explorer.AddressHistory
		err := json.Unmarshal(hit.Source, &history)
		if err == nil {
			chartPoint := &entity.ChartPoint{
				Time:  history.Time,
				Value: float64(history.Balance.Spending) / 100000000,
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
	query = query.Must(elastic.NewMatchQuery("is_stake", true))

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

		changesAgg := elastic.NewNestedAggregation().Path("changes")
		changesAgg.SubAggregation("staking", elastic.NewSumAggregation().Field("changes.staking"))
		changesAgg.SubAggregation("spending", elastic.NewSumAggregation().Field("changes.spending"))
		changesAgg.SubAggregation("voting", elastic.NewSumAggregation().Field("changes.voting"))

		timeAgg := elastic.NewRangeAggregation().Field("time").AddRange(group.Start, group.End)
		timeAgg.SubAggregation("changes", changesAgg)

		agg.SubAggregation(fmt.Sprintf("group-%d", i), timeAgg)

		groups = append(groups, group)
	}

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
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

					if nested, found := bucket.Aggregations.Nested("changes"); found {
						if stakingValue, found := nested.Aggregations.Sum("staking"); found {
							groups[i].Staking = int64(*stakingValue.Value)
						}
						if spendingValue, found := nested.Aggregations.Sum("spending"); found {
							groups[i].Spending = int64(*spendingValue.Value)
						}
						if votingValue, found := nested.Aggregations.Sum("voting"); found {
							groups[i].Voting = int64(*votingValue.Value)
						}
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

func (r *AddressTransactionRepository) GetStakingReport(report *entity.StakingReport) error {

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("time").Gte(report.From))
	query = query.Must(elastic.NewTermsQuery("is_stake", true))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Size(10000).
		Sort("height", false).
		Collapse(elastic.NewCollapseBuilder("hash.keyword")).
		Do(context.Background())
	if err != nil {
		return err
	}

	for _, hit := range results.Hits.Hits {
		history := &explorer.AddressHistory{}
		err := json.Unmarshal(hit.Source, &history)
		if err != nil {
			return err
		}
		var reporter entity.Reporter
		reporter.Address = history.Hash
		reporter.Balance = float64(history.Balance.Staking) / 100000000
		report.Addresses = append(report.Addresses, reporter)

		report.Staking += reporter.Balance
	}

	return nil
}

func (r *AddressTransactionRepository) GetStakingRange(from uint64, to uint64) (*entity.StakingBlocks, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewRangeQuery("height").Gt(from).Lte(to))

	hashAgg := elastic.NewTermsAggregation().Field("hash.keyword")
	hashAgg.SubAggregation("balance.staking", elastic.NewMaxAggregation().Field("balance.staking"))

	stakeAgg := elastic.NewFilterAggregation().Filter(
		elastic.NewBoolQuery().
			Must(elastic.NewTermQuery("is_stake", true)).
			MustNot(elastic.NewTermQuery("changes.spending", 0)),
	)
	stakeAgg.SubAggregation("hash", hashAgg)

	coldStakeAgg := elastic.NewFilterAggregation().Filter(
		elastic.NewBoolQuery().
			Must(elastic.NewTermQuery("is_stake", true)).
			Must(elastic.NewTermQuery("changes.spending", 0)),
	).SubAggregation("hash", hashAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
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
		From:        from,
		To:          to,
	}, nil
}

func (r *AddressTransactionRepository) StakingRewardsForAddresses(addresses []string) ([]*entity.StakingReward, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("hash", strings.Join(addresses, " ")))
	query = query.Must(elastic.NewMatchQuery("is_stake", true))

	now := time.Now().UTC().Truncate(time.Second)

	agg := elastic.NewTermsAggregation().Field("hash.keyword")
	agg.SubAggregation("last24Hours", dateGroupAgg(now.Add(-(time.Hour*24)), now))
	agg.SubAggregation("last7Days", dateGroupAgg(now.Add(-(time.Hour*24*7)), now))
	agg.SubAggregation("last30Days", dateGroupAgg(now.Add(-(time.Hour*24*30)), now))
	agg.SubAggregation("lastYear", dateGroupAgg(now.Add(-(time.Hour*24*365)), now))
	agg.SubAggregation("all", dateGroupAgg(now.AddDate(-100, 0, 0), now))

	service := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get())
	service.Query(query)
	service.Size(0)
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

			rewards = append(rewards, reward)
		}
	}

	return rewards, nil
}

func dateGroupAgg(from time.Time, to time.Time) (aggregation *elastic.RangeAggregation) {
	aggregation = elastic.NewRangeAggregation().Field("time").AddRange(from, to)
	aggregation.SubAggregation("staked", elastic.NewSumAggregation().Field("total"))

	return
}

func stakingPeriodResults(bucket *elastic.AggregationBucketKeyItem, periodName string) *entity.StakingRewardPeriod {
	rewardPeriod := &entity.StakingRewardPeriod{Period: periodName}

	if period, found := bucket.Aggregations.Range(rewardPeriod.Period); found {
		aggBucket := period.Buckets[0]

		balance := int64(0)
		if stakedValue, found := aggBucket.Aggregations.Sum("staked"); found {
			balance += int64(*stakedValue.Value)
		}

		rewardPeriod.Stakes = aggBucket.DocCount
		rewardPeriod.Balance = balance
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
