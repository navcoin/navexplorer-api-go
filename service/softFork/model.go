package softFork

type SoftForks struct {
	BlockCycle      int      `json:"blockCycle"`
	BlocksInCycle   int      `json:"blocksInCycle"`
	FirstBlock      int      `json:"firstBlock"`
	CurrentBlock    int      `json:"currentBlock"`
	BlocksRemaining int      `json:"blocksRemaining"`
	SoftForks []    SoftFork `json:"softForks"`
}

type SoftFork struct {
	Name             string `json:"name"`
	SignalBit        int    `json:"signalBit"`
	State            string `json:"state"`
	BlocksSignalling int    `json:"blocksSignalling"`
	SignalledToBlock int    `json:"signalledToBlock"`
	LockedInHeight   int    `json:"lockedInHeight,omitempty"`
	ActivationHeight int    `json:"activationHeight,omitempty"`
}