package communityFund

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type Proposal struct {
	ID                  bson.ObjectId `bson:"_id" json:"-"`

	Version             int           `bson:"version" json:"version"`
	Hash                string        `bson:"hash" json:"hash"`
	BlockHash           string        `bson:"blockHash" json:"block-hash"`
	Height              int           `json:"height"`
	Description         string        `bson:"description" json:"description"`
	RequestedAmount     float64       `bson:"requestedAmount" json:"requested-amount"`
	NotPaidYet          float64       `bson:"notPaidYet" json:"not-paid-yet"`
	UserPaidFee         float64       `bson:"userPaidFee" json:"user-paid-fee"`
	PaymentAddress      string        `bson:"paymentAddress" json:"payment-address"`
	ProposalDuration    int           `bson:"proposalDuration" json:"proposal-duration"`

	VotesYes            int           `bson:"votesYes" json:"votes-yes"`
	VotesNo             int           `bson:"votesNo" json:"votes-no"`
	VotingCycle         int           `bson:"votingCycle" json:"voting-cycle"`

	Status              string        `bson:"status" json:"status"`
	State               string        `bson:"state" json:"state"`
	StateChangedOnBlock string        `bson:"stateChangedOnBlock" json:"state-changed-on-block"`

	ExpiresOn           time.Time     `bson:"expiresOn" json:"expires-on"`
	CreatedAt           time.Time     `bson:"createdAt" json:"created-at"`
}

type PaymentRequest struct {
	ID                  bson.ObjectId `bson:"_id" json:"-"`

	Version             int           `bson:"version" json:"version"`
	Hash                string        `bson:"hash" json:"hash"`
	BlockHash           string        `bson:"blockHash" json:"block-hash"`
	ProposalHash        string        `bson:"proposalHash" json:"proposal-hash"`
	Description         string        `bson:"description" json:"description"`
	RequestedAmount     float64       `bson:"requestedAmount" json:"requested-amount"`

	VotesYes            int           `bson:"votesYes" json:"votes-yes"`
	VotesNo             int           `bson:"votesNo" json:"votes-no"`
	VotingCycle         int           `bson:"votingCycle" json:"voting-cycle"`

	Status              string        `bson:"status" json:"status"`
	State               string        `bson:"state" json:"state"`
	StateChangedOnBlock string        `bson:"stateChangedOnBlock" json:"state-changed-on-block"`
	PaidOnBlock         string        `bson:"paidOnBlock" json:"paid-on-block"`

	CreatedAt           time.Time     `bson:"createdAt" json:"created-at"`
}

type BlockCycle struct {
	BlocksInCycle   int              `json:"blocks-in-cycle"`
	MinQuorum       float64          `json:"min-quorum"`
	ProposalVoting  BlockCycleVoting `json:"proposal-voting"`
	PaymentVoting   BlockCycleVoting `json:"payment-voting"`
	Height          int              `json:"height"`

	Cycle           int              `json:"cycle"`
	FirstBlock      int              `json:"first-block"`
	CurrentBlock    int              `json:"current-block"`
	BlocksRemaining int              `json:"blocks-remaining"`
}

type BlockCycleVoting struct {
	Cycles int     `json:"cycles"`
	Accept float64 `json:"accept"`
	Reject float64 `json:"reject"`
}