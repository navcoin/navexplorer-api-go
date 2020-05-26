package entity

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	log "github.com/sirupsen/logrus"
)

type VotingCycle struct {
	group.BlockGroup

	Index int
}

func CreateVotingCycles(segments uint, size uint, firstBlock uint, count uint) []*VotingCycle {
	log.WithFields(log.Fields{"segments": segments, "size": size, "firstBlock": firstBlock, "count": count}).Info("CreateVotingCycles")

	votingCycles := make([]*VotingCycle, 0)

	for i := 0; i <= int(segments)-1; i++ {
		votingCycle := &VotingCycle{Index: i}
		if i == 0 {
			votingCycle.Start = firstBlock
		} else {
			votingCycle.Start = votingCycles[i-1].End + 1
		}
		votingCycle.End = votingCycle.Start + size - 1

		if int(count) == len(votingCycles) {
			return votingCycles
		}

		log.Infof("Creating Voting Cycle: %d %d - %d -- Max:%d", votingCycle.Index, votingCycle.Start, votingCycle.End, int(count))
		votingCycles = append(votingCycles, votingCycle)
	}

	return votingCycles
}
