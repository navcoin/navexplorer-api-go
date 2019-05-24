package staking

import (
	"time"
)

type Report struct {
	TotalSupply float64   `json:"totalSupply"`
	Staking     float64   `json:"staking"`
	Addresses   []Reporter `json:"addresses"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
}

type Reporter struct {
	Address            string  `json:"address"`
	Balance            float64 `json:"balance"`
}

type StakingBlocks struct {
	BlockCount  int     `json:"blockCount"`
	Staking     float64 `json:"staking"`
	ColdStaking float64 `json:"coldStaking"`
	Fees        float64 `json:"fees"`
}

type Reward struct {
	Address string
	Periods []RewardPeriod `json:"periods"`
}

type RewardPeriod struct {
	Period  string `json:"period"`
	Stakes  int64  `json:"stakes"`
	Balance int64  `json:"balance"`
}