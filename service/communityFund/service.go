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
)

var IndexProposal = config.Get().Network + ".communityfundproposal"
var IndexProposalVote = config.Get().Network + ".communityfundproposalvote"
var IndexPaymentRequest = config.Get().Network + ".communityfundpaymentrequest"
var IndexPaymentRequestVote = config.Get().Network + ".communityfundpaymentrequestvote"

func GetBlockCycle() (blockCycle BlockCycle) {
	communityFund := config.Get().CommunityFund

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

func GetProposalsByState(state string, size int, ascending bool, offset int) (proposals []Proposal, total int64, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	query := elastic.NewBoolQuery()
	if state != "" {
		query = query.Must(elastic.NewMatchQuery("state", state))
	}

	if ascending == false && offset > 0 {
		query = query.Must(elastic.NewRangeQuery("height").Lt(offset))
	} else {
		query = query.Must(elastic.NewRangeQuery("height").Gt(offset))
	}

	results, err := client.Search().Index(IndexProposal).
		Pretty(true).
		Query(query).
		Sort("height", ascending).
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

	results, _ := client.Search(IndexProposal).
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

	results, err := client.Search().Index(IndexPaymentRequest).
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

	aggregation := elastic.NewTermsAggregation().Field("address").OrderByCountDesc().Size(2147483647)

	results, err := client.Search(IndexProposalVote).
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
		return trends, err
	}

	blockCycle := GetBlockCycle()

	height := blockCycle.Height
	segments := 10
	segmentSize := blockCycle.BlocksInCycle / segments

	for segment := 0; segment < segments; segment++ {
		var trend Trend

		trend.End = height - (segment * segmentSize)
		trend.Start = trend.End - segmentSize

		query := elastic.NewBoolQuery()
		query = query.Must(elastic.NewMatchQuery("proposal", hash))
		query = query.Must(elastic.NewRangeQuery("height").From(trend.Start).To(trend.End).IncludeLower(false))

		results, err := client.Search().Index(IndexProposalVote).
			Query(query).
			Size(0).
			Aggregation("VotesFor", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", true))).
			Aggregation("VotesAgainst", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", false))).
			Do(context.Background())

		if err != nil {
			log.Fatal(err)
		}

		if agg, found := results.Aggregations.Filter("VotesFor"); found {
			trend.VotesYes = int(agg.DocCount)
		}

		if agg, found := results.Aggregations.Filter("VotesAgainst"); found {
			trend.VotesNo = int(agg.DocCount)
		}

		trend.TrendYes = (float64(trend.VotesYes) / float64(segmentSize)) * 100
		trend.TrendNo = (float64(trend.VotesNo) / float64(segmentSize)) * 100
		trend.TrendAbstain = (float64(segmentSize - trend.VotesYes - trend.VotesNo) / float64(segmentSize)) * 100

		trends = append(trends, trend)
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

	results, err := client.Search().Index(IndexPaymentRequest).
		Query(query).
		Sort("createdAt", false).
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

	results, _ := client.Search(IndexPaymentRequest).
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

	aggregation := elastic.NewTermsAggregation().Field("address").OrderByCountDesc().Size(2147483647)

	results, err := client.Search(IndexPaymentRequestVote).
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

func GetPaymentRequestTrend(hash string) (trends []Trend, err error) {
	client, err := elasticsearch.NewClient()
	if err != nil {
		return
	}

	blockCycle := GetBlockCycle()

	height := blockCycle.Height
	segments := 10
	segmentSize := blockCycle.BlocksInCycle / segments

	for segment := 0; segment < segments; segment++ {
		var trend Trend

		trend.End = height - (segment * segmentSize)
		trend.Start = trend.End - segmentSize

		query := elastic.NewBoolQuery()
		query = query.Must(elastic.NewMatchQuery("paymentRequest", hash))
		query = query.Must(elastic.NewRangeQuery("height").From(trend.Start).To(trend.End).IncludeLower(false))

		results, err := client.Search().Index(IndexPaymentRequestVote).Pretty(true).
			Query(query).
			Size(0).
			Aggregation("VotesFor", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", true))).
			Aggregation("VotesAgainst", elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("vote", false))).
			Do(context.Background())

		if err != nil {
			log.Print(err)
			return
		}

		if agg, found := results.Aggregations.Filter("VotesFor"); found {
			trend.VotesYes = int(agg.DocCount)
		}

		if agg, found := results.Aggregations.Filter("VotesAgainst"); found {
			trend.VotesNo = int(agg.DocCount)
		}

		trend.TrendYes = (float64(trend.VotesYes) / float64(segmentSize)) * 100
		trend.TrendNo = (float64(trend.VotesNo) / float64(segmentSize)) * 100
		trend.TrendAbstain = (float64(segmentSize - trend.VotesYes - trend.VotesNo) / float64(segmentSize)) * 100

		trends = append(trends, trend)
	}

	return trends, err
}

var (
	ErrProposalNotFound = errors.New("proposal not found")
	ErrPaymentRequestNotFound = errors.New("payment request not found")
)