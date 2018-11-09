package softFork

import "github.com/globalsign/mgo/bson"

type SoftForks struct {
	BlockCycle      int      `json:"blockCycle"`
	BlocksInCycle   int      `json:"blocksInCycle"`
	FirstBlock      int      `json:"firstBlock"`
	CurrentBlock    int      `json:"currentBlock"`
	BlocksRemaining int      `json:"blocksRemaining"`
	BlocksRequired  int      `json:"blocksRequired"`
	SoftForks []    SoftFork `json:"softForks"`
}

type SoftFork struct {
	ID               bson.ObjectId `bson:"_id" json:"-"`
	Name             string        `bson:"name" json:"name"`
	SignalBit        int           `bson:"signalBit" json:"signalBit"`
	State            string        `bson:"state" json:"state"`
	BlocksSignalling int           `bson:"blocksSignalling" json:"blocksSignalling"`
	SignalledToBlock int           `bson:"signalledToBlock" json:"signalledToBlock"`
	LockedInHeight   int           `bson:"lockedInHeight" json:"lockedInHeight,omitempty"`
	ActivationHeight int           `bson:"activationHeight" json:"activationHeight,omitempty"`
}