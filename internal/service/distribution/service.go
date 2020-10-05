package distribution

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
)

type Service interface {
	GetTotalSupply(network string) (float64, error)
}

type service struct {
	addressRepository *repository.AddressRepository
}

func NewDistributionService(addressRepository *repository.AddressRepository) Service {
	return &service{addressRepository}
}

func (s *service) GetTotalSupply(network string) (float64, error) {
	return s.addressRepository.Network(network).GetTotalSupply()
}
