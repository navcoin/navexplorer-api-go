package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"strings"
)

type AddressRepository interface {
	GetAddresses(n network.Network, size, page int) ([]*explorer.Address, int64, error)
	GetAddressByHash(n network.Network, hash string) (*explorer.Address, error)
	GetBalancesForAddresses(n network.Network, addresses []string) ([]*explorer.Address, error)
	GetWealthDistribution(n network.Network, groups []int, totalSupply uint64) ([]*entity.Wealth, error)
	UpdateAddress(n network.Network, address *explorer.Address) error
}

var (
	ErrAddressNotFound = errors.New("Address not found")
	ErrAddressInvalid  = errors.New("Address is invalid")
)

type addressRepository struct {
	elastic *elastic_cache.Index
}

func NewAddressRepository(elastic *elastic_cache.Index) AddressRepository {
	return &addressRepository{elastic: elastic}
}

func (r *addressRepository) GetAddresses(n network.Network, size, page int) ([]*explorer.Address, int64, error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get(n)).
		Sort("spendable", false).
		From((page * size) - size).
		Size(size).
		Do(context.Background())

	addresses, total, err := r.findMany(results, err)
	for idx := range addresses {
		addresses[idx].RichList = explorer.RichList{
			Spendable: uint64(idx + 1 + (page * size) - size),
		}
	}

	return addresses, total, err
}

func (r *addressRepository) GetAddressByHash(n network.Network, hash string) (*explorer.Address, error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get(n)).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(n, results, err)
}

func (r *addressRepository) GetBalancesForAddresses(n network.Network, addresses []string) ([]*explorer.Address, error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get(n)).
		Query(elastic.NewMatchQuery("hash", strings.Join(addresses, " "))).
		Size(5000).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	a, _, err := r.findMany(results, err)
	return a, err
}

func (r *addressRepository) GetWealthDistribution(n network.Network, groups []int, totalSupply uint64) ([]*entity.Wealth, error) {

	totalWealth := &entity.Wealth{
		Balance:    float64(totalSupply) / 100000000,
		Percentage: 100,
	}

	distribution := make([]*entity.Wealth, 0)

	for i := 0; i < len(groups); i++ {
		results, _ := r.elastic.Client.Search(elastic_cache.AddressIndex.Get(n)).
			From(0).
			Size(groups[i]).
			Sort("spendable", false).
			Do(context.Background())

		wealth := &entity.Wealth{Group: groups[i]}

		for _, element := range results.Hits.Hits {
			address := new(explorer.Address)
			if err := json.Unmarshal(element.Source, &address); err != nil {
				return nil, err
			}

			wealth.Balance += float64(address.Spendable) / 100000000
			wealth.Percentage = int64((wealth.Balance / totalWealth.Balance) * 100)
		}

		distribution = append(distribution, wealth)
	}

	distribution = append(distribution, totalWealth)

	return distribution, nil
}

func (r *addressRepository) UpdateAddress(n network.Network, address *explorer.Address) error {
	_, err := r.elastic.Client.
		Index().
		Index(elastic_cache.AddressIndex.Get(n)).
		Id(address.Slug()).
		BodyJson(address).
		Do(context.Background())

	return err
}

func (r *addressRepository) populateRichListPosition(n network.Network, address *explorer.Address) error {
	spendable, err := r.elastic.Client.Count(elastic_cache.AddressIndex.Get(n)).
		Query(elastic.NewRangeQuery("spendable").Gt(address.Spendable)).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to get rich list position")
		return err
	}

	stakable, err := r.elastic.Client.Count(elastic_cache.AddressIndex.Get(n)).
		Query(elastic.NewRangeQuery("stakable").Gt(address.Stakable)).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to get rich list position")
		return err
	}

	votingWeight, err := r.elastic.Client.Count(elastic_cache.AddressIndex.Get(n)).
		Query(elastic.NewRangeQuery("voting_weight").Gt(address.VotingWeight)).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to get rich list position")
		return err
	}

	address.RichList = explorer.RichList{
		Spendable:    uint64(spendable) + 1,
		Stakable:     uint64(stakable) + 1,
		VotingWeight: uint64(votingWeight) + 1,
	}

	return nil
}

func (r *addressRepository) findOne(n network.Network, results *elastic.SearchResult, err error) (*explorer.Address, error) {
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

	err = r.populateRichListPosition(n, address)
	if err != nil {
		return nil, err
	}

	return address, err
}

func (r *addressRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.Address, int64, error) {
	if err != nil {
		return nil, 0, err
	}

	addresses := make([]*explorer.Address, 0)
	for _, hit := range results.Hits.Hits {
		var address *explorer.Address
		if err := json.Unmarshal(hit.Source, &address); err == nil {
			addresses = append(addresses, address)
		}
	}

	return addresses, results.TotalHits(), err
}
