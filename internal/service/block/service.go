package block

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

type Service interface {
	GetBestBlock(n network.Network) (*explorer.Block, error)
	GetBlockGroups(n network.Network, period *group.Period, count int) (*entity.BlockGroups, error)
	GetBlock(n network.Network, hash string) (*explorer.Block, error)
	GetRawBlock(n network.Network, hash string) (*explorer.RawBlock, error)
	GetBlocks(n network.Network, config *pagination.Config) ([]*explorer.Block, int64, error)
	GetTransactions(n network.Network, config *pagination.Config, ignoreCoinbase, ignoreStaking bool) ([]*explorer.BlockTransaction, int64, error)
	GetTransactionsByBlockHash(n network.Network, blockHash string) ([]*explorer.BlockTransaction, error)
	GetTransactionByHash(n network.Network, hash string) (*explorer.BlockTransaction, error)
	GetRawTransactionByHash(n network.Network, hash string) (*explorer.RawBlockTransaction, error)
}

type service struct {
	blockRepo       repository.BlockRepository
	transactionRepo repository.BlockTransactionRepository
}

func NewBlockService(
	blockRepo repository.BlockRepository,
	transactionRepo repository.BlockTransactionRepository,
) Service {
	return &service{blockRepo, transactionRepo}
}

func (s *service) GetBestBlock(n network.Network) (*explorer.Block, error) {
	return s.blockRepo.GetBestBlock(n)
}

func (s *service) GetBlockGroups(n network.Network, period *group.Period, count int) (*entity.BlockGroups, error) {
	timeGroups := group.CreateTimeGroup(period, count)

	blockGroups := new(entity.BlockGroups)
	for i := range timeGroups {
		blockGroup := &entity.BlockGroup{
			TimeGroup: *timeGroups[i],
			Period:    *period,
		}
		blockGroups.Items = append(blockGroups.Items, blockGroup)
	}

	err := s.blockRepo.PopulateBlockGroups(n, blockGroups)

	return blockGroups, err
}

func (s *service) GetBlock(n network.Network, hash string) (*explorer.Block, error) {
	return s.blockRepo.GetBlockByHashOrHeight(n, hash)
}

func (s *service) GetRawBlock(n network.Network, hash string) (*explorer.RawBlock, error) {
	return s.blockRepo.GetRawBlockByHashOrHeight(n, hash)
}

func (s *service) GetBlocks(n network.Network, config *pagination.Config) ([]*explorer.Block, int64, error) {
	return s.blockRepo.GetBlocks(n, config.Ascending, config.Size, config.Page)
}

func (s *service) GetTransactions(n network.Network, config *pagination.Config, ignoreCoinbase, ignoreStaking bool) ([]*explorer.BlockTransaction, int64, error) {
	return s.transactionRepo.GetTransactions(n, config.Ascending, config.Size, config.Page, ignoreCoinbase, ignoreStaking)
}

func (s *service) GetTransactionsByBlockHash(n network.Network, blockHash string) ([]*explorer.BlockTransaction, error) {
	block, err := s.blockRepo.GetBlockByHashOrHeight(n, blockHash)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.GetTransactionsByBlock(n, block)
}

func (s *service) GetTransactionByHash(n network.Network, hash string) (*explorer.BlockTransaction, error) {
	return s.transactionRepo.GetTransactionByHash(n, hash)
}

func (s *service) GetRawTransactionByHash(n network.Network, hash string) (*explorer.RawBlockTransaction, error) {
	return s.transactionRepo.GetRawTransactionByHash(n, hash)
}
