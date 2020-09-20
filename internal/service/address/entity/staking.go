package entity

import (
	"time"
)

type StakingGroup struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Stakes   int64     `json:"stakes"`
	Staking  int64     `json:"staking"`
	Spending int64     `json:"spending"`
	Voting   int64     `json:"voting"`
}

type StakingReport struct {
	TotalSupply float64    `json:"totalSupply"`
	Staking     float64    `json:"staking"`
	Addresses   []Reporter `json:"addresses"`
	From        time.Time  `json:"from"`
	To          time.Time  `json:"to"`
}

type Reporter struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

type StakingBlocks struct {
	BlockCount  int     `json:"blockCount"`
	Staking     float64 `json:"staking"`
	ColdStaking float64 `json:"coldStaking"`
	Fees        float64 `json:"fees"`
	From        uint64  `json:"from,omitempty"`
	To          uint64  `json:"to,omitempty"`
}

type StakingReward struct {
	Address string                 `json:"address"`
	Periods []*StakingRewardPeriod `json:"periods"`
}

type StakingRewardPeriod struct {
	Period  string `json:"period"`
	Stakes  int64  `json:"stakes"`
	Balance int64  `json:"balance"`
}
