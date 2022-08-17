package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/navcoin/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/navcoin/navexplorer-api-go/v2/internal/framework"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/address/entity"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/group"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/network"
	"github.com/navcoin/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
	"sync"
	"time"
)

type AddressHistoryRepository interface {
	GetLatestByHash(n network.Network, hash string) (*explorer.AddressHistory, error)
	GetFirstByHash(n network.Network, hash string) (*explorer.AddressHistory, error)
	GetCountByHash(n network.Network, hash string) (int64, error)
	GetStakingSummary(n network.Network, hash string) (count, stakable, spendable, votingWeight int64, err error)
	GetSpendSummary(n network.Network, hash string) (spendableReceive, spendableSent, stakableReceive, stakableSent, votingWeightReceive, votingWeightSent int64, err error)
	GetHistoryByHash(n network.Network, hash string, p framework.Pagination, s framework.Sort, f framework.Filters) ([]*explorer.AddressHistory, int64, error)
	GetAddressGroups(n network.Network, period *group.Period, count int) ([]entity.AddressGroup, error)
	GetAddressGroupsTotal(n network.Network, period *group.Period, count int) ([]entity.AddressGroupTotal, error)
	GetStakingChart(n network.Network, period, hash string) (groups []*entity.StakingGroup, err error)
	GetStakingRange(n network.Network, from, to uint64, address []string) (*entity.StakingBlocks, error)
	StakingRewardsForAddresses(n network.Network, addresses []string) ([]*entity.StakingReward, error)
}

var (
	ErrAddressHistoryNotFound = errors.New("Address history not found")
)

type addressHistoryRepository struct {
	elastic *elastic_cache.Index
}

func NewAddressHistoryRepository(elastic *elastic_cache.Index) AddressHistoryRepository {
	return &addressHistoryRepository{elastic: elastic}
}

func (r *addressHistoryRepository) GetLatestByHash(n network.Network, hash string) (*explorer.AddressHistory, error) {
	query := elastic.NewBoolQuery().Filter(elastic.NewMatchPhraseQuery("hash", hash))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Sort("height", false).
		Sort("txindex", false).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *addressHistoryRepository) GetFirstByHash(n network.Network, hash string) (*explorer.AddressHistory, error) {
	query := elastic.NewBoolQuery().Filter(elastic.NewMatchPhraseQuery("hash", hash))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Sort("height", true).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *addressHistoryRepository) GetCountByHash(n network.Network, hash string) (int64, error) {
	query := elastic.NewBoolQuery().Filter(elastic.NewMatchPhraseQuery("hash", hash))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		TrackTotalHits(true).
		Size(0).
		Do(context.Background())

	if err != nil {
		return 0, err
	}

	return results.TotalHits(), nil
}

func (r *addressHistoryRepository) GetStakingSummary(n network.Network, hash string) (count, stakable, spendable, votingWeight int64, err error) {
	query := elastic.NewBoolQuery().Filter(elastic.NewMatchPhraseQuery("hash", hash))

	changeAgg := elastic.NewNestedAggregation().Path("changes")
	changeAgg.SubAggregation("stakable", elastic.NewSumAggregation().Field("changes.stakable"))
	changeAgg.SubAggregation("spendable", elastic.NewSumAggregation().Field("changes.spendable"))
	changeAgg.SubAggregation("voting_weight", elastic.NewSumAggregation().Field("changes.voting_weight"))

	stakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("is_stake", true))
	stakeAgg.SubAggregation("changes", changeAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Aggregation("stake", stakeAgg).
		Sort("height", false).
		Sort("txindex", false).
		Size(0).
		Do(context.Background())

	if err == nil && results != nil {
		if agg, found := results.Aggregations.Filter("stakable"); found {
			count = agg.DocCount
			if changes, found := agg.Nested("changes"); found {
				if stakableValue, found := changes.Sum("stakable"); found {
					stakable = int64(*stakableValue.Value)
				}
				if spendableValue, found := changes.Sum("spendable"); found {
					spendable = int64(*spendableValue.Value)
				}
				if votingWeightValue, found := changes.Sum("voting_weight"); found {
					votingWeight = int64(*votingWeightValue.Value)
				}
			}
		}
	}

	return
}

func (r *addressHistoryRepository) GetSpendSummary(n network.Network, hash string) (spendableReceive, spendableSent, stakableReceive, stakableSent, votingWeightReceive, votingWeightSent int64, err error) {
	query := elastic.NewBoolQuery().
		Filter(elastic.NewMatchPhraseQuery("hash", hash)).
		Must(elastic.NewTermQuery("is_stake", false))

	spendableReceiveAgg := elastic.NewRangeAggregation().Field("changes.spendable").Gt(0)
	spendableReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.spendable"))

	spendableSentAgg := elastic.NewRangeAggregation().Field("changes.spendable").Lt(0)
	spendableSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.spendable"))

	stakableReceiveAgg := elastic.NewRangeAggregation().Field("changes.stakable").Gt(0)
	stakableReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.stakable"))

	stakableSentAgg := elastic.NewRangeAggregation().Field("changes.stakable").Lt(0)
	stakableSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.stakable"))

	votingWeightReceiveAgg := elastic.NewRangeAggregation().Field("changes.voting_weight").Gt(0)
	votingWeightReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.voting_weight"))

	votingWeightSentAgg := elastic.NewRangeAggregation().Field("changes.voting_weight").Lt(0)
	votingWeightSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.voting_weight"))

	changeAgg := elastic.NewNestedAggregation().Path("changes")
	changeAgg.SubAggregation("spendableReceive", spendableReceiveAgg)
	changeAgg.SubAggregation("spendableSent", spendableSentAgg)

	changeAgg.SubAggregation("stakableSent", stakableSentAgg)
	changeAgg.SubAggregation("votingWeightSent", votingWeightSentAgg)
	changeAgg.SubAggregation("stakableReceive", stakableReceiveAgg)
	changeAgg.SubAggregation("votingWeightReceive", votingWeightReceiveAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Aggregation("changes", changeAgg).
		Sort("height", false).
		Sort("txindex", false).
		Size(0).
		Do(context.Background())

	if err == nil && results != nil {
		if changes, found := results.Aggregations.Nested("changes"); found {
			if spendableReceiveResult, found := changes.Range("spendableReceive"); found {
				if spendableReceiveSum, found := spendableReceiveResult.Buckets[0].Sum("sum"); found {
					spendableReceive = int64(*spendableReceiveSum.Value)
				}
			}

			if spendableSentResult, found := changes.Range("spendableSent"); found {
				if spendableSentSum, found := spendableSentResult.Buckets[0].Sum("sum"); found {
					spendableSent = int64(*spendableSentSum.Value)
				}
			}

			if stakableReceiveResult, found := changes.Range("stakableReceive"); found {
				if stakableReceiveSum, found := stakableReceiveResult.Buckets[0].Sum("sum"); found {
					stakableReceive = int64(*stakableReceiveSum.Value)
				}
			}

			if stakableSentResult, found := changes.Range("stakableSent"); found {
				if stakableSentSum, found := stakableSentResult.Buckets[0].Sum("sum"); found {
					stakableSent = int64(*stakableSentSum.Value)
				}
			}

			if votingWeightReceiveResult, found := changes.Range("votingWeightReceive"); found {
				if votingWeightReceiveSum, found := votingWeightReceiveResult.Buckets[0].Sum("sum"); found {
					votingWeightReceive = int64(*votingWeightReceiveSum.Value)
				}
			}

			if votingWeightSentResult, found := changes.Range("votingWeightSent"); found {
				if votingWeightSentSum, found := votingWeightSentResult.Buckets[0].Sum("sum"); found {
					votingWeightSent = int64(*votingWeightSentSum.Value)
				}
			}
		}
	}

	return
}

func (r *addressHistoryRepository) GetHistoryByHash(n network.Network, hash string, p framework.Pagination, s framework.Sort, f framework.Filters) ([]*explorer.AddressHistory, int64, error) {
	query := elastic.NewBoolQuery().Filter(elastic.NewMatchPhraseQuery("hash", hash))

	options := f.OnlySupportedOptions([]string{"type"})
	if option, err := options.Get("type"); err == nil {
		switch option.Values()[0] {
		case "staking":
			query.Must(elastic.NewTermQuery("is_stake", true))
			break
		case "sending":
			query.Must(elastic.NewTermQuery("is_stake", false))
			query.Filter(elastic.NewNestedQuery("changes", elastic.NewBoolQuery().
				Should(elastic.NewRangeQuery("changes.spendable").Lt(0)).
				Should(elastic.NewRangeQuery("changes.stakable").Lt(0)).
				Should(elastic.NewRangeQuery("changes.voting_weight").Lt(0))),
			)
			break
		case "receiving":
			query.Must(elastic.NewTermQuery("is_stake", false))
			query.Filter(elastic.NewNestedQuery("changes", elastic.NewBoolQuery().
				Should(elastic.NewRangeQuery("changes.spendable").Gt(0)).
				Should(elastic.NewRangeQuery("changes.stakable").Gt(0)).
				Should(elastic.NewRangeQuery("changes.voting_weight").Gt(0))),
			)
			break
		}
	}
	service := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n))
	service.Query(query)
	sort(service, s, &defaultSort{"height", false})

	service.Size(p.Size())
	service.From(p.From())
	service.TrackTotalHits(true)

	results, err := service.Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return r.findMany(results, err)
}

func (r *addressHistoryRepository) GetAddressGroups(n network.Network, period *group.Period, count int) ([]entity.AddressGroup, error) {
	timeGroups := group.CreateTimeGroup(period, count)

	addressGroups := make([]entity.AddressGroup, 0)
	for i := range timeGroups {
		blockGroup := entity.AddressGroup{
			TimeGroup: *timeGroups[i],
			Period:    *period,
		}
		addressGroups = append(addressGroups, blockGroup)
	}

	var wg sync.WaitGroup
	wg.Add(len(addressGroups))

	for i := range addressGroups {
		go func(idx int) {
			defer wg.Done()
			spendAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("is_stake", false))
			spendAgg.SubAggregation("hash", elastic.NewCardinalityAggregation().Field("hash.keyword"))

			stakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("is_stake", true))
			stakeAgg.SubAggregation("hash", elastic.NewCardinalityAggregation().Field("hash.keyword"))

			results, err := r.elastic.Client.
				Search(elastic_cache.AddressHistoryIndex.Get(n)).
				Query(elastic.NewRangeQuery("time").From(addressGroups[idx].Start).To(addressGroups[idx].End)).
				Size(0).
				Aggregation("spend", spendAgg).
				Aggregation("stake", stakeAgg).
				Do(context.Background())
			if err != nil {
				return
			}

			if agg, found := results.Aggregations.Filter("spend"); found {
				if hash, found := agg.Cardinality("hash"); found {
					addressGroups[idx].Spend = int64(*hash.Value)
				}
			}

			if agg, found := results.Aggregations.Filter("stake"); found {
				if hash, found := agg.Cardinality("hash"); found {
					addressGroups[idx].Stake = int64(*hash.Value)
				}
			}
		}(i)
	}
	wg.Wait()

	return addressGroups, nil
}

func (r *addressHistoryRepository) GetAddressGroupsTotal(n network.Network, period *group.Period, count int) ([]entity.AddressGroupTotal, error) {
	timeGroups := group.CreateTimeGroup(period, count)

	addressGroups := make([]entity.AddressGroupTotal, 0)
	for i := range timeGroups {
		blockGroup := entity.AddressGroupTotal{
			TimeGroup: *timeGroups[i],
			Period:    *period,
		}
		addressGroups = append(addressGroups, blockGroup)
	}

	var wg sync.WaitGroup
	wg.Add(len(addressGroups))

	for i := range addressGroups {
		go func(idx int) {
			defer wg.Done()
			totalAgg := elastic.NewCardinalityAggregation().Field("hash.keyword")

			results, err := r.elastic.Client.
				Search(elastic_cache.AddressHistoryIndex.Get(n)).
				Query(elastic.NewRangeQuery("time").From(addressGroups[idx].Start).To(addressGroups[idx].End)).
				Size(0).
				Aggregation("total", totalAgg).
				Do(context.Background())
			if err != nil {
				return
			}

			if agg, found := results.Aggregations.Cardinality("total"); found {
				addressGroups[idx].Addresses = int64(*agg.Value)
			}
		}(i)
	}
	wg.Wait()

	return addressGroups, nil
}

func (r *addressHistoryRepository) GetStakingChart(n network.Network, period string, hash string) (groups []*entity.StakingGroup, err error) {
	count := 12
	now := time.Now().UTC().Truncate(time.Second)

	query := elastic.NewBoolQuery().
		Filter(elastic.NewMatchPhraseQuery("hash", hash)).
		Must(elastic.NewMatchQuery("is_stake", true))

	agg := elastic.NewFilterAggregation().Filter(query)

	for i := 0; i < count; i++ {
		g := &entity.StakingGroup{End: now}

		switch period {
		case "hourly":
			{
				if i == 0 {
					g.Start = now.Truncate(time.Hour)
				} else {
					g.End = groups[i-1].Start
					g.Start = g.End.Add(-time.Hour)
				}
				break
			}
		case "daily":
			{
				if i == 0 {
					g.Start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				} else {
					g.End = groups[i-1].Start
					g.Start = g.End.AddDate(0, 0, -1)
				}
				break
			}
		case "monthly":
			{
				if i == 0 {
					g.Start = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location())
					g.Start = g.Start.AddDate(0, 0, 1)
				} else {
					g.End = groups[i-1].Start
					g.Start = g.End.AddDate(0, -1, 0)
				}
				break
			}
		}

		changesAgg := elastic.NewNestedAggregation().Path("changes")
		changesAgg.SubAggregation("stakable", elastic.NewSumAggregation().Field("changes.stakable"))
		changesAgg.SubAggregation("spendable", elastic.NewSumAggregation().Field("changes.spendable"))
		changesAgg.SubAggregation("voting_weight", elastic.NewSumAggregation().Field("changes.voting_weight"))

		timeAgg := elastic.NewRangeAggregation().Field("time").AddRange(g.Start, g.End)
		timeAgg.SubAggregation("changes", changesAgg)

		agg.SubAggregation(fmt.Sprintf("group-%d", i), timeAgg)

		groups = append(groups, g)
	}

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
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
						if stakableValue, found := nested.Aggregations.Sum("stakable"); found {
							groups[i].Stakable = int64(*stakableValue.Value)
						}
						if spendableValue, found := nested.Aggregations.Sum("spendable"); found {
							groups[i].Spendable = int64(*spendableValue.Value)
						}
						if votingWeightValue, found := nested.Aggregations.Sum("voting_weight"); found {
							groups[i].VotingWeight = int64(*votingWeightValue.Value)
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

func (r *addressHistoryRepository) GetStakingRange(n network.Network, from, to uint64, addresses []string) (*entity.StakingBlocks, error) {
	zap.S().Infof("AddressHistory: GetStakingRange(%s, %d, %d)", n.String(), from, to)

	values := make([]interface{}, len(addresses))
	for i, v := range addresses {
		values[i] = v
	}

	query := elastic.NewBoolQuery().
		Must(elastic.NewRangeQuery("height").Gt(from).Lte(to)).
		Must(elastic.NewTermsQuery("hash.keyword", values...)).
		MustNot(elastic.NewNestedQuery("balance", elastic.NewTermQuery("balance.stakable", 0)))

	hashAgg := elastic.NewTermsAggregation().Field("hash.keyword").Size(int(to - from))
	hashAgg.SubAggregation("balance", elastic.NewNestedAggregation().Path("balance").
		SubAggregation("stakable", elastic.NewMaxAggregation().Field("balance.stakable")))

	stakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("is_stake", true)).
		Must(elastic.NewTermQuery("is_coldstake", false)))
	stakeAgg.SubAggregation("hash", hashAgg)

	coldStakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("is_stake", true)).
		Must(elastic.NewTermQuery("is_coldstake", true)))
	coldStakeAgg.SubAggregation("hash", hashAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Aggregation("stake", stakeAgg).
		Aggregation("coldStake", coldStakeAgg).
		Size(0).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	stakingBlocks := &entity.StakingBlocks{
		From: from,
		To:   to,
		//Addresses: make(map[string]entity.StakingBlocksAddress),
	}

	stakeBalance := float64(0)
	if stake, found := results.Aggregations.Filter("stake"); found {
		if hash, found := stake.Aggregations.Terms("hash"); found {
			for _, bucket := range hash.Buckets {
				stakingBlocks.BlockCount += bucket.DocCount
				if balance, found := bucket.Aggregations.Nested("balance"); found {
					if stakable, found := balance.Max("stakable"); found {
						zap.S().With(
							zap.String("hash", bucket.Key.(string)),
							zap.Float64("balance", *stakable.Value/100000000),
						).Infof("Staking balance found")

						//stakingBlocks.Addresses[bucket.Key.(string)] = entity.StakingBlocksAddress{
						//	Txs:     int(bucket.DocCount),
						//	Balance: *stakable.Value / 100000000,
						//}
						stakeBalance += *stakable.Value
					}
				}
			}
		}
	}

	coldStakeBalance := float64(0)
	if stake, found := results.Aggregations.Filter("coldStake"); found {
		if hash, found := stake.Aggregations.Terms("hash"); found {
			for _, bucket := range hash.Buckets {
				stakingBlocks.BlockCount += bucket.DocCount

				if balance, found := bucket.Aggregations.Nested("balance"); found {
					if stakable, found := balance.Max("stakable"); found {
						zap.S().With(
							zap.String("hash", bucket.Key.(string)),
							zap.Float64("balance", *stakable.Value/100000000),
						).Infof("ColdStaking balance found")

						//stakingBlocks.Addresses[bucket.Key.(string)] = entity.StakingBlocksAddress{
						//	Txs:     int(bucket.DocCount),
						//	Balance: *stakable.Value / 100000000,
						//}
						coldStakeBalance += *stakable.Value
					}
				}
			}
		}
	}

	stakingBlocks.Staking = stakeBalance / 100000000
	stakingBlocks.ColdStaking = coldStakeBalance / 100000000

	return stakingBlocks, nil
}

func (r *addressHistoryRepository) StakingRewardsForAddresses(n network.Network, addresses []string) ([]*entity.StakingReward, error) {
	values := make([]interface{}, len(addresses))
	for i, v := range addresses {
		values[i] = v
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermsQuery("hash.keyword", values...))
	query = query.Must(elastic.NewTermQuery("is_stake", "true"))

	now := time.Now().UTC().Truncate(time.Second)

	agg := elastic.NewTermsAggregation().Field("hash.keyword")
	agg.SubAggregation("last24Hours", dateGroupAgg(now.Add(-(time.Hour*24)), now))
	agg.SubAggregation("last7Days", dateGroupAgg(now.Add(-(time.Hour*24*7)), now))
	agg.SubAggregation("last30Days", dateGroupAgg(now.Add(-(time.Hour*24*30)), now))
	agg.SubAggregation("lastYear", dateGroupAgg(now.Add(-(time.Hour*24*365)), now))
	agg.SubAggregation("all", dateGroupAgg(now.AddDate(-100, 0, 0), now))

	service := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n))
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
	aggregation.SubAggregation("changes", elastic.NewNestedAggregation().Path("changes").
		SubAggregation("stakable", elastic.NewSumAggregation().Field("changes.stakable")))

	return
}

func stakingPeriodResults(bucket *elastic.AggregationBucketKeyItem, periodName string) *entity.StakingRewardPeriod {
	rewardPeriod := &entity.StakingRewardPeriod{Period: periodName}

	if period, found := bucket.Aggregations.Range(rewardPeriod.Period); found {
		aggBucket := period.Buckets[0]

		balance := int64(0)
		changes, _ := aggBucket.Aggregations.Nested("changes")
		if stakedValue, found := changes.Aggregations.Sum("stakable"); found {
			balance += int64(*stakedValue.Value)
		}

		rewardPeriod.Stakes = aggBucket.DocCount
		rewardPeriod.Balance = balance
	}

	return rewardPeriod
}

func (r *addressHistoryRepository) findOne(results *elastic.SearchResult, err error) (*explorer.AddressHistory, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrAddressHistoryNotFound
		return nil, err
	}

	var history *explorer.AddressHistory
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &history)
	if err != nil {
		return nil, err
	}

	return history, err
}

func (r *addressHistoryRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.AddressHistory, int64, error) {
	if err != nil {
		return nil, 0, err
	}

	historys := make([]*explorer.AddressHistory, 0)
	for _, hit := range results.Hits.Hits {
		var history *explorer.AddressHistory
		if err := json.Unmarshal(hit.Source, &history); err == nil {
			historys = append(historys, history)
		}
	}

	return historys, results.TotalHits(), err
}
