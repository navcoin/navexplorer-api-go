package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/navcoin/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/dao/entity"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/network"
	"github.com/navcoin/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type DaoProposalRepository interface {
	GetProposals(n network.Network, status *explorer.ProposalStatus, dir bool, size int, page int) ([]*explorer.Proposal, int64, error)
	GetLegacyProposals(n network.Network, status *explorer.ProposalStatus, dir bool, size int, page int) ([]*entity.LegacyProposal, int64, error)
	GetProposal(n network.Network, hash string) (*explorer.Proposal, error)
	GetValueLocked(n network.Network) (*float64, error)
}

type daoProposalRepository struct {
	elastic *elastic_cache.Index
}

var (
	ErrProposalNotFound = errors.New("Proposal not found")
)

func NewDaoProposalRepository(elastic *elastic_cache.Index) DaoProposalRepository {
	return &daoProposalRepository{elastic: elastic}
}

func (r *daoProposalRepository) GetProposals(n network.Network, status *explorer.ProposalStatus, dir bool, size int, page int) ([]*explorer.Proposal, int64, error) {
	query := elastic.NewBoolQuery()
	if *status == explorer.ProposalAccepted {
		query = query.Should(elastic.NewTermQuery("status.keyword", status.Status))
		query = query.Should(elastic.NewTermQuery("status.keyword", explorer.ProposalPendingVotingPreq.Status))
		query = query.Should(elastic.NewTermQuery("status.keyword", explorer.ProposalAcceptedExpired.Status))
	} else {
		query = query.Must(elastic.NewTermQuery("status.keyword", status.Status))
	}

	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get(n)).
		Query(query).
		Sort("state.keyword", !dir).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return r.findMany(results, err)
}

func (r *daoProposalRepository) GetLegacyProposals(n network.Network, status *explorer.ProposalStatus, dir bool, size int, page int) ([]*entity.LegacyProposal, int64, error) {
	query := elastic.NewBoolQuery()
	if status != nil {
		query = query.Must(elastic.NewTermQuery("status.keyword", status))
	}

	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get(n)).
		Query(query).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	proposals := make([]*entity.LegacyProposal, 0)
	for _, hit := range results.Hits.Hits {
		var proposal *entity.LegacyProposal
		if err := json.Unmarshal(hit.Source, &proposal); err == nil {
			proposals = append(proposals, proposal)
		}
	}

	return proposals, results.TotalHits(), err
}

func (r *daoProposalRepository) GetProposal(n network.Network, hash string) (*explorer.Proposal, error) {
	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get(n)).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *daoProposalRepository) GetValueLocked(n network.Network) (*float64, error) {
	query := elastic.NewBoolQuery()
	query = query.Should(elastic.NewTermQuery("state.keyword", explorer.ProposalAccepted.State))
	query = query.Should(elastic.NewMatchQuery("state", explorer.ProposalPendingVotingPreq.State))

	lockedAgg := elastic.NewFilterAggregation().Filter(query)
	lockedAgg.SubAggregation("notPaidYet", elastic.NewSumAggregation().Field("notPaidYet"))

	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get(n)).
		Aggregation("locked", lockedAgg).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("Failed to get value details")
		return nil, err
	}

	if stats, found := results.Aggregations.Filter("locked"); found {
		if notPaidYet, found := stats.Aggregations.Sum("notPaidYet"); found {
			return notPaidYet.Value, nil
		}
	}

	return nil, errors.New("Could not find locked aggregation")
}

func (r *daoProposalRepository) findOne(results *elastic.SearchResult, err error) (*explorer.Proposal, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrProposalNotFound
		return nil, err
	}

	var proposal *explorer.Proposal
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &proposal)
	if err != nil {
		return nil, err
	}

	return proposal, err
}

func (r *daoProposalRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.Proposal, int64, error) {
	if err != nil {
		return nil, 0, err
	}

	proposals := make([]*explorer.Proposal, 0)
	for _, hit := range results.Hits.Hits {
		var proposal *explorer.Proposal
		if err := json.Unmarshal(hit.Source, &proposal); err == nil {
			proposals = append(proposals, proposal)
		}
	}

	return proposals, results.TotalHits(), err
}
