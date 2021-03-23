package repository

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type BlockTransactionRepository interface {
	GetTransactions(n network.Network, p framework.Pagination, s framework.Sort, f framework.Filter) ([]*explorer.BlockTransaction, int64, error)
	GetTransactionsByBlock(n network.Network, block *explorer.Block) ([]*explorer.BlockTransaction, error)
	GetTransactionByHash(n network.Network, hash string) (*explorer.BlockTransaction, error)
	GetRawTransactionByHash(n network.Network, hash string) (*explorer.RawBlockTransaction, error)
	GetAssociatedStakingAddresses(n network.Network, address string) ([]string, error)
}

type blockTransactionRepository struct {
	elastic *elastic_cache.Index
}

func NewBlockTransactionRepository(elastic *elastic_cache.Index) BlockTransactionRepository {
	return &blockTransactionRepository{elastic: elastic}
}

func isSupportedFilter() []string {
	return []string{"include_coinbase_txs", "include_staking_txs"}
}

func (r *blockTransactionRepository) GetTransactions(n network.Network, p framework.Pagination, s framework.Sort, f framework.Filter) ([]*explorer.BlockTransaction, int64, error) {
	query := elastic.NewBoolQuery()

	options := f.OnlySupportedOptions([]string{"include_coinbase_txs", "include_staking_txs"})

	if option, err := options.GetAsBool("include_coinbase_txs"); err == nil && option == false {
		query = query.MustNot(elastic.NewTermQuery("type", "coinbase"))
	}

	if option, err := options.GetAsBool("include_staking_txs"); err == nil && option == false {
		query = query.MustNot(elastic.NewTermsQuery("type", "staking", "cold_staking", "cold_staking_v2"))
	}

	service := r.elastic.Client.Search(elastic_cache.BlockTransactionIndex.Get(n))
	service.Query(query)
	sort(service, s)

	service.Size(p.Size())
	service.TrackTotalHits(true)

	results, err := service.Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	var txs = make([]*explorer.BlockTransaction, 0)
	for _, hit := range results.Hits.Hits {
		var tx *explorer.BlockTransaction
		if err := json.Unmarshal(hit.Source, &tx); err == nil {
			txs = append(txs, tx)
		}
	}

	return txs, results.TotalHits(), err
}

func (r *blockTransactionRepository) GetTransactionsByBlock(n network.Network, block *explorer.Block) ([]*explorer.BlockTransaction, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockTransactionIndex.Get(n)).
		Query(elastic.NewMatchPhraseQuery("blockhash", block.Hash)).
		Sort("index", true).
		Size(10000).
		Do(context.Background())

	return r.findMany(results, err)
}

func (r *blockTransactionRepository) GetTransactionByHash(n network.Network, hash string) (*explorer.BlockTransaction, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockTransactionIndex.Get(n)).
		Query(elastic.NewTermQuery("hash", hash)).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *blockTransactionRepository) GetRawTransactionByHash(n network.Network, hash string) (*explorer.RawBlockTransaction, error) {
	tx, err := r.GetTransactionByHash(n, hash)
	if err != nil {
		return nil, err
	}

	txJson, _ := json.Marshal(tx)
	rawTx := new(explorer.RawBlockTransaction)
	err = json.Unmarshal(txJson, rawTx)

	return rawTx, err
}

func (r *blockTransactionRepository) GetAssociatedStakingAddresses(n network.Network, address string) ([]string, error) {
	stakingAddresses := make([]string, 0)

	outputsQuery := elastic.NewBoolQuery()
	outputsQuery = outputsQuery.Must(elastic.NewMatchQuery("outputs.type", "COLD_STAKING"))
	outputsQuery = outputsQuery.Must(elastic.NewMatchQuery("outputs.addresses.keyword", address))

	query := elastic.NewNestedQuery("outputs", outputsQuery)

	results, err := r.elastic.Client.Search(elastic_cache.BlockTransactionIndex.Get(n)).
		Query(query).
		Size(50000000).
		Sort("time", false).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("Failed to get staking addresses")
		return stakingAddresses, err
	}

	for _, hit := range results.Hits.Hits {
		transaction := new(explorer.BlockTransaction)
		err := json.Unmarshal(hit.Source, &transaction)
		if err == nil {
			for _, output := range transaction.Vout {
				if len(output.ScriptPubKey.Addresses) == 2 && output.ScriptPubKey.Addresses[1] == address {
					if !contains(stakingAddresses, output.ScriptPubKey.Addresses[0]) {
						stakingAddresses = append(stakingAddresses, output.ScriptPubKey.Addresses[0])
					}
				}
			}
		}
	}

	return stakingAddresses, err
}

func (r *blockTransactionRepository) findOne(results *elastic.SearchResult, err error) (*explorer.BlockTransaction, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrBlockNotFound
		return nil, err
	}

	var tx explorer.BlockTransaction
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &tx)

	return &tx, err
}

func (r *blockTransactionRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.BlockTransaction, error) {
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

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
