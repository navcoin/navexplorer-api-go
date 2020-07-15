package elastic_cache

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
)

type Indices string

var (
	AddressIndex            Indices = "address"
	AddressTransactionIndex Indices = "addresstransaction"
	BlockIndex              Indices = "block"
	BlockTransactionIndex   Indices = "blocktransaction"
	ConsensusIndex          Indices = "consensus"
	ProposalIndex           Indices = "proposal"
	DaoVoteIndex            Indices = "daovote"
	DaoConsultationIndex    Indices = "consultation"
	PaymentRequestIndex     Indices = "paymentrequest"
	SignalIndex             Indices = "signal"
	SoftForkIndex           Indices = "softfork"
)

func (i *Indices) Get() string {
	network := param.GetGlobalParam("network", "mainnet").(string)
	index := config.Get().Index[network]

	if network == "mainnet" && string(*i) == "softfork" {
		return fmt.Sprintf("%s.%s", network, string(*i))
	}

	return fmt.Sprintf("%s.%s.%s", network, index, string(*i))
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
		DaoConsultationIndex,
		SignalIndex,
		SoftForkIndex,
	}
}
