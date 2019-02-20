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
	"time"
)

var IndexBlock = ".block"
var IndexBlockTransaction = ".blocktransaction"

func GetBlocks(size int, ascending bool, page int) (blocks []Block, total int64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	if size > 1000 {
		size = 1000
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexBlock).
		Sort("height", ascending).
		From((page * size) - size).
		Size(size).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	bestBlock, err := GetBestBlock()
	if err != nil {
		log.Print(err)
		return
	}

	for _, hit := range results.Hits.Hits {
		var block Block
		err := json.Unmarshal(*hit.Source, &block)
		if err == nil {
			block.Confirmations = bestBlock.Height - block.Height + 1
			block.Best = block.Height == bestBlock.Height

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

	if err != nil {
		log.Print(err)
		return
	}

	bestBlock, err := GetBestBlock()
	if err != nil {
		log.Print(err)
		return
	}

	block.Best = block.Height == bestBlock.Height
	block.Confirmations = bestBlock.Height - block.Height + 1

	return block, err
}

func GetBlockByHash(hash string) (block Block, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, _ := client.Search(config.Get().SelectedNetwork + IndexBlock).
		Query(elastic.NewTermQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		err = ErrBlockNotFound
		return
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &block)

	return block, err
}

func GetBlockByHeight(height int) (block Block, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, _ := client.Search(config.Get().SelectedNetwork + IndexBlock).
		Query(elastic.NewTermQuery("height", height)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		err = ErrBlockNotFound
		return
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &block)

	return block, err
}

func GetBestBlock() (block Block, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, _ := client.Search().Index(config.Get().SelectedNetwork + IndexBlock).
		Sort("height", false).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		err = ErrBlockNotFound
		return
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &block)

	return block, err
}

func GetTransactionsByHash(blockHash string) (transactions []Transaction, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, _ := client.Search(config.Get().SelectedNetwork + IndexBlockTransaction).
		Query(elastic.NewTermQuery("blockHash", blockHash)).
		Do(context.Background())

	if results.Hits.TotalHits == 0 {
		return make([]Transaction, 0), err
	}

	for _, hit := range results.Hits.Hits {
		var transaction Transaction
		json.Unmarshal(*hit.Source, &transaction)

		transactions = append(transactions, transaction)
	}

	return transactions, err
}

func GetTransactionByHash(hash string) (transaction Transaction, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexBlockTransaction).
		Query(elastic.NewTermQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if err != nil {
		return
	}

	if results.TotalHits() == 1 {
		hit := results.Hits.Hits[0]
		err = json.Unmarshal(*hit.Source, &transaction)
	}

	return transaction, err
}

func GetBlockGroups(period string, count int) (groups []Group, err error) {
	groups, err = GetGroupsForPeriod(period, count)

	return groups, err
}

func GetGroupsForPeriod(period string, count int) (groups []Group, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	service := client.Search(config.Get().SelectedNetwork + IndexBlock).Size(0)

	now := time.Now().UTC().Truncate(time.Second)

	for i := 0; i < count; i++ {
		var group Group
		group.End = now

		switch period {
		case "hourly":
			{
				if i == 0 {
					group.Start = now.Truncate(time.Hour)
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.Add(- time.Hour)
				}
				break
			}
		case "daily":
			{
				if i == 0 {
					group.Start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.AddDate(0,0, -1)
				}
				break
			}
		case "monthly":
			{
				if i == 0 {
					group.Start = time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location())
					group.Start = group.Start.AddDate(0, 0, 1)
				} else {
					group.End = groups[i-1].Start
					group.Start = group.End.AddDate(0,-1, 0)
				}
				break
			}
		}

		agg := elastic.NewRangeAggregation().Field("created").AddRange(group.Start, group.End)
		agg.SubAggregation("stake", elastic.NewSumAggregation().Field("stake"))
		agg.SubAggregation("fees", elastic.NewSumAggregation().Field("fees"))
		agg.SubAggregation("spend", elastic.NewSumAggregation().Field("spend"))
		agg.SubAggregation("transactions", elastic.NewSumAggregation().Field("transactions"))
		agg.SubAggregation("height", elastic.NewMaxAggregation().Field("height"))

		service.Aggregation(string(i), agg)

		groups = append(groups, group)
	}

	results, err := service.Do(context.Background())

	for i := 0; i < count; i++ {
		if agg, found := results.Aggregations.Range(string(i)); found {
			bucket := agg.Buckets[0]
			groups[i].Blocks = bucket.DocCount
			if stake, found := bucket.Aggregations.Sum("stake"); found {
				groups[i].Stake = int64(*stake.Value)
			}
			if fees, found := bucket.Aggregations.Sum("fees"); found {
				groups[i].Fees = int64(*fees.Value)
			}

			if spend, found := bucket.Aggregations.Sum("spend"); found {
				groups[i].Spend = int64(*spend.Value)
			}

			if transactions, found := bucket.Aggregations.Sum("transactions"); found {
				groups[i].Transactions = int64(*transactions.Value)
			}

			if height, found := bucket.Aggregations.Max("height"); found {
				groups[i].Height = int64(*height.Value)
			}
		}
	}

	return groups, err
}

var (
	ErrNoBlocksFound = errors.New("no blocks not found")
	ErrBlockNotFound = errors.New("block not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)