package navcoind

import (
	"errors"
	"fmt"
	"github.com/NavPool/navpool-api/config"
)

const (
	VERSION           = 0.1
	RPCCLIENT_TIMEOUT = 30
)

type Navcoind struct {
	client *rpcClient
}

func New() (*Navcoind, error) {
	network := config.Get().Networks[0]
	if config.Get().SelectedNetwork == "testnet" {
		network = config.Get().Networks[1]
	}

	rpcClient, err := newClient(network.Host, network.Port, network.Username, network.Password)
	if err != nil {
		return nil, err
	}

	return &Navcoind{rpcClient}, nil
}

func HandleError(err error, r *rpcResponse) error {
	if err != nil {
		return err
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		return errors.New(fmt.Sprintf("(%v) %s", rr["code"].(float64), rr["message"].(string)))
	}

	return nil
}
