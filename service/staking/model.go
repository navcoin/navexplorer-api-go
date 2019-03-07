package staking

type Report struct {
	TotalSupply float64 `json:"totalSupply"`
	Staking     float64 `json:"staking"`
	From        int     `json:"blockFrom"`
	To          int     `json:"blockTo"`
}
