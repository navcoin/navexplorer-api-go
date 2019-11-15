package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic"
)

type DaoProposalRepository struct {
	elastic *elastic_cache.Index
	index   string
}

type DaoProposalState string

var (
	ProposalPending  DaoProposalState = "pending"
	ProposalAccepted DaoProposalState = "accepted"
	ProposalExpired  DaoProposalState = "Expired"
)

var (
	ErrProposalNotFound = errors.New("Proposal not found")
)

func NewDaoProposalRepository(elastic *elastic_cache.Index, network string) *DaoProposalRepository {
	return &DaoProposalRepository{elastic, fmt.Sprintf("%s.%s", network, "proposal")}
}

func (r *DaoProposalRepository) StateFromString(state string) (*DaoProposalState, error) {
	switch true {
	case state == string(ProposalPending):
		return &ProposalPending, nil
	case state == string(ProposalAccepted):
		return &ProposalAccepted, nil
	case state == string(ProposalExpired):
		return &ProposalExpired, nil
	}

	return nil, errors.New(fmt.Sprintf("Proposal state %s not found", state))
}

func (r *DaoProposalRepository) Proposals(state DaoProposalState, dir bool, size int, page int) ([]*explorer.Proposal, int, error) {
	results, err := r.elastic.Client.Search(r.index).
		Query(elastic.NewMatchQuery("state", state)).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	var proposals = make([]*explorer.Proposal, 0)

	for _, hit := range results.Hits.Hits {
		var proposal *explorer.Proposal
		if err := json.Unmarshal(hit.Source, proposal); err == nil {
			proposals = append(proposals, proposal)
		}
	}

	return proposals, int(results.Hits.TotalHits.Value), err
}

func (r *DaoProposalRepository) Proposal(hash string) (*explorer.Proposal, error) {
	results, err := r.elastic.Client.Search(r.index).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if err != nil || results.TotalHits() == 0 {
		return nil, ErrProposalNotFound
	}

	var proposal *explorer.Proposal
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, proposal)

	return proposal, err
}
