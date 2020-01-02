package di

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
)

var Definitions = []dingo.Def{
	{
		Name: "elastic",
		Build: func() (*elastic_cache.Index, error) {
			elastic, err := elastic_cache.New()
			if err != nil {
				log.WithError(err).Fatal("Failed toStart ES")
			}

			return elastic, nil
		},
	},
	{
		Name: "address.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.AddressRepository, error) {
			return repository.NewAddressRepository(elastic), nil
		},
	},
	{
		Name: "address.transaction.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.AddressTransactionRepository, error) {
			return repository.NewAddressTransactionRepository(elastic), nil
		},
	},
	{
		Name: "address.service",
		Build: func(addressRepository *repository.AddressRepository, addressTransactionRepository *repository.AddressTransactionRepository) (*address.AddressService, error) {
			return address.NewAddressService(addressRepository, addressTransactionRepository), nil
		},
	},
	{
		Name: "block.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.BlockRepository, error) {
			return repository.NewBlockRepository(elastic), nil
		},
	},
	{
		Name: "block.transaction.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.BlockTransactionRepository, error) {
			return repository.NewBlockTransactionRepository(elastic), nil
		},
	},
	{
		Name: "block.service",
		Build: func(blockRepository *repository.BlockRepository, blockTransactionRepository *repository.BlockTransactionRepository) (*block.BlockService, error) {
			return block.NewBlockService(blockRepository, blockTransactionRepository), nil
		},
	},
	{
		Name: "dao.proposal.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.DaoProposalRepository, error) {
			return repository.NewDaoProposalRepository(elastic), nil
		},
	},
	{
		Name: "dao.payment-request.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.DaoPaymentRequestRepository, error) {
			return repository.NewDaoPaymentRequestRepository(elastic), nil
		},
	},
	{
		Name: "dao.consensus.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.DaoConsensusRepository, error) {
			return repository.NewDaoConsensusRepository(elastic), nil
		},
	},
	{
		Name: "softfork.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.SoftForkRepository, error) {
			return repository.NewSoftForkRepository(elastic), nil
		},
	},
	{
		Name: "dao.service",
		Build: func(
			proposalRepo *repository.DaoProposalRepository,
			paymentRequestRepo *repository.DaoPaymentRequestRepository,
			consensusRepo *repository.DaoConsensusRepository,
			blockTxRepo *repository.BlockTransactionRepository,
		) (*dao.DaoService, error) {
			return dao.NewDaoService(proposalRepo, paymentRequestRepo, consensusRepo, blockTxRepo), nil
		},
	},
}
