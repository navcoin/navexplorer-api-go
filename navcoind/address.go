package navcoind

import (
	"encoding/json"
	"log"
)

type ValidateAddress struct {
	Valid bool `json:"isvalid"`
}

func (nav *Navcoind) GetRawTransaction(hash string) (data string, err error) {
	response, err := nav.client.call("getrawtransaction", []interface{}{hash, 1})
	if err = HandleError(err, &response); err != nil {
		return
	}

	result, err := response.Result.MarshalJSON()

	return string(result), err
}

func (nav *Navcoind) ValidateAddress(address string) (isValid bool) {
	response, err := nav.client.call("validateaddress", []interface{}{address})
	if err = HandleError(err, &response); err != nil {
		log.Println(err)
		return false
	}

	var validateAddress ValidateAddress
	err = json.Unmarshal(response.Result, &validateAddress)

	if err != nil {
		return false
	}

	return validateAddress.Valid
}