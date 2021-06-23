package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type AddressRepository interface {
	GetAddresses(n network.Network, size, page int, sort framework.Sort) ([]*explorer.Address, int64, error)
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

func (r *addressRepository) GetAddresses(n network.Network, size, page int, s framework.Sort) ([]*explorer.Address, int64, error) {
	service := r.elastic.Client.Search(elastic_cache.AddressIndex.Get(n)).
		From((page * size) - size).
		Size(size)

	sort(service, s, &defaultSort{"spendable", false})

	results, err := service.Do(context.Background())

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
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(n, results, err)
}

func (r *addressRepository) GetBalancesForAddresses(n network.Network, addresses []string) ([]*explorer.Address, error) {
	values := make([]interface{}, len(addresses))
	for i, v := range addresses {
		values[i] = v
	}

	results, err := r.elastic.Client.Search(elastic_cache.AddressIndex.Get(n)).
		Query(elastic.NewTermsQuery("hash.keyword", values...)).
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

	// Exclude the wNav miltisig staking address
	query := elastic.NewBoolQuery().MustNot(
		elastic.NewTermQuery("hash.keyword", "a456b36048ce2e732ef729d044a1f744738df5fa-0277fa3f4f6d447c5914d8d69c259f94c76aa6eae829c5bd54e3cd6fc3f7e12f2f-033a0879f9ab601b4ee20ec9fed77ea1a48e9026b48e0d2a425d874b40ef13d022-034a51aa6aafbd6c6075ecaee0fbcf2c9ffbac05a49007a0f02c9d6680dccee6d4-03ad915271a0b327f5379585c00c42a732530f246b60f9bb1c19af7db59363897e"))

	for i := 0; i < len(groups); i++ {
		results, _ := r.elastic.Client.Search(elastic_cache.AddressIndex.Get(n)).
			From(0).
			Query(query).
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
