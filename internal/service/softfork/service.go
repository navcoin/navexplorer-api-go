package softfork

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/softfork/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type Service struct {
	blockRepo          *repository.BlockRepository
	softForkRepository *repository.SoftForkRepository
	cycleSize          uint64
}

func NewSoftForkService(blockRepo *repository.BlockRepository, softForkRepo *repository.SoftForkRepository, cycleSize uint64) *Service {
	return &Service{blockRepo, softForkRepo, cycleSize}
}

func (s *Service) GetCycle() (*entity.SoftForkCycle, error) {
	block, err := s.blockRepo.BestBlock()
	if err != nil {
		return nil, err
	}

	cycle := &entity.SoftForkCycle{
		BlocksInCycle:   s.cycleSize,
		BlockCycle:      (block.Height)/(s.cycleSize) + 1,
		CurrentBlock:    block.Height,
		FirstBlock:      (block.Height / s.cycleSize) * s.cycleSize,
		RemainingBlocks: ((block.Height / s.cycleSize) * s.cycleSize) + s.cycleSize - block.Height,
	}

	return cycle, nil
}

func (s *Service) GetSoftForks() (softForks []*explorer.SoftFork, err error) {
	return s.softForkRepository.SoftForks()
}
