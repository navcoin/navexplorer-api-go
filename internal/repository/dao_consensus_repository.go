package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type DaoConsensusRepository struct {
	elastic *elastic_cache.Index
}

var (
	ErrConsensusNotFound = errors.New("Consensus not found")
)

func NewDaoConsensusRepository(elastic *elastic_cache.Index) *DaoConsensusRepository {
	return &DaoConsensusRepository{elastic}
}

func (r *DaoConsensusRepository) GetConsensus() (*explorer.Consensus, error) {
	results, err := r.elastic.Client.Search(elastic_cache.ConsensusIndex.Get()).
		Size(1).
		Do(context.Background())

	if err != nil || results.TotalHits() == 0 {
		err = ErrConsensusNotFound
		return nil, err
	}

	var consensus *explorer.Consensus
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &consensus)
	if err != nil {
		return nil, err
	}

	return consensus, err
}
