package group

import "time"

func CreateGroups(period string, size int) []*Group {
	var groups = make([]*Group, 0)
	var now = time.Now().UTC().Truncate(time.Second)

	for i := 0; i < size; i++ {
		var group = Group{Start: now, End: now}

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
