package distribution

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
)

type Service interface {
	GetTotalSupply(n network.Network) (float64, error)
}

type service struct {
	addressRepository repository.AddressRepository
}

func NewDistributionService(addressRepository repository.AddressRepository) Service {
	return &service{addressRepository}
}

func (s *service) GetTotalSupply(n network.Network) (float64, error) {
	return s.addressRepository.GetTotalSupply(n)
}
