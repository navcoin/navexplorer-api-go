package elastic_cache

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
)

type Indices string

var (
	AddressIndex            Indices = "address"
	AddressTransactionIndex Indices = "addresstransaction"
	BlockIndex              Indices = "block"
	BlockTransactionIndex   Indices = "blocktransaction"
	ConsensusIndex          Indices = "consensus"
	CfundIndex              Indices = "cfund"
	ProposalIndex           Indices = "proposal"
	DaoVoteIndex            Indices = "daovote"
	DaoConsultationIndex    Indices = "consultation"
	PaymentRequestIndex     Indices = "paymentrequest"
	SignalIndex             Indices = "signal"
	SoftForkIndex           Indices = "softfork"
)

// Sets the network and returns the full string
func (i *Indices) Get() string {
	return fmt.Sprintf("%s.%s", param.GetGlobalParam("network", "mainnet"), string(*i))
}

func All() []Indices {
	return []Indices{
		AddressIndex,
		AddressTransactionIndex,
		BlockIndex,
		BlockTransactionIndex,
		CfundIndex,
		ConsensusIndex,
		ProposalIndex,
		PaymentRequestIndex,
		DaoVoteIndex,
		DaoConsultationIndex,
		SignalIndex,
		SoftForkIndex,
	}
}
