package entity

type AddressSummary struct {
	Height       uint64          `json:"height"`
	Hash         string          `json:"hash"`
	Spendable    *AddressBalance `json:"spendable"`
	Stakable     *AddressBalance `json:"stakable"`
	VotingWeight *AddressBalance `json:"voting_weight"`
	Txs          int64           `json:"txs"`
}

type AddressBalance struct {
	Balance  int64 `json:"balance"`
	Staked   int64 `json:"staked"`
	Sent     int64 `json:"sent"`
	Received int64 `json:"received"`
}
