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

var IndexAddress = ".address"
var IndexAddressTransaction = ".addresstransaction"

func GetAddresses(size int) (addresses []Address, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddress).
		Sort("balance", false).
		Size(size).
		Do(context.Background())
	if err != nil {
		return
	}

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
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddress).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		err = ErrAddressNotFound
		return
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
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	position, err = client.Count(config.Get().SelectedNetwork + IndexAddress).
		Query(elastic.NewRangeQuery("balance").Gte(balance)).
		Do(context.Background())

	if err != nil {
		log.Print(err)
	}

	return position, err
}

func GetTransactions(address string, types string, size int, page int) (transactions []Transaction, total int64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", address))
	query = query.MustNot(elastic.NewTermQuery("standard", false))

	if len(types) != 0 {
		query = query.Must(elastic.NewMatchQuery("type", types))
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).
		Query(query).
		Sort("height", false).
		From((page * size) - size).
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

var (
	ErrAddressNotFound = errors.New("address not found")
)