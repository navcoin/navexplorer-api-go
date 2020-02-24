package entity

import "github.com/NavExplorer/navexplorer-api-go/internal/service/group"

type VotingCycle struct {
	group.BlockGroup

	Index int
}

func CreateVotingCycles(segments int, size int, firstBlock int, bestHeight uint64) []*VotingCycle {
	votingCycles := make([]*VotingCycle, 0)

	for i := 0; i <= segments-1; i++ {
		votingCycle := &VotingCycle{Index: i}
		if i == 0 {
			votingCycle.Start = firstBlock
		} else {
			votingCycle.Start = votingCycles[i-1].End + 1
		}
		votingCycle.End = votingCycle.Start + size - 1

		if votingCycle.Start > int(bestHeight) {
			break
		}
		votingCycles = append(votingCycles, votingCycle)
	}

	return votingCycles
}
