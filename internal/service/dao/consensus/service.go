package consensus

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetParameters(n network.Network) (*explorer.ConsensusParameters, error)
	GetParameter(n network.Network, parameter Parameter) *explorer.ConsensusParameter
}

type service struct {
	consensusRepository repository.DaoConsensusRepository
}

func NewConsensusService(consensusRepository repository.DaoConsensusRepository) Service {
	return &service{consensusRepository}
}

func (s *service) GetParameters(n network.Network) (*explorer.ConsensusParameters, error) {
	p, err := s.consensusRepository.GetConsensusParameters(n)
	if err != nil {
		log.WithError(err).Error("Failed to get consensus parameters")
		return nil, err
	}

	return p, nil
}

func (s *service) GetParameter(n network.Network, parameter Parameter) *explorer.ConsensusParameter {
	parameters, _ := s.GetParameters(n)

	return parameters.Get(int(parameter))
}
