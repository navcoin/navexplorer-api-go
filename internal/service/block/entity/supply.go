package entity

import "time"

type Supply struct {
	Height  uint64        `json:"height"`
	Time    time.Time     `json:"time"`
	Balance SupplyBalance `json:"balance"`
	Change  SupplyChange  `json:"change"`
}

type SupplyBalance struct {
	Public  uint64 `json:"public"`
	Private uint64 `json:"private"`
	Wrapped uint64 `json:"wrapped"`
}

type SupplyChange struct {
	Public  int64 `json:"public"`
	Private int64 `json:"private"`
	Wrapped int64 `json:"wrapped"`
}
