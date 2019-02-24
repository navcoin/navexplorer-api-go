package address

import (
	"time"
)

type Address struct {
	Hash               string  `json:"hash"`
	Received           float64 `json:"received"`
	ReceivedCount      int     `json:"receivedCount"`
	Sent               float64 `json:"sent"`
	SentCount          int     `json:"sentCount"`
	Staked             float64 `json:"staked"`
	StakedCount        int     `json:"stakedCount"`
	StakedSent         float64 `json:"stakedSent"`
	StakedReceived     float64 `json:"stakedReceived"`
	ColdStaked         float64 `json:"coldStaked"`
	ColdStakedCount    int     `json:"coldStakedCount"`
	ColdStakedSent     float64 `json:"coldStakedSent"`
	ColdStakedReceived float64 `json:"coldStakedReceived"`
	ColdStakedBalance  float64 `json:"coldStakedBalance"`
	Balance            float64 `json:"balance"`
	BlockIndex         int     `json:"blockIndex"`
	RichListPosition   int64   `json:"richListPosition"`
}

type Transaction struct {
	Time                time.Time `json:"time"`
	Address             string    `json:"address"`
	Type                string    `json:"type"`
	Transaction         string    `json:"transaction"`
	Height              int       `json:"height"`
	Balance             float64   `json:"balance"`
	Sent                float64   `json:"sent"`
	Received            float64   `json:"received"`
	ColdStaking         bool      `json:"coldStaking"`
	ColdStakingBalance  float64   `json:"coldStakingBalance"`
	ColdStakingSent     float64   `json:"coldStakingSent"`
	ColdStakingReceived float64   `json:"coldStakingReceived"`
}

type Chart struct {
	Points []ChartPoint `json:"points"`
}

type ChartPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}