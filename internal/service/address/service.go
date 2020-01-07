package address

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type Service struct {
	addressRepository            *repository.AddressRepository
	addressTransactionRepository *repository.AddressTransactionRepository
	blockTransactionRepository   *repository.BlockTransactionRepository
}

func NewAddressService(
	addressRepository *repository.AddressRepository,
	addressTransactionRepository *repository.AddressTransactionRepository,
	blockTransactionRepository *repository.BlockTransactionRepository,
) *Service {
	return &Service{
		addressRepository,
		addressTransactionRepository,
		blockTransactionRepository,
	}
}

func (s *Service) GetAddress(hash string) (*explorer.Address, error) {
	return s.addressRepository.AddressByHash(hash)
}

func (s *Service) GetAddresses(config *pagination.Config) ([]*explorer.Address, int, error) {
	return s.addressRepository.Addresses(config.Size, config.Page)
}

func (s *Service) GetTransactions(hash string, cold bool, config *pagination.Config) ([]*explorer.AddressTransaction, int, error) {
	return s.addressTransactionRepository.TransactionsByHash(hash, cold, config.Dir, config.Size, config.Page)
}

func (s *Service) GetStakingReport(hash string, period *group.Period) ([]*entity.StakingReport, error) {
	timeGroups := group.CreateTimeGroup(period, 12)

	stakingReport := make([]*entity.StakingReport, 0)
	for i := range timeGroups {
		addressStaking := &entity.StakingReport{
			TimeGroup: *timeGroups[i],
			Stakes:    0,
			Amount:    0,
		}
		stakingReport = append(stakingReport, addressStaking)
	}

	err := s.addressTransactionRepository.GetStakingReport(hash, stakingReport)

	return stakingReport, err
}

func (s *Service) GetAssociatedStakingAddresses(address string) ([]string, error) {
	return s.blockTransactionRepository.GetAssociatedStakingAddresses(address)
}

func (s *Service) ValidateAddress(hash string) (bool, error) {
	return true, nil
}
