package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	log "github.com/sirupsen/logrus"
)

type VotingCycle struct {
	group.BlockGroup

	Index int
}

func CreateVotingCycles(segments int, size int, firstBlock int, bestHeight uint64, maxStart uint64) []*VotingCycle {
	votingCycles := make([]*VotingCycle, 0)

	for i := 0; i <= segments-1; i++ {
		votingCycle := &VotingCycle{Index: i}
		if i == 0 {
			votingCycle.Start = firstBlock
		} else {
			votingCycle.Start = votingCycles[i-1].End + 1
		}
		votingCycle.End = votingCycle.Start + size

		if votingCycle.Start > int(bestHeight) {
			log.Debugf("VotingCycle Start > best height (%d > %d)", votingCycle.Start, bestHeight)
			// Dont continue if cycles in the future
			return votingCycles
		}
		log.Debugf("Creating Voting Cycle: %d %d - %d -- Max:%d", votingCycle.Index, votingCycle.Start, votingCycle.End, maxStart)

		votingCycles = append(votingCycles, votingCycle)

		if int(maxStart) != 0 && votingCycle.Start > int(maxStart) {
			// Dont continue if entity transitioned in the previous cycle
			return votingCycles
		}
	}

	return votingCycles
}
