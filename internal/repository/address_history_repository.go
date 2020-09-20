package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

var (
	ErrAddressHistoryNotFound = errors.New("Address history not found")
)

type AddressHistoryRepository struct {
	elastic *elastic_cache.Index
}

func NewAddressHistoryRepository(elastic *elastic_cache.Index) *AddressHistoryRepository {
	return &AddressHistoryRepository{elastic}
}

func (r *AddressHistoryRepository) LatestByHash(hash string) (*explorer.AddressHistory, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Sort("height", false).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *AddressHistoryRepository) StakingSummary(hash string) (count, staking, spending, voting uint64, err error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))

	changeAgg := elastic.NewNestedAggregation().Path("changes")
	changeAgg.SubAggregation("staking", elastic.NewSumAggregation().Field("changes.staking"))
	changeAgg.SubAggregation("spending", elastic.NewSumAggregation().Field("changes.spending"))
	changeAgg.SubAggregation("voting", elastic.NewSumAggregation().Field("changes.voting"))

	stakeAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("is_stake", true))
	stakeAgg.SubAggregation("changes", changeAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Aggregation("stake", stakeAgg).
		Sort("height", false).
		Size(0).
		Do(context.Background())

	if err == nil && results != nil {
		if agg, found := results.Aggregations.Filter("stake"); found {
			count = uint64(agg.DocCount)
			if changes, found := agg.Nested("changes"); found {
				if stakingValue, found := changes.Sum("staking"); found {
					staking = uint64(*stakingValue.Value)
				}
				if spendingValue, found := changes.Sum("spending"); found {
					spending = uint64(*spendingValue.Value)
				}
				if votingValue, found := changes.Sum("voting"); found {
					voting = uint64(*votingValue.Value)
				}
			}
		}
	}

	return
}

func (r *AddressHistoryRepository) SpendSummary(hash string) (spendingSent, spendingReceive, spending, voting uint64, err error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))
	query = query.Must(elastic.NewTermQuery("is_stake", false))

	spendingReceiveAgg := elastic.NewRangeAggregation().Field("changes.spending").Gt(0)
	spendingReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.spending"))

	spendingSentAgg := elastic.NewRangeAggregation().Field("changes.spending").Lt(0)
	spendingSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.spending"))

	stakingSentAgg := elastic.NewRangeAggregation().Field("changes.staking").Lt(0)
	stakingSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.staking"))

	votingSentAgg := elastic.NewRangeAggregation().Field("changes.voting").Lt(0)
	votingSentAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.voting"))

	stakingReceiveAgg := elastic.NewRangeAggregation().Field("changes.staking").Gt(0)
	stakingReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.staking"))

	votingReceiveAgg := elastic.NewRangeAggregation().Field("changes.voting").Gt(0)
	votingReceiveAgg.SubAggregation("sum", elastic.NewSumAggregation().Field("changes.voting"))

	changeAgg := elastic.NewNestedAggregation().Path("changes")
	changeAgg.SubAggregation("spendingReceive", spendingReceiveAgg)
	changeAgg.SubAggregation("spendingSent", spendingSentAgg)

	changeAgg.SubAggregation("stakingSent", stakingSentAgg)
	changeAgg.SubAggregation("votingSent", votingSentAgg)
	changeAgg.SubAggregation("stakingReceive", stakingReceiveAgg)
	changeAgg.SubAggregation("votingReceive", votingReceiveAgg)

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Aggregation("changes", changeAgg).
		Sort("height", false).
		Size(0).
		Do(context.Background())

	if err == nil && results != nil {
		if changes, found := results.Aggregations.Nested("changes"); found {
			spendingReceiveResult, _ := changes.Sum("spendingReceive")
			spendingReceiveSum, _ := spendingReceiveResult.Aggregations.Sum("sum")
			spendingReceive = uint64(*spendingReceiveSum.Value)

			spendingSentResult, _ := changes.Sum("spendingSent")
			spendingSentSum, _ := spendingSentResult.Aggregations.Sum("sum")
			spendingSent = uint64(*spendingSentSum.Value)

			if spendingValue, found := changes.Sum("spending"); found {
				spending = uint64(*spendingValue.Value)
			}
			if votingValue, found := changes.Sum("voting"); found {
				voting = uint64(*votingValue.Value)
			}
		}
	}

	return
}

func (r *AddressHistoryRepository) HistoryByHash(hash string, txType string, dir bool, size int, page int) ([]*explorer.AddressHistory, int64, error) {
	log.Info(txType)
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

	//if len(txTypes) != 0 {
	//	sendQuery := elastic.NewBoolQuery().
	//		Must(elastic.NewTermQuery("is_stake", false)).
	//		Must(elastic.NewBoolQuery().Filter(elastic.NewNestedQuery("changes", elastic.NewBoolQuery().
	//			Must(elastic.NewRangeQuery("changes.spending").Lt(0))),
	//		))
	//	if hasType("send", txTypes) {
	//		query.Must(sendQuery)
	//	} else {
	//		query.MustNot(sendQuery)
	//	}
	//
	//	receiveQuery := elastic.NewBoolQuery().
	//		Must(elastic.NewTermQuery("is_stake", false)).
	//		Filter(elastic.NewNestedQuery("changes", elastic.NewBoolQuery().Must(elastic.NewRangeQuery("changes.spending").Gt(0))))
	//	if hasType("receive", txTypes) {
	//		query.Should(receiveQuery)
	//	} else {
	//		query.MustNot(receiveQuery)
	//	}
	//
	//	stakeQuery := elastic.NewTermQuery("is_stake", true)
	//	if hasType("stake", txTypes) {
	//		query.Should(stakeQuery)
	//	} else {
	//		query.MustNot(stakeQuery)
	//	}
	//}

	results, err := r.elastic.Client.Search(elastic_cache.AddressHistoryIndex.Get()).
		Query(query).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		TrackTotalHits(true).
		Do(context.Background())

	return r.findMany(results, err)
}

func (r *AddressHistoryRepository) findOne(results *elastic.SearchResult, err error) (*explorer.AddressHistory, error) {
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

func (r *AddressHistoryRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.AddressHistory, int64, error) {
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

func hasType(txType string, txTypes []string) bool {
	for _, t := range txTypes {
		if t == txType {
			return true
		}
	}
	return false
}
