package block

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type Service interface {
	GetBestBlock() (*explorer.Block, error)
	GetBlockGroups(period *group.Period, count int) (*entity.BlockGroups, error)
	GetBlock(hash string) (*explorer.Block, error)
	GetRawBlock(hash string) (*explorer.RawBlock, error)
	GetBlocks(config *pagination.Config) ([]*explorer.Block, int64, error)
	GetTransactions(blockHash string) ([]*explorer.BlockTransaction, error)
	GetTransactionByHash(hash string) (*explorer.BlockTransaction, error)
	GetRawTransactionByHash(hash string) (*explorer.RawBlockTransaction, error)
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

func (s *service) GetBestBlock() (*explorer.Block, error) {
	return s.blockRepo.BestBlock()
}

func (s *service) GetBlockGroups(period *group.Period, count int) (*entity.BlockGroups, error) {
	timeGroups := group.CreateTimeGroup(period, count)

	blockGroups := new(entity.BlockGroups)
	for i := range timeGroups {
		blockGroup := &entity.BlockGroup{
			TimeGroup: *timeGroups[i],
			Period:    *period,
		}
		blockGroups.Items = append(blockGroups.Items, blockGroup)
	}

	err := s.blockRepo.GetBlockGroups(blockGroups)

	return blockGroups, err
}

func (s *service) GetBlock(hash string) (*explorer.Block, error) {
	return s.blockRepo.BlockByHashOrHeight(hash)
}

func (s *service) GetRawBlock(hash string) (*explorer.RawBlock, error) {
	return s.blockRepo.RawBlockByHashOrHeight(hash)
}

func (s *service) GetBlocks(config *pagination.Config) ([]*explorer.Block, int64, error) {
	return s.blockRepo.Blocks(config.Ascending, config.Size, config.Page)
}

func (s *service) GetTransactions(blockHash string) ([]*explorer.BlockTransaction, error) {
	block, err := s.blockRepo.BlockByHashOrHeight(blockHash)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.TransactionsByBlock(block)
}

func (s *service) GetTransactionByHash(hash string) (*explorer.BlockTransaction, error) {
	return s.transactionRepo.TransactionByHash(hash)
}

func (s *service) GetRawTransactionByHash(hash string) (*explorer.RawBlockTransaction, error) {
	return s.transactionRepo.RawTransactionByHash(hash)
}
