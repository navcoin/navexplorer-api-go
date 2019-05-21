package network

import (
	"context"
	"encoding/json"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
)

var IndexNodes = ".nodes"

func GetNodes() (nodes []Node, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, _ := client.Search(config.Get().SelectedNetwork + IndexNodes).
		Size(10000).
		Do(context.Background())

	if results.Hits.TotalHits.Value == 0 {
		return make([]Node, 0), err
	}

	for _, hit := range results.Hits.Hits {
		var node Node
		err = json.Unmarshal(*&hit.Source, &node)
		if err == nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, err
}
