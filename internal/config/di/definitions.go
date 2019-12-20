package di

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
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
		Name: "dao.proposal.repo",
		Build: func(elastic *elastic_cache.Index) (*repository.DaoProposalRepository, error) {
			return repository.NewDaoProposalRepository(elastic), nil
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
}
