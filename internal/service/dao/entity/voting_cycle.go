package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	log "github.com/sirupsen/logrus"
)

type VotingCycle struct {
	group.BlockGroup

	Index int
}

func CreateVotingCycles(segments uint, size uint, firstBlock uint, maxStart uint) []*VotingCycle {
	log.WithFields(log.Fields{"segments": segments, "size": size, "firstBlock": firstBlock, "maxStart": maxStart}).Info("CreateVotingCycles")

	votingCycles := make([]*VotingCycle, 0)

	for i := 0; i <= int(segments)-1; i++ {
		votingCycle := &VotingCycle{Index: i}
		if i == 0 {
			votingCycle.Start = firstBlock
		} else {
			votingCycle.Start = votingCycles[i-1].End + 1
		}
		votingCycle.End = votingCycle.Start + size - 1

		if maxStart != 0 && votingCycle.Start > maxStart {
			log.Debugf("VotingCycle Start > maxStart (%d > %d)", votingCycle.Start, maxStart)
			// Dont continue if cycles in the future
			return votingCycles
		}

		log.Debugf("Creating Voting Cycle: %d %d - %d -- Max:%d", votingCycle.Index, votingCycle.Start, votingCycle.End, int(maxStart))
		votingCycles = append(votingCycles, votingCycle)
	}

	return votingCycles
}
