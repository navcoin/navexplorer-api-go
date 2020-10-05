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
	GetBlockCycleByHeight(network string, height uint64) (*entity.LegacyBlockCycle, error)
	GetBlockCycleByBlock(network string, block *explorer.Block) (*entity.LegacyBlockCycle, error)
	GetConsensus(network string) (*explorer.ConsensusParameters, error)
	GetCfundStats(network string) (*entity.CfundStats, error)

	GetProposals(network string, parameters ProposalParameters, config *pagination.Config) ([]*explorer.Proposal, int64, error)
	GetProposal(network, hash string) (*explorer.Proposal, error)
	GetVotingCycles(network string, element explorer.ChainHeight, count uint) ([]*entity.VotingCycle, error)
	GetProposalVotes(network, hash string) ([]*entity.CfundVote, error)
	GetProposalTrend(network, hash string) ([]*entity.CfundTrend, error)

	GetPaymentRequests(network string, parameters PaymentRequestParameters, config *pagination.Config) ([]*explorer.PaymentRequest, int64, error)
	GetPaymentRequestsForProposal(network string, proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error)
	GetPaymentRequest(network, hash string) (*explorer.PaymentRequest, error)
	GetPaymentRequestVotes(network, hash string) ([]*entity.CfundVote, error)
	GetPaymentRequestTrend(network, hash string) ([]*entity.CfundTrend, error)

	GetConsultations(network string, parameters ConsultationParameters, config *pagination.Config) ([]*explorer.Consultation, int64, error)
	GetConsultation(network, hash string) (*explorer.Consultation, error)
	GetAnswer(network, hash string) (*explorer.Answer, error)
	GetAnswerVotes(network, consultationHash string, hash string) ([]*entity.CfundVote, error)
	GetConsensusConsultations(network string, config *pagination.Config) ([]*explorer.Consultation, int64, error)
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

func (s *service) GetBlockCycleByHeight(network string, height uint64) (*entity.LegacyBlockCycle, error) {
	return s.GetBlockCycleByBlock(network, &explorer.Block{RawBlock: explorer.RawBlock{Height: height}})
}

func (s *service) GetBlockCycleByBlock(network string, block *explorer.Block) (*entity.LegacyBlockCycle, error) {
	blockCycle := &entity.LegacyBlockCycle{
		BlocksInCycle: uint(s.consensusService.GetParameter(network, consensus.VOTING_CYCLE_LENGTH).Value),
		ProposalVoting: entity.Voting{
			Quorum: float64(s.consensusService.GetParameter(network, consensus.PROPOSAL_MIN_QUORUM).Value) / 100,
			Cycles: uint(s.consensusService.GetParameter(network, consensus.PROPOSAL_MAX_VOTING_CYCLES).Value),
			Accept: uint(s.consensusService.GetParameter(network, consensus.PROPOSAL_MIN_ACCEPT).Value),
			Reject: uint(s.consensusService.GetParameter(network, consensus.PROPOSAL_MIN_REJECT).Value),
		},
		PaymentVoting: entity.Voting{
			Quorum: float64(s.consensusService.GetParameter(network, consensus.PAYMENT_REQUEST_MIN_QUORUM).Value) / 100,
			Cycles: uint(s.consensusService.GetParameter(network, consensus.PAYMENT_REQUEST_MAX_VOTING_CYCLES).Value),
			Accept: uint(s.consensusService.GetParameter(network, consensus.PAYMENT_REQUEST_MIN_ACCEPT).Value),
			Reject: uint(s.consensusService.GetParameter(network, consensus.PAYMENT_REQUEST_MIN_REJECT).Value),
		},
		Cycle:      block.BlockCycle.Cycle,
		FirstBlock: (block.BlockCycle.Cycle * block.BlockCycle.Size) - block.BlockCycle.Size,
	}
	blockCycle.CurrentBlock = uint(block.Height)
	blockCycle.BlocksRemaining = blockCycle.BlocksInCycle + blockCycle.FirstBlock - blockCycle.CurrentBlock - 1

	return blockCycle, nil
}

func (s *service) GetConsensus(network string) (*explorer.ConsensusParameters, error) {
	return s.consensusService.GetParameters(network)
}

func (s *service) GetCfundStats(network string) (*entity.CfundStats, error) {
	cfundStats := new(entity.CfundStats)

	if block, _ := s.blockRepository.Network(network).BestBlock(); block != nil {
		cfundStats.Available = block.Cfund.Available
		cfundStats.Locked = block.Cfund.Locked
	}

	if paid, _ := s.paymentRequestRepository.Network(network).ValuePaid(); paid != nil {
		cfundStats.Paid = *paid
	}

	return cfundStats, nil
}

func (s *service) GetProposals(network string, parameters ProposalParameters, config *pagination.Config) ([]*explorer.Proposal, int64, error) {
	var status *explorer.ProposalStatus
	if parameters.State != nil && explorer.IsProposalStateValid(*parameters.State) {
		s := explorer.GetProposalStatusByState(*parameters.State)
		status = &s
	}

	return s.proposalRepository.Network(network).Proposals(status, config.Ascending, config.Size, config.Page)
}

func (s *service) GetProposal(network, hash string) (*explorer.Proposal, error) {
	return s.proposalRepository.Network(network).Proposal(hash)
}

func (s *service) GetVotingCycles(network string, element explorer.ChainHeight, count uint) ([]*entity.VotingCycle, error) {
	log.Infof("GetVotingCycles for %T", element)

	block, err := s.blockRepository.Network(network).BlockByHeight(element.GetHeight())
	if err != nil {
		return nil, err
	}

	var segments uint
	switch e := element.(type) {
	case *explorer.Proposal:
		segments = uint(s.consensusService.GetParameter(network, consensus.PROPOSAL_MAX_VOTING_CYCLES).Value) + 2
	case *explorer.PaymentRequest:
		segments = uint(s.consensusService.GetParameter(network, consensus.PAYMENT_REQUEST_MAX_VOTING_CYCLES).Value) + 2
	case *explorer.Consultation:
		segments = uint(s.consensusService.GetParameter(network, consensus.CONSULTATION_MAX_VOTING_CYCLES).Value) + 2
	default:
		log.Fatalf("Unable to get Max voting cycles from %T", e)
	}

	return entity.CreateVotingCycles(
		segments,
		uint(s.consensusService.GetParameter(network, consensus.VOTING_CYCLE_LENGTH).Value),
		uint(block.Height)-block.BlockCycle.Index,
		count+1,
	), nil
}

func (s *service) GetProposalVotes(network, hash string) ([]*entity.CfundVote, error) {
	log.WithField("hash", hash).Info("GetProposalVotes")

	proposal, err := s.GetProposal(network, hash)
	if err != nil {
		return nil, err
	}

	votingCycles, err := s.GetVotingCycles(network, proposal, proposal.VotingCycle)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.Network(network).GetVotes(explorer.ProposalVote, hash, votingCycles)
}

func (s *service) GetProposalTrend(network, hash string) ([]*entity.CfundTrend, error) {
	log.WithField("hash", hash).Info("GetProposalTrend")

	proposal, err := s.GetProposal(network, hash)
	if err != nil {
		return nil, err
	}

	max, err := s.getMax(network, proposal.Status, proposal.StateChangedOnBlock)
	if err != nil {
		return nil, err
	}

	size := uint(s.consensusService.GetParameter(network, consensus.VOTING_CYCLE_LENGTH).Value)

	cfundVotes, err := s.voteRepository.Network(network).GetVotes(
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

func (s *service) GetPaymentRequests(network string, parameters PaymentRequestParameters, config *pagination.Config) ([]*explorer.PaymentRequest, int64, error) {
	var status *explorer.PaymentRequestStatus
	if parameters.State != nil && explorer.IsPaymentRequestStateValid(*parameters.State) {
		s := explorer.GetPaymentRequestStatusByState(*parameters.State)
		status = &s
	}

	var proposalHash string
	if parameters.Proposal != "" {
		proposalHash = parameters.Proposal
	}

	return s.paymentRequestRepository.Network(network).PaymentRequests(proposalHash, status, config.Ascending, config.Size, config.Page)
}

func (s *service) GetPaymentRequestsForProposal(network string, proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.Network(network).PaymentRequestsForProposal(proposal)
}

func (s *service) GetPaymentRequest(network, hash string) (*explorer.PaymentRequest, error) {
	return s.paymentRequestRepository.Network(network).PaymentRequest(hash)
}

func (s *service) GetPaymentRequestVotes(network, hash string) ([]*entity.CfundVote, error) {
	log.Debugf("GetPaymentRequestVotes(hash:%s)", hash)

	paymentRequest, err := s.GetPaymentRequest(network, hash)
	if err != nil {
		return nil, err
	}

	votingCycles, err := s.GetVotingCycles(network, paymentRequest, paymentRequest.VotingCycle)
	if err != nil {
		return nil, err
	}

	return s.voteRepository.GetVotes(explorer.PaymentRequestVote, hash, votingCycles)
}

func (s *service) GetPaymentRequestTrend(network, hash string) ([]*entity.CfundTrend, error) {
	log.Debugf("GetPaymentRequestTrend(hash:%s)", hash)
	paymentRequest, err := s.GetPaymentRequest(network, hash)
	if err != nil {
		return nil, err
	}

	max, err := s.getMax(network, paymentRequest.Status, paymentRequest.StateChangedOnBlock)
	if err != nil {
		return nil, err
	}

	size := uint(s.consensusService.GetParameter(network, consensus.VOTING_CYCLE_LENGTH).Value)

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

func (s *service) GetConsultations(network string, parameters ConsultationParameters, config *pagination.Config) ([]*explorer.Consultation, int64, error) {
	if parameters.State != nil && explorer.IsConsultationStateValid(*parameters.State) {
		s := explorer.GetConsultationStatusByState(*parameters.State)
		parameters.Status = &s
	}

	return s.consultationRepository.
		Network(network).
		Consultations(parameters.Status, parameters.Consensus, parameters.Min, config.Ascending, config.Size, config.Page)
}

func (s *service) GetConsultation(network, hash string) (*explorer.Consultation, error) {
	return s.consultationRepository.Network(network).Consultation(hash)
}

func (s *service) GetAnswer(network, hash string) (*explorer.Answer, error) {
	return s.consultationRepository.Network(network).Answer(hash)
}

func (s *service) GetAnswerVotes(network, consultationHash string, hash string) ([]*entity.CfundVote, error) {
	log.Debugf("GetAnswerVotes(hash:%s)", hash)

	consultation, err := s.GetConsultation(network, consultationHash)
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

	votingCycles, err := s.GetVotingCycles(network, consultation, uint(consultation.VotingCyclesFromCreation))
	if err != nil {
		return nil, err
	}

	return s.voteRepository.Network(network).GetVotes(explorer.DaoVote, hash, votingCycles)
}

func (s *service) GetConsensusConsultations(network string, config *pagination.Config) ([]*explorer.Consultation, int64, error) {
	return s.consultationRepository.Network(network).ConsensusConsultations(config.Ascending, config.Size, config.Page)
}

func (s *service) getMax(network, status string, stateChangedOnBlock string) (uint, error) {
	var block *explorer.Block
	var err error

	if status == explorer.ProposalPending.Status {
		block, err = s.blockRepository.Network(network).BestBlock()
	} else {
		block, err = s.blockRepository.Network(network).BlockByHash(stateChangedOnBlock)
	}

	if err != nil {
		return 0, err
	}

	log.Info("max ", block.Height)
	return uint(block.Height), nil
}
