package address

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type AddressService struct {
	addressRepository            *repository.AddressRepository
	addressTransactionRepository *repository.AddressTransactionRepository
}

func NewAddressService(
	addressRepository *repository.AddressRepository,
	addressTransactionRepository *repository.AddressTransactionRepository,
) *AddressService {
	return &AddressService{addressRepository, addressTransactionRepository}
}

func (s *AddressService) GetAddress(hash string) (*explorer.Address, error) {
	return s.addressRepository.AddressByHash(hash)
}

func (s *AddressService) GetAddresses(config *pagination.Config) ([]*explorer.Address, int, error) {
	return s.addressRepository.Addresses(config.Size, config.Page)
}

func (s *AddressService) GetTransactions(hash string, cold bool, config *pagination.Config) ([]*explorer.AddressTransaction, int, error) {
	return s.addressTransactionRepository.TransactionsByHash(hash, cold, config.Dir, config.Size, config.Page)
}

func (s *AddressService) ValidateAddress(hash string) (bool, error) {
	return true, nil
}
