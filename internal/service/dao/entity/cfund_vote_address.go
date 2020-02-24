package entity

type CfundVoteAddresses struct {
	Cycle int                          `json:"cycle"`
	Yes   []*CfundVoteAddressesElement `json:"yes"`
	No    []*CfundVoteAddressesElement `json:"no"`
}

type CfundVoteAddressesElement struct {
	Address string `json:"address"`
	Votes   int    `json:"votes"`
}
