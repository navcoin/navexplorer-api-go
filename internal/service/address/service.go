package address

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetAddress(network string, hash string) (*explorer.Address, error)
	GetAddresses(network string, config *pagination.Config) ([]*explorer.Address, int64, error)
	GetAddressSummary(network string, hash string) (*entity.AddressSummary, error)
	GetStakingChart(period string, address string) ([]*entity.StakingGroup, error)
	GetAddressGroups(network string, period *group.Period, count int) ([]entity.AddressGroup, error)
	GetHistory(network string, hash string, txType string, config *pagination.Config) ([]*explorer.AddressHistory, int64, error)
	GetAssociatedStakingAddresses(network string, address string) ([]string, error)
	GetNamedAddresses(network string, addresses []string) ([]*explorer.Address, error)
	ValidateAddress(network string, hash string) (bool, error)
}

type service struct {
	addressRepository          *repository.AddressRepository
	addressHistoryRepository   repository.AddressHistoryRepository
	blockRepository            *repository.BlockRepository
	blockTransactionRepository *repository.BlockTransactionRepository
}

func NewAddressService(
	addressRepository *repository.AddressRepository,
	addressHistoryRepository repository.AddressHistoryRepository,
	blockRepository *repository.BlockRepository,
	blockTransactionRepository *repository.BlockTransactionRepository,
) Service {
	return &service{
		addressRepository,
		addressHistoryRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *service) GetAddress(network string, hash string) (*explorer.Address, error) {
	address, err := s.addressRepository.Network(network).AddressByHash(hash)
	if err != nil {
		return nil, err
	}

	s.UpdateCreatedAt(network, address)

	return address, err
}

func (s *service) GetAddresses(network string, config *pagination.Config) ([]*explorer.Address, int64, error) {
	return s.addressRepository.Network(network).Addresses(config.Size, config.Page)
}

func (s *service) GetHistory(network string, hash string, txType string, config *pagination.Config) ([]*explorer.AddressHistory, int64, error) {
	return s.addressHistoryRepository.Network(network).HistoryByHash(hash, txType, config.Ascending, config.Size, config.Page)
}

func (s *service) GetAddressSummary(network string, hash string) (*entity.AddressSummary, error) {
	h, err := s.addressRepository.Network(network).AddressByHash(hash)
	if err != nil {
		return nil, err
	}

	summary := &entity.AddressSummary{Height: h.Height, Hash: h.Hash}

	txs, err := s.addressHistoryRepository.Network(network).CountByHash(h.Hash)
	if err == nil {
		summary.Txs = txs
	}

	_, stakeStaking, stakeSpending, stakeVoting, err := s.addressHistoryRepository.Network(network).StakingSummary(hash)
	if err != nil {
		return nil, err
	}

	spendingReceive, spendingSent, stakingReceive, stakingSent, votingReceive, votingSent, err := s.addressHistoryRepository.Network(network).SpendSummary(hash)

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

func (s *service) GetAddressGroups(network string, period *group.Period, count int) ([]entity.AddressGroup, error) {
	return s.addressHistoryRepository.Network(network).GetAddressGroups(period, count)
}

//func (s *service) GetBalanceChart(address string) (entity.Chart, error) {
//	return s.addressTransactionRepository.BalanceChart(address)
//}
//
func (s *service) GetStakingChart(period string, address string) ([]*entity.StakingGroup, error) {
	return s.addressHistoryRepository.StakingChart(period, address)
}

//
//func (s *service) GetStakingReport() (*entity.StakingReport, error) {
//	report := new(entity.StakingReport)
//	report.To = time.Now().UTC().Truncate(time.Second)
//	report.From = report.To.AddDate(0, 0, -1)
//
//	if err := s.addressTransactionRepository.GetStakingReport(report); err != nil {
//		return nil, err
//	}
//
//	totalSupply, err := s.addressRepository.GetTotalSupply()
//	if err == nil {
//		report.TotalSupply = totalSupply
//	}
//
//	return report, nil
//}
//
//func (s *service) GetStakingByBlockCount(blockCount int, extended bool) (*entity.StakingBlocks, error) {
//	bestBlock, err := s.blockRepository.BestBlock()
//	if err != nil {
//		return nil, err
//	}
//
//	height := uint64(0)
//	if blockCount < int(bestBlock.Height) {
//		height = bestBlock.Height - uint64(blockCount)
//	}
//
//	stakingBlocks, err := s.addressTransactionRepository.GetStakingRange(height, bestBlock.Height)
//	if err != nil {
//		return nil, err
//	}
//
//	fees, err := s.blockRepository.FeesForLastBlocks(blockCount)
//	if err == nil {
//		stakingBlocks.Fees = fees
//	}
//
//	stakingBlocks.BlockCount = blockCount
//
//	return stakingBlocks, err
//}

func (s *service) GetAssociatedStakingAddresses(network, address string) ([]string, error) {
	return s.blockTransactionRepository.Network(network).AssociatedStakingAddresses(address)
}

func (s *service) GetNamedAddresses(network string, addresses []string) ([]*explorer.Address, error) {
	return s.addressRepository.Network(network).BalancesForAddresses(addresses)
}

func (s *service) ValidateAddress(network, hash string) (bool, error) {
	return true, nil
}

func (s *service) UpdateCreatedAt(network string, address *explorer.Address) {
	if address.CreatedBlock != 0 {
		return
	}

	history, err := s.addressHistoryRepository.Network(network).FirstByHash(address.Hash)
	if err != nil {
		return
	}

	address.CreatedBlock = history.Height
	address.CreatedTime = history.Time

	err = s.addressRepository.Network(network).UpdateAddress(address)
	if err != nil {
		log.WithField("hash", address.Hash).Error("Failed to update address created fields")
	} else {
		log.WithField("hash", address.Hash).Info("Updated address created fields")
	}
}
