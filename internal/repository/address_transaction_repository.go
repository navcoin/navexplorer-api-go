package repository

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type AddressTransactionRepository struct {
	elastic *elastic_cache.Index
}

func NewAddressTransactionRepository(elastic *elastic_cache.Index) *AddressTransactionRepository {
	return &AddressTransactionRepository{elastic}
}

func (r *AddressTransactionRepository) TransactionsByHash(hash string, cold bool, dir bool, size int, page int) ([]*explorer.AddressTransaction, int64, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))
	query = query.Must(elastic.NewTermQuery("cold", cold))

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

func (r *AddressTransactionRepository) GetStakingReport(hash string, stakingReport []*entity.StakingReport) error {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))
	query = query.Must(elastic.NewTermQuery("type.keyword", "stake"))

	service := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).Query(query).Size(0)

	for i := range stakingReport {
		agg := elastic.NewRangeAggregation().Field("time").AddRange(stakingReport[i].Start, stakingReport[i].End)
		agg.SubAggregation("amount", elastic.NewSumAggregation().Field("total"))
		service.Aggregation(string(i), agg)
	}

	results, err := service.Do(context.Background())
	if err != nil {
		return err
	}

	for i := range stakingReport {
		if agg, found := results.Aggregations.Range(string(i)); found {
			bucket := agg.Buckets[0]
			stakingReport[i].Stakes = uint(bucket.DocCount)
			if amount, found := bucket.Aggregations.Sum("amount"); found {
				stakingReport[i].Amount = uint64(*amount.Value)
			}
		}
	}

	return nil
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
