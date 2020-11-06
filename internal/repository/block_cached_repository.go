package repository

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/block/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

type cachingBlockRepository struct {
	repository BlockRepository
	cache      *cache.Cache
}

func NewCachingBlockRepository(repository BlockRepository, cache *cache.Cache) BlockRepository {
	return &cachingBlockRepository{repository: repository, cache: cache}
}

func (r *cachingBlockRepository) GetBestBlock(n network.Network) (*explorer.Block, error) {
	result, err := r.cache.Get(
		fmt.Sprintf("%s.best-block", n.ToString()),
		func() (interface{}, error) {
			return r.repository.GetBestBlock(n)
		},
		cache.RefreshingExpiration,
	)
	if err != nil {
		return nil, err

	}

	return result.(*explorer.Block), err
}

func (r *cachingBlockRepository) GetBlocks(n network.Network, asc bool, size int, page int, bestBlock *explorer.Block) ([]*explorer.Block, int64, error) {
	return r.repository.GetBlocks(n, asc, size, page, bestBlock)
}

func (r *cachingBlockRepository) GetBlockGroups(n network.Network, period string, count int) ([]*entity.BlockGroup, error) {
	return r.repository.GetBlockGroups(n, period, count)
}

func (r *cachingBlockRepository) PopulateBlockGroups(n network.Network, blockGroups *entity.BlockGroups) error {
	return r.repository.PopulateBlockGroups(n, blockGroups)
}

func (r *cachingBlockRepository) GetBlockByHashOrHeight(n network.Network, hash string) (*explorer.Block, error) {
	return r.repository.GetBlockByHashOrHeight(n, hash)
}

func (r *cachingBlockRepository) GetBlockByHash(n network.Network, hash string) (*explorer.Block, error) {
	return r.repository.GetBlockByHash(n, hash)
}

func (r *cachingBlockRepository) GetBlockByHeight(n network.Network, height uint64) (*explorer.Block, error) {
	return r.repository.GetBlockByHeight(n, height)
}

func (r *cachingBlockRepository) GetRawBlockByHashOrHeight(n network.Network, hash string) (*explorer.RawBlock, error) {
	return r.repository.GetRawBlockByHashOrHeight(n, hash)
}

func (r *cachingBlockRepository) GetFeesForLastBlocks(n network.Network, blocks int) (fees float64, err error) {
	return r.repository.GetFeesForLastBlocks(n, blocks)
}
