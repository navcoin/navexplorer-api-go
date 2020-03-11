package dao

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	consensusService           *consensus.Service
	proposalRepository         *repository.DaoProposalRepository
	paymentRequestRepository   *repository.DaoPaymentRequestRepository
	consensusRepository        *repository.DaoConsensusRepository
	voteRepository             *repository.DaoVoteRepository
	blockRepository            *repository.BlockRepository
	blockTransactionRepository *repository.BlockTransactionRepository
}

func NewDaoService(
	consensusService *consensus.Service,
	proposalRepository *repository.DaoProposalRepository,
	paymentRequestRepository *repository.DaoPaymentRequestRepository,
	consensusRepository *repository.DaoConsensusRepository,
	voteRepository *repository.DaoVoteRepository,
	blockRepository *repository.BlockRepository,
	blockTransactionRepository *repository.BlockTransactionRepository,
) *Service {
	return &Service{
		consensusService,
		proposalRepository,
		paymentRequestRepository,
		consensusRepository,
		voteRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *Service) GetBlockCycleByHeight(height uint64) (*entity.LegacyBlockCycle, error) {
	return s.GetBlockCycleByBlock(&explorer.Block{RawBlock: explorer.RawBlock{Height: height}})
}

func (s *Service) GetBlockCycleByBlock(block *explorer.Block) (*entity.LegacyBlockCycle, error) {
	bc := block.BlockCycle(s.consensusService.GetParameter(consensus.VOTING_CYCLE_LENGTH).Value)

	blockCycle := &entity.LegacyBlockCycle{
		BlocksInCycle: uint(s.consensusService.GetParameter(consensus.VOTING_CYCLE_LENGTH).Value),
		ProposalVoting: entity.Voting{
			Quorum: float64(s.consensusService.GetParameter(consensus.PROPOSAL_MIN_QUORUM).Value) / 100,
			Cycles: uint(s.consensusService.GetParameter(consensus.PROPOSAL_MAX_VOTING_CYCLES).Value),
			Accept: uint(s.consensusService.GetParameter(consensus.PROPOSAL_MIN_ACCEPT).Value),
			Reject: uint(s.consensusService.GetParameter(consensus.PROPOSAL_MIN_REJECT).Value),
		},
		PaymentVoting: entity.Voting{
			Quorum: float64(s.consensusService.GetParameter(consensus.PAYMENT_REQUEST_MIN_QUORUM).Value) / 100,
			Cycles: uint(s.consensusService.GetParameter(consensus.PAYMENT_REQUEST_MAX_VOTING_CYCLES).Value),
			Accept: uint(s.consensusService.GetParameter(consensus.PAYMENT_REQUEST_MIN_ACCEPT).Value),
			Reject: uint(s.consensusService.GetParameter(consensus.PAYMENT_REQUEST_MIN_REJECT).Value),
		},
		Cycle:      uint(bc.Cycle),
		FirstBlock: uint((bc.Cycle * bc.Size) - bc.Size),
	}
	blockCycle.CurrentBlock = uint(block.Height)
	blockCycle.BlocksRemaining = blockCycle.BlocksInCycle + blockCycle.FirstBlock - blockCycle.CurrentBlock - 1

	return blockCycle, nil
}

func (s *Service) GetConsensus() (*explorer.ConsensusParameters, error) {
	return s.consensusService.GetParameters()
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

func (s *Service) GetProposals(status *explorer.ProposalStatus, config *pagination.Config) ([]*explorer.Proposal, int64, error) {
	return s.proposalRepository.Proposals(status, config.Dir, config.Size, config.Page)
}

func (s *Service) GetProposal(hash string) (*explorer.Proposal, error) {
	return s.proposalRepository.Proposal(hash)
}

func (s *Service) GetVotingCycles(element explorer.ChainHeight, max uint64) ([]*entity.VotingCycle, error) {
	bestBlock, err := s.blockRepository.BestBlock()
	if err != nil {
		return nil, err
	}

	bc, err := s.GetBlockCycleByHeight(element.GetHeight())
	if err != nil {
		return nil, err
	}

	return entity.CreateVotingCycles(
		s.consensusService.GetParameter(consensus.PROPOSAL_MAX_VOTING_CYCLES).Value,
		s.consensusService.GetParameter(consensus.VOTING_CYCLE_LENGTH).Value,
		int(bc.FirstBlock),
		bestBlock.Height,
		max,
	), nil
}

func (s *Service) GetProposalVotes(hash string) ([]*entity.CfundVote, error) {
	proposal, err := s.GetProposal(hash)
	if err != nil {
		return nil, err
	}

	votingCycles, err := s.GetVotingCycles(proposal, proposal.UpdatedOnBlock)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(explorer.ProposalVote, hash, votingCycles)
}

func (s *Service) GetProposalTrend(hash string) ([]*entity.CfundTrend, error) {
	proposal, err := s.GetProposal(hash)
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

	firstBlock := int(bc.CurrentBlock - bc.BlocksInCycle)
	if proposal.UpdatedOnBlock != 0 {
		firstBlock = int(proposal.UpdatedOnBlock)
	}
	cfundVotes, err := s.voteRepository.GetVotes(
		explorer.ProposalVote,
		proposal.Hash,
		entity.CreateVotingCycles(10, int(bc.BlocksInCycle/10), firstBlock, bestBlock.Height, 0),
	)
	if err != nil {
		return nil, err
	}

	cfundTrends := make([]*entity.CfundTrend, 0)
	for _, cfundVote := range cfundVotes {
		cfundTrend := &entity.CfundTrend{
			BlockGroup: cfundVote.BlockGroup,
			Votes: entity.Votes{
				Yes:     cfundVote.Yes,
				No:      cfundVote.No,
				Abstain: cfundVote.Abstain,
			},
			Trend: entity.Votes{
				Yes:     int(float64(cfundVote.Yes*10) / float64(bc.BlocksInCycle) * 100),
				No:      int(float64(cfundVote.No*10) / float64(bc.BlocksInCycle) * 100),
				Abstain: int(float64(cfundVote.Abstain*10) / float64(bc.BlocksInCycle) * 100),
			},
		}
		cfundTrends = append(cfundTrends, cfundTrend)
	}

	return cfundTrends, nil
}

func (s *Service) GetPaymentRequests(status *explorer.PaymentRequestStatus, config *pagination.Config) ([]*explorer.PaymentRequest, int64, error) {
	return s.paymentRequestRepository.PaymentRequests(status, config.Dir, config.Size, config.Page)
}

func (s *Service) GetPaymentRequestsForProposal(proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.PaymentRequestsForProposal(proposal)
}

func (s *Service) GetPaymentRequest(hash string) (*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.PaymentRequest(hash)
}

func (s *Service) GetPaymentRequestVotes(hash string) ([]*entity.CfundVote, error) {
	log.Debugf("GetPaymentRequestVotes(hash:%s)", hash)

	p, err := s.GetPaymentRequest(hash)
	if err != nil {
		return nil, err
	}

	votingCycles, err := s.GetVotingCycles(p, p.UpdatedOnBlock)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(explorer.PaymentRequestVote, p.Hash, votingCycles)
}

func (s *Service) GetPaymentRequestTrend(hash string) ([]*entity.CfundTrend, error) {
	log.Debugf("GetPaymentRequestTrend(hash:%s)", hash)
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

	firstBlock := int(bc.CurrentBlock - bc.BlocksInCycle)
	if p.UpdatedOnBlock != 0 {
		firstBlock = int(p.UpdatedOnBlock)
	}

	cfundVotes, err := s.voteRepository.GetVotes(
		explorer.PaymentRequestVote,
		p.Hash,
		entity.CreateVotingCycles(10, int(bc.BlocksInCycle/10), firstBlock, bestBlock.Height, 0),
	)
	if err != nil {
		return nil, err
	}

	cfundTrends := make([]*entity.CfundTrend, 0)
	for _, cfundVote := range cfundVotes {
		cfundTrend := &entity.CfundTrend{
			BlockGroup: cfundVote.BlockGroup,
			Votes: entity.Votes{
				Yes:     cfundVote.Yes,
				No:      cfundVote.No,
				Abstain: cfundVote.Abstain,
			},
			Trend: entity.Votes{
				Yes:     int(float64(cfundVote.Yes*10) / float64(bc.BlocksInCycle) * 100),
				No:      int(float64(cfundVote.No*10) / float64(bc.BlocksInCycle) * 100),
				Abstain: int(float64(cfundVote.Abstain*10) / float64(bc.BlocksInCycle) * 100),
			},
		}
		cfundTrends = append(cfundTrends, cfundTrend)
	}

	return cfundTrends, nil
}
