package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
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

func (r *DaoProposalRepository) Proposals(status explorer.ProposalStatus, dir bool, size int, page int) ([]*explorer.Proposal, int, error) {
	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get()).
		Query(elastic.NewTermQuery("status.keyword", status)).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return r.findMany(results, err)
}

func (r *DaoProposalRepository) Proposal(hash string) (*explorer.Proposal, error) {
	results, err := r.elastic.Client.Search(elastic_cache.ProposalIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
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

func (r *DaoProposalRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.Proposal, int, error) {
	if err != nil {
		return nil, 0, err
	}

	var proposals []*explorer.Proposal
	for _, hit := range results.Hits.Hits {
		var proposal *explorer.Proposal
		if err := json.Unmarshal(hit.Source, &proposal); err == nil {
			proposals = append(proposals, proposal)
		}
	}

	return proposals, int(results.Hits.TotalHits.Value), err
}
