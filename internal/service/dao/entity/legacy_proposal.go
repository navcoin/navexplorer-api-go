package entity

import "github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"

type LegacyProposal struct {
	explorer.RawProposal

	Height         uint64 `json:"height"`
	UpdatedOnBlock uint64 `json:"updatedOnBlock"`

	VotesYes    int `json:"votesYes"`
	VotesNo     int `json:"votesNo"`
	VotingCycle int `json:"votingCycle"`
}

func (p *LegacyProposal) GetHeight() uint64 {
	return p.Height
}
