package softfork

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/softfork/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
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
