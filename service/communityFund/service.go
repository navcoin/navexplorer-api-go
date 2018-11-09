package communityFund

import (
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
)

type Service struct{
	repository *Repository
}

var repository = new(Repository)

var blockService = new(block.Service)

func (s *Service) GetBlockCycle() (blockCycle BlockCycle) {
	cfConfig := config.Get().CommunityFund

	blockCycle.BlocksInCycle = cfConfig.BlocksInCycle
	blockCycle.MinQuorum = cfConfig.MinQuorum
	blockCycle.ProposalVoting.Cycles = cfConfig.ProposalVoting.Cycles
	blockCycle.ProposalVoting.Accept = cfConfig.ProposalVoting.Accept
	blockCycle.ProposalVoting.Reject = cfConfig.ProposalVoting.Reject
	blockCycle.PaymentVoting.Cycles = cfConfig.PaymentVoting.Cycles
	blockCycle.PaymentVoting.Accept = cfConfig.PaymentVoting.Accept
	blockCycle.PaymentVoting.Reject = cfConfig.PaymentVoting.Reject
	blockCycle.Height = blockService.GetBestBlock().Height

	blockCycle.Cycle = (blockCycle.Height) / (blockCycle.BlocksInCycle) + 1
	blockCycle.FirstBlock = (blockCycle.Height / blockCycle.BlocksInCycle) * blockCycle.BlocksInCycle
	blockCycle.CurrentBlock = blockCycle.Height - blockCycle.FirstBlock + 1
	blockCycle.BlocksRemaining = blockCycle.FirstBlock + blockCycle.BlocksInCycle - blockCycle.Height - 1

	return blockCycle
}

func (s *Service) GetProposalsByState(state string) (proposals []Proposal, err error) {
	proposals, err = repository.FindProposalsByState(state)
	if proposals == nil {
		proposals = make([]Proposal, 0)
	}

	return proposals, err
}

func (s *Service) GetProposalByHash(hash string) (proposal Proposal, err error) {
	return repository.FindProposalByHash(hash)
}

func (s *Service) GetPaymentRequests(proposalHash string) (paymentRequests []PaymentRequest, err error) {
	paymentRequests, err = repository.FindPaymentRequestsByProposalHash(proposalHash)
	if paymentRequests == nil {
		paymentRequests = make([]PaymentRequest, 0)
	}

	return paymentRequests, err
}