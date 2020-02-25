package entity

import "github.com/NavExplorer/navexplorer-api-go/internal/framework"

type SoftForkCycle struct {
	BlocksInCycle   uint64 `json:"blocksInCycle"`
	BlockCycle      uint64 `json:"blockCycle"`
	CurrentBlock    uint64 `json:"currentBlock"`
	FirstBlock      uint64 `json:"firstBlock"`
	RemainingBlocks uint64 `json:"remainingBlocks"`
}

func GetBlocksInCycle() uint64 {
	if framework.GetParameter("network", "mainnet") == "testnet" {
		return 800
	} else {
		return 20160
	}
}
