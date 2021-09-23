package dao

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetBlockCycleByHeight(n network.Network, height uint64) (*entity.LegacyBlockCycle, error)
	GetBlockCycleByBlock(n network.Network, block *explorer.Block) (*entity.LegacyBlockCycle, error)
	GetConsensus(n network.Network) (explorer.ConsensusParameters, error)
	GetCfundStats(n network.Network) (*entity.CfundStats, error)
	GetExcludedVotes(n network.Network, cycle uint) (uint, error)

	GetProposals(n network.Network, parameters ProposalParameters, pagination framework.Pagination) ([]*explorer.Proposal, int64, error)
	GetProposal(n network.Network, hash string) (*explorer.Proposal, error)
	GetVotingCycles(n network.Network, element explorer.ChainHeight, count uint) ([]*entity.VotingCycle, error)
	GetProposalVotes(n network.Network, hash string) ([]*entity.CfundVote, []*entity.VotingCycle, error)
	GetProposalTrend(n network.Network, hash string) ([]*entity.CfundTrend, error)

	GetPaymentRequests(n network.Network, parameters PaymentRequestParameters, pagination framework.Pagination) ([]*explorer.PaymentRequest, int64, error)
	GetPaymentRequestsForProposal(n network.Network, proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error)
	GetPaymentRequest(n network.Network, hash string) (*explorer.PaymentRequest, error)
	GetPaymentRequestVotes(n network.Network, hash string) ([]*entity.CfundVote, []*entity.VotingCycle, error)
	GetPaymentRequestTrend(n network.Network, hash string) ([]*entity.CfundTrend, error)

	GetConsultations(n network.Network, parameters ConsultationParameters, pagination framework.Pagination) ([]*explorer.Consultation, int64, error)
	GetConsultation(n network.Network, hash string) (*explorer.Consultation, error)
	GetAnswer(n network.Network, hash string) (*explorer.Answer, error)
	GetAnswerVotes(n network.Network, consultationHash string, hash string) ([]*entity.CfundVote, []*entity.VotingCycle, error)
	GetConsensusConsultations(n network.Network, pagination framework.Pagination) ([]*explorer.Consultation, int64, error)
}

type service struct {
	consensusService           consensus.Service
	proposalRepository         repository.DaoProposalRepository
	paymentRequestRepository   repository.DaoPaymentRequestRepository
	consultationRepository     repository.DaoConsultationRepository
	consensusRepository        repository.DaoConsensusRepository
	voteRepository             repository.DaoVoteRepository
	blockRepository            repository.BlockRepository
	blockTransactionRepository repository.BlockTransactionRepository
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
	proposalRepository repository.DaoProposalRepository,
	paymentRequestRepository repository.DaoPaymentRequestRepository,
	consultationRepository repository.DaoConsultationRepository,
	consensusRepository repository.DaoConsensusRepository,
	voteRepository repository.DaoVoteRepository,
	blockRepository repository.BlockRepository,
	blockTransactionRepository repository.BlockTransactionRepository,
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

func (s *service) GetBlockCycleByHeight(n network.Network, height uint64) (*entity.LegacyBlockCycle, error) {
	return s.GetBlockCycleByBlock(n, &explorer.Block{RawBlock: explorer.RawBlock{Height: height}})
}

func (s *service) GetBlockCycleByBlock(n network.Network, block *explorer.Block) (*entity.LegacyBlockCycle, error) {
	blockCycle := &entity.LegacyBlockCycle{
		BlocksInCycle: uint(s.consensusService.GetParameter(n, consensus.VOTING_CYCLE_LENGTH).Value),
		ProposalVoting: entity.Voting{
			Quorum: float64(s.consensusService.GetParameter(n, consensus.PROPOSAL_MIN_QUORUM).Value) / 100,
			Cycles: uint(s.consensusService.GetParameter(n, consensus.PROPOSAL_MAX_VOTING_CYCLES).Value),
			Accept: uint(s.consensusService.GetParameter(n, consensus.PROPOSAL_MIN_ACCEPT).Value),
			Reject: uint(s.consensusService.GetParameter(n, consensus.PROPOSAL_MIN_REJECT).Value),
		},
		PaymentVoting: entity.Voting{
			Quorum: float64(s.consensusService.GetParameter(n, consensus.PAYMENT_REQUEST_MIN_QUORUM).Value) / 100,
			Cycles: uint(s.consensusService.GetParameter(n, consensus.PAYMENT_REQUEST_MAX_VOTING_CYCLES).Value),
			Accept: uint(s.consensusService.GetParameter(n, consensus.PAYMENT_REQUEST_MIN_ACCEPT).Value),
			Reject: uint(s.consensusService.GetParameter(n, consensus.PAYMENT_REQUEST_MIN_REJECT).Value),
		},
		Cycle:      block.BlockCycle.Cycle,
		FirstBlock: (block.BlockCycle.Cycle * block.BlockCycle.Size) - block.BlockCycle.Size,
	}
	blockCycle.CurrentBlock = uint(block.Height)
	blockCycle.BlocksRemaining = blockCycle.BlocksInCycle + blockCycle.FirstBlock - blockCycle.CurrentBlock - 1

	return blockCycle, nil
}

func (s *service) GetConsensus(n network.Network) (explorer.ConsensusParameters, error) {
	return s.consensusService.GetParameters(n)
}

func (s *service) GetCfundStats(n network.Network) (*entity.CfundStats, error) {
	cfundStats := new(entity.CfundStats)

	if block, _ := s.blockRepository.GetBestBlock(n); block != nil {
		cfundStats.Available = block.Cfund.Available
		cfundStats.Locked = block.Cfund.Locked
	}

	if paid, _ := s.paymentRequestRepository.GetValuePaid(n); paid != nil {
		cfundStats.Paid = *paid
	}

	return cfundStats, nil
}

func (s *service) GetExcludedVotes(n network.Network, cycle uint) (uint, error) {
	return s.voteRepository.GetExcludedVotes(n, cycle)
}

func (s *service) GetProposals(n network.Network, parameters ProposalParameters, pagination framework.Pagination) ([]*explorer.Proposal, int64, error) {
	var status *explorer.ProposalStatus
	if parameters.State != nil && explorer.IsProposalStateValid(*parameters.State) {
		s := explorer.GetProposalStatusByState(*parameters.State)
		status = &s
	}

	proposals, total, err := s.proposalRepository.GetProposals(n, status, false, pagination.Size(), pagination.Page())
	if err == nil {
		for _, proposal := range proposals {
			proposal.VotesExcluded = s.getExcludedVotesForProposal(n, *proposal)
		}
	}

	return proposals, total, err
}

func (s *service) GetProposal(n network.Network, hash string) (*explorer.Proposal, error) {
	proposal, err := s.proposalRepository.GetProposal(n, hash)
	if err == nil {
		proposal.VotesExcluded = s.getExcludedVotesForProposal(n, *proposal)
	}

	return proposal, err
}

func (s *service) GetVotingCycles(n network.Network, element explorer.ChainHeight, count uint) ([]*entity.VotingCycle, error) {
	log.Infof("GetVotingCycles for %T", element)

	block, err := s.blockRepository.GetBlockByHeight(n, element.GetHeight())
	if err != nil {
		return nil, err
	}

	var segments uint
	switch e := element.(type) {
	case *explorer.Proposal:
		segments = uint(s.consensusService.GetParameter(n, consensus.PROPOSAL_MAX_VOTING_CYCLES).Value) + 2
	case *explorer.PaymentRequest:
		segments = uint(s.consensusService.GetParameter(n, consensus.PAYMENT_REQUEST_MAX_VOTING_CYCLES).Value) + 2
	case *explorer.Consultation:
		segments = uint(s.consensusService.GetParameter(n, consensus.CONSULTATION_MAX_VOTING_CYCLES).Value) + 2
	default:
		log.Fatalf("Unable to get Max voting cycles from %T", e)
	}

	return entity.CreateVotingCycles(
		segments,
		uint(s.consensusService.GetParameter(n, consensus.VOTING_CYCLE_LENGTH).Value),
		uint(block.Height)-block.BlockCycle.Index,
		count+1,
	), nil
}

func (s *service) GetProposalVotes(n network.Network, hash string) ([]*entity.CfundVote, []*entity.VotingCycle, error) {
	log.WithField("hash", hash).Info("GetProposalVotes")

	proposal, err := s.GetProposal(n, hash)
	if err != nil {
		return nil, nil, err
	}

	votingCycles, err := s.GetVotingCycles(n, proposal, proposal.VotingCycle)
	if err != nil {
		return nil, nil, err
	}

	votes, err := s.voteRepository.GetVotes(n, explorer.ProposalVote, hash, votingCycles)
	if err != nil {
		return nil, nil, err
	}

	return votes, votingCycles, err
}

func (s *service) GetProposalTrend(n network.Network, hash string) ([]*entity.CfundTrend, error) {
	log.WithField("hash", hash).Info("GetProposalTrend")

	proposal, err := s.GetProposal(n, hash)
	if err != nil {
		return nil, err
	}

	max, err := s.getMax(n, proposal.Status, proposal.StateChangedOnBlock)
	if err != nil {
		return nil, err
	}

	size := uint(s.consensusService.GetParameter(n, consensus.VOTING_CYCLE_LENGTH).Value)

	cfundVotes, err := s.voteRepository.GetVotes(
		n,
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
				Exclude: cfundVote.Exclude,
			},
			Trend: entity.Votes{
				Yes:     int(float64(cfundVote.Yes*10) / float64(size-uint(cfundVote.Exclude*10)) * 100),
				No:      int(float64(cfundVote.No*10) / float64(size-uint(cfundVote.Exclude*10)) * 100),
				Abstain: int(float64(cfundVote.Abstain*10) / float64(size-uint(cfundVote.Exclude*10)) * 100),
				Exclude: int(float64(cfundVote.Exclude*10) / float64(size) * 100),
			},
		}
		cfundTrends = append(cfundTrends, cfundTrend)
	}

	return cfundTrends, nil
}

func (s *service) GetPaymentRequests(n network.Network, parameters PaymentRequestParameters, pagination framework.Pagination) ([]*explorer.PaymentRequest, int64, error) {
	var status *explorer.PaymentRequestStatus
	if parameters.State != nil && explorer.IsPaymentRequestStateValid(*parameters.State) {
		s := explorer.GetPaymentRequestStatusByState(*parameters.State)
		status = &s
	}

	var hash string
	if parameters.Proposal != "" {
		hash = parameters.Proposal
	}

	paymentRequests, total, err := s.paymentRequestRepository.GetPaymentRequests(n, hash, status, false, pagination.Size(), pagination.Page())
	if err == nil {
		for _, paymentRequest := range paymentRequests {
			paymentRequest.VotesExcluded = s.getExcludedVotesForPaymentRequest(n, *paymentRequest)
		}
	}

	return paymentRequests, total, err
}

func (s *service) GetPaymentRequestsForProposal(n network.Network, proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error) {
	paymentRequests, err := s.paymentRequestRepository.GetPaymentRequestsForProposal(n, proposal)

	for _, paymentRequest := range paymentRequests {
		paymentRequest.VotesExcluded = s.getExcludedVotesForPaymentRequest(n, *paymentRequest)
	}

	return paymentRequests, err
}

func (s *service) GetPaymentRequest(n network.Network, hash string) (*explorer.PaymentRequest, error) {
	paymentRequest, err := s.paymentRequestRepository.GetPaymentRequest(n, hash)
	if err == nil {
		paymentRequest.VotesExcluded = s.getExcludedVotesForPaymentRequest(n, *paymentRequest)
	}
	return paymentRequest, err
}

func (s *service) GetPaymentRequestVotes(n network.Network, hash string) ([]*entity.CfundVote, []*entity.VotingCycle, error) {
	paymentRequest, err := s.GetPaymentRequest(n, hash)
	if err != nil {
		return nil, nil, err
	}

	votingCycles, err := s.GetVotingCycles(n, paymentRequest, paymentRequest.VotingCycle)
	if err != nil {
		return nil, nil, err
	}

	votes, err := s.voteRepository.GetVotes(n, explorer.PaymentRequestVote, hash, votingCycles)
	if err != nil {
		return nil, nil, err
	}

	return votes, votingCycles, err
}

func (s *service) GetPaymentRequestTrend(n network.Network, hash string) ([]*entity.CfundTrend, error) {
	log.Debugf("GetPaymentRequestTrend(hash:%s)", hash)
	paymentRequest, err := s.GetPaymentRequest(n, hash)
	if err != nil {
		return nil, err
	}

	max, err := s.getMax(n, paymentRequest.Status, paymentRequest.StateChangedOnBlock)
	if err != nil {
		return nil, err
	}

	size := uint(s.consensusService.GetParameter(n, consensus.VOTING_CYCLE_LENGTH).Value)

	cfundVotes, err := s.voteRepository.GetVotes(
		n,
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
				Exclude: cfundVote.Exclude,
			},
			Trend: entity.Votes{
				Yes:     int(float64(cfundVote.Yes*10) / float64(size-uint(cfundVote.Exclude*10)) * 100),
				No:      int(float64(cfundVote.No*10) / float64(size-uint(cfundVote.Exclude*10)) * 100),
				Abstain: int(float64(cfundVote.Abstain*10) / float64(size-uint(cfundVote.Exclude*10)) * 100),
				Exclude: int(float64(cfundVote.Exclude*10) / float64(size) * 100),
			},
		}
		cfundTrends = append(cfundTrends, cfundTrend)
	}

	return cfundTrends, nil
}

func (s *service) GetConsultations(n network.Network, parameters ConsultationParameters, pagination framework.Pagination) ([]*explorer.Consultation, int64, error) {
	if parameters.State != nil && explorer.IsConsultationStateValid(*parameters.State) {
		s := explorer.GetConsultationStatusByState(*parameters.State)
		parameters.Status = &s
	}

	return s.consultationRepository.GetConsultations(n, parameters.Status, parameters.Consensus, parameters.Min, false, pagination.Size(), pagination.Page())
}

func (s *service) GetConsultation(n network.Network, hash string) (*explorer.Consultation, error) {
	return s.consultationRepository.GetConsultation(n, hash)
}

func (s *service) GetAnswer(n network.Network, hash string) (*explorer.Answer, error) {
	return s.consultationRepository.GetAnswer(n, hash)
}

func (s *service) GetAnswerVotes(n network.Network, consultationHash string, hash string) ([]*entity.CfundVote, []*entity.VotingCycle, error) {
	consultation, err := s.GetConsultation(n, consultationHash)
	if err != nil {
		return nil, nil, err
	}

	votingCycles, err := s.GetVotingCycles(n, consultation, uint(consultation.VotingCyclesFromCreation))
	if err != nil {
		return nil, nil, err
	}

	votes, err := s.voteRepository.GetVotes(n, explorer.DaoVote, hash, votingCycles)
	if err != nil {
		return nil, nil, err
	}

	return votes, votingCycles, err
}

func (s *service) GetConsensusConsultations(n network.Network, pagination framework.Pagination) ([]*explorer.Consultation, int64, error) {
	return s.consultationRepository.GetConsensusConsultations(n, false, pagination.Size(), pagination.Page())
}

func (s *service) getMax(n network.Network, status string, stateChangedOnBlock string) (uint, error) {
	var block *explorer.Block
	var err error

	if status == explorer.ProposalPending.Status {
		block, err = s.blockRepository.GetBestBlock(n)
	} else {
		block, err = s.blockRepository.GetBlockByHash(n, stateChangedOnBlock)
	}

	if err != nil {
		return 0, err
	}

	log.Info("max ", block.Height)
	return uint(block.Height), nil
}

func (s *service) getExcludedVotesForProposal(n network.Network, proposal explorer.Proposal) uint {
	creationBlock, err := s.blockRepository.GetBlockByHeight(n, proposal.Height)
	if err != nil {
		return 0
	}

	excluded, err := s.voteRepository.GetExcludedVotes(n, creationBlock.BlockCycle.Cycle+proposal.VotingCycle)
	if err != nil {
		return 0
	}

	return excluded
}

func (s *service) getExcludedVotesForPaymentRequest(n network.Network, paymentRequest explorer.PaymentRequest) uint {
	creationBlock, err := s.blockRepository.GetBlockByHeight(n, paymentRequest.Height)
	if err != nil {
		return 0
	}

	excluded, err := s.voteRepository.GetExcludedVotes(n, creationBlock.BlockCycle.Cycle+paymentRequest.VotingCycle)
	if err != nil {
		return 0
	}

	return excluded
}
