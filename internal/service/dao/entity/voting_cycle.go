package entity

import "github.com/NavExplorer/navexplorer-api-go/internal/service/group"

type VotingCycle struct {
	group.BlockGroup

	Index int
}

func CreateVotingCycles(segments int, size int, firstBlock int) []*VotingCycle {
	votingCycles := make([]*VotingCycle, segments+1)

	for i := 0; i <= segments; i++ {
		votingCycles[i] = &VotingCycle{Index: i}
		if i == 0 {
			votingCycles[i].Start = firstBlock
		} else {
			votingCycles[i].Start = votingCycles[i-1].End + 1
		}
		votingCycles[i].End = votingCycles[i].Start + size - 1
	}

	return votingCycles
}
