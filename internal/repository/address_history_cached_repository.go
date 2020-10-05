package repository

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
	"reflect"
)

type addressHistoryCachedRepository struct {
	repository AddressHistoryRepository
	cache      *cache.Cache
	network    string
}

func NewAddressHistoryCachedRepository(repository AddressHistoryRepository, cache *cache.Cache) AddressHistoryRepository {
	return &addressHistoryCachedRepository{repository: repository, cache: cache}
}

func (r *addressHistoryCachedRepository) Network(network string) AddressHistoryRepository {
	r.repository.Network(network)

	return r
}

func (r *addressHistoryCachedRepository) LatestByHash(hash string) (*explorer.AddressHistory, error) {
	return r.repository.LatestByHash(hash)
}

func (r *addressHistoryCachedRepository) FirstByHash(hash string) (*explorer.AddressHistory, error) {
	return r.repository.FirstByHash(hash)
}

func (r *addressHistoryCachedRepository) CountByHash(hash string) (int64, error) {
	return r.repository.CountByHash(hash)
}

func (r *addressHistoryCachedRepository) StakingSummary(hash string) (count, staking, spending, voting int64, err error) {
	return r.repository.StakingSummary(hash)
}

func (r *addressHistoryCachedRepository) SpendSummary(hash string) (spendingReceive, spendingSent, stakingReceive, stakingSent, votingReceive, votingSent int64, err error) {
	return r.repository.SpendSummary(hash)
}

func (r *addressHistoryCachedRepository) HistoryByHash(hash, txType string, dir bool, size, page int) ([]*explorer.AddressHistory, int64, error) {
	return r.repository.HistoryByHash(hash, txType, dir, size, page)
}

func (r *addressHistoryCachedRepository) GetAddressGroups(period *group.Period, count int) ([]entity.AddressGroup, error) {
	addressGroup := make([]entity.AddressGroup, count)

	callback := func() (interface{}, error) {
		return r.repository.GetAddressGroups(period, count)
	}

	cacheId := fmt.Sprintf("address.groups.%s.%d", string(*period), count)
	cacheResult, err := r.cache.Get(cacheId, callback, cache.RefreshingExpiration)
	if err != nil {
		log.WithError(err).Error("Failed to get cache")
		return addressGroup, err
	}

	for i, v := range InterfaceSlice(cacheResult) {
		addressGroup[i] = v.(entity.AddressGroup)
	}

	return addressGroup, nil
}

func (r *addressHistoryCachedRepository) StakingChart(period string, hash string) (groups []*entity.StakingGroup, err error) {
	return r.repository.StakingChart(period, hash)
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
