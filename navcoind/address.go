package navcoind

import (
	"encoding/json"
)

type ValidateAddress struct {
	Valid           bool   `json:"isvalid"`
	Address         string `json:"address"`
	StakingAddress  string `json:"stakingaddress"`
	SpendingAddress string `json:"spendingaddress"`
	ColdStaking     bool   `json:"iscoldstaking"`
}

func (nav *Navcoind) GetRawTransaction(hash string) (data string, err error) {
	response, err := nav.client.call("getrawtransaction", []interface{}{hash, 1})
	if err = HandleError(err, &response); err != nil {
		return
	}

	result, err := response.Result.MarshalJSON()

	return string(result), err
}

func (nav *Navcoind) ValidateAddress(address string) (validateAddress ValidateAddress, err error) {
	response, err := nav.client.call("validateaddress", []interface{}{address})
	if err = HandleError(err, &response); err != nil {
		return
	}

	err = json.Unmarshal(response.Result, &validateAddress)

	return
}