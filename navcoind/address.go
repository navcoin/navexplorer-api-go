package navcoind

func (nav *Navcoind) GetRawTransaction(hash string) (data string, err error) {
	response, err := nav.client.call("getrawtransaction", []interface{}{hash, 1})
	if err = HandleError(err, &response); err != nil {
		return
	}

	result, err := response.Result.MarshalJSON()

	return string(result), err
}