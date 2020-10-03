package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
)

type AddressGroups struct {
	Items []*AddressGroup `json:"items"`
}

type AddressGroup struct {
	group.TimeGroup
	Period    group.Period `json:"period"`
	Addresses int64        `json:"addresses"`
	Spend     int64        `json:"spend"`
}
