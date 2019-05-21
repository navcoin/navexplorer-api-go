package communityFund

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/elasticsearch"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"github.com/olivere/elastic"
	"log"
	"strconv"
	"strings"
)

var IndexBlockTransaction = ".blocktransaction"
var IndexProposal = ".communityfundproposal"
var IndexProposalVote = ".communityfundproposalvote"
var IndexPaymentRequest = ".communityfundpaymentrequest"
var IndexPaymentRequestVote = ".communityfundpaymentrequestvote"

func GetBlockCycle() (blockCycle BlockCycle) {
	network, _ := config.Get().Network()

	communityFund := network.CommunityFund

	blockCycle.BlocksInCycle = communityFund.BlocksInCycle
	blockCycle.MinQuorum = communityFund.MinQuorum
	blockCycle.ProposalVoting.Cycles = communityFund.ProposalVoting.Cycles
	blockCycle.ProposalVoting.Accept = communityFund.ProposalVoting.Accept
	blockCycle.ProposalVoting.Reject = communityFund.ProposalVoting.Reject
	blockCycle.PaymentVoting.Cycles = communityFund.PaymentVoting.Cycles
	blockCycle.PaymentVoting.Accept = communityFund.PaymentVoting.Accept
	blockCycle.PaymentVoting.Reject = communityFund.PaymentVoting.Reject

	bestBlock, _ := block.GetBestBlock()
	blockCycle.Height = bestBlock.Height

	blockCycle.Cycle = (blockCycle.Height) / (blockCycle.BlocksInCycle) + 1
	blockCycle.FirstBlock = (blockCycle.Height / blockCycle.BlocksInCycle) * blockCycle.BlocksInCycle
	blockCycle.CurrentBlock = blockCycle.Height - blockCycle.FirstBlock + 1
	blockCycle.BlocksRemaining = blockCycle.FirstBlock + blockCycle.BlocksInCycle - blockCycle.Height - 1

	return blockCycle
}

func GetStats() (stats Stats) {
	contributed, err := GetCommunityFundContributed()
	if err == nil {
		stats.Contributed = contributed
	}

	paid, locked, requested, err := GetCommunityFundPaidAndLocked()
	if err == nil {
		stats.Paid = paid
		stats.Locked = locked
		stats.Requested = requested
	}

	return stats
}

func GetCommunityFundContributed() (contributed float64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	contributionAgg := elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("outputs.type", "CFUND_CONTRIBUTION"))
	contributionAgg.SubAggregation("amount", elastic.NewSumAggregation().Field("outputs.amount"))

	agg := elastic.NewNestedAggregation().Path("outputs")
	agg.SubAggregation("outputs", contributionAgg)

	results, err := client.Search(config.Get().SelectedNetwork + IndexBlockTransaction).
		Aggregation("contribution", agg).
		Size(0).
		Do(context.Background())

	if agg, found := results.Aggregations.Nested("contribution"); found {
		if agg, found = agg.Aggregations.Filter("outputs"); found {
			if amount, found := agg.Aggregations.Sum("amount"); found {
				contributed = *amount.Value / 100000000
			}
		}
	}

	return contributed, err
}

func GetCommunityFundPaidAndLocked() (paid float64, locked float64, requested float64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	query = query.Should(
		elastic.NewMatchQuery("state", "ACCEPTED"),
		elastic.NewMatchQuery("state", "EXPIRED"))

	paidAgg := elastic.NewFilterAggregation().Filter(query)
	paidAgg.SubAggregation("requestedAmount", elastic.NewSumAggregation().Field("requestedAmount"))
	paidAgg.SubAggregation("notPaidYet", elastic.NewSumAggregation().Field("notPaidYet"))

	lockedAgg := elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("state", "ACCEPTED"))
	lockedAgg.SubAggregation("notPaidYet", elastic.NewSumAggregation().Field("notPaidYet"))

	requestedAgg := elastic.NewSumAggregation().Field("requestedAmount")

	results, err := client.Search(config.Get().SelectedNetwork + IndexProposal).
		Aggregation("paid", paidAgg).
		Aggregation("locked", lockedAgg).
		Aggregation("requested", requestedAgg).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	if stats, found := results.Aggregations.Filter("paid"); found {
		var requestedAmount float64
		if amount, found := stats.Aggregations.Sum("requestedAmount"); found {
			requestedAmount = *amount.Value
		}

		var notPaidYet float64
		if amount, found := stats.Aggregations.Sum("notPaidYet"); found {
			notPaidYet = *amount.Value
		}

		paid = requestedAmount - notPaidYet
	}

	if stats, found := results.Aggregations.Filter("locked"); found {
		if notPaidYet, found := stats.Aggregations.Sum("notPaidYet"); found {
			locked = *notPaidYet.Value
		}
	}

	if stats, found := results.Aggregations.Sum("requested"); found {
		requested = *stats.Value
	}

	return paid, locked, requested, err
}

func GetProposalsByState(state string, size int, ascending bool, page int) (proposals []Proposal, total int64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	if state != "" {
		query = query.Must(elastic.NewMatchQuery("state", state))
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexProposal).
		Query(query).
		Sort("height", ascending).
		From((page * size) - size).
		Size(size).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	for _, hit := range results.Hits.Hits {
		var proposal Proposal
		err := json.Unmarshal(*hit.Source, &proposal)
		if err == nil {
			proposals = append(proposals, proposal)
		}
	}

	return proposals, results.Hits.TotalHits, err
}

func GetProposalByHash(hash string) (proposal Proposal, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, _ := client.Search(config.Get().SelectedNetwork + IndexProposal).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		err = ErrProposalNotFound
		return
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &proposal)

	return proposal, err
}

func GetProposalPaymentRequests(hash string) (paymentRequests []PaymentRequest, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("proposalHash", hash))

	results, err := client.Search(config.Get().SelectedNetwork + IndexPaymentRequest).
		Query(query).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	for _, hit := range results.Hits.Hits {
		var paymentRequest PaymentRequest
		err := json.Unmarshal(*hit.Source, &paymentRequest)
		if err == nil {
			paymentRequests = append(paymentRequests, paymentRequest)
		}
	}

	return paymentRequests, err
}

func GetProposalVotes(hash string, vote bool) (votes []Votes, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	blockCycle := GetBlockCycle()

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("proposal", hash))
	query = query.Must(	elastic.NewMatchQuery("vote", vote))
	query = query.Must(elastic.NewRangeQuery("height").Gte(blockCycle.FirstBlock))

	aggregation := elastic.NewTermsAggregation().Field("address.keyword").OrderByCountDesc().Size(2147483647)

	results, err := client.Search(config.Get().SelectedNetwork + IndexProposalVote).
		Query(query).
		Aggregation("address", aggregation).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	if agg, found := results.Aggregations.Terms("address"); found {
		for _, bucket := range agg.Buckets {
			votes = append(votes, Votes{
				Address: bucket.Key.(string),
				Votes: bucket.DocCount,
			})
		}
	}

	return votes, err
}

func GetProposalTrend(hash string) (trends []Trend, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	blockCycle := GetBlockCycle()

	segments := 10
	segmentSize := blockCycle.BlocksInCycle / segments

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("proposal", hash))

	agg := elastic.NewRangeAggregation().Field("height")
	for segment := 0; segment < segments; segment++ {
		var end = blockCycle.Height - (segment * segmentSize)
		agg = agg.AddRange(end - segmentSize, end)
	}
	agg.SubAggregation("yes", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", true)))
	agg.SubAggregation("no", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", false)))

	results, _ := client.Search(config.Get().SelectedNetwork + IndexProposalVote).
		Query(query).
		Size(0).
		Aggregation("votes", agg).
		Do(context.Background())

	if agg, found := results.Aggregations.Terms("votes"); found {
		for _, bucket := range agg.Buckets {
			var trend Trend

			fromData, _ := bucket.Aggregations["from"].MarshalJSON()
			start, err := strconv.ParseFloat(strings.Trim(string(fromData[:]), "\""), 64)
			if err == nil {
				trend.Start = int(start)
			}

			toData, _ := bucket.Aggregations["to"].MarshalJSON()
			end, err := strconv.ParseFloat(strings.Trim(string(toData[:]), "\""), 64)
			if err == nil {
				trend.End = int(end)
			}

			yes, found := bucket.Filter("yes")
			if found == true {
				trend.VotesYes = int(yes.DocCount)
			}

			no, found := bucket.Filter("no")
			if found == true {
				trend.VotesNo = int(no.DocCount)
			}

			trend.TrendYes = (float64(trend.VotesYes) / float64(segmentSize)) * 100
			trend.TrendNo = (float64(trend.VotesNo) / float64(segmentSize)) * 100
			trend.TrendAbstain = (float64(segmentSize - trend.VotesYes - trend.VotesNo) / float64(segmentSize)) * 100

			trends = append(trends, trend)
		}
	}

	return trends, err
}

func GetPaymentRequestsByState(state string) (paymentRequests []PaymentRequest, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	if state != "" {
		query = query.Must(elastic.NewMatchQuery("state", state))
	}

	results, err := client.Search(config.Get().SelectedNetwork + IndexPaymentRequest).
		Query(query).
		Sort("createdAt", false).
		Size(1000).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	for _, hit := range results.Hits.Hits {
		var paymentRequest PaymentRequest
		err := json.Unmarshal(*hit.Source, &paymentRequest)
		if err == nil {
			paymentRequests = append(paymentRequests, paymentRequest)
		}
	}

	return paymentRequests, err
}

func GetPaymentRequestByHash(hash string) (paymentRequest PaymentRequest, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	results, _ := client.Search(config.Get().SelectedNetwork + IndexPaymentRequest).
		Query(elastic.NewMatchQuery("hash", hash)).
		Size(1).
		Do(context.Background())

	if results.TotalHits() == 0 {
		err = ErrPaymentRequestNotFound
		return
	}

	hit := results.Hits.Hits[0]
	err = json.Unmarshal(*hit.Source, &paymentRequest)

	return paymentRequest, err
}

func GetPaymentRequestVotes(hash string, vote bool) (votes []Votes, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	blockCycle := GetBlockCycle()

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("paymentRequest", hash))
	query = query.Must(elastic.NewMatchQuery("vote", vote))
	query = query.Must(elastic.NewRangeQuery("height").Gte(blockCycle.FirstBlock))

	aggregation := elastic.NewTermsAggregation().Field("address.keyword").OrderByCountDesc().Size(2147483647)

	results, err := client.Search(config.Get().SelectedNetwork + IndexPaymentRequestVote).
		Query(query).
		Aggregation("address", aggregation).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Print(err)
		return
	}

	if agg, found := results.Aggregations.Terms("address"); found {
		for _, bucket := range agg.Buckets {
			votes = append(votes, Votes{
				Address: bucket.Key.(string),
				Votes: bucket.DocCount,
			})
		}
	}

	if votes == nil {
		votes = make([]Votes, 0)
	}

	return votes, err
}

func GetPaymentRequestTrend(hash string) (trends []Trend, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	blockCycle := GetBlockCycle()

	segments := 10
	segmentSize := blockCycle.BlocksInCycle / segments

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewMatchQuery("paymentRequest", hash))

	agg := elastic.NewRangeAggregation().Field("height")
	for segment := 0; segment < segments; segment++ {
		var end = blockCycle.Height - (segment * segmentSize)
		agg = agg.AddRange(end - segmentSize, end)
	}
	agg.SubAggregation("yes", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", true)))
	agg.SubAggregation("no", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", false)))

	results, _ := client.Search(config.Get().SelectedNetwork + IndexPaymentRequestVote).
		Query(query).
		Size(0).
		Aggregation("votes", agg).
		Do(context.Background())

	if agg, found := results.Aggregations.Terms("votes"); found {
		for _, bucket := range agg.Buckets {
			var trend Trend

			fromData, _ := bucket.Aggregations["from"].MarshalJSON()
			start, err := strconv.ParseFloat(strings.Trim(string(fromData[:]), "\""), 64)
			if err == nil {
				trend.Start = int(start)
			}

			toData, _ := bucket.Aggregations["to"].MarshalJSON()
			end, err := strconv.ParseFloat(strings.Trim(string(toData[:]), "\""), 64)
			if err == nil {
				trend.End = int(end)
			}

			yes, found := bucket.Filter("yes")
			if found == true {
				trend.VotesYes = int(yes.DocCount)
			}

			no, found := bucket.Filter("no")
			if found == true {
				trend.VotesNo = int(no.DocCount)
			}

			trend.TrendYes = (float64(trend.VotesYes) / float64(segmentSize)) * 100
			trend.TrendNo = (float64(trend.VotesNo) / float64(segmentSize)) * 100
			trend.TrendAbstain = (float64(segmentSize - trend.VotesYes - trend.VotesNo) / float64(segmentSize)) * 100

			trends = append(trends, trend)
		}
	}

	return trends, err
}

var (
	ErrProposalNotFound = errors.New("proposal not found")
	ErrPaymentRequestNotFound = errors.New("payment request not found")
)