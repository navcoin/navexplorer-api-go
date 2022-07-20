package softfork

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/repository"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/network"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/softfork/entity"
	"github.com/navcoin/navexplorer-indexer-go/v2/pkg/explorer"
)

type Service interface {
	GetCycle(n network.Network) (*entity.SoftForkCycle, error)
	GetSoftForks(n network.Network) ([]*explorer.SoftFork, error)
}

type service struct {
	blockRepo          repository.BlockRepository
	softForkRepository repository.SoftForkRepository
}

func NewSoftForkService(blockRepo repository.BlockRepository, softForkRepo repository.SoftForkRepository) Service {
	return &service{blockRepo, softForkRepo}
}

func (s *service) GetCycle(n network.Network) (*entity.SoftForkCycle, error) {
	block, err := s.blockRepo.GetBestBlock(n)
	if err != nil {
		return nil, err
	}

	cycleSize := entity.GetBlocksInCycle(n)

	cycle := &entity.SoftForkCycle{
		BlocksInCycle:   cycleSize,
		BlockCycle:      (block.Height / cycleSize) + 1,
		CurrentBlock:    block.Height,
		FirstBlock:      (block.Height / cycleSize) * cycleSize,
		RemainingBlocks: ((block.Height / cycleSize) * cycleSize) + cycleSize - block.Height,
	}

	return cycle, nil
}

func (s *service) GetSoftForks(n network.Network) ([]*explorer.SoftFork, error) {
	return s.softForkRepository.GetSoftForks(n)
}
