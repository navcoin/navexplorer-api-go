package group

import (
	"time"
)

type TimeGroup struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (t *TimeGroup) SetRange(start time.Time, end time.Time) {
	t.Start = start
	t.End = end
}

func CreateTimeGroup(period *Period, size int) []*TimeGroup {
	groups := make([]*TimeGroup, 0)
	start := time.Now().UTC().Truncate(time.Second)

	for i := 0; i < size; i++ {
		var group *TimeGroup

		switch period {
		case &PeriodHourly:
			if i == 0 {
				group = &TimeGroup{
					Start: start.Truncate(time.Hour),
					End:   start,
				}
			} else {
				group = &TimeGroup{
					Start: groups[i-1].Start.Add(-time.Hour),
					End:   groups[i-1].Start,
				}
			}
			break
		case &PeriodDaily:
			if i == 0 {
				group = &TimeGroup{
					Start: time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()),
					End:   start,
				}
			} else {
				group = &TimeGroup{Start: groups[i-1].Start.AddDate(0, 0, -1), End: groups[i-1].Start}
			}
			break
		case &PeriodWeekly:
			if i == 0 {
				group = &TimeGroup{
					Start: time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()),
					End:   start,
				}
			} else {
				group = &TimeGroup{Start: groups[i-1].Start.AddDate(0, 0, -7), End: groups[i-1].Start}
			}
			break
		case &PeriodMonthly:
			if i == 0 {
				group = &TimeGroup{
					Start: time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location()),
					End:   start,
				}
			} else {
				group = &TimeGroup{Start: groups[i-1].Start.AddDate(0, -1, 0), End: groups[i-1].Start}
			}
			break
		}
		groups = append(groups, group)
	}

	return groups
}
