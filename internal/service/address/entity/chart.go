package entity

import "time"

type Chart struct {
	Points []*ChartPoint `json:"points"`
}

type ChartPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}
