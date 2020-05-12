package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type DaoProposalRepository struct {
	elastic *elastic_cache.Index
}

var (
	ErrProposalNotFound = errors.New("Proposal not found")
)

func NewDaoProposalRepository(elastic *elastic_cache.Index) *DaoProposalRepository {
	return &DaoProposalRepository{elastic}
}

func (r *DaoProposalRepository) Proposals(status *explorer.ProposalStatus, dir bool, size int, page int) ([]*explorer.Proposal, int64, error) {
	query := elastic.NewBoolQuery()
	if status != nil {
		query = query.Must(elastic.NewTermQuery("status.keyword", status.Status))
	}

	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get()).
		Query(query).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return r.findMany(results, err)
}

func (r *DaoProposalRepository) LegacyProposals(status *explorer.ProposalStatus, dir bool, size int, page int) ([]*entity.LegacyProposal, int64, error) {
	query := elastic.NewBoolQuery()
	if status != nil {
		query = query.Must(elastic.NewTermQuery("status.keyword", status))
	}

	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get()).
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

func (r *DaoProposalRepository) Proposal(hash string) (*explorer.Proposal, error) {
	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *DaoProposalRepository) ValueLocked() (*float64, error) {

	query := elastic.NewBoolQuery()
	//query = query.Must(elastic.NewMatchQuery("status", explorer.ProposalAccepted))
	query = query.Should(elastic.NewMatchQuery("state", explorer.ProposalAccepted.State))
	query = query.Should(elastic.NewMatchQuery("state", explorer.ProposalPendingVotingPreq.State))

	lockedAgg := elastic.NewFilterAggregation().Filter(query)
	lockedAgg.SubAggregation("notPaidYet", elastic.NewSumAggregation().Field("notPaidYet"))

	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get()).
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

func (r *DaoProposalRepository) findOne(results *elastic.SearchResult, err error) (*explorer.Proposal, error) {
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

func (r *DaoProposalRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.Proposal, int64, error) {
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
