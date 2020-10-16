package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
)

type DaoConsensusRepository interface {
	GetConsensusParameters(n network.Network) (*explorer.ConsensusParameters, error)
}

type daoConsensusRepository struct {
	elastic *elastic_cache.Index
}

var (
	ErrConsensusNotFound = errors.New("Consensus not found")
)

func NewDaoConsensusRepository(elastic *elastic_cache.Index) DaoConsensusRepository {
	return &daoConsensusRepository{elastic: elastic}
}

func (r *daoConsensusRepository) GetConsensusParameters(n network.Network) (*explorer.ConsensusParameters, error) {
	results, err := r.elastic.Client.Search(elastic_cache.ConsensusIndex.Get(n)).
		Size(1000).
		Sort("id", true).
		Do(context.Background())
	if err != nil || results == nil {
		raven.CaptureError(err, nil)
		return nil, err
	}

	if len(results.Hits.Hits) == 0 {
		return nil, ErrConsensusNotFound
	}

	consensusParameters := new(explorer.ConsensusParameters)
	for _, hit := range results.Hits.Hits {
		var consensusParameter *explorer.ConsensusParameter
		if err = json.Unmarshal(hit.Source, &consensusParameter); err != nil {
			return nil, err
		}
		consensusParameters.Add(consensusParameter)
	}

	return consensusParameters, nil
}
