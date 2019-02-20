package communityFund

import (
	"time"
)

type Proposal struct {
	Version             int       `json:"version"`
	Hash                string    `json:"hash"`
	BlockHash           string    `json:"blockHash"`
	Height              int       `json:"height"`
	Description         string    `json:"description"`
	RequestedAmount     float64   `json:"requestedAmount"`
	NotPaidYet          float64   `json:"notPaidYet"`
	UserPaidFee         float64   `json:"userPaidFee"`
	PaymentAddress      string    `json:"paymentAddress"`
	ProposalDuration    int       `json:"proposalDuration"`
	VotesYes            int       `json:"votesYes"`
	VotesNo             int       `json:"votesNo"`
	VotingCycle         int       `json:"votingCycle"`
	Status              string    `json:"status"`
	State               string    `json:"state"`
	StateChangedOnBlock string    `json:"stateChangedOnBlock,omitempty"`
	ExpiresOn           *time.Time `json:"expiresOn,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
}

type PaymentRequest struct {
	Version             int       `json:"version"`
	Hash                string    `json:"hash"`
	BlockHash           string    `json:"blockHash"`
	ProposalHash        string    `json:"proposalHash"`
	Description         string    `json:"description"`
	RequestedAmount     float64   `json:"requestedAmount"`
	VotesYes            int       `json:"votesYes"`
	VotesNo             int       `json:"votesNo"`
	VotingCycle         int       `json:"votingCycle"`
	Status              string    `json:"status"`
	State               string    `json:"state"`
	StateChangedOnBlock string    `json:"stateChangedOnBlock,omitempty"`
	PaidOnBlock         string    `json:"paidOnBlock,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
}

type Votes struct {
	Address  string `json:"address"`
	Votes    int64  `json:"votes"`
}

type Trend struct {
	Start        int     `json:"start"`
	End          int     `json:"end"`
	VotesYes     int     `json:"votesYes"`
	VotesNo      int     `json:"votesNo"`
	TrendYes     float64 `json:"trendYes"`
	TrendNo      float64 `json:"trendNo"`
	TrendAbstain float64 `json:"trendAbstain"`
}

type BlockCycle struct {
	BlocksInCycle   int              `json:"blocksInCycle"`
	MinQuorum       float64          `json:"minQuorum"`
	ProposalVoting  BlockCycleVoting `json:"proposalVoting"`
	PaymentVoting   BlockCycleVoting `json:"paymentVoting"`
	Height          int              `json:"height"`
	Cycle           int              `json:"cycle"`
	FirstBlock      int              `json:"firstBlock"`
	CurrentBlock    int              `json:"currentBlock"`
	BlocksRemaining int              `json:"blocksRemaining"`
}

type BlockCycleVoting struct {
	Cycles int     `json:"cycles"`
	Accept float64 `json:"accept"`
	Reject float64 `json:"reject"`
}

type Stats struct {
	Contributed float64 `json:"contributed"`
	Requested   float64 `json:"requested"`
	Paid        float64 `json:"paid"`
	Locked      float64 `json:"locked"`
}