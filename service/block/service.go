package block

import (
	"github.com/NavExplorer/navexplorer-api-go/db/pagination"
	"strconv"
)

type Service struct{
	repository *Repository
}

var repository = new(Repository)

func (s *Service) GetBlocks(dir string, size int, offset string) (blocks []Block, paginator pagination.Paginator, err error) {
	blocks, total, err := repository.FindBlocks(dir, size, offset)

	paginator = pagination.NewPaginator(len(blocks), total, size, dir, offset)

	return blocks, paginator, err
}

func (s * Service) GetBestBlock() (block Block) {
	blocks, _, err := s.GetBlocks("DESC", 1, "")

	if err != nil || blocks == nil {
		return
	}

	return blocks[0]
}

func (s *Service) GetBlockByHashOrHeight(hashOrHeight string) (block Block, err error) {
	block, err = repository.FindOneBlockByHash(hashOrHeight)

	if err != nil {
		height, _ := strconv.Atoi(hashOrHeight)
		block, err = repository.FindOneBlockByHeight(height)
	}

	bestBlock := service.GetBestBlock()
	block.Confirmations = (bestBlock.Height - block.Height) + 1

	return block, err
}

func (s *Service) GetTransactions(dir string, size int, offset string, types []string) (transactions []Transaction, paginator pagination.Paginator, err error) {
	transactions, total, err := repository.FindTransactions(dir, size, offset, types)
	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	paginator = pagination.NewPaginator(len(transactions), total, size, dir, offset)

	return transactions, paginator, err
}

func (s *Service) GetTransactionsByHash(hash string) (transactions []Transaction, err error) {
	return repository.FindAllTransactionsByBlockHash(hash)
}

func (s *Service) GetTransactionByHash(hash string) (transaction Transaction, err error) {
	return repository.FindOneTransactionByHash(hash)
}