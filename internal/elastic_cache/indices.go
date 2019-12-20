package elastic_cache

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework"
)

type Indices string

var (
	AddressIndex            Indices = "address"
	AddressTransactionIndex Indices = "addresstransaction"
	BlockIndex              Indices = "block"
	BlockTransactionIndex   Indices = "blocktransaction"
	ConsensusIndex          Indices = "daoconsensus"
	ProposalIndex           Indices = "proposal"
	DaoVoteIndex            Indices = "daovote"
	PaymentRequestIndex     Indices = "paymentrequest"
	SignalIndex             Indices = "signal"
	SoftForkIndex           Indices = "softfork"
)

// Sets the network and returns the full string
func (i *Indices) Get() string {
	return fmt.Sprintf("%s.%s", framework.GetParameter("network", "mainnet"), string(*i))
}

func All() []Indices {
	return []Indices{
		AddressIndex,
		AddressTransactionIndex,
		BlockIndex,
		BlockTransactionIndex,
		ConsensusIndex,
		ProposalIndex,
		PaymentRequestIndex,
		DaoVoteIndex,
		SignalIndex,
		SoftForkIndex,
	}
}
