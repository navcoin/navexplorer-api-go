package network

import "time"

type Node struct {
	Address     string    `json:"address"`
	Good        bool      `json:"good"`
	LastSuccess time.Time `json:"lastSuccess"`
	Percent2h   float64   `json:"percent2h"`
	Percent8h   float64   `json:"percent8h"`
	Percent1d   float64   `json:"percent1d"`
	Percent7d   float64   `json:"percent7d"`
	Percent30d  float64   `json:"percent30d"`
	Blocks      int64     `json:"blocks"`
	Svcs        string    `json:"svcs"`
	Version     string    `json:"version"`
	UserAgent   string    `json:"userAgent"`
}
