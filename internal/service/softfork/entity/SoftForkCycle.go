package entity

import "github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"

type SoftForkCycle struct {
	BlocksInCycle   uint64 `json:"blocksInCycle"`
	BlockCycle      uint64 `json:"blockCycle"`
	CurrentBlock    uint64 `json:"currentBlock"`
	FirstBlock      uint64 `json:"firstBlock"`
	RemainingBlocks uint64 `json:"remainingBlocks"`
}

func GetBlocksInCycle(network network.Network) uint64 {
	if network.Name != "mainnet" {
		return 800
	}

	return 20160
}
