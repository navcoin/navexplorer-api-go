package address

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/olivere/elastic"
	"log"
)

var IndexAddress = config.Get().Network + ".address"
var IndexAddressTransaction = config.Get().Network + ".addresstransaction"

func GetAddresses(size int) (addresses []Address, err error) {
	client := elasticsearch.NewClient()

	if size > 1000 {
		size = 1000
	}

	results, err := client.Search().Index(IndexAddress).
		Sort("balance", false).
		Size(size).
		Do(context.Background())

	for index, hit := range results.Hits.Hits {
		var address Address
		err := json.Unmarshal(*hit.Source, &address)
		if err == nil {
			address.RichListPosition = int64(index+1)
			addresses = append(addresses, address)
		}
	}

	return addresses, err
}

func GetAddress(hash string) (address Address, err error) {
	client := elasticsearch.NewClient()

	results, err := client.Search().Index(IndexAddress).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		return address, errors.New("address not found")
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &address)

	richListPosition, err := GetRichListPosition(address.Balance)
	if err == nil {
		address.RichListPosition = richListPosition
	}

	return address, err
}

func GetRichListPosition(balance float64) (position int64, err error) {
	client := elasticsearch.NewClient()

	return client.Count().Index(IndexAddress).
		Query(elastic.NewRangeQuery("balance").Gte(balance)).
		Do(context.Background())
}

func GetTransactions(address string, types string, size int, ascending bool, offset int) (transactions []Transaction, total int64, err error) {
	client := elasticsearch.NewClient()

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", address))
	query = query.MustNot(elastic.NewTermQuery("standard", false))

	if len(types) != 0 {
		query = query.Must(elastic.NewMatchQuery("type", types))
	}

	if ascending == false && offset > 0 {
		query = query.Must(elastic.NewRangeQuery("height").Lt(offset))
	} else {
		query = query.Must(elastic.NewRangeQuery("height").Gt(offset))
	}

	results, err := client.Search().Index(IndexAddressTransaction).
		Query(query).
		Sort("height", ascending).
		Size(size).
		Do(context.Background())

	if err != nil {
		log.Print(err)
	}

	for _, hit := range results.Hits.Hits {
		var transaction Transaction
		err := json.Unmarshal(*hit.Source, &transaction)
		if err == nil {
			transactions = append(transactions, transaction)
		}
	}

	return transactions, results.Hits.TotalHits, err
}
