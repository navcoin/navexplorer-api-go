package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

var (
	ErrAddressNotFound = errors.New("Address not found")
	ErrAddressInvalid  = errors.New("Address is invalid")
)

type AddressRepository struct {
	elastic *elastic_cache.Index
}

func NewAddressRepository(elastic *elastic_cache.Index) *AddressRepository {
	return &AddressRepository{elastic}
}

func (r *AddressRepository) Addresses(size int, page int) ([]*explorer.Address, int, error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Sort("balance", false).
		From((page * size) - size).
		Size(size).
		Do(context.Background())

	return r.findMany(results, err)
}

func (r *AddressRepository) AddressByHash(hash string) (*explorer.Address, error) {
	if valid, err := r.Validate(hash); valid == false || err != nil {
		if err != nil {
			log.WithError(err).Error("Failed to validate address")
		}
		return nil, ErrAddressInvalid
	}

	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *AddressRepository) Validate(hash string) (bool, error) {
	return true, nil
}

func (r *AddressRepository) getRichListPosition(balance uint64) (uint, error) {
	position, err := r.elastic.Client.Count(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("balance").Gt(balance)).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Infof("Failed to get rich list position")
	}

	return uint(position + 1), err
}

func (r *AddressRepository) findOne(results *elastic.SearchResult, err error) (*explorer.Address, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrAddressNotFound
		return nil, err
	}

	var address *explorer.Address
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &address)
	if err != nil {
		return nil, err
	}

	address.Position, err = r.getRichListPosition(address.Balance)
	if err != nil {
		return nil, err
	}

	return address, err
}

func (r *AddressRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.Address, int, error) {
	if err != nil {
		return nil, 0, err
	}

	addresses := make([]*explorer.Address, 0)
	for index, hit := range results.Hits.Hits {
		var address *explorer.Address
		if err := json.Unmarshal(hit.Source, &address); err == nil {
			address.Position = uint(index + 1)
			addresses = append(addresses, address)
		}
	}

	return addresses, int(results.Hits.TotalHits.Value), err
}
