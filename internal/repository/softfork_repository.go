package repository

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

type SoftForkRepository interface {
	GetSoftForks(n network.Network) ([]*explorer.SoftFork, error)
}

type softForkRepository struct {
	elastic *elastic_cache.Index
}

func NewSoftForkRepository(elastic *elastic_cache.Index) SoftForkRepository {
	return &softForkRepository{elastic: elastic}
}
func (r *softForkRepository) GetSoftForks(n network.Network) ([]*explorer.SoftFork, error) {
	results, err := r.elastic.Client.Search(elastic_cache.SoftForkIndex.Get(n)).
		Size(9999).
		Sort("signalBit", false).
		Do(context.Background())
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
