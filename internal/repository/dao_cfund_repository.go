package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/getsentry/raven-go"
	"log"
)

type DaoCfundRepository struct {
	elastic *elastic_cache.Index
}

func NewDaoCfundRepository(elastic *elastic_cache.Index) *DaoCfundRepository {
	return &DaoCfundRepository{elastic}
}

var (
	ErrCfundNotFound = errors.New("Cfund not found")
)

func (r *DaoCfundRepository) GetStats() (*explorer.Cfund, error) {
	network := fmt.Sprintf("%v", param.GetGlobalParam("network", nil))
	if network == "" {
		log.Fatal("No network specified to get consensus parameters")
	}

	results, err := r.elastic.Client.Search(fmt.Sprintf("%s.%s", network, elastic_cache.CfundIndex)).
		Size(1).
		Do(context.Background())
	if err != nil || results == nil {
		raven.CaptureError(err, nil)
		return nil, err
	}

	if len(results.Hits.Hits) != 1 {
		return nil, ErrCfundNotFound
	}

	cfund := new(explorer.Cfund)
	if err = json.Unmarshal(results.Hits.Hits[0].Source, &cfund); err != nil {
		return nil, err
	}

	return cfund, nil
}
