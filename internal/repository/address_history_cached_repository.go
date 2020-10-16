package repository

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"reflect"
)

type cachingAddressHistoryRepository struct {
	repository AddressHistoryRepository
	cache      *cache.Cache
}

func NewCachingAddressHistoryRepository(repository AddressHistoryRepository, cache *cache.Cache) AddressHistoryRepository {
	return &cachingAddressHistoryRepository{repository: repository, cache: cache}
}

func (r *cachingAddressHistoryRepository) GetLatestByHash(n network.Network, hash string) (*explorer.AddressHistory, error) {
	return r.repository.GetLatestByHash(n, hash)
}

func (r *cachingAddressHistoryRepository) GetFirstByHash(n network.Network, hash string) (*explorer.AddressHistory, error) {
	return r.repository.GetFirstByHash(n, hash)
}

func (r *cachingAddressHistoryRepository) GetCountByHash(n network.Network, hash string) (int64, error) {
	return r.repository.GetCountByHash(n, hash)
}

func (r *cachingAddressHistoryRepository) GetStakingSummary(n network.Network, hash string) (count, staking, spending, voting int64, err error) {
	return r.repository.GetStakingSummary(n, hash)
}

func (r *cachingAddressHistoryRepository) GetSpendSummary(n network.Network, hash string) (spendingReceive, spendingSent, stakingReceive, stakingSent, votingReceive, votingSent int64, err error) {
	return r.repository.GetSpendSummary(n, hash)
}

func (r *cachingAddressHistoryRepository) GetHistoryByHash(n network.Network, hash, txType string, dir bool, size, page int) ([]*explorer.AddressHistory, int64, error) {
	return r.repository.GetHistoryByHash(n, hash, txType, dir, size, page)
}

func (r *cachingAddressHistoryRepository) GetAddressGroups(n network.Network, period *group.Period, count int) ([]entity.AddressGroup, error) {
	addressGroup := make([]entity.AddressGroup, count)

	result, err := r.cache.Get(
		fmt.Sprintf("%s.address.groups.%s.%d", n.ToString(), string(*period), count),
		func() (interface{}, error) {
			return r.repository.GetAddressGroups(n, period, count)
		},
		cache.RefreshingExpiration,
	)
	if err != nil {
		log.WithError(err).Error("Failed to get cache")
		return addressGroup, err
	}

	for i, v := range InterfaceSlice(result) {
		addressGroup[i] = v.(entity.AddressGroup)
	}

	return addressGroup, nil
}

func (r *cachingAddressHistoryRepository) GetStakingChart(n network.Network, period string, hash string) (groups []*entity.StakingGroup, err error) {
	return r.repository.GetStakingChart(n, period, hash)
}

func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}
