package navcoind

func (nav *Navcoind) GetBlock(hash string) (data string, err error) {
	response, err := nav.client.call("getblock", []interface{}{hash})
	if err = HandleError(err, &response); err != nil {
		return
	}

	result, err := response.Result.MarshalJSON()

	return string(result), err
}