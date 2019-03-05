package staking

type Addresses struct {
	TotalSupply float64          `json:"stakingSupply"`
	Addresses   []StakingAddress `json:"addresses"`
}

type StakingAddress struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}
