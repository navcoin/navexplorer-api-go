package dto

type CfundVote struct {
	Cycle int `json:"cycle"`
	Start int
	End   int
	Vote  CfundVoteGroup `json:"vote"`
}

type CfundVoteGroup struct {
	Yes     int `json:"yes"`
	No      int `json:"no"`
	Abstain int `json:"abstain"`
}
