package repository

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type SoftForkRepository struct {
	elastic *elastic_cache.Index
}

func NewSoftForkRepository(elastic *elastic_cache.Index) *SoftForkRepository {
	return &SoftForkRepository{elastic}
}

func (r *SoftForkRepository) SoftForks() ([]*explorer.SoftFork, error) {
	results, err := r.elastic.Client.Search(elastic_cache.SoftForkIndex.Get()).Size(9999).Do(context.Background())
	if err != nil || results.Hits.TotalHits.Value == 0 {
		return nil, err
	}

	softForks := make([]*explorer.SoftFork, 0)
	for _, hit := range results.Hits.Hits {
		var softFork *explorer.SoftFork
		err = json.Unmarshal(hit.Source, &softFork)
		if err == nil {
			softForks = append(softForks, softFork)
		}
	}

	return softForks, nil
}
