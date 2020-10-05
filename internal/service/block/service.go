package block

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type Service interface {
	GetBestBlock(network string) (*explorer.Block, error)
	GetBlockGroups(network string, period *group.Period, count int) (*entity.BlockGroups, error)
	GetBlock(network, hash string) (*explorer.Block, error)
	GetRawBlock(network, hash string) (*explorer.RawBlock, error)
	GetBlocks(network string, config *pagination.Config) ([]*explorer.Block, int64, error)
	GetTransactions(network string, config *pagination.Config, ignoreCoinbase, ignoreStaking bool) ([]*explorer.BlockTransaction, int64, error)
	GetTransactionsByBlockHash(network, blockHash string) ([]*explorer.BlockTransaction, error)
	GetTransactionByHash(network, hash string) (*explorer.BlockTransaction, error)
	GetRawTransactionByHash(network, hash string) (*explorer.RawBlockTransaction, error)
}

type service struct {
	blockRepo       *repository.BlockRepository
	transactionRepo *repository.BlockTransactionRepository
}

func NewBlockService(
	blockRepo *repository.BlockRepository,
	transactionRepo *repository.BlockTransactionRepository,
) Service {
	return &service{blockRepo, transactionRepo}
}

func (s *service) GetBestBlock(network string) (*explorer.Block, error) {
	return s.blockRepo.Network(network).BestBlock()
}

func (s *service) GetBlockGroups(network string, period *group.Period, count int) (*entity.BlockGroups, error) {
	timeGroups := group.CreateTimeGroup(period, count)

	blockGroups := new(entity.BlockGroups)
	for i := range timeGroups {
		blockGroup := &entity.BlockGroup{
			TimeGroup: *timeGroups[i],
			Period:    *period,
		}
		blockGroups.Items = append(blockGroups.Items, blockGroup)
	}

	err := s.blockRepo.Network(network).GetBlockGroups(blockGroups)

	return blockGroups, err
}

func (s *service) GetBlock(network, hash string) (*explorer.Block, error) {
	return s.blockRepo.Network(network).BlockByHashOrHeight(hash)
}

func (s *service) GetRawBlock(network, hash string) (*explorer.RawBlock, error) {
	return s.blockRepo.Network(network).RawBlockByHashOrHeight(hash)
}

func (s *service) GetBlocks(network string, config *pagination.Config) ([]*explorer.Block, int64, error) {
	return s.blockRepo.Network(network).Blocks(config.Ascending, config.Size, config.Page)
}

func (s *service) GetTransactions(network string, config *pagination.Config, ignoreCoinbase, ignoreStaking bool) ([]*explorer.BlockTransaction, int64, error) {
	return s.transactionRepo.Network(network).Transactions(config.Ascending, config.Size, config.Page, ignoreCoinbase, ignoreStaking)
}

func (s *service) GetTransactionsByBlockHash(network, blockHash string) ([]*explorer.BlockTransaction, error) {
	block, err := s.blockRepo.Network(network).BlockByHashOrHeight(blockHash)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.Network(network).TransactionsByBlock(block)
}

func (s *service) GetTransactionByHash(network, hash string) (*explorer.BlockTransaction, error) {
	return s.transactionRepo.Network(network).TransactionByHash(hash)
}

func (s *service) GetRawTransactionByHash(network, hash string) (*explorer.RawBlockTransaction, error) {
	return s.transactionRepo.Network(network).RawTransactionByHash(hash)
}
