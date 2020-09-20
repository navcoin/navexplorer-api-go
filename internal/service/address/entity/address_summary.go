package entity

type AddressSummary struct {
	Height   uint64          `json:"height"`
	Hash     string          `json:"hash"`
	Sent     *AddressBalance `json:"sent"`
	Received *AddressBalance `json:"received"`
	Staked   *AddressBalance `json:"staked"`
}

type AddressBalance struct {
	Balance  uint64 `json:"balance"`
	Staking  uint64 `json:"staking"`
	Spending uint64 `json:"spending"`
	Voting   uint64 `json:"voting"`
}
