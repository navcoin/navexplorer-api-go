package repository

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
)

type AddressTransactionRepository struct {
	elastic *elastic_cache.Index
}

func NewAddressTransactionRepository(elastic *elastic_cache.Index) *AddressTransactionRepository {
	return &AddressTransactionRepository{elastic}
}

func (r *AddressTransactionRepository) TransactionsByHash(hash string, cold bool, dir bool, size int, page int) ([]*explorer.AddressTransaction, int, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("hash.keyword", hash))
	query = query.Must(elastic.NewTermQuery("cold", cold))

	results, err := r.elastic.Client.Search(elastic_cache.AddressTransactionIndex.Get()).
		Query(query).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())

	return r.findMany(results, err)
}

func (r *AddressTransactionRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.AddressTransaction, int, error) {
	if err != nil {
		return nil, 0, err
	}

	txs := make([]*explorer.AddressTransaction, 0)
	for _, hit := range results.Hits.Hits {
		var tx *explorer.AddressTransaction
		if err := json.Unmarshal(hit.Source, &tx); err == nil {
			txs = append(txs, tx)
		}
	}

	return txs, int(results.Hits.TotalHits.Value), err
}
