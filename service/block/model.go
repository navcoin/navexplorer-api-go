package block

import (
	"time"
)

type Block struct {
	Hash          string    `json:"hash"`
	MerkleRoot    string    `json:"merkleRoot"`
	Bits          string    `json:"bits"`
	Size          int       `json:"size"`
	Version       int       `json:"version"`
	VersionHex    string    `json:"versionHex"`
	Nonce         int       `json:"nonce"`
	Height        int       `json:"height"`
	Difficulty    float64   `json:"difficulty"`
	Confirmations int       `json:"confirmations"`
	Created       time.Time `json:"created"`
	Stake         int       `json:"stake"`
	StakedBy      string    `json:"stakedBy"`
	Fees          int       `json:"fees"`
	Spend         int       `json:"spend"`
	Transactions  int       `json:"transactions"`
	Balance       int       `json:"balance"`
	CfundPayout   int       `json:"cfundPayout"`
	Best          bool      `json:"cfundPayout"`
}

type Transaction struct {
	Hash                string                `json:"hash"`
	BlockHash           string                `json:"-"`
	Type                string                `json:"type"`
	Height              int                   `json:"height"`
	Time                time.Time             `json:"time"`
	Stake               int                   `json:"stake"`
	Fees                int                   `json:"fees"`
	Version             int                   `json:"version"`
	AnonDestination     string                `json:"anonDestination"`
	Inputs              []Input               `json:"inputs"`
	Outputs             []Output              `json:"outputs"`
	ProposalVotes       []ProposalVote        `json:"proposalVotes"`
	PaymentRequestVotes []PaymentRequestVotes `json:"paymentRequestVotes"`
}

type Input struct {
	Index               int      `json:"index"`
	Addresses           []string `json:"addresses"`
	Amount              float64  `json:"amount"`
	PreviousOutput      string   `json:"previousOutput"`
	PreviousOutputBlock int      `json:"previousOutputBlock"`
}

type Output struct {
	Index      int         `json:"index"`
	Type       string      `json:"type"`
	Addresses  []string    `json:"addresses,omitempty"`
	Amount     float64     `json:"amount,omitempty"`
	RedeemedIn *RedeemedIn `json:"redeemedIn,omitempty"`
	Hash       string      `json:"hash,omitempty"`
}

type RedeemedIn struct {
	Hash   string `json:"hash,omitempty"`
	Height string `json:"height,omitempty"`
}

type ProposalVote struct {
	Hash string `json:"hash"`
	Vote string `json:"vote"`
}

type PaymentRequestVotes struct {
	Hash string `json:"hash"`
	Vote string `json:"vote"`
}