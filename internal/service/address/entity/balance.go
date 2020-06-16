package entity

type Balance struct {
	Address           string `json:"address"`
	Balance           int64  `json:"balance"`
	ColdStakedBalance int64  `json:"coldStakedBalance"`
}
