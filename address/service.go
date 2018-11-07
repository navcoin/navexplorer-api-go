package address

type Service struct{
	repository *Repository
}

var repository = new(Repository)

func (s *Service) GetAddress(hash string) (address Address, err error) {
	address, err = repository.FindOneAddressByHash(hash)

	if err == nil {
		richListPosition := repository.GetRichListPosition(address)
		address.RichListPosition = &richListPosition
	}

	return address, err
}

func(s *Service) GetAddresses(count int) (addresses []Address, err error) {
	addresses, err = repository.FindTopAddressesOrderByBalanceDesc(count)

	for index, address := range addresses {
		richListPosition := repository.GetRichListPosition(address)
		addresses[index].RichListPosition = &richListPosition
	}

	return addresses, err
}

func(s *Service) GetAddressTransactions(address string, types []string, dir string, size int, offset string) (transactions []Transaction, err error) {
	transactions, err = repository.FindTransactionsByAddress(address, types, dir, size, offset)

	return transactions, err
}
