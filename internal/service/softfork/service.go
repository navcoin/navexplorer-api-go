package softfork

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/softfork/entity"
)

type Service struct {
	blockRepo *repository.BlockRepository
	cycleSize uint64
}

func NewSoftForkService(blockRepo *repository.BlockRepository, cycleSize uint64) *Service {
	return &Service{blockRepo, cycleSize}
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
