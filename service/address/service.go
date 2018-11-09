package address

import (
	"github.com/NavExplorer/navexplorer-api-go/db/pagination"
)

type Service struct{
	repository *Repository
}

var repository = new(Repository)

func (s *Service) GetAddress(hash string) (address Address, err error) {
	address, err = repository.FindOneAddressByHash(hash)

	if err == nil {
		richListPosition := repository.GetRichListPosition(address)
		address.RichListPosition = richListPosition
	}

	return address, err
}

func(s *Service) GetAddresses(count int) (addresses []Address, err error) {
	addresses, err = repository.FindTopAddressesOrderByBalanceDesc(count)

	for index := range addresses {
		addresses[index].RichListPosition = index + 1
	}

	return addresses, err
}

func(s *Service) GetTransactions(address string, dir string, size int, offset string, types []string) (txs []Transaction, paginator pagination.Paginator, err error) {
	transactions, total, err := repository.FindTransactionsByAddress(address, dir, size, offset, types)
	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	paginator = pagination.NewPaginator(len(transactions), total, size, dir, offset)

	return transactions, paginator, err
}
