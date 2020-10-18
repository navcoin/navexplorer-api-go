package coin

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/coin/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
)

type Service interface {
	GetWealthDistribution(n network.Network, groups []int) ([]*entity.Wealth, error)
}

type service struct {
	addressRepo repository.AddressRepository
}

func NewCoinService(addressRepo repository.AddressRepository) Service {
	return &service{addressRepo}
}

func (s *service) GetWealthDistribution(n network.Network, groups []int) ([]*entity.Wealth, error) {
	return s.addressRepo.GetWealthDistribution(n, groups)
}
