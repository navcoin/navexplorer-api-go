package address

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"time"
)

type Service interface {
	GetAddress(hash string) (*explorer.Address, error)
	GetAddresses(config *pagination.Config) ([]*explorer.Address, int64, error)
	GetAddressSummary(hash string) (*entity.AddressSummary, error)
	GetAddressGroups(period *group.Period, count int) (*entity.AddressGroups, error)
	GetHistory(hash string, txType string, config *pagination.Config) ([]*explorer.AddressHistory, int64, error)
	GetBalanceChart(address string) (entity.Chart, error)
	GetStakingChart(period string, address string) ([]*entity.StakingGroup, error)
	GetStakingReport() (*entity.StakingReport, error)
	GetStakingByBlockCount(blockCount int, extended bool) (*entity.StakingBlocks, error)
	GetAssociatedStakingAddresses(address string) ([]string, error)
	GetNamedAddresses(addresses []string) ([]*explorer.Address, error)
	GetStakingRewardsForAddresses(addresses []string) ([]*entity.StakingReward, error)
	ValidateAddress(hash string) (bool, error)
}

type service struct {
	addressRepository            *repository.AddressRepository
	addressHistoryRepository     *repository.AddressHistoryRepository
	addressTransactionRepository *repository.AddressTransactionRepository
	blockRepository              *repository.BlockRepository
	blockTransactionRepository   *repository.BlockTransactionRepository
}

func NewAddressService(
	addressRepository *repository.AddressRepository,
	addressHistoryRepository *repository.AddressHistoryRepository,
	addressTransactionRepository *repository.AddressTransactionRepository,
	blockRepository *repository.BlockRepository,
	blockTransactionRepository *repository.BlockTransactionRepository,
) Service {
	return &service{
		addressRepository,
		addressHistoryRepository,
		addressTransactionRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *service) GetAddress(hash string) (*explorer.Address, error) {
	address, err := s.addressRepository.AddressByHash(hash)
	if err != nil {
		return nil, err
	}

	s.UpdateCreatedAt(address)

	return address, err
}

func (s *service) GetAddresses(config *pagination.Config) ([]*explorer.Address, int64, error) {
	return s.addressRepository.Addresses(config.Size, config.Page)
}

func (s *service) GetHistory(hash string, txType string, config *pagination.Config) ([]*explorer.AddressHistory, int64, error) {
	return s.addressHistoryRepository.HistoryByHash(hash, txType, config.Ascending, config.Size, config.Page)
}

func (s *service) GetAddressSummary(hash string) (*entity.AddressSummary, error) {
	h, err := s.addressRepository.AddressByHash(hash)
	if err != nil {
		return nil, err
	}

	summary := &entity.AddressSummary{Height: h.Height, Hash: h.Hash}

	txs, err := s.addressHistoryRepository.CountByHash(h.Hash)
	if err == nil {
		summary.Txs = txs
	}

	_, stakeStaking, stakeSpending, stakeVoting, err := s.addressHistoryRepository.StakingSummary(hash)
	if err != nil {
		return nil, err
	}

	spendingReceive, spendingSent, stakingReceive, stakingSent, votingReceive, votingSent, err := s.addressHistoryRepository.SpendSummary(hash)

	summary.Spending = &entity.AddressBalance{
		Balance:  h.Spending,
		Sent:     spendingSent,
		Received: spendingReceive,
		Staked:   stakeSpending,
	}

	summary.Staking = &entity.AddressBalance{
		Balance:  h.Staking,
		Received: stakingReceive,
		Sent:     stakingSent,
		Staked:   stakeStaking,
	}

	summary.Voting = &entity.AddressBalance{
		Balance:  h.Voting,
		Received: votingReceive,
		Sent:     votingSent,
		Staked:   stakeVoting,
	}

	return summary, nil
}

func (s *service) GetAddressGroups(period *group.Period, count int) (*entity.AddressGroups, error) {
	timeGroups := group.CreateTimeGroup(period, count)

	addressGroups := new(entity.AddressGroups)
	for i := range timeGroups {
		blockGroup := &entity.AddressGroup{
			TimeGroup: *timeGroups[i],
			Period:    *period,
		}
		addressGroups.Items = append(addressGroups.Items, blockGroup)
	}

	err := s.addressHistoryRepository.GetAddressGroups(addressGroups)

	return addressGroups, err
}

func (s *service) GetBalanceChart(address string) (entity.Chart, error) {
	return s.addressTransactionRepository.BalanceChart(address)
}

func (s *service) GetStakingChart(period string, address string) ([]*entity.StakingGroup, error) {
	return s.addressTransactionRepository.StakingChart(period, address)
}

func (s *service) GetStakingReport() (*entity.StakingReport, error) {
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

func (s *service) GetStakingByBlockCount(blockCount int, extended bool) (*entity.StakingBlocks, error) {
	bestBlock, err := s.blockRepository.BestBlock()
	if err != nil {
		return nil, err
	}

	height := uint64(0)
	if blockCount < int(bestBlock.Height) {
		height = bestBlock.Height - uint64(blockCount)
	}

	stakingBlocks, err := s.addressTransactionRepository.GetStakingRange(height, bestBlock.Height)
	if err != nil {
		return nil, err
	}

	fees, err := s.blockRepository.FeesForLastBlocks(blockCount)
	if err == nil {
		stakingBlocks.Fees = fees
	}

	stakingBlocks.BlockCount = blockCount

	return stakingBlocks, err
}

func (s *service) GetAssociatedStakingAddresses(address string) ([]string, error) {
	return s.blockTransactionRepository.AssociatedStakingAddresses(address)
}

func (s *service) GetNamedAddresses(addresses []string) ([]*explorer.Address, error) {
	return s.addressRepository.BalancesForAddresses(addresses)
}

func (s *service) GetStakingRewardsForAddresses(addresses []string) ([]*entity.StakingReward, error) {
	return s.addressTransactionRepository.StakingRewardsForAddresses(addresses)
}

func (s *service) ValidateAddress(hash string) (bool, error) {
	return true, nil
}

func (s *service) UpdateCreatedAt(address *explorer.Address) {
	if address.CreatedBlock != 0 {
		return
	}

	history, err := s.addressHistoryRepository.FirstByHash(address.Hash)
	if err != nil {
		return
	}

	address.CreatedBlock = history.Height
	address.CreatedTime = history.Time

	err = s.addressRepository.UpdateAddress(address)
	if err != nil {
		log.WithField("hash", address.Hash).Error("Failed to update address created fields")
	} else {
		log.WithField("hash", address.Hash).Info("Updated address created fields")
	}
}
