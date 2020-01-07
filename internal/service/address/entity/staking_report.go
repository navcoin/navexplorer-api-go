package entity

import "github.com/NavExplorer/navexplorer-api-go/internal/service/group"

type StakingReport struct {
	group.TimeGroup

	Stakes uint   `json:"stakes"`
	Amount uint64 `json:"amount"`
}
