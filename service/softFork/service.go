package softFork

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"log"
)

var IndexSoftFork = config.Get().Network + ".softfork"

func GetSoftForks() (softForks SoftForks, err error) {
	client := elasticsearch.NewClient()

	results, err := client.Search().Index(IndexSoftFork).Do(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	bestBlock, _ := block.GetBestBlock()

	softForks.BlocksInCycle = config.Get().SoftFork.BlocksInCycle
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
