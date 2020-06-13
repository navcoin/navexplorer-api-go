package consensus

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetParameters() (*explorer.ConsensusParameters, error)
	GetParameter(parameter Parameter) *explorer.ConsensusParameter
}

type service struct {
	consensusRepository *repository.DaoConsensusRepository
}

func NewConsensusService(consensusRepository *repository.DaoConsensusRepository) Service {
	return &service{consensusRepository}
}

func (s *service) GetParameters() (*explorer.ConsensusParameters, error) {
	network := fmt.Sprintf("%v", param.GetGlobalParam("network", nil))
	if network == "" {
		log.Fatal("No network specified to get consensus parameters")
	}

	p, err := s.consensusRepository.GetConsensusParameters(network)
	if err != nil {
		log.WithError(err).Error("Failed to get consensus parameters")
		return nil, err
	}

	return p, nil
}

func (s *service) GetParameter(parameter Parameter) *explorer.ConsensusParameter {
	parameters, _ := s.GetParameters()

	return parameters.Get(int(parameter))
}
