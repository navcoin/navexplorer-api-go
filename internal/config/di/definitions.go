package di

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/event"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/address"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/coin"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/dao"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/dao/consensus"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/distribution"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/softfork"
	"github.com/sarulabs/dingo/v3"
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
			return repository.NewCachingAddressHistoryRepository(
				repository.NewAddressHistoryRepository(elastic),
				cache,
			), nil
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
			return repository.NewCachingBlockRepository(
				repository.NewBlockRepository(elastic),
				cache,
			), nil
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
		Name: "event.consumer",
		Build: func() (*event.Consumer, error) {
			return event.NewConsumer(
				config.Get().RabbitMq.User,
				config.Get().RabbitMq.Password,
				config.Get().RabbitMq.Host,
				config.Get().RabbitMq.Port,
				config.Get().RabbitMq.Prefix,
			), nil
		},
	},
	{
		Name: "block.subscriber",
		Build: func(consumer *event.Consumer, cache2 *cache.Cache) (*block.Subscriber, error) {
			return block.NewBlockSubscriber(network.GetNetworks(), consumer, cache2), nil
		},
	},
	{
		Name: "coin.service",
		Build: func(addressRepository repository.AddressRepository) (coin.Service, error) {
			return coin.NewCoinService(addressRepository), nil
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
		Name: "distribution.service",
		Build: func(addressRepo repository.AddressRepository) (distribution.Service, error) {
			return distribution.NewDistributionService(addressRepo), nil
		},
	},
	{
		Name: "cache",
		Build: func() (*cache.Cache, error) {
			return cache.New(5*time.Minute, 10*time.Minute), nil
		},
	},
}
