package softFork

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"log"
)

var IndexSoftFork = ".softfork"

func GetSoftForks() (softForks SoftForks, err error) {
	var network = 0
	if config.Get().SelectedNetwork == "testnet" {
		network = 1
	}

	client, err := elasticsearch.NewClient()
	if err != nil {
		return softForks, err
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexSoftFork).Do(context.Background())
	if err != nil {
		log.Print(err)
		return
	}

	bestBlock, err := block.GetBestBlock()
	if err != nil {
		log.Print(err)
		return
	}

	softForks.BlocksInCycle = config.Get().Networks[network].SoftFork.BlocksInCycle
	softForks.CurrentBlock = bestBlock.Height
	softForks.BlockCycle = (softForks.CurrentBlock) / (softForks.BlocksInCycle) + 1
	softForks.FirstBlock = (softForks.CurrentBlock / softForks.BlocksInCycle) * softForks.BlocksInCycle
	softForks.BlocksRemaining = softForks.FirstBlock + softForks.BlocksInCycle - softForks.CurrentBlock

	for _, hit := range results.Hits.Hits {
		var softFork SoftFork
		err := json.Unmarshal(*hit.Source, &softFork)
		if err == nil {
			softForks.SoftForks = append(softForks.SoftForks, softFork)
		}
	}

	return softForks, err
}
