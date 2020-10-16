package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"time"
)

type AddressHistoryRepository interface {
	GetLatestByHash(n network.Network, hash string) (*explorer.AddressHistory, error)
	GetFirstByHash(n network.Network, hash string) (*explorer.AddressHistory, error)
	GetCountByHash(n network.Network, hash string) (int64, error)
	GetStakingSummary(n network.Network, hash string) (count, staking, spending, voting int64, err error)
	GetSpendSummary(n network.Network, hash string) (spendingReceive, spendingSent, stakingReceive, stakingSent, votingReceive, votingSent int64, err error)
	GetHistoryByHash(n network.Network, hash, txType string, dir bool, size, page int) ([]*explorer.AddressHistory, int64, error)
	GetAddressGroups(n network.Network, period *group.Period, count int) ([]entity.AddressGroup, error)
	GetStakingChart(n network.Network, period, hash string) (groups []*entity.StakingGroup, err error)
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
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Sort("height", false).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *addressHistoryRepository) GetFirstByHash(n network.Network, hash string) (*explorer.AddressHistory, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Sort("height", true).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *addressHistoryRepository) GetCountByHash(n network.Network, hash string) (int64, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

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

func (r *addressHistoryRepository) GetStakingSummary(n network.Network, hash string) (count, staking, spending, voting int64, err error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	changeAgg := elastic.NewNestedAggregation().Path("changes")
	changeAgg.SubAggregation("staking", elastic.NewSumAggregation().Field("changes.staking"))
	changeAgg.SubAggregation("spending", elastic.NewSumAggregation().Field("changes.spending"))
	changeAgg.SubAggregation("voting", elastic.NewSumAggregation().Field("changes.voting"))

	stakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("is_stake", true))
	stakeAgg.SubAggregation("changes", changeAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Aggregation("stake", stakeAgg).
		Sort("height", false).
		Size(0).
		Do(context.Background())

	if err == nil && results != nil {
		if agg, found := results.Aggregations.Filter("stake"); found {
			count = agg.DocCount
			if changes, found := agg.Nested("changes"); found {
				if stakingValue, found := changes.Sum("staking"); found {
					staking = int64(*stakingValue.Value)
				}
				if spendingValue, found := changes.Sum("spending"); found {
					spending = int64(*spendingValue.Value)
				}
				if votingValue, found := changes.Sum("voting"); found {
					voting = int64(*votingValue.Value)
				}
			}
		}
	}

	return
}

func (r *addressHistoryRepository) GetSpendSummary(n network.Network, hash string) (spendingReceive, spendingSent, stakingReceive, stakingSent, votingReceive, votingSent int64, err error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))
	query = query.Must(elastic.NewTermQuery("is_stake", false))

	spendingReceiveAgg := elastic.NewRangeAggregation().Field("changes.spending").Gt(0)
	spendingReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.spending"))

	spendingSentAgg := elastic.NewRangeAggregation().Field("changes.spending").Lt(0)
	spendingSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.spending"))

	stakingReceiveAgg := elastic.NewRangeAggregation().Field("changes.staking").Gt(0)
	stakingReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.staking"))

	stakingSentAgg := elastic.NewRangeAggregation().Field("changes.staking").Lt(0)
	stakingSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.staking"))

	votingReceiveAgg := elastic.NewRangeAggregation().Field("changes.voting").Gt(0)
	votingReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.voting"))

	votingSentAgg := elastic.NewRangeAggregation().Field("changes.voting").Lt(0)
	votingSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.voting"))

	changeAgg := elastic.NewNestedAggregation().Path("changes")
	changeAgg.SubAggregation("spendingReceive", spendingReceiveAgg)
	changeAgg.SubAggregation("spendingSent", spendingSentAgg)

	changeAgg.SubAggregation("stakingSent", stakingSentAgg)
	changeAgg.SubAggregation("votingSent", votingSentAgg)
	changeAgg.SubAggregation("stakingReceive", stakingReceiveAgg)
	changeAgg.SubAggregation("votingReceive", votingReceiveAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Aggregation("changes", changeAgg).
		Sort("height", false).
		Size(0).
		Do(context.Background())

	if err == nil && results != nil {
		if changes, found := results.Aggregations.Nested("changes"); found {
			if spendingReceiveResult, found := changes.Range("spendingReceive"); found {
				if spendingReceiveSum, found := spendingReceiveResult.Buckets[0].Sum("sum"); found {
					spendingReceive = int64(*spendingReceiveSum.Value)
				}
			}

			if spendingSentResult, found := changes.Range("spendingSent"); found {
				if spendingSentSum, found := spendingSentResult.Buckets[0].Sum("sum"); found {
					spendingSent = int64(*spendingSentSum.Value)
				}
			}

			if stakingReceiveResult, found := changes.Range("stakingReceive"); found {
				if stakingReceiveSum, found := stakingReceiveResult.Buckets[0].Sum("sum"); found {
					stakingReceive = int64(*stakingReceiveSum.Value)
				}
			}

			if stakingSentResult, found := changes.Range("stakingSent"); found {
				if stakingSentSum, found := stakingSentResult.Buckets[0].Sum("sum"); found {
					stakingSent = int64(*stakingSentSum.Value)
				}
			}

			if votingReceiveResult, found := changes.Range("votingReceive"); found {
				if votingReceiveSum, found := votingReceiveResult.Buckets[0].Sum("sum"); found {
					votingReceive = int64(*votingReceiveSum.Value)
				}
			}

			if votingSentResult, found := changes.Range("votingSent"); found {
				if votingSentSum, found := votingSentResult.Buckets[0].Sum("sum"); found {
					votingSent = int64(*votingSentSum.Value)
				}
			}
		}
	}

	return
}

func (r *addressHistoryRepository) GetHistoryByHash(n network.Network, hash, txType string, dir bool, size, page int) ([]*explorer.AddressHistory, int64, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	switch txType {
	case "stake":
		{
			query.Must(elastic.NewTermQuery("is_stake", true))
		}
	case "send":
		{
			query.Must(elastic.NewTermQuery("is_stake", false))
			query.Filter(elastic.NewNestedQuery("changes", elastic.NewBoolQuery().
				Should(elastic.NewRangeQuery("changes.spending").Lt(0)).
				Should(elastic.NewRangeQuery("changes.staking").Lt(0)).
				Should(elastic.NewRangeQuery("changes.voting").Lt(0))),
			)
			break
		}
	case "receive":
		{
			query.Must(elastic.NewTermQuery("is_stake", false))
			query.Filter(elastic.NewNestedQuery("changes", elastic.NewBoolQuery().
				Should(elastic.NewRangeQuery("changes.spending").Gt(0)).
				Should(elastic.NewRangeQuery("changes.staking").Gt(0)).
				Should(elastic.NewRangeQuery("changes.voting").Gt(0))),
			)
			break
		}
	}

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).
		Query(query).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		TrackTotalHits(true).
		Do(context.Background())

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

	service := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get(n)).Size(0)

	for i, item := range addressGroups {
		hashAgg := elastic.NewCardinalityAggregation().Field("hash.keyword")

		spendAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("is_stake", false))
		spendAgg.SubAggregation("hash", elastic.NewCardinalityAggregation().Field("hash.keyword"))

		stakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("is_stake", true))
		stakeAgg.SubAggregation("hash", elastic.NewCardinalityAggregation().Field("hash.keyword"))

		agg := elastic.NewRangeAggregation().Field("time").AddRange(item.Start, item.End)
		agg.SubAggregation("hash", hashAgg)
		agg.SubAggregation("spend", spendAgg)
		agg.SubAggregation("stake", stakeAgg)

		service.Aggregation(string(rune(i)), agg)
	}

	results, err := service.Do(context.Background())
	if err != nil {
		return nil, err
	}

	for i := range addressGroups {
		if agg, found := results.Aggregations.Range(string(rune(i))); found {

			if hash, found := agg.Buckets[0].Cardinality("hash"); found {
				addressGroups[i].Addresses = int64(*hash.Value)
			}

			if spend, found := agg.Buckets[0].Filter("spend"); found {
				if hash, found := spend.Cardinality("hash"); found {
					addressGroups[i].Spend = int64(*hash.Value)
				}
			}

			if spend, found := agg.Buckets[0].Filter("stake"); found {
				if hash, found := spend.Cardinality("hash"); found {
					addressGroups[i].Stake = int64(*hash.Value)
				}
			}
		}
	}

	return addressGroups, nil
}

func (r *addressHistoryRepository) GetStakingChart(n network.Network, period string, hash string) (groups []*entity.StakingGroup, err error) {
	count := 12
	now := time.Now().UTC().Truncate(time.Second)

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("hash", hash))
	query = query.Must(elastic.NewMatchQuery("is_stake", true))

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
		changesAgg.SubAggregation("staking", elastic.NewSumAggregation().Field("changes.staking"))
		changesAgg.SubAggregation("spending", elastic.NewSumAggregation().Field("changes.spending"))
		changesAgg.SubAggregation("voting", elastic.NewSumAggregation().Field("changes.voting"))

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
