package entity

import "github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"

type LegacyProposal struct {
	explorer.Proposal

	VotesYes    int `json:"votesYes"`
	VotesNo     int `json:"votesNo"`
	VotingCycle int `json:"votingCycle"`
}

func (p *LegacyProposal) GetHeight() uint64 {
	return p.Height
}
