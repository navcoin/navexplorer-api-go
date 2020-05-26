package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
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

func (r *DaoConsensusRepository) GetConsensusParameters(network string) (*explorer.ConsensusParameters, error) {
	results, err := r.elastic.Client.Search(fmt.Sprintf("%s.%s", network, elastic_cache.ConsensusIndex)).
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
