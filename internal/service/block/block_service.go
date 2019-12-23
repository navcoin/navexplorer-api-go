package block

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type BlockService struct {
	blockRepo       *repository.BlockRepository
	transactionRepo *repository.BlockTransactionRepository
}

func NewBlockService(
	blockRepo *repository.BlockRepository,
	transactionRepo *repository.BlockTransactionRepository,
) *BlockService {
	return &BlockService{blockRepo, transactionRepo}
}

func (s *BlockService) GetBestBlock() (*explorer.Block, error) {
	return s.blockRepo.BestBlock()
}

func (s *BlockService) GetBlock(hash string) (*explorer.Block, error) {
	return s.blockRepo.BlockByHashOrHeight(hash)
}

func (s *BlockService) GetRawBlock(hash string) (*explorer.RawBlock, error) {
	return s.blockRepo.RawBlockByHashOrHeight(hash)
}

func (s *BlockService) GetBlocks(config *pagination.Config) ([]*explorer.Block, int, error) {
	return s.blockRepo.Blocks(config.Dir, config.Size, config.Page)
}

func (s *BlockService) GetTransactions(blockHash string) ([]*explorer.BlockTransaction, error) {
	block, err := s.blockRepo.BlockByHashOrHeight(blockHash)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.TransactionsByBlock(block)
}

func (s *BlockService) GetTransactionByHash(hash string) (*explorer.BlockTransaction, error) {
	return s.transactionRepo.TransactionByHash(hash)
}

func (s *BlockService) GetRawTransactionByHash(hash string) (*explorer.RawBlockTransaction, error) {
	return s.transactionRepo.RawTransactionByHash(hash)
}
