package service

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/repository"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/address/entity"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/network"
)

type StakingService interface {
	GetStakingRewardsForAddresses(n network.Network, addresses []string) ([]*entity.StakingReward, error)
}

func NewStakingService(
	addressHistoryRepository repository.AddressHistoryRepository,
) StakingService {
	return &stakingService{addressHistoryRepository}
}

type stakingService struct {
	addressHistoryRepository repository.AddressHistoryRepository
}

func (s *stakingService) GetStakingRewardsForAddresses(n network.Network, addresses []string) ([]*entity.StakingReward, error) {
	return s.addressHistoryRepository.StakingRewardsForAddresses(n, addresses)
}
