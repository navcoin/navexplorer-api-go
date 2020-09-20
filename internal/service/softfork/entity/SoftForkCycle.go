package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
)

type SoftForkCycle struct {
	BlocksInCycle   uint64 `json:"blocksInCycle"`
	BlockCycle      uint64 `json:"blockCycle"`
	CurrentBlock    uint64 `json:"currentBlock"`
	FirstBlock      uint64 `json:"firstBlock"`
	RemainingBlocks uint64 `json:"remainingBlocks"`
}

func GetBlocksInCycle() uint64 {
	if param.GetGlobalParam("network", config.Get().DefaultNetwork) == "testnet" {
		return 800
	}

	return 20160
}
