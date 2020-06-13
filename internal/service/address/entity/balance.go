package entity

type Balance struct {
	Address           string  `json:"address"`
	Balance           float64 `json:"balance"`
	ColdStakedBalance float64 `json:"coldStakedBalance"`
}
