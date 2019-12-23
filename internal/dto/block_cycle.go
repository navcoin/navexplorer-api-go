package dto

type BlockCycle struct {
	BlocksInCycle   uint    `json:"blocksInCycle"`
	Quorum          float64 `json:"minQuorum"`
	ProposalVoting  Voting  `json:"proposalVoting"`
	PaymentVoting   Voting  `json:"paymentVoting"`
	Height          uint64  `json:"height"`
	Cycle           uint    `json:"cycle"`
	FirstBlock      uint    `json:"firstBlock"`
	CurrentBlock    uint    `json:"currentBlock"`
	BlocksRemaining uint    `json:"blocksRemaining"`
}

type Voting struct {
	Cycles uint `json:"cycles"`
	Accept uint `json:"accept"`
	Reject uint `json:"reject"`
}
