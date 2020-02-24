package coin

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/coin/entity"
)

type Service struct {
	addressRepo *repository.AddressRepository
}

func NewCoinService(addressRepo *repository.AddressRepository) *Service {
	return &Service{addressRepo}
}

func (s *Service) GetWealthDistribution(groups []int) ([]*entity.Wealth, error) {
	return s.addressRepo.WealthDistribution(groups)
}
