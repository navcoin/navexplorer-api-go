package service

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
)

type SupplyService interface {
	GetPublicSupply(n network.Network) (float64, error)
	GetPrivateSupply(n network.Network) (float64, error)
}

func NewSupplyService(
	addressRepository repository.AddressRepository,
	blockTransactionRepository repository.BlockTransactionRepository,
) SupplyService {
	return &supplyService{addressRepository, blockTransactionRepository}
}

type supplyService struct {
	addressRepository          repository.AddressRepository
	blockTransactionRepository repository.BlockTransactionRepository
}

func (s *supplyService) GetPublicSupply(n network.Network) (float64, error) {
	return s.addressRepository.GetTotalSupply(n)
}

func (s *supplyService) GetPrivateSupply(n network.Network) (float64, error) {
	return s.blockTransactionRepository.GetPrivateSupply(n)
}
