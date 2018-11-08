package softFork

import "github.com/NavExplorer/navexplorer-api-go/block"

type Service struct{
	repository *Repository
}

var repository = new(Repository)
var blocksInCycle = 400

var blockService = new(block.Service)

func (s *Service) GetSoftForks() (softForks SoftForks, err error) {
	softFork, err := repository.FindAll()

	softForks.SoftForks = softFork
	softForks.BlocksInCycle = blocksInCycle
	softForks.CurrentBlock = blockService.GetBestBlock().Height
	softForks.BlockCycle = (softForks.CurrentBlock) / (blocksInCycle) + 1
	softForks.FirstBlock = (softForks.CurrentBlock / blocksInCycle) * blocksInCycle
	softForks.BlocksRemaining = softForks.FirstBlock + blocksInCycle - softForks.CurrentBlock

	return softForks, err
}
