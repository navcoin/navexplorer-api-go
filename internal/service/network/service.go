package network

import (
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/config"
	log "github.com/sirupsen/logrus"
)

type Network struct {
	Name  string
	Index string
}

var (
	ErrNetworkNotFound = errors.New("Network not found")
)

func (n *Network) ToString() string {
	return fmt.Sprintf("%s.%s", n.Name, n.Index)
}

func GetNetworks() []Network {
	networks := make([]Network, 0)
	for network, index := range config.Get().Index {
		networks = append(networks, Network{
			Name:  network,
			Index: index,
		})
	}

	return networks
}

func GetNetwork(name string) (Network, error) {
	for _, n := range GetNetworks() {
		if n.Name == name {
			return n, nil
		}
	}
	log.Errorf("Failed to find network (%s)", name)

	return Network{}, ErrNetworkNotFound
}
