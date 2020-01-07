package block

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type Service struct {
	blockRepo       *repository.BlockRepository
	transactionRepo *repository.BlockTransactionRepository
}

func NewBlockService(
	blockRepo *repository.BlockRepository,
	transactionRepo *repository.BlockTransactionRepository,
) *Service {
	return &Service{blockRepo, transactionRepo}
}

func (s *Service) GetBestBlock() (*explorer.Block, error) {
	return s.blockRepo.BestBlock()
}

func (s *Service) GetBlock(hash string) (*explorer.Block, error) {
	return s.blockRepo.BlockByHashOrHeight(hash)
}

func (s *Service) GetRawBlock(hash string) (*explorer.RawBlock, error) {
	return s.blockRepo.RawBlockByHashOrHeight(hash)
}

func (s *Service) GetBlocks(config *pagination.Config) ([]*explorer.Block, int, error) {
	return s.blockRepo.Blocks(config.Dir, config.Size, config.Page)
}

func (s *Service) GetTransactions(blockHash string) ([]*explorer.BlockTransaction, error) {
	block, err := s.blockRepo.BlockByHashOrHeight(blockHash)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.TransactionsByBlock(block)
}

func (s *Service) GetTransactionByHash(hash string) (*explorer.BlockTransaction, error) {
	return s.transactionRepo.TransactionByHash(hash)
}

func (s *Service) GetRawTransactionByHash(hash string) (*explorer.RawBlockTransaction, error) {
	return s.transactionRepo.RawTransactionByHash(hash)
}
