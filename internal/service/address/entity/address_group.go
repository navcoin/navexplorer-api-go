package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
)

type AddressGroup struct {
	group.TimeGroup
	Period    group.Period `json:"period"`
	Addresses int64        `json:"addresses"`
	Stake     int64        `json:"stake"`
	Spend     int64        `json:"spend"`
}
