package elastic_cache

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
	log "github.com/sirupsen/logrus"
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

func (i *Indices) Get() string {
	network := param.GetGlobalParam("network", config.Get().DefaultNetwork).(string)
	index := config.Get().Index[network]

	if network == "mainnet" && string(*i) == "softfork" {
		return fmt.Sprintf("%s.%s", network, string(*i))
	}

	indexName := fmt.Sprintf("%s.%s.%s", network, index, string(*i))
	log.Info("Using index ", indexName)

	return indexName
}

func All() []Indices {
	return []Indices{
		AddressIndex,
		AddressHistoryIndex,
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
