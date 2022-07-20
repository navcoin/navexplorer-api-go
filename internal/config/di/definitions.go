package di

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/cache"
	"github.com/navcoin/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/navcoin/navexplorer-api-go/v2/internal/repository"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/address"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/block"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/dao"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/dao/consensus"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/softfork"
	"github.com/sarulabs/dingo/v4"
	log "github.com/sirupsen/logrus"
	"time"
)

var Definitions = []dingo.Def{
	{
		Name: "elastic",
		Build: func() (*elastic_cache.Index, error) {
			elastic, err := elastic_cache.New()
			if err != nil {
				log.WithError(err).Fatal("Failed to start ES")
			}

			return elastic, nil
		},
	},
	{
		Name: "address.repo",
		Build: func(elastic *elastic_cache.Index) (repository.AddressRepository, error) {
			return repository.NewAddressRepository(elastic), nil
		},
	},
	{
		Name: "address.history.repo",
		Build: func(elastic *elastic_cache.Index, cache *cache.Cache) (repository.AddressHistoryRepository, error) {
			return repository.NewAddressHistoryRepository(elastic), nil
		},
	},
	{
		Name: "address.service",
		Build: func(
			addressRepository repository.AddressRepository,
			addressHistoryRepository repository.AddressHistoryRepository,
			blockRepository repository.BlockRepository,
			blockTransactionRepository repository.BlockTransactionRepository,
		) (address.Service, error) {
			return address.NewAddressService(addressRepository, addressHistoryRepository, blockRepository, blockTransactionRepository), nil
		},
	},
	{
		Name: "block.repo",
		Build: func(elastic *elastic_cache.Index, cache *cache.Cache) (repository.BlockRepository, error) {
			return repository.NewBlockRepository(elastic), nil
		},
	},
	{
		Name: "block.transaction.repo",
		Build: func(elastic *elastic_cache.Index) (repository.BlockTransactionRepository, error) {
			return repository.NewBlockTransactionRepository(elastic), nil
		},
	},
	{
		Name: "block.service",
		Build: func(blockRepository repository.BlockRepository, blockTransactionRepository repository.BlockTransactionRepository) (block.Service, error) {
			return block.NewBlockService(blockRepository, blockTransactionRepository), nil
		},
	},
	{
		Name: "dao.proposal.repo",
		Build: func(elastic *elastic_cache.Index) (repository.DaoProposalRepository, error) {
			return repository.NewDaoProposalRepository(elastic), nil
		},
	},
	{
		Name: "dao.payment-request.repo",
		Build: func(elastic *elastic_cache.Index) (repository.DaoPaymentRequestRepository, error) {
			return repository.NewDaoPaymentRequestRepository(elastic), nil
		},
	},
	{
		Name: "dao.vote.repo",
		Build: func(elastic *elastic_cache.Index) (repository.DaoVoteRepository, error) {
			return repository.NewDaoVoteRepository(elastic), nil
		},
	},
	{
		Name: "dao.consultation.repo",
		Build: func(elastic *elastic_cache.Index) (repository.DaoConsultationRepository, error) {
			return repository.NewDaoConsultationRepository(elastic), nil
		},
	},
	{
		Name: "dao.consensus.repo",
		Build: func(elastic *elastic_cache.Index) (repository.DaoConsensusRepository, error) {
			return repository.NewDaoConsensusRepository(elastic), nil
		},
	},
	{
		Name: "softfork.repo",
		Build: func(elastic *elastic_cache.Index) (repository.SoftForkRepository, error) {
			return repository.NewSoftForkRepository(elastic), nil
		},
	},
	{
		Name: "softfork.service",
		Build: func(blockRepository repository.BlockRepository, softforkRepository repository.SoftForkRepository) (softfork.Service, error) {
			return softfork.NewSoftForkService(blockRepository, softforkRepository), nil
		},
	},
	{
		Name: "dao.service",
		Build: func(
			consensusService consensus.Service,
			proposalRepo repository.DaoProposalRepository,
			paymentRequestRepo repository.DaoPaymentRequestRepository,
			consultationRepo repository.DaoConsultationRepository,
			consensusRepo repository.DaoConsensusRepository,
			voteRepo repository.DaoVoteRepository,
			blockRepo repository.BlockRepository,
			blockTxRepo repository.BlockTransactionRepository,
		) (dao.Service, error) {
			return dao.NewDaoService(consensusService, proposalRepo, paymentRequestRepo, consultationRepo, consensusRepo, voteRepo, blockRepo, blockTxRepo), nil
		},
	},
	{
		Name: "dao.consensus.service",
		Build: func(consensusRepo repository.DaoConsensusRepository) (consensus.Service, error) {
			return consensus.NewConsensusService(consensusRepo), nil
		},
	},
	{
		Name: "staking.service",
		Build: func(addressHistoryRepo repository.AddressHistoryRepository) (service.StakingService, error) {
			return service.NewStakingService(addressHistoryRepo), nil
		},
	},
	{
		Name: "cache",
		Build: func() (*cache.Cache, error) {
			return cache.New(5*time.Minute, 10*time.Minute), nil
		},
	},
}
