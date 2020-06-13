package group

type Period string

var (
	PeriodHourly  Period = "hourly"
	PeriodDaily   Period = "daily"
	PeriodWeekly  Period = "weekly"
	PeriodMonthly Period = "monthly"
)

func GetPeriod(period string) *Period {
	if string(PeriodHourly) == period {
		return &PeriodHourly
	}
	if string(PeriodDaily) == period {
		return &PeriodDaily
	}
	if string(PeriodWeekly) == period {
		return &PeriodWeekly
	}
	if string(PeriodMonthly) == period {
		return &PeriodMonthly
	}

	return nil
}
