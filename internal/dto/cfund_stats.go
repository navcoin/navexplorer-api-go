package dto

type CfundStats struct {
	Contributed float64 `json:"contributed"`
	Available   float64 `json:"available"`
	Paid        float64 `json:"paid"`
	Locked      float64 `json:"locked"`
}
