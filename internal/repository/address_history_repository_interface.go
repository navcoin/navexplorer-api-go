package repository

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address/entity"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

type AddressHistoryRepository interface {
	Network(network string) AddressHistoryRepository
	LatestByHash(hash string) (*explorer.AddressHistory, error)
	FirstByHash(hash string) (*explorer.AddressHistory, error)
	CountByHash(hash string) (int64, error)
	StakingSummary(hash string) (count, staking, spending, voting int64, err error)
	SpendSummary(hash string) (spendingReceive, spendingSent, stakingReceive, stakingSent, votingReceive, votingSent int64, err error)
	HistoryByHash(hash string, txType string, dir bool, size int, page int) ([]*explorer.AddressHistory, int64, error)
	GetAddressGroups(period *group.Period, count int) ([]entity.AddressGroup, error)
	StakingChart(period string, hash string) (groups []*entity.StakingGroup, err error)
}
