package entity

type AddressSummary struct {
	Height   uint64          `json:"height"`
	Hash     string          `json:"hash"`
	Spending *AddressBalance `json:"spending"`
	Staking  *AddressBalance `json:"staking"`
	Voting   *AddressBalance `json:"voting"`
	Txs      int64           `json:"txs"`
}

type AddressBalance struct {
	Balance  int64 `json:"balance"`
	Staked   int64 `json:"staked"`
	Sent     int64 `json:"sent"`
	Received int64 `json:"received"`
}
