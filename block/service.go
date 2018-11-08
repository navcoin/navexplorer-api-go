package block

import "strconv"

type Service struct{
	repository *Repository
}

var repository = new(Repository)

func (s *Service) GetBlocks(dir string, size int, offset string) (blocks []Block, err error) {
	blocks, err = repository.FindBlocks(dir, size, offset)

	return blocks, err
}

func (s * Service) GetBestBlock() (block Block) {
	blocks, err := s.GetBlocks("DESC", 1, "")

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