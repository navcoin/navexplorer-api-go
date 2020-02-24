package address

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"time"
)

type Service struct {
	addressRepository            *repository.AddressRepository
	addressTransactionRepository *repository.AddressTransactionRepository
	blockRepository              *repository.BlockRepository
	blockTransactionRepository   *repository.BlockTransactionRepository
}

func NewAddressService(
	addressRepository *repository.AddressRepository,
	addressTransactionRepository *repository.AddressTransactionRepository,
	blockRepository *repository.BlockRepository,
	blockTransactionRepository *repository.BlockTransactionRepository,
) *Service {
	return &Service{
		addressRepository,
		addressTransactionRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *Service) GetAddress(hash string) (*explorer.Address, error) {
	return s.addressRepository.AddressByHash(hash)
}

func (s *Service) GetAddresses(config *pagination.Config) ([]*explorer.Address, int64, error) {
	return s.addressRepository.Addresses(config.Size, config.Page)
}

func (s *Service) GetTransactions(hash string, types string, cold bool, config *pagination.Config) ([]*explorer.AddressTransaction, int64, error) {
	return s.addressTransactionRepository.TransactionsByHash(hash, types, cold, config.Dir, config.Size, config.Page)
}

func (s *Service) GetBalanceChart(address string) (entity.Chart, error) {
	return s.addressTransactionRepository.BalanceChart(address)
}

func (s *Service) GetStakingChart(period string, address string) ([]*entity.StakingGroup, error) {
	return s.addressTransactionRepository.StakingChart(period, address)
}

func (s *Service) GetStakingReport() (*entity.StakingReport, error) {
	report := new(entity.StakingReport)
	report.To = time.Now().UTC().Truncate(time.Second)
	report.From = report.To.AddDate(0, 0, -1)

	if err := s.addressTransactionRepository.GetStakingReport(report); err != nil {
		return nil, err
	}

	totalSupply, err := s.addressRepository.GetTotalSupply()
	if err == nil {
		report.TotalSupply = totalSupply
	}

	return report, nil
}

func (s *Service) GetStakingByBlockCount(blockCount int) (*entity.StakingBlocks, error) {
	bestBlock, err := s.blockRepository.BestBlock()
	if err != nil {
		return nil, err
	}

	if blockCount > int(bestBlock.Height) {
		blockCount = int(bestBlock.Height)
	}

	stakingBlocks, err := s.addressTransactionRepository.GetStakingHigherThan(blockCount)
	if err != nil {
		return nil, err
	}

	fees, err := s.blockRepository.FeesForLastBlocks(blockCount)
	if err == nil {
		stakingBlocks.Fees = fees
	}

	return stakingBlocks, err
}

func (s *Service) GetTransactionsForAddresses(addresses []string, txType string, start *time.Time, end *time.Time) ([]*explorer.AddressTransaction, error) {
	return s.addressTransactionRepository.TransactionsForAddresses(addresses, txType, start, end)
}

func (s *Service) GetAssociatedStakingAddresses(address string) ([]string, error) {
	return s.blockTransactionRepository.AssociatedStakingAddresses(address)
}

func (s *Service) GetBalancesForAddresses(addresses []string) ([]*entity.Balance, error) {
	return s.addressRepository.BalancesForAddresses(addresses)
}

func (s *Service) GetStakingRewardsForAddresses(addresses []string) ([]*entity.StakingReward, error) {
	return s.addressTransactionRepository.StakingRewardsForAddresses(addresses)
}

func (s *Service) ValidateAddress(hash string) (bool, error) {
	return true, nil
}
