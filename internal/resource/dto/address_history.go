package dto

type HistoryParameters struct {
	TxType TxType `form:"type"`
}

type TxType string

const (
	Stake   TxType = "stake"
	Send    TxType = "send"
	Receive TxType = "receive"
)
