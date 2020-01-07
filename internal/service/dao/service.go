package dao

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type Service struct {
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
) *Service {
	return &Service{
		proposalRepository,
		paymentRequestRepository,
		consensusRepository,
		voteRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *Service) GetBlockCycleByHeight(height uint64) (*entity.BlockCycle, error) {
	return s.GetBlockCycleByBlock(&explorer.Block{RawBlock: explorer.RawBlock{Height: height}})
}

func (s *Service) GetBlockCycleByBlock(block *explorer.Block) (*entity.BlockCycle, error) {
	consensus, err := s.GetConsensus()
	if err != nil {
		return nil, err
	}

	bc := block.BlockCycle(consensus.BlocksPerVotingCycle, consensus.MinSumVotesPerVotingCycle)

	blockCycle := &entity.BlockCycle{
		BlocksInCycle: consensus.BlocksPerVotingCycle,
		Quorum:        float64(consensus.MinSumVotesPerVotingCycle) / 100,
		ProposalVoting: entity.Voting{
			Cycles: consensus.MaxCountVotingCycleProposals,
			Accept: consensus.VotesAcceptProposalPercentage,
			Reject: consensus.VotesRejectProposalPercentage,
		},
		PaymentVoting: entity.Voting{
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

func (s *Service) GetConsensus() (*explorer.Consensus, error) {
	return s.consensusRepository.GetConsensus()
}

func (s *Service) GetCfundStats() (*entity.CfundStats, error) {
	cfundStats := new(entity.CfundStats)

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

func (s *Service) GetProposals(status *explorer.ProposalStatus, config *pagination.Config) ([]*explorer.Proposal, int, error) {
	return s.proposalRepository.Proposals(status, config.Dir, config.Size, config.Page)
}

func (s *Service) GetProposal(hash string) (*explorer.Proposal, error) {
	return s.proposalRepository.Proposal(hash)
}

func (s *Service) GetProposalVotes(hash string) ([]*entity.CfundVote, error) {
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
		entity.CreateVotingCycles(int(bc.ProposalVoting.Cycles), int(bc.BlocksInCycle), int(bc.FirstBlock)),
		bestBlock.Height,
	)
}

func (s *Service) GetProposalTrend(hash string) ([]*entity.CfundVote, error) {
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

	cfundVotes, err := s.voteRepository.GetVotes(
		explorer.ProposalVote,
		p.Hash,
		entity.CreateVotingCycles(10, int(bc.BlocksInCycle/10), int(bc.CurrentBlock-bc.BlocksInCycle)),
		bestBlock.Height,
	)
	if err != nil {
		return nil, err
	}

	for _, cfundVote := range cfundVotes {
		cfundVote.Yes = int(float64(cfundVote.Yes)/10) * 100
		cfundVote.No = int(float64(cfundVote.No)/10) * 100
		cfundVote.Abstain = int(float64(cfundVote.Abstain)/10) * 100
	}

	return cfundVotes, nil
}

func (s *Service) GetPaymentRequests(status *explorer.PaymentRequestStatus, config *pagination.Config) ([]*explorer.PaymentRequest, int, error) {
	return s.paymentRequestRepository.PaymentRequests(status, config.Dir, config.Size, config.Page)
}

func (s *Service) GetPaymentRequest(hash string) (*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.PaymentRequest(hash)
}

func (s *Service) GetPaymentRequestVotes(hash string) ([]*entity.CfundVote, error) {
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
		entity.CreateVotingCycles(int(bc.ProposalVoting.Cycles), int(bc.BlocksInCycle), int(bc.FirstBlock)),
		bestBlock.Height,
	)
}

func (s *Service) GetPaymentRequestTrend(hash string) ([]*entity.CfundVote, error) {
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

	cfundVotes, err := s.voteRepository.GetVotes(
		explorer.PaymentRequestVote,
		p.Hash,
		entity.CreateVotingCycles(10, int(bc.BlocksInCycle/10), int(bc.CurrentBlock-bc.BlocksInCycle)),
		bestBlock.Height,
	)
	if err != nil {
		return nil, err
	}

	for _, cfundVote := range cfundVotes {
		cfundVote.Yes = int(float64(cfundVote.Yes)/10) * 100
		cfundVote.No = int(float64(cfundVote.No)/10) * 100
		cfundVote.Abstain = int(float64(cfundVote.Abstain)/10) * 100
	}

	return cfundVotes, nil
}
