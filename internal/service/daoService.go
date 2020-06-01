package service

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type DaoService interface {
	GetBlockCycleByHeight(height uint64) (*entity.LegacyBlockCycle, error)
	GetBlockCycleByBlock(block *explorer.Block) (*entity.LegacyBlockCycle, error)
	GetConsensus() (*explorer.ConsensusParameters, error)
	GetCfundStats() (*entity.CfundStats, error)
	GetProposals(parameters dao.ProposalParameters, config *pagination.Config) ([]*explorer.Proposal, int64, error)
	GetProposal(hash string) (*explorer.Proposal, error)
	GetVotingCycles(element explorer.ChainHeight, max uint) ([]*entity.VotingCycle, error)
	GetProposalVotes(hash string) ([]*entity.CfundVote, error)
	GetProposalTrend(hash string) ([]*entity.CfundTrend, error)
	GetPaymentRequests(parameters dao.PaymentRequestParameters, config *pagination.Config) ([]*explorer.PaymentRequest, int64, error)
	GetPaymentRequestsForProposal(proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error)
	GetPaymentRequest(hash string) (*explorer.PaymentRequest, error)
	GetPaymentRequestVotes(hash string) ([]*entity.CfundVote, error)
	GetPaymentRequestTrend(hash string) ([]*entity.CfundTrend, error)
	GetConsultations(parameters dao.ConsultationParameters, config *pagination.Config) ([]*explorer.Consultation, int64, error)
	GetConsultation(hash string) (*explorer.Consultation, error)
	GetConsensusConsultations(config *pagination.Config) ([]*explorer.Consultation, int64, error)
	GetAnswer(hash string) (*explorer.Answer, error)
	GetAnswerVotes(consultationHash string, hash string) ([]*entity.CfundVote, error)
}
