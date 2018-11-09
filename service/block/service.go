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

func (s *Service) GetTransactions(dir string, size int, offset string, types []string) (transactions []Transaction, err error) {
	transactions, err = repository.FindTransactions(dir, size, offset, types)

	return transactions, err
}

func (s *Service) GetTransactionsByBlock(hash string) (transactions []Transaction, err error) {
	transactions, err = repository.FindAllTransactionsByBlockHash(hash)

	return transactions, err
}

func (s *Service) GetTransactionByHash(hash string) (transaction Transaction, err error) {
	transaction, err = repository.FindOneTransactionByHash(hash)

	return transaction, err
}