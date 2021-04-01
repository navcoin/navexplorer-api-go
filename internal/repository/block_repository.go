package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/block/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"strconv"
)

type BlockRepository interface {
	GetBestBlock(n network.Network) (*explorer.Block, error)
	GetBlocks(n network.Network, p framework.Pagination, s framework.Sort, f framework.Filters, bestBlock *explorer.Block) ([]*explorer.Block, int64, error)
	GetBlockGroups(n network.Network, period string, count int) ([]*entity.BlockGroup, error)
	PopulateBlockGroups(n network.Network, blockGroups *entity.BlockGroups) error
	GetBlockByHashOrHeight(n network.Network, hash string) (*explorer.Block, error)
	GetBlockByHash(n network.Network, hash string) (*explorer.Block, error)
	GetBlockByHeight(n network.Network, height uint64) (*explorer.Block, error)
	GetRawBlockByHashOrHeight(n network.Network, hash string) (*explorer.RawBlock, error)
	GetFeesForLastBlocks(n network.Network, blocks int) (fees float64, err error)
}

var (
	ErrBlockNotFound = errors.New("Block not found")
)

type blockRepository struct {
	elastic *elastic_cache.Index
}

func NewBlockRepository(elastic *elastic_cache.Index) BlockRepository {
	return &blockRepository{elastic: elastic}
}

func (r *blockRepository) GetBestBlock(n network.Network) (*explorer.Block, error) {
	results, err := r.elastic.Client.Search().Index(elastic_cache.BlockIndex.Get(n)).
		Sort("height", false).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *blockRepository) GetBlocks(n network.Network, p framework.Pagination, s framework.Sort, f framework.Filters, bestBlock *explorer.Block) ([]*explorer.Block, int64, error) {
	service := r.elastic.Client.Search(elastic_cache.BlockIndex.Get(n))
	sort(service, s, &defaultSort{"height", false})

	//from := int(bestBlock.Height+1) - ((p.Page() - 1) * p.Size())
	//if from <= 0 {
	//	from = p.Size()
	//}

	service.Size(p.Size())
	service.From(p.From())
	service.TrackTotalHits(true)

	results, err := service.Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	var blocks = make([]*explorer.Block, 0)
	for _, hit := range results.Hits.Hits {
		var block *explorer.Block
		if err := json.Unmarshal(hit.Source, &block); err == nil {
			block.Best = block.Height == bestBlock.Height
			if bestBlock.Height >= block.Height {
				block.Confirmations = bestBlock.Height - block.Height
			}

			blocks = append(blocks, block)
		}
	}

	return blocks, results.TotalHits(), err
}

func (r *blockRepository) GetBlockGroups(n network.Network, period string, count int) ([]*entity.BlockGroup, error) {
	service := r.elastic.Client.Search(elastic_cache.BlockIndex.Get(n)).Size(0)

	timeGroups := group.CreateTimeGroup(group.GetPeriod(period), count)
	for i := range timeGroups {
		agg := elastic.NewRangeAggregation().Field("created").AddRange(timeGroups[i].Start, timeGroups[i].End)
		agg.SubAggregation("stake", elastic.NewSumAggregation().Field("stake"))
		agg.SubAggregation("fees", elastic.NewSumAggregation().Field("fees"))
		agg.SubAggregation("spend", elastic.NewSumAggregation().Field("spend"))
		agg.SubAggregation("transactions", elastic.NewSumAggregation().Field("transactions"))
		agg.SubAggregation("height", elastic.NewMaxAggregation().Field("height"))

		service.Aggregation(string(rune(i)), agg)
	}

	results, err := service.Do(context.Background())
	if err != nil || results == nil {
		return nil, err
	}

	blockGroups := make([]*entity.BlockGroup, 0)
	for i := range timeGroups {
		blockGroup := &entity.BlockGroup{TimeGroup: *timeGroups[i], Period: *group.GetPeriod(period)}

		if agg, found := results.Aggregations.Range(string(rune(i))); found {
			blockGroup.Blocks = agg.Buckets[0].DocCount

			if stake, found := agg.Buckets[0].Aggregations.Sum("stake"); found {
				blockGroup.Stake = int64(*stake.Value)
			}
			if fees, found := agg.Buckets[0].Aggregations.Sum("fees"); found {
				blockGroup.Fees = int64(*fees.Value)
			}

			if spend, found := agg.Buckets[0].Aggregations.Sum("spend"); found {
				blockGroup.Spend = int64(*spend.Value)
			}

			if transactions, found := agg.Buckets[0].Aggregations.Sum("transactions"); found {
				blockGroup.Transactions = int64(*transactions.Value)
			}

			if height, found := agg.Buckets[0].Aggregations.Max("height"); found {
				if height.Value != nil {
					blockGroup.Height = int64(*height.Value)
				}
			}
			blockGroups = append(blockGroups, blockGroup)
		}
	}

	return blockGroups, err
}

func (r *blockRepository) PopulateBlockGroups(n network.Network, blockGroups *entity.BlockGroups) error {
	service := r.elastic.Client.Search(elastic_cache.BlockIndex.Get(n)).Size(0)

	for i, item := range blockGroups.Items {
		agg := elastic.NewRangeAggregation().Field("time").AddRange(item.Start, item.End)
		agg.SubAggregation("stake", elastic.NewSumAggregation().Field("stake"))
		agg.SubAggregation("fees", elastic.NewSumAggregation().Field("fees"))
		agg.SubAggregation("spend", elastic.NewSumAggregation().Field("spend"))
		agg.SubAggregation("tx", elastic.NewSumAggregation().Field("tx_count"))
		agg.SubAggregation("height", elastic.NewMaxAggregation().Field("height"))

		service.Aggregation(string(rune(i)), agg)
	}

	results, err := service.Do(context.Background())
	if err != nil {
		return err
	}

	for i, item := range blockGroups.Items {
		if agg, found := results.Aggregations.Range(string(rune(i))); found {
			bucket := agg.Buckets[0]
			item.Blocks = bucket.DocCount
			if stake, found := bucket.Aggregations.Sum("stake"); found {
				item.Stake = int64(*stake.Value)
			}
			if fees, found := bucket.Aggregations.Sum("fees"); found {
				item.Fees = int64(*fees.Value)
			}

			if spend, found := bucket.Aggregations.Sum("spend"); found {
				item.Spend = int64(*spend.Value)
			}

			if transactions, found := bucket.Aggregations.Sum("tx"); found {
				item.Transactions = int64(*transactions.Value)
			}

			if height, found := bucket.Aggregations.Max("height"); found {
				if height.Value != nil {
					item.Height = int64(*height.Value)
				}
			}
		}
	}

	return nil
}

func (r *blockRepository) GetBlockByHashOrHeight(n network.Network, hash string) (*explorer.Block, error) {
	block, err := r.GetBlockByHash(n, hash)
	if err != nil {
		height, _ := strconv.Atoi(hash)
		block, err = r.GetBlockByHeight(n, uint64(height))
	}

	if err != nil {
		return nil, err
	}

	bestBlock, err := r.GetBestBlock(n)
	if err != nil {
		return nil, err
	}

	nextBlock, _ := r.GetBlockByHeight(n, block.Height+1)
	if nextBlock != nil {
		block.Nextblockhash = nextBlock.Hash
	}

	block.Best = block.Height == bestBlock.Height
	block.Confirmations = bestBlock.Height - block.Height

	return block, err
}

func (r *blockRepository) GetBlockByHash(n network.Network, hash string) (*explorer.Block, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockIndex.Get(n)).
		Query(elastic.NewTermQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *blockRepository) GetBlockByHeight(n network.Network, height uint64) (*explorer.Block, error) {
	results, err := r.elastic.Client.Search(elastic_cache.BlockIndex.Get(n)).
		Query(elastic.NewTermQuery("height", height)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *blockRepository) GetRawBlockByHashOrHeight(n network.Network, hash string) (*explorer.RawBlock, error) {
	block, err := r.GetBlockByHashOrHeight(n, hash)
	if err != nil {
		return nil, err
	}

	blockJson, _ := json.Marshal(block)
	rawBlock := new(explorer.RawBlock)
	err = json.Unmarshal(blockJson, rawBlock)

	return rawBlock, err
}

func (r *blockRepository) GetFeesForLastBlocks(n network.Network, blocks int) (fees float64, err error) {
	bestBlock, err := r.GetBestBlock(n)
	if err != nil {
		return
	}

	results, err := r.elastic.Client.Search(elastic_cache.BlockIndex.Get(n)).
		Query(elastic.NewRangeQuery("height").Gt(bestBlock.Height-uint64(blocks))).
		Aggregation("fees", elastic.NewSumAggregation().Field("fees")).
		Size(0).
		Do(context.Background())
	if err != nil {
		return 0, err
	}

	if feesValue, found := results.Aggregations.Sum("fees"); found {
		fees = *feesValue.Value / 100000000
	}

	return
}

func (r *blockRepository) findOne(results *elastic.SearchResult, err error) (*explorer.Block, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrBlockNotFound
		return nil, err
	}

	var block explorer.Block
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &block)

	return &block, err
}
