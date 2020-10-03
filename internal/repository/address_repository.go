package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	entitycoin "github.com/NavExplorer/navexplorer-api-go/internal/service/coin/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
	"strings"
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

func (r *AddressRepository) Addresses(size int, page int) ([]*explorer.Address, int64, error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Sort("spending", false).
		From((page * size) - size).
		Size(size).
		Do(context.Background())

	addresses, total, err := r.findMany(results, err)
	for idx := range addresses {
		addresses[idx].RichList = explorer.RichList{
			Spending: uint64(idx + 1 + (page * size) - size),
		}
	}

	return addresses, total, err
}

func (r *AddressRepository) AddressByHash(hash string) (*explorer.Address, error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *AddressRepository) BalancesForAddresses(addresses []string) ([]*explorer.Address, error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewMatchQuery("hash", strings.Join(addresses, " "))).
		Size(5000).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	a, _, err := r.findMany(results, err)
	return a, err
}

func (r *AddressRepository) WealthDistribution(groups []int) ([]*entitycoin.Wealth, error) {
	totalSupply, err := r.GetTotalSupply()
	if err != nil {
		return nil, err
	}

	totalWealth := &entitycoin.Wealth{
		Balance:    totalSupply,
		Percentage: 100,
	}

	distribution := make([]*entitycoin.Wealth, 0)

	for i := 0; i < len(groups); i++ {
		results, _ := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
			From(0).
			Size(groups[i]).
			Sort("spending", false).
			Do(context.Background())

		wealth := &entitycoin.Wealth{Group: groups[i]}

		for _, element := range results.Hits.Hits {
			address := new(explorer.Address)
			err = json.Unmarshal(element.Source, &address)

			wealth.Balance += float64(address.Spending) / 100000000
			wealth.Percentage = int64((wealth.Balance / totalWealth.Balance) * 100)
		}

		distribution = append(distribution, wealth)
	}

	distribution = append(distribution, totalWealth)

	return distribution, err
}

func (r *AddressRepository) GetTotalSupply() (totalSupply float64, err error) {
	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get()).
		Aggregation("totalWealth", elastic.NewSumAggregation().Field("spending")).
		Size(0).
		Do(context.Background())
	if err != nil {
		return
	}

	if total, found := results.Aggregations.Sum("totalWealth"); found {
		totalSupply = *total.Value / 100000000
	}

	return
}

func (r *AddressRepository) UpdateAddress(address *explorer.Address) error {
	_, err := r.elastic.Client.
		Index().
		Index(elastic_cache.AddressIndex.Get()).
		Id(address.Slug()).
		BodyJson(address).
		Do(context.Background())

	return err
}

func (r *AddressRepository) populateRichListPosition(address *explorer.Address) error {
	spending, err := r.elastic.Client.Count(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("spending").Gt(address.Spending)).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to get rich list position")
		return err
	}

	staking, err := r.elastic.Client.Count(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("staking").Gt(address.Staking)).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to get rich list position")
		return err
	}

	voting, err := r.elastic.Client.Count(elastic_cache.AddressIndex.Get()).
		Query(elastic.NewRangeQuery("spending").Gt(address.Voting)).
		Do(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to get rich list position")
		return err
	}

	address.RichList = explorer.RichList{
		Spending: uint64(spending) + 1,
		Staking:  uint64(staking) + 1,
		Voting:   uint64(voting) + 1,
	}

	return nil
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

	err = r.populateRichListPosition(address)
	if err != nil {
		return nil, err
	}

	return address, err
}

func (r *AddressRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.Address, int64, error) {
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
