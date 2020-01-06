package dao

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/dto"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/voting_cycle"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type DaoService struct {
	proposalRepository         *repository.DaoProposalRepository
	paymentRequestRepository   *repository.DaoPaymentRequestRepository
	consensusRepository        *repository.DaoConsensusRepository
	voteRepository             *repository.DaoVoteRepository
	blockRepository            *repository.BlockRepository
	blockTransactionRepository *repository.BlockTransactionRepository
}

func NewDaoService(
	proposalRepository *repository.DaoProposalRepository,
	paymentRequestRepository *repository.DaoPaymentRequestRepository,
	consensusRepository *repository.DaoConsensusRepository,
	voteRepository *repository.DaoVoteRepository,
	blockRepository *repository.BlockRepository,
	blockTransactionRepository *repository.BlockTransactionRepository,
) *DaoService {
	return &DaoService{
		proposalRepository,
		paymentRequestRepository,
		consensusRepository,
		voteRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *DaoService) GetBlockCycleByHeight(height uint64) (*dto.BlockCycle, error) {
	return s.GetBlockCycleByBlock(&explorer.Block{RawBlock: explorer.RawBlock{Height: height}})
}

func (s *DaoService) GetBlockCycleByBlock(block *explorer.Block) (*dto.BlockCycle, error) {
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
		Cycle:      bc.Cycle,
		FirstBlock: (bc.Cycle * bc.Size) - bc.Size + 1,
	}
	blockCycle.CurrentBlock = uint(block.Height)
	blockCycle.BlocksRemaining = blockCycle.BlocksInCycle + blockCycle.FirstBlock - blockCycle.CurrentBlock

	return blockCycle, nil
}

func (s *DaoService) GetConsensus() (*explorer.Consensus, error) {
	return s.consensusRepository.GetConsensus()
}

func (s *DaoService) GetCfundStats() (*dto.CfundStats, error) {
	cfundStats := new(dto.CfundStats)

	if contributed, err := s.blockTransactionRepository.TotalAmountByOutputType(explorer.VoutCfundContribution); err == nil {
		cfundStats.Contributed = *contributed
	}

	if locked, err := s.proposalRepository.ValueLocked(); err == nil {
		cfundStats.Locked = *locked
	}

	if paid, err := s.paymentRequestRepository.ValuePaid(); err == nil {
		cfundStats.Paid = *paid
	}

	cfundStats.Available = cfundStats.Contributed - cfundStats.Paid - cfundStats.Locked

	return cfundStats, nil
}

func (s *DaoService) GetProposals(status *explorer.ProposalStatus, config *pagination.Config) ([]*explorer.Proposal, int, error) {
	return s.proposalRepository.Proposals(status, config.Dir, config.Size, config.Page)
}

func (s *DaoService) GetProposal(hash string) (*explorer.Proposal, error) {
	return s.proposalRepository.Proposal(hash)
}

func (s *DaoService) GetProposalVotes(hash string) ([]*dto.CfundVote, error) {
	p, err := s.GetProposal(hash)
	if err != nil {
		return nil, err
	}

	bestBlock, err := s.blockRepository.BestBlock()
	if err != nil {
		return nil, err
	}

	bc, err := s.GetBlockCycleByHeight(p.Height)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(
		explorer.ProposalVote,
		p.Hash,
		voting_cycle.CreateVotingCycles(int(bc.ProposalVoting.Cycles), int(bc.BlocksInCycle), int(bc.FirstBlock)),
		bestBlock.Height,
	)
}

func (s *DaoService) GetProposalTrend(hash string) ([]*dto.CfundVote, error) {
	p, err := s.GetProposal(hash)
	if err != nil {
		return nil, err
	}

	bestBlock, err := s.blockRepository.BestBlock()
	if err != nil {
		return nil, err
	}

	bc, err := s.GetBlockCycleByBlock(bestBlock)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetTrend(
		explorer.ProposalVote,
		p.Hash,
		voting_cycle.CreateVotingCycles(10, int(bc.BlocksInCycle/10), int(bc.CurrentBlock-bc.BlocksInCycle)),
		bestBlock.Height,
	)
}

func (s *DaoService) GetPaymentRequests(status *explorer.PaymentRequestStatus, config *pagination.Config) ([]*explorer.PaymentRequest, int, error) {
	return s.paymentRequestRepository.PaymentRequests(status, config.Dir, config.Size, config.Page)
}

func (s *DaoService) GetPaymentRequest(hash string) (*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.PaymentRequest(hash)
}

func (s *DaoService) GetPaymentRequestVotes(hash string) ([]*dto.CfundVote, error) {
	p, err := s.GetPaymentRequest(hash)
	if err != nil {
		return nil, err
	}

	bestBlock, err := s.blockRepository.BestBlock()
	if err != nil {
		return nil, err
	}

	bc, err := s.GetBlockCycleByHeight(p.Height)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(
		explorer.PaymentRequestVote,
		p.Hash,
		voting_cycle.CreateVotingCycles(int(bc.ProposalVoting.Cycles), int(bc.BlocksInCycle), int(bc.FirstBlock)),
		bestBlock.Height,
	)
}

func (s *DaoService) GetPaymentRequestTrend(hash string) ([]*dto.CfundVote, error) {
	p, err := s.GetPaymentRequest(hash)
	if err != nil {
		return nil, err
	}

	bestBlock, err := s.blockRepository.BestBlock()
	if err != nil {
		return nil, err
	}

	bc, err := s.GetBlockCycleByBlock(bestBlock)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetTrend(
		explorer.PaymentRequestVote,
		p.Hash,
		voting_cycle.CreateVotingCycles(10, int(bc.BlocksInCycle/10), int(bc.CurrentBlock-bc.BlocksInCycle)),
		bestBlock.Height,
	)
}
