package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block_group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"strconv"
)

var (
	ErrBlockNotFound = errors.New("Block not found")
)

type BlockRepository struct {
	elastic *elastic_cache.Index
}

func NewBlockRepository(elastic *elastic_cache.Index) *BlockRepository {
	return &BlockRepository{elastic}
}

func (r *BlockRepository) BestBlock() (*explorer.Block, error) {
	results, err := r.elastic.Client.Search().Index(elastic_cache.BlockIndex.Get()).
		Sort("height", false).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *BlockRepository) Blocks(asc bool, size int, page int) ([]*explorer.Block, int, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockIndex.Get()).
		Sort("height", asc).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	bestBlock, err := r.BestBlock()
	if err != nil {
		return nil, 0, err
	}

	var blocks = make([]*explorer.Block, 0)
	for _, hit := range results.Hits.Hits {
		var block *explorer.Block
		if err := json.Unmarshal(hit.Source, &block); err == nil {
			block.Best = block.Height == bestBlock.Height
			block.Confirmations = bestBlock.Height - block.Height + 1

			blocks = append(blocks, block)
		}
	}

	return blocks, int(results.Hits.TotalHits.Value), err
}

func (r *BlockRepository) BlockGroups(period string, count int) ([]*block_group.BlockGroup, error) {
	groups := block_group.CreateGroups(period, count)

	service := r.elastic.Client.Search(elastic_cache.BlockIndex.Get()).Size(0)

	for idx, group := range groups {
		agg := elastic.NewRangeAggregation().Field("created").AddRange(group.Start, group.End)
		agg.SubAggregation("stake", elastic.NewSumAggregation().Field("stake"))
		agg.SubAggregation("fees", elastic.NewSumAggregation().Field("fees"))
		agg.SubAggregation("spend", elastic.NewSumAggregation().Field("spend"))
		agg.SubAggregation("transactions", elastic.NewSumAggregation().Field("transactions"))
		agg.SubAggregation("height", elastic.NewMaxAggregation().Field("height"))

		service.Aggregation(string(idx), agg)
	}

	results, err := service.Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	for idx, _ := range groups {
		if agg, found := results.Aggregations.Range(string(idx)); found {
			groups[idx].Blocks = agg.Buckets[0].DocCount

			if stake, found := agg.Buckets[0].Aggregations.Sum("stake"); found {
				groups[idx].Stake = int64(*stake.Value)
			}
			if fees, found := agg.Buckets[0].Aggregations.Sum("fees"); found {
				groups[idx].Fees = int64(*fees.Value)
			}

			if spend, found := agg.Buckets[0].Aggregations.Sum("spend"); found {
				groups[idx].Spend = int64(*spend.Value)
			}

			if transactions, found := agg.Buckets[0].Aggregations.Sum("transactions"); found {
				groups[idx].Transactions = int64(*transactions.Value)
			}

			if height, found := agg.Buckets[0].Aggregations.Max("height"); found {
				if height.Value != nil {
					groups[idx].Height = int64(*height.Value)
				}
			}
		}
	}

	return groups, err
}

func (r *BlockRepository) BlockByHashOrHeight(hash string) (*explorer.Block, error) {
	block, err := r.BlockByHash(hash)
	if err != nil {
		height, _ := strconv.Atoi(hash)
		block, err = r.BlockByHeight(uint64(height))
	}

	if err != nil {
		return nil, err
	}

	bestBlock, err := r.BestBlock()
	if err != nil {
		return nil, err
	}

	block.Best = block.Height == bestBlock.Height
	block.Confirmations = bestBlock.Height - block.Height + 1

	return block, err
}

func (r *BlockRepository) BlockByHash(hash string) (*explorer.Block, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockIndex.Get()).
		Query(elastic.NewTermQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *BlockRepository) BlockByHeight(height uint64) (*explorer.Block, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockIndex.Get()).
		Query(elastic.NewTermQuery("height", height)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *BlockRepository) RawBlockByHashOrHeight(hash string) (*explorer.RawBlock, error) {
	block, err := r.BlockByHashOrHeight(hash)
	if err != nil {
		return nil, err
	}

	blockJson, _ := json.Marshal(block)
	rawBlock := new(explorer.RawBlock)
	err = json.Unmarshal(blockJson, rawBlock)

	return rawBlock, err
}

func (r *BlockRepository) findOne(results *elastic.SearchResult, err error) (*explorer.Block, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrBlockNotFound
		return nil, err
	}

	var block explorer.Block
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &block)

	return &block, err
}
