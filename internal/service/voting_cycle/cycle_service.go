package voting_cycle

type Cycle struct {
	Index int
	Start int
	End   int
}

func CreateVotingCycles(segments int, size int, firstBlock int) []*Cycle {
	votingCycles := make([]*Cycle, segments+1)

	for i := 0; i <= segments; i++ {
		votingCycles[i] = &Cycle{Index: i}
		if i == 0 {
			votingCycles[i].Start = firstBlock
		} else {
			votingCycles[i].Start = votingCycles[i-1].End + 1
		}
		votingCycles[i].End = votingCycles[i].Start + size - 1
	}

	return votingCycles
}
