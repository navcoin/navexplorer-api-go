package group

import "time"

type Group struct {
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	Blocks       int64     `json:"blocks"`
	Stake        int64     `json:"stake"`
	Fees         int64     `json:"fees"`
	Spend        int64     `json:"spend"`
	Transactions int64     `json:"transactions"`
	Height       int64     `json:"height"`
}

func (g *Group) Window(start time.Time, end time.Time) *Group {
	g.Start = start
	g.End = end

	return g
}
