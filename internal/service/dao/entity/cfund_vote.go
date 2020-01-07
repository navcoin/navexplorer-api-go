package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
)

type CfundVote struct {
	group.BlockGroup

	Cycle   int `json:"cycle"`
	Yes     int `json:"yes"`
	No      int `json:"no"`
	Abstain int `json:"abstain"`
}

func NewCfundVote(cycle int, start int, end int) *CfundVote {
	return &CfundVote{
		BlockGroup: group.BlockGroup{Start: start, End: end},
		Cycle:      cycle,
	}
}
