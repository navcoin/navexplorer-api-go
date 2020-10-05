package coin

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/coin/entity"
)

type Service interface {
	GetWealthDistribution(network string, groups []int) ([]*entity.Wealth, error)
}

type service struct {
	addressRepo *repository.AddressRepository
}

func NewCoinService(addressRepo *repository.AddressRepository) Service {
	return &service{addressRepo}
}

func (s *service) GetWealthDistribution(network string, groups []int) ([]*entity.Wealth, error) {
	return s.addressRepo.Network(network).WealthDistribution(groups)
}
