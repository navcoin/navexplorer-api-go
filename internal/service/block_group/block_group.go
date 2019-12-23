package block_group

import "time"

type BlockGroup struct {
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	Blocks       int64     `json:"blocks"`
	Stake        int64     `json:"stake"`
	Fees         int64     `json:"fees"`
	Spend        int64     `json:"spend"`
	Transactions int64     `json:"transactions"`
	Height       int64     `json:"height"`
}

func (g *BlockGroup) Window(start time.Time, end time.Time) *BlockGroup {
	g.Start = start
	g.End = end

	return g
}

func CreateGroups(period string, size int) []*BlockGroup {
	var groups = make([]*BlockGroup, 0)
	var now = time.Now().UTC().Truncate(time.Second)

	for i := 0; i < size; i++ {
		var group = BlockGroup{Start: now, End: now}

		switch period {
		case "hourly":
			{
				if i == 0 {
					group.Start = now.Truncate(time.Hour)
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.Add(-time.Hour)
				}
				break
			}
		case "daily":
			{
				if i == 0 {
					group.Start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.AddDate(0, 0, -1)
				}
				break
			}
		case "monthly":
			{
				if i == 0 {
					group.Start = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location())
					group.Start = group.Start.AddDate(0, 0, 1)
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.AddDate(0, -1, 0)
				}
				break
			}
		}

		groups = append(groups, &group)
	}

	return groups
}
