package entity

type BlockCycle struct {
	Cycle           int `json:"cycle"`
	BlocksInCycle   int `json:"blocksInCycle"`
	FirstBlock      int `json:"firstBlock"`
	CurrentBlock    int `json:"currentBlock"`
	BlocksRemaining int `json:"blocksRemaining"`
}

func (bc *BlockCycle) LastBlock() int {
	return bc.FirstBlock + bc.BlocksInCycle - 1
}

type LegacyBlockCycle struct {
	BlocksInCycle   uint   `json:"blocksInCycle"`
	ProposalVoting  Voting `json:"proposalVoting"`
	PaymentVoting   Voting `json:"paymentVoting"`
	Cycle           uint   `json:"cycle"`
	FirstBlock      uint   `json:"firstBlock"`
	CurrentBlock    uint   `json:"currentBlock"`
	BlocksRemaining uint   `json:"blocksRemaining"`
}

type Voting struct {
	Quorum float64 `json:"quorum"`
	Cycles uint    `json:"cycles"`
	Accept uint    `json:"accept"`
	Reject uint    `json:"reject"`
}
