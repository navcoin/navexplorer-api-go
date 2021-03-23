package block

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/block/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

type Service interface {
	GetBestBlock(n network.Network) (*explorer.Block, error)
	GetBlockGroups(n network.Network, period *group.Period, count int) (*entity.BlockGroups, error)
	GetBlock(n network.Network, hash string) (*explorer.Block, error)
	GetRawBlock(n network.Network, hash string) (*explorer.RawBlock, error)
	GetBlocks(n network.Network, pagination framework.Pagination) ([]*explorer.Block, int64, error)
	GetTransactions(n network.Network, request framework.RestRequest) ([]*explorer.BlockTransaction, int64, error)
	GetTransactionsByBlockHash(n network.Network, blockHash string) ([]*explorer.BlockTransaction, error)
	GetTransactionByHash(n network.Network, hash string) (*explorer.BlockTransaction, error)
	GetRawTransactionByHash(n network.Network, hash string) (*explorer.RawBlockTransaction, error)
}

type service struct {
	blockRepo       repository.BlockRepository
	transactionRepo repository.BlockTransactionRepository
}

func NewBlockService(blockRepo repository.BlockRepository, transactionRepo repository.BlockTransactionRepository) Service {
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

func (s *service) GetBlocks(n network.Network, pagination framework.Pagination) ([]*explorer.Block, int64, error) {
	bestBlock, err := s.blockRepo.GetBestBlock(n)
	if err != nil {
		return nil, 0, err
	}

	return s.blockRepo.GetBlocks(n, false, pagination.Size(), pagination.Page(), bestBlock)
}

func (s *service) GetTransactions(n network.Network, request framework.RestRequest) ([]*explorer.BlockTransaction, int64, error) {
	return s.transactionRepo.GetTransactions(n, request.Pagination(), request.Sort(), request.Filter())

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
