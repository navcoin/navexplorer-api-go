package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
)

type BlockGroups struct {
	Items []*BlockGroup `json:"items"`
}

type BlockGroup struct {
	group.TimeGroup
	Period       group.Period `json:"period"`
	Blocks       int64        `json:"blocks"`
	Stake        int64        `json:"stake"`
	Fees         int64        `json:"fees"`
	Spend        int64        `json:"spend"`
	Transactions int64        `json:"transactions"`
	Height       int64        `json:"height"`
}
