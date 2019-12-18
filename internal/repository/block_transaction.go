package repository

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
)

type BlockTransactionRepository struct {
	elastic *elastic_cache.Index
}

func NewBlockTransactionRepository(elastic *elastic_cache.Index) *BlockTransactionRepository {
	return &BlockTransactionRepository{elastic}
}

func (r *BlockTransactionRepository) TransactionsByBlock(block *explorer.Block) ([]*explorer.BlockTransaction, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockTransactionIndex.Get()).
		Query(elastic.NewMatchPhraseQuery("blockhash", block.Hash)).
		Size(10000).
		Do(context.Background())

	return r.findMany(results, err)
}

func (r *BlockTransactionRepository) TransactionByHash(hash string) (*explorer.BlockTransaction, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockTransactionIndex.Get()).
		Query(elastic.NewTermQuery("hash", hash)).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *BlockTransactionRepository) RawTransactionByHash(hash string) (*explorer.RawBlockTransaction, error) {
	tx, err := r.TransactionByHash(hash)
	if err != nil {
		return nil, err
	}

	txJson, _ := json.Marshal(tx)
	rawTx := new(explorer.RawBlockTransaction)
	err = json.Unmarshal(txJson, rawTx)

	return rawTx, err
}

func (r *BlockTransactionRepository) findOne(results *elastic.SearchResult, err error) (*explorer.BlockTransaction, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrBlockNotFound
		return nil, err
	}

	var tx explorer.BlockTransaction
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &tx)

	return &tx, err
}

func (r *BlockTransactionRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.BlockTransaction, error) {
	if err != nil || results.Hits.TotalHits.Value == 0 {
		return nil, err
	}

	transactions := make([]*explorer.BlockTransaction, 0)
	for _, hit := range results.Hits.Hits {
		var transaction explorer.BlockTransaction
		err = json.Unmarshal(hit.Source, &transaction)
		if err == nil {
			transactions = append(transactions, &transaction)
		}
	}

	return transactions, nil
}
