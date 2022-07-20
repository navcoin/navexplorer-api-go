package entity

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/group"
)

type AddressGroup struct {
	Period group.Period `json:"period"`
	group.TimeGroup
	Stake int64 `json:"stake"`
	Spend int64 `json:"spend"`
}
