package address

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetAddress(n network.Network, hash string) (*explorer.Address, error)
	GetAddresses(n network.Network, pagination framework.Pagination) ([]*explorer.Address, int64, error)
	GetAddressSummary(n network.Network, hash string) (*entity.AddressSummary, error)
	GetStakingChart(n network.Network, period string, address string) ([]*entity.StakingGroup, error)
	GetAddressGroups(n network.Network, period *group.Period, count int) ([]entity.AddressGroup, error)
	GetHistory(n network.Network, hash string, request framework.RestRequest) ([]*explorer.AddressHistory, int64, error)
	GetAssociatedStakingAddresses(n network.Network, address string) ([]string, error)
	GetNamedAddresses(n network.Network, addresses []string) ([]*explorer.Address, error)
	ValidateAddress(n network.Network, hash string) (bool, error)
	GetPublicWealthDistribution(n network.Network, groups []int) ([]*entity.Wealth, error)
}

type service struct {
	addressRepository          repository.AddressRepository
	addressHistoryRepository   repository.AddressHistoryRepository
	blockRepository            repository.BlockRepository
	blockTransactionRepository repository.BlockTransactionRepository
}

func NewAddressService(
	addressRepository repository.AddressRepository,
	addressHistoryRepository repository.AddressHistoryRepository,
	blockRepository repository.BlockRepository,
	blockTransactionRepository repository.BlockTransactionRepository,
) Service {
	return &service{
		addressRepository,
		addressHistoryRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *service) GetAddress(n network.Network, hash string) (*explorer.Address, error) {
	address, err := s.addressRepository.GetAddressByHash(n, hash)
	if err != nil {
		return nil, err
	}

	s.UpdateCreatedAt(n, address)

	return address, err
}

func (s *service) GetAddresses(n network.Network, pagination framework.Pagination) ([]*explorer.Address, int64, error) {
	return s.addressRepository.GetAddresses(n, pagination.Size(), pagination.Page())
}

func (s *service) GetHistory(n network.Network, hash string, request framework.RestRequest) ([]*explorer.AddressHistory, int64, error) {
	return s.addressHistoryRepository.GetHistoryByHash(n, hash, request.Pagination(), request.Sort(), request.Filters())
}

func (s *service) GetAddressSummary(n network.Network, hash string) (*entity.AddressSummary, error) {
	h, err := s.addressRepository.GetAddressByHash(n, hash)
	if err != nil {
		return nil, err
	}

	summary := &entity.AddressSummary{Height: h.Height, Hash: h.Hash}

	txs, err := s.addressHistoryRepository.GetCountByHash(n, h.Hash)
	if err == nil {
		summary.Txs = txs
	}

	_, stakeStakable, stakeSpendable, stakeVotingWeight, err := s.addressHistoryRepository.GetStakingSummary(n, hash)
	if err != nil {
		return nil, err
	}

	spendableReceive, spendableSent, stakableReceive, stakableSent, votingWeightReceive, votingWeightSent, err := s.addressHistoryRepository.GetSpendSummary(n, hash)

	summary.Spendable = &entity.AddressBalance{
		Balance:  h.Spendable,
		Sent:     spendableSent,
		Received: spendableReceive,
		Staked:   stakeSpendable,
	}

	summary.Stakable = &entity.AddressBalance{
		Balance:  h.Stakable,
		Received: stakableReceive,
		Sent:     stakableSent,
		Staked:   stakeStakable,
	}

	summary.VotingWeight = &entity.AddressBalance{
		Balance:  h.VotingWeight,
		Received: votingWeightReceive,
		Sent:     votingWeightSent,
		Staked:   stakeVotingWeight,
	}

	return summary, nil
}

func (s *service) GetAddressGroups(n network.Network, period *group.Period, count int) ([]entity.AddressGroup, error) {
	return s.addressHistoryRepository.GetAddressGroups(n, period, count)
}

//func (s *service) GetBalanceChart(address string) (entity.Chart, error) {
//	return s.addressTransactionRepository.BalanceChart(address)
//}
//
func (s *service) GetStakingChart(n network.Network, period string, address string) ([]*entity.StakingGroup, error) {
	return s.addressHistoryRepository.GetStakingChart(n, period, address)
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

func (s *service) GetAssociatedStakingAddresses(n network.Network, address string) ([]string, error) {
	return s.blockTransactionRepository.GetAssociatedStakingAddresses(n, address)
}

func (s *service) GetNamedAddresses(n network.Network, addresses []string) ([]*explorer.Address, error) {
	return s.addressRepository.GetBalancesForAddresses(n, addresses)
}

func (s *service) ValidateAddress(n network.Network, hash string) (bool, error) {
	return true, nil
}

func (s *service) GetPublicWealthDistribution(n network.Network, groups []int) ([]*entity.Wealth, error) {
	block, err := s.blockRepository.GetBestBlock(n)
	if err != nil {
		return nil, err
	}
	return s.addressRepository.GetWealthDistribution(n, groups, block.SupplyBalance.Public)
}

func (s *service) UpdateCreatedAt(n network.Network, address *explorer.Address) {
	if address.CreatedBlock != 0 {
		return
	}

	history, err := s.addressHistoryRepository.GetFirstByHash(n, address.Hash)
	if err != nil {
		return
	}

	address.CreatedBlock = history.Height
	address.CreatedTime = history.Time

	err = s.addressRepository.UpdateAddress(n, address)
	if err != nil {
		log.WithField("hash", address.Hash).Error("Failed to update address created fields")
	} else {
		log.WithField("hash", address.Hash).Info("Updated address created fields")
	}
}
