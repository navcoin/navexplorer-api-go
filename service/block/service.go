package block

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/olivere/elastic"
	"log"
	"strconv"
)

var IndexBlock = config.Get().Network + ".block"
var IndexBlockTransaction = config.Get().Network + ".blocktransaction"

func GetBlocks(size int, ascending bool, offset int) (blocks []Block, total int64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return blocks, 0, err
	}

	if size > 1000 {
		size = 1000
	}

	var offsetQuery *elastic.RangeQuery
	if ascending == false && offset > 0 {
		offsetQuery = elastic.NewRangeQuery("height").Lt(offset)
	} else {
		offsetQuery = elastic.NewRangeQuery("height").Gt(offset)
	}

	results, err := client.Search().Index(IndexBlock).
		Query(offsetQuery).
		Sort("height", ascending).
		Size(size).
		Do(context.Background())

	if err != nil {
		log.Print(err)
	}

	bestBlock, err := GetBestBlock()
	if err != nil {
		panic(err)
	}

	for _, hit := range results.Hits.Hits {
		var block Block
		err := json.Unmarshal(*hit.Source, &block)
		if err == nil {
			block.Confirmations = bestBlock.Height - block.Height + 1
			blocks = append(blocks, block)
		}
	}

	return blocks, results.Hits.TotalHits, err
}

func GetBlockByHashOrHeight(hash string) (block Block, err error) {
	block, err = GetBlockByHash(hash)
	if err != nil {
		height, _ := strconv.Atoi(hash)
		block, err = GetBlockByHeight(height)
	}

	bestBlock, err := GetBestBlock()
	if err != nil {
		panic(err)
	}

	block.Confirmations = bestBlock.Height - block.Height + 1

	return block, err
}

func GetBlockByHash(hash string) (block Block, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return block, err
	}

	results, _ := client.Search().Index(IndexBlock).
		Query(elastic.NewTermQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		return block, errors.New("block not found")
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &block)

	return block, err
}

func GetBlockByHeight(height int) (block Block, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return block, err
	}

	results, _ := client.Search().Index(IndexBlock).
		Query(elastic.NewTermQuery("height", height)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		return block, errors.New("block not found")
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &block)

	return block, err
}

func GetBestBlock() (block Block, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return block, err
	}

	results, _ := client.Search().Index(IndexBlock).
		Sort("height", false).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		return block, errors.New("block not found")
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &block)

	return block, err
}

func GetTransactionsByHash(blockHash string) (transactions []Transaction, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return transactions, err
	}

	results, _ := client.Search().Index(IndexBlockTransaction).
		Query(elastic.NewTermQuery("blockHash", blockHash)).
		Do(context.Background())

	if results.Hits.TotalHits == 0 {
		return make([]Transaction, 0), err
	}

	for _, hit := range results.Hits.Hits {
		var transaction Transaction
		err := json.Unmarshal(*hit.Source, &transaction)
		if err != nil {
		}

		transactions = append(transactions, transaction)
	}

	return transactions, err
}

func GetTransactionByHash(hash string) (transaction Transaction, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return transaction, err
	}

	results, _ := client.Search().Index(IndexBlockTransaction).
		Query(elastic.NewTermQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 1 {
		hit := results.Hits.Hits[0]
		err = json.Unmarshal(*hit.Source, &transaction)
	}

	return transaction, err
}
