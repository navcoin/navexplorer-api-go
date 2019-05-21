package helper

import (
	"errors"
	"time"
)

var (
	ErrPeriodNotFound = errors.New("Unable to retrieve stats")
)

func PeriodToStartDate(period string) (start time.Time, err error) {
	now := time.Now()

	switch period {
	case "30d":
		start = now.Add(- (time.Hour * 24 * 30))
		break
	case "7d":
		start = now.Add(- (time.Hour * 24 * 7))
		break
	case "24h":
		start = now.Add(- (time.Hour * 24))
		break
	default:
		err = ErrPeriodNotFound
	}

	return
}
