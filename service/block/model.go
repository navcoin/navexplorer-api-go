package block

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type Block struct {
	ID            bson.ObjectId `bson:"_id" json:"id"`
	Hash          string        `bson:"hash" json:"hash"`
	MerkleRoot    string        `bson:"merkleRoot" json:"merkleRoot"`
	Bits          string        `bson:"bits" json:"bits"`
	Size          int           `bson:"size" json:"size"`
	Version       int           `bson:"version" json:"version"`
	VersionHex    string        `bson:"versionHex" json:"versionHex"`
	Nonce         int           `bson:"nonce" json:"nonce"`
	Height        int           `bson:"height" json:"height"`
	Difficulty    float64       `bson:"difficulty" json:"difficulty"`
	Confirmations int           `bson:"confirmations" json:"confirmations"`
	Created       time.Time     `bson:"created" json:"created"`
	Stake         int           `bson:"stake" json:"stake"`
	StakedBy      string        `bson:"stakedBy" json:"stakedBy"`
	Fees          int           `bson:"fees" json:"fees"`
	Spend         int           `bson:"spend" json:"spend"`
	Transactions  int           `bson:"transactions" json:"transactions"`
	Signals       []Signal      `bson:"signals" json:"signals"`
}

type Signal struct {
	Name       string `bson:"name" json:"name"`
	Signalling bool   `bson:"signalling" json:"signalling"`
}

type Transaction struct {
	ID                  bson.ObjectId         `bson:"_id" json:"id"`
	Hash                string                `bson:"hash" json:"hash"`
	BlockHash           string                `bson:"blockHash" json:"-"`
	Type                string                `bson:"type" json:"type"`
	Height              int                   `bson:"height" json:"height"`
	Time                time.Time             `bson:"time" json:"time"`
	Stake               int                   `bson:"stake" json:"stake"`
	Fees                int                   `bson:"fees" json:"fees"`
	Version             int                   `bson:"version" json:"version"`
	AnonDestination     string                `bson:"anonDestination" json:"anonDestination"`
	Inputs              []Input               `bson:"inputs" json:"inputs"`
	Outputs             []Output              `bson:"outputs" json:"outputs"`
	ProposalVotes       []ProposalVote        `bson:"proposalVotes" json:"proposalVotes"`
	PaymentRequestVotes []PaymentRequestVotes `bson:"paymentRequestVotes" json:"paymentRequestVotes"`
}

type Input struct {
	Index               int      `bson:"index" json:"index"`
	Addresses           []string `bson:"addresses" json:"addresses"`
	Amount              float64  `bson:"amount" json:"amount"`
	PreviousOutput      string   `bson:"previousOutput" json:"previousOutput"`
	PreviousOutputBlock int      `bson:"previousOutputBlock" json:"previousOutputBlock"`
}

type Output struct {
	Index      int        `bson:"index" json:"index"`
	Type       string     `bson:"type" json:"type"`
	Addresses  []string   `bson:"addresses" json:"addresses"`
	Amount     float64    `bson:"amount" json:"amount"`
	RedeemedIn RedeemedIn `bson:"redeemedIn" json:"redeemedIn"`
	Hash       string     `bson:"hash" json:"hash,omitempty"`
}

type RedeemedIn struct {
	Hash   string `bson:"hash" json:"hash,omitempty"`
	Height string `bson:"height" json:"height,omitempty"`
}

type ProposalVote struct {
	Hash string `bson:"hash" json:"hash"`
	Vote string `bson:"vote" json:"vote"`
}

type PaymentRequestVotes struct {
	Hash string `bson:"hash" json:"hash"`
	Vote string `bson:"vote" json:"vote"`
}