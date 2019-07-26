package address

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/navcoind"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"github.com/olivere/elastic"
	"log"
	"strings"
	"time"
)

var IndexAddress = ".address"
var IndexAddressTransaction = ".addresstransaction"

func GetAddresses(size int, page int) (addresses []Address, total int64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddress).
		Sort("balance", false).
		From((page * size) - size).
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

	return addresses, results.Hits.TotalHits, err
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
	if err != nil {
		return
	}

	if results.TotalHits() == 0 {
		validateAddress, err := ValidateAddress(hash)
		if err != nil {
			return address, err
		}

		if !validateAddress.Valid {
			err = ErrAddressNotValid
		} else {
			err = ErrAddressNotFound
		}

		return address, err
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &address)

	richListPosition, err := GetRichListPosition(address.Balance)
	if err == nil {
		address.RichListPosition = richListPosition
	}

	return address, err
}

func ValidateAddress(hash string) (validateAddress navcoind.ValidateAddress, err error) {
	nav, err := navcoind.New(config.Get().SelectedNetwork)
	if err != nil {
		return
	}

	return nav.ValidateAddress(hash)
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
		if strings.Contains(types,"staking") {
			types += " cold_staking"
		}
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

func GetColdTransactions(address string, types string, size int, page int) (transactions []Transaction, total int64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", address))
	query = query.Must(elastic.NewTermQuery("coldStaking", true))

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

func GetBalanceChart(address string) (chart Chart, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	now := time.Now().UTC().Truncate(time.Second)
	from := time.Date(now.Year(), now.Month(), now.Day()-30, 0, 0, 0, 0, now.Location())

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", address))
	query = query.Must(elastic.NewRangeQuery("time").Gte(from))

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).
		Query(query).
		Sort("height", false).
		Size(10000).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	for _, hit := range results.Hits.Hits {
		var transaction Transaction
		err := json.Unmarshal(*hit.Source, &transaction)
		if err == nil {
			var chartPoint ChartPoint
			chartPoint.Time = transaction.Time
			chartPoint.Value = transaction.Balance
			chart.Points = append(chart.Points, chartPoint)
		}
	}

	return chart, err
}

func GetStakingChart(period string, address string) (groups []StakingGroup, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	service := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).Size(0)

	count := 12
	now := time.Now().UTC().Truncate(time.Second)

	for i := 0; i < count; i++ {
		var group StakingGroup
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

		agg := elastic.NewRangeAggregation().Field("time").AddRange(group.Start, group.End)
		agg.SubAggregation("sent", elastic.NewSumAggregation().Field("sent"))
		agg.SubAggregation("received", elastic.NewSumAggregation().Field("received"))
		agg.SubAggregation("coldStakingSent", elastic.NewSumAggregation().Field("coldStakingSent"))
		agg.SubAggregation("coldStakingReceived", elastic.NewSumAggregation().Field("coldStakingReceived"))
		agg.SubAggregation("delegateStake", elastic.NewSumAggregation().Field("delegateStake"))

		query := elastic.NewBoolQuery()
		query = query.Must(elastic.NewMatchQuery("address", address))
		query = query.Must(elastic.NewMatchQuery("type", "STAKING COLD_STAKING"))
		service.Query(query)
		service.Aggregation(string(i), agg)

		groups = append(groups, group)
	}

	results, err := service.Do(context.Background())

	for i := 0; i < count; i++ {
		if agg, found := results.Aggregations.Range(string(i)); found {
			bucket := agg.Buckets[0]
			groups[i].Stakes = bucket.DocCount
			sent := int64(0)
			received := int64(0)
			if sentValue, found := bucket.Aggregations.Sum("sent"); found {
				sent = sent + int64(*sentValue.Value)
			}
			if coldStakingSentValue, found := bucket.Aggregations.Sum("coldStakingSent"); found {
				sent = sent + int64(*coldStakingSentValue.Value)
			}
			if receivedValue, found := bucket.Aggregations.Sum("received"); found {
				received = received + int64(*receivedValue.Value)
			}
			if coldStakingReceivedValue, found := bucket.Aggregations.Sum("coldStakingReceived"); found {
				received = received + int64(*coldStakingReceivedValue.Value)
			}
			if delegateStakeValue, found := bucket.Aggregations.Sum("delegateStake"); found {
				received = received + int64(*delegateStakeValue.Value)
			}

			groups[i].Amount = int64(received - sent)
		}
	}

	return groups, err
}

func GetBalancesForAddresses(addresses []string) (balances []Balance, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddress).
		Query(elastic.NewMatchQuery("hash", strings.Join(addresses, " "))).
		Size(5000).
		Do(context.Background())
	if err != nil {
		return
	}

	for _, hit := range results.Hits.Hits {
		var address Address
		err := json.Unmarshal(*hit.Source, &address)
		if err == nil {
			var balance Balance
			balance.Address = address.Hash
			balance.Balance = address.Balance
			balance.ColdStakedBalance = address.ColdStakedBalance
			balances = append(balances, balance)
		}
	}

	return balances, err
}

func GetTransactionsForAddresses(addresses []string, txType string, start *time.Time, end *time.Time) (transactions []Transaction, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("address", strings.Join(addresses, " ")))
	query = query.Must(elastic.NewMatchQuery("type", txType))
	query = query.Must(elastic.NewRangeQuery("time").Gt(&start).Lte(&end))

	results, err := client.Search(config.Get().SelectedNetwork + IndexAddressTransaction).
		Query(query).
		Size(5000).
		Sort("time", false).
		Do(context.Background())
	if err != nil {
		return
	}

	for _, hit := range results.Hits.Hits {
		var transaction Transaction
		err := json.Unmarshal(*hit.Source, &transaction)
		if err == nil {
			transactions = append(transactions, transaction)
		}
	}

	return
}

func GetAssociatedStakingAddresses(address string) (stakingAddresses []string, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	agg := elastic.NewNestedAggregation().Path("outputs")
	agg.SubAggregation("addresses", elastic.NewTermsAggregation().Field("outputs.addresses.keyword"))

	outputsQuery := elastic.NewBoolQuery()
	outputsQuery = outputsQuery.Must(elastic.NewMatchQuery("outputs.type", "COLD_STAKING"))
	outputsQuery = outputsQuery.Must(elastic.NewScriptQuery(elastic.NewScript("doc['outputs.addresses.keyword'][0] == params.address").Param("address", address)))

	query := elastic.NewNestedQuery("outputs", outputsQuery)

	results, err := client.Search(config.Get().SelectedNetwork + block.IndexBlockTransaction).
		Query(query).
		Size(50000000).
		Sort("time", false).
		Do(context.Background())
	if err != nil {
		return
	}

	stakingAddresses = make([]string, 0)
	for _, hit := range results.Hits.Hits {
		var transaction block.Transaction
		err := json.Unmarshal(*hit.Source, &transaction)
		if err == nil {
			for _, output := range transaction.Outputs {
				if len(output.Addresses) == 2 && output.Addresses[1] == address {
					if !contains(stakingAddresses, output.Addresses[0]) {
						stakingAddresses = append(stakingAddresses, output.Addresses[0])
					}
				}
			}
		}
	}

	return
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

var (
	ErrAddressNotFound = errors.New("address not found")
	ErrAddressNotValid = errors.New("address not valid")
)