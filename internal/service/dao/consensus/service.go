package consensus

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	consensusRepository *repository.DaoConsensusRepository
}

func NewConsensusService(consensusRepository *repository.DaoConsensusRepository) *Service {
	return &Service{consensusRepository}
}

func (s *Service) GetParameters() (*explorer.ConsensusParameters, error) {
	network := fmt.Sprintf("%v", param.GetGlobalParam("network", nil))
	if network == "" {
		log.Fatal("No network specified to get consensus parameters")
	}

	parameters := param.GetNetworkParam(network, "consensus", nil)
	if parameters != nil {
		return parameters.(*explorer.ConsensusParameters), nil
	}

	p, err := s.consensusRepository.GetConsensusParameters(network)
	if err != nil {
		log.WithError(err).Error("Failed to get consensus parameters")
		return nil, err
	}
	log.Info(p.Get(1))

	param.SetNetworkParam(network, "consensus", p)

	return p, nil
}

func (s *Service) GetParameter(parameter Parameter) *explorer.ConsensusParameter {
	parameters, _ := s.GetParameters()

	return parameters.Get(int(parameter))
}
