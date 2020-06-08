package dao

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetBlockCycleByHeight(height uint64) (*entity.LegacyBlockCycle, error)
	GetBlockCycleByBlock(block *explorer.Block) (*entity.LegacyBlockCycle, error)
	GetConsensus() (*explorer.ConsensusParameters, error)
	GetCfundStats() (*entity.CfundStats, error)

	GetProposals(parameters ProposalParameters, config *pagination.Config) ([]*explorer.Proposal, int64, error)
	GetProposal(hash string) (*explorer.Proposal, error)
	GetVotingCycles(element explorer.ChainHeight, count uint) ([]*entity.VotingCycle, error)
	GetProposalVotes(hash string) ([]*entity.CfundVote, error)
	GetProposalTrend(hash string) ([]*entity.CfundTrend, error)

	GetPaymentRequests(parameters PaymentRequestParameters, config *pagination.Config) ([]*explorer.PaymentRequest, int64, error)
	GetPaymentRequestsForProposal(proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error)
	GetPaymentRequest(hash string) (*explorer.PaymentRequest, error)
	GetPaymentRequestVotes(hash string) ([]*entity.CfundVote, error)
	GetPaymentRequestTrend(hash string) ([]*entity.CfundTrend, error)

	GetConsultations(parameters ConsultationParameters, config *pagination.Config) ([]*explorer.Consultation, int64, error)
	GetConsultation(hash string) (*explorer.Consultation, error)
	GetAnswer(hash string) (*explorer.Answer, error)
	GetAnswerVotes(consultationHash string, hash string) ([]*entity.CfundVote, error)
	GetConsensusConsultations(config *pagination.Config) ([]*explorer.Consultation, int64, error)
}

type service struct {
	consensusService           consensus.Service
	proposalRepository         *repository.DaoProposalRepository
	paymentRequestRepository   *repository.DaoPaymentRequestRepository
	consultationRepository     *repository.DaoConsultationRepository
	consensusRepository        *repository.DaoConsensusRepository
	voteRepository             *repository.DaoVoteRepository
	blockRepository            *repository.BlockRepository
	blockTransactionRepository *repository.BlockTransactionRepository
}

type ConsultationParameters struct {
	State     *uint                        `form:"state"`
	Status    *explorer.ConsultationStatus `form:"-"`
	Consensus *bool                        `form:"consensus"`
	Min       *uint                        `form:"min"`
}

type ProposalParameters struct {
	State *uint `form:"state"`
	Votes bool  `form:"votes"`
}

type PaymentRequestParameters struct {
	Proposal string `form:"proposal"`
	State    *uint  `form:"state"`
	Votes    bool   `form:"votes"`
}

func NewDaoService(
	consensusService consensus.Service,
	proposalRepository *repository.DaoProposalRepository,
	paymentRequestRepository *repository.DaoPaymentRequestRepository,
	consultationRepository *repository.DaoConsultationRepository,
	consensusRepository *repository.DaoConsensusRepository,
	voteRepository *repository.DaoVoteRepository,
	blockRepository *repository.BlockRepository,
	blockTransactionRepository *repository.BlockTransactionRepository,
) Service {
	return &service{
		consensusService,
		proposalRepository,
		paymentRequestRepository,
		consultationRepository,
		consensusRepository,
		voteRepository,
		blockRepository,
		blockTransactionRepository,
	}
}

func (s *service) GetBlockCycleByHeight(height uint64) (*entity.LegacyBlockCycle, error) {
	return s.GetBlockCycleByBlock(&explorer.Block{RawBlock: explorer.RawBlock{Height: height}})
}

func (s *service) GetBlockCycleByBlock(block *explorer.Block) (*entity.LegacyBlockCycle, error) {
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
		Cycle:      block.BlockCycle.Cycle,
		FirstBlock: (block.BlockCycle.Cycle * block.BlockCycle.Size) - block.BlockCycle.Size,
	}
	blockCycle.CurrentBlock = uint(block.Height)
	blockCycle.BlocksRemaining = blockCycle.BlocksInCycle + blockCycle.FirstBlock - blockCycle.CurrentBlock - 1

	return blockCycle, nil
}

func (s *service) GetConsensus() (*explorer.ConsensusParameters, error) {
	return s.consensusService.GetParameters()
}

func (s *service) GetCfundStats() (*entity.CfundStats, error) {
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

func (s *service) GetProposals(parameters ProposalParameters, config *pagination.Config) ([]*explorer.Proposal, int64, error) {
	var status *explorer.ProposalStatus
	if parameters.State != nil && explorer.IsProposalStateValid(*parameters.State) {
		s := explorer.GetProposalStatusByState(*parameters.State)
		status = &s
	}

	return s.proposalRepository.Proposals(status, config.Ascending, config.Size, config.Page)
}

func (s *service) GetProposal(hash string) (*explorer.Proposal, error) {
	return s.proposalRepository.Proposal(hash)
}

func (s *service) GetVotingCycles(element explorer.ChainHeight, count uint) ([]*entity.VotingCycle, error) {
	log.Infof("GetVotingCycles for %T", element)

	block, err := s.blockRepository.BlockByHeight(element.GetHeight())
	if err != nil {
		return nil, err
	}

	var segments uint
	switch e := element.(type) {
	case *explorer.Proposal:
		segments = uint(s.consensusService.GetParameter(consensus.PROPOSAL_MAX_VOTING_CYCLES).Value) + 2
	case *explorer.PaymentRequest:
		segments = uint(s.consensusService.GetParameter(consensus.PAYMENT_REQUEST_MAX_VOTING_CYCLES).Value) + 2
	case *explorer.Consultation:
		segments = uint(s.consensusService.GetParameter(consensus.CONSULTATION_MAX_VOTING_CYCLES).Value) + 2
	default:
		log.Fatalf("Unable to get Max voting cycles from %T", e)
	}

	return entity.CreateVotingCycles(
		segments,
		uint(s.consensusService.GetParameter(consensus.VOTING_CYCLE_LENGTH).Value),
		uint(block.Height)-block.BlockCycle.Index,
		count+1,
	), nil
}

func (s *service) GetProposalVotes(hash string) ([]*entity.CfundVote, error) {
	log.WithField("hash", hash).Info("GetProposalVotes")

	proposal, err := s.GetProposal(hash)
	if err != nil {
		return nil, err
	}

	votingCycles, err := s.GetVotingCycles(proposal, proposal.VotingCycle)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(explorer.ProposalVote, hash, votingCycles)
}

func (s *service) GetProposalTrend(hash string) ([]*entity.CfundTrend, error) {
	log.WithField("hash", hash).Info("GetProposalTrend")

	proposal, err := s.GetProposal(hash)
	if err != nil {
		return nil, err
	}

	max, err := s.getMax(proposal.Status, proposal.StateChangedOnBlock)
	if err != nil {
		return nil, err
	}

	size := uint(s.consensusService.GetParameter(consensus.VOTING_CYCLE_LENGTH).Value)

	cfundVotes, err := s.voteRepository.GetVotes(
		explorer.ProposalVote,
		proposal.Hash,
		entity.CreateVotingCycles(10, size/10, max-size+1, 10),
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
				Yes:     int(float64(cfundVote.Yes*10) / float64(size) * 100),
				No:      int(float64(cfundVote.No*10) / float64(size) * 100),
				Abstain: int(float64(cfundVote.Abstain*10) / float64(size) * 100),
			},
		}
		cfundTrends = append(cfundTrends, cfundTrend)
	}

	return cfundTrends, nil
}

func (s *service) GetPaymentRequests(parameters PaymentRequestParameters, config *pagination.Config) ([]*explorer.PaymentRequest, int64, error) {
	var status *explorer.PaymentRequestStatus
	if parameters.State != nil && explorer.IsPaymentRequestStateValid(*parameters.State) {
		s := explorer.GetPaymentRequestStatusByState(*parameters.State)
		status = &s
	}

	var proposalHash string
	if parameters.Proposal != "" {
		proposalHash = parameters.Proposal
	}

	return s.paymentRequestRepository.PaymentRequests(proposalHash, status, config.Ascending, config.Size, config.Page)
}

func (s *service) GetPaymentRequestsForProposal(proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.PaymentRequestsForProposal(proposal)
}

func (s *service) GetPaymentRequest(hash string) (*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.PaymentRequest(hash)
}

func (s *service) GetPaymentRequestVotes(hash string) ([]*entity.CfundVote, error) {
	log.Debugf("GetPaymentRequestVotes(hash:%s)", hash)

	paymentRequest, err := s.GetPaymentRequest(hash)
	if err != nil {
		return nil, err
	}

	votingCycles, err := s.GetVotingCycles(paymentRequest, paymentRequest.VotingCycle)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(explorer.PaymentRequestVote, hash, votingCycles)
}

func (s *service) GetPaymentRequestTrend(hash string) ([]*entity.CfundTrend, error) {
	log.Debugf("GetPaymentRequestTrend(hash:%s)", hash)
	paymentRequest, err := s.GetPaymentRequest(hash)
	if err != nil {
		return nil, err
	}

	max, err := s.getMax(paymentRequest.Status, paymentRequest.StateChangedOnBlock)
	if err != nil {
		return nil, err
	}

	size := uint(s.consensusService.GetParameter(consensus.VOTING_CYCLE_LENGTH).Value)

	cfundVotes, err := s.voteRepository.GetVotes(
		explorer.PaymentRequestVote,
		paymentRequest.Hash,
		entity.CreateVotingCycles(10, size/10, max-size, 10),
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
				Yes:     int(float64(cfundVote.Yes*10) / float64(size) * 100),
				No:      int(float64(cfundVote.No*10) / float64(size) * 100),
				Abstain: int(float64(cfundVote.Abstain*10) / float64(size) * 100),
			},
		}
		cfundTrends = append(cfundTrends, cfundTrend)
	}

	return cfundTrends, nil
}

func (s *service) GetConsultations(parameters ConsultationParameters, config *pagination.Config) ([]*explorer.Consultation, int64, error) {
	if parameters.State != nil && explorer.IsConsultationStateValid(*parameters.State) {
		s := explorer.GetConsultationStatusByState(*parameters.State)
		parameters.Status = &s
	}

	return s.consultationRepository.Consultations(parameters.Status, parameters.Consensus, parameters.Min, config.Ascending, config.Size, config.Page)
}

func (s *service) GetConsultation(hash string) (*explorer.Consultation, error) {
	return s.consultationRepository.Consultation(hash)
}

func (s *service) GetAnswer(hash string) (*explorer.Answer, error) {
	return s.consultationRepository.Answer(hash)
}

func (s *service) GetAnswerVotes(consultationHash string, hash string) ([]*entity.CfundVote, error) {
	log.Debugf("GetAnswerVotes(hash:%s)", hash)

	consultation, err := s.GetConsultation(consultationHash)
	if err != nil {
		return nil, err
	}

	var answer *explorer.Answer
	for _, a := range consultation.Answers {
		if a.Hash == hash {
			answer = a
		}
	}
	if answer == nil {

	}

	votingCycles, err := s.GetVotingCycles(consultation, uint(consultation.VotingCyclesFromCreation))
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(explorer.DaoVote, hash, votingCycles)
}

func (s *service) GetConsensusConsultations(config *pagination.Config) ([]*explorer.Consultation, int64, error) {
	return s.consultationRepository.ConsensusConsultations(config.Ascending, config.Size, config.Page)
}

func (s *service) getMax(status string, stateChangedOnBlock string) (uint, error) {
	var block *explorer.Block
	var err error

	if status == explorer.ProposalPending.Status {
		block, err = s.blockRepository.BestBlock()
	} else {
		block, err = s.blockRepository.BlockByHash(stateChangedOnBlock)
	}

	if err != nil {
		return 0, err
	}

	log.Info("max ", block.Height)
	return uint(block.Height), nil
}
