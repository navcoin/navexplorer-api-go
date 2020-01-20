package entity

type SoftForkCycle struct {
	BlocksInCycle   uint64 `json:"blocksInCycle"`
	BlockCycle      uint64 `json:"blockCycle"`
	CurrentBlock    uint64 `json:"currentBlock"`
	FirstBlock      uint64 `json:"firstBlock"`
	RemainingBlocks uint64 `json:"remainingBlocks"`
}
