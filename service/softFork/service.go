package softFork

import "github.com/NavExplorer/navexplorer-api-go/service/block"

type Service struct{
	repository *Repository
}

var repository = new(Repository)
var blocksInCycle = 20160

var blockService = new(block.Service)

func (s *Service) GetSoftForks() (softForks SoftForks, err error) {
	softFork, err := repository.FindAll()
	if softFork == nil {
		softFork = make([]SoftFork, 0)
	}

	softForks.SoftForks = softFork
	softForks.BlocksInCycle = blocksInCycle
	softForks.CurrentBlock = blockService.GetBestBlock().Height
	softForks.BlockCycle = (softForks.CurrentBlock) / (blocksInCycle) + 1
	softForks.FirstBlock = (softForks.CurrentBlock / blocksInCycle) * blocksInCycle
	softForks.BlocksRemaining = softForks.FirstBlock + blocksInCycle - softForks.CurrentBlock
	softForks.BlocksRequired = int(float64(softForks.BlocksInCycle) * float64(0.75))

	return softForks, err
}
