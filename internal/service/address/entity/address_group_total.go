package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
)

type AddressGroupTotal struct {
	Period group.Period `json:"period"`
	group.TimeGroup
	Addresses int64 `json:"addresses"`
}
