package dao

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/dto"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type DaoService struct {
	proposalRepository       *repository.DaoProposalRepository
	paymentRequestRepository *repository.DaoPaymentRequestRepository
	consensusRepository      *repository.DaoConsensusRepository
}

func NewDaoService(proposalRepository *repository.DaoProposalRepository, paymentRequestRepository *repository.DaoPaymentRequestRepository, consensusRepository *repository.DaoConsensusRepository) *DaoService {
	return &DaoService{
		proposalRepository,
		paymentRequestRepository,
		consensusRepository,
	}
}

func (s *DaoService) GetConsensus() (*explorer.Consensus, error) {
	return s.consensusRepository.GetConsensus()
}

func (s *DaoService) GetBlockCycle(block *explorer.Block) (*dto.BlockCycle, error) {
	consensus, err := s.GetConsensus()
	if err != nil {
		return nil, err
	}

	bc := block.BlockCycle(consensus.BlocksPerVotingCycle, consensus.MinSumVotesPerVotingCycle)

	blockCycle := &dto.BlockCycle{
		BlocksInCycle: consensus.BlocksPerVotingCycle,
		Quorum:        float64(consensus.MinSumVotesPerVotingCycle) / 100,
		ProposalVoting: dto.Voting{
			Cycles: consensus.MaxCountVotingCycleProposals,
			Accept: consensus.VotesAcceptProposalPercentage,
			Reject: consensus.VotesRejectProposalPercentage,
		},
		PaymentVoting: dto.Voting{
			Cycles: consensus.MaxCountVotingCyclePaymentRequests,
			Accept: consensus.VotesAcceptPaymentRequestPercentage,
			Reject: consensus.VotesRejectPaymentRequestPercentage,
		},
		Height:     block.Height,
		Cycle:      bc.Cycle,
		FirstBlock: bc.Cycle * bc.Size,
	}
	blockCycle.CurrentBlock = uint(blockCycle.Height) - (bc.Cycle * bc.Size)
	blockCycle.BlocksRemaining = blockCycle.BlocksInCycle - blockCycle.CurrentBlock

	return blockCycle, nil
}

func (s *DaoService) GetProposal(hash string) (*explorer.Proposal, error) {
	return s.proposalRepository.Proposal(hash)
}

func (s *DaoService) GetProposals(status *explorer.ProposalStatus, config *pagination.Config) ([]*explorer.Proposal, int, error) {
	return s.proposalRepository.Proposals(status, config.Dir, config.Size, config.Page)
}

func (s *DaoService) GetPaymentRequest(hash string) (*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.PaymentRequest(hash)
}

func (s *DaoService) GetPaymentRequests(status *explorer.PaymentRequestStatus, config *pagination.Config) ([]*explorer.PaymentRequest, int, error) {
	return s.paymentRequestRepository.PaymentRequests(status, config.Dir, config.Size, config.Page)
}
