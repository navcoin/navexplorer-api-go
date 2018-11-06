package address

import (
	"time"
	"github.com/globalsign/mgo/bson"
)

type Address struct {
	ID               bson.ObjectId `bson:"_id" json:"id"`
	Hash             string        `bson:"hash" json:"hash"`
	Received	     float64       `bson:"received" json:"received"`
	ReceivedCount    int           `bson:"receivedCount" json:"receivedCount"`
	Sent             float64       `bson:"sent" json:"sent"`
	SentCount        int           `bson:"sentCount" json:"sentCount"`
	Staked           float64       `bson:"staked" json:"staked"`
	StakedCount      int           `bson:"stakedCount" json:"stakedCount"`
	StakedSent       float64       `bson:"stakedSent" json:"stakedSent"`
	StakedReceived   float64       `bson:"stakedReceived" json:"stakedReceived"`
	Balance          float64       `bson:"balance" json:"balance"`
	BlockIndex       int           `bson:"blockIndex" json:"blockIndex"`

	RichListPosition *int          `bson:"-" json:"richListPosition"`
}

type Transaction struct {
	ID               bson.ObjectId `bson:"_id" json:"id"`
	Time             time.Time     `bson:"time" json:"time"`
	Address          string        `bson:"address" json:"address"`
	Type             string        `bson:"type" json:"type"`
	Transaction      string        `bson:"transaction" json:"transaction"`
	Height           int           `bson:"height" json:"height"`
	Balance          float64       `bson:"balance" json:"balance"`
	Sent             float64       `bson:"sent" json:"sent"`
	Received         float64       `bson:"received" json:"received"`
}
