package elastic_cache

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/network"
)

type Indices string

var (
	AddressIndex          Indices = "address"
	AddressHistoryIndex   Indices = "addresshistory"
	BlockIndex            Indices = "block"
	BlockTransactionIndex Indices = "blocktransaction"
	ConsensusIndex        Indices = "consensus"
	ProposalIndex         Indices = "proposal"
	DaoVoteIndex          Indices = "daovote"
	DaoConsultationIndex  Indices = "consultation"
	PaymentRequestIndex   Indices = "paymentrequest"
	SignalIndex           Indices = "signal"
	SoftForkIndex         Indices = "softfork"
)

func (i *Indices) Get(network network.Network) string {
	return fmt.Sprintf("%s.%s", network.ToString(), string(*i))
}
