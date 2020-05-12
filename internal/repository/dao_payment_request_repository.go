package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type DaoPaymentRequestRepository struct {
	elastic *elastic_cache.Index
}

var (
	ErrPaymentRequestNotFound = errors.New("Payment request not found")
)

func NewDaoPaymentRequestRepository(elastic *elastic_cache.Index) *DaoPaymentRequestRepository {
	return &DaoPaymentRequestRepository{elastic}
}

func (r *DaoPaymentRequestRepository) PaymentRequests(status *explorer.PaymentRequestStatus, dir bool, size int, page int) ([]*explorer.PaymentRequest, int64, error) {
	query := elastic.NewBoolQuery()
	if status != nil {
		query = query.Must(elastic.NewTermQuery("status.keyword", status.Status))
	}

	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get()).
		Query(query).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return r.findMany(results, err)
}

func (r *DaoPaymentRequestRepository) PaymentRequestsForProposal(proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error) {
	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get()).
		Query(elastic.NewTermQuery("proposalHash.keyword", proposal.Hash)).
		Size(999).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	paymentRequests, _, err := r.findMany(results, err)

	return paymentRequests, err
}

func (r *DaoPaymentRequestRepository) PaymentRequest(hash string) (*explorer.PaymentRequest, error) {
	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get()).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *DaoPaymentRequestRepository) ValuePaid() (*float64, error) {
	paidAgg := elastic.NewFilterAggregation().Filter(elastic.NewMatchQuery("status", explorer.PaymentRequestPaid))
	paidAgg.SubAggregation("requestedAmount", elastic.NewSumAggregation().Field("requestedAmount"))

	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get()).
		Aggregation("paid", paidAgg).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.WithError(err).Error("Failed to get value details")
		return nil, err
	}

	if stats, found := results.Aggregations.Filter("paid"); found {
		if requestedAmount, found := stats.Aggregations.Sum("requestedAmount"); found {
			return requestedAmount.Value, nil
		}
	}

	return nil, errors.New("Could not find paid aggregation")
}

func (r *DaoPaymentRequestRepository) findOne(results *elastic.SearchResult, err error) (*explorer.PaymentRequest, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrPaymentRequestNotFound
		return nil, err
	}

	var paymentRequest *explorer.PaymentRequest
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &paymentRequest)
	if err != nil {
		return nil, err
	}

	return paymentRequest, err
}

func (r *DaoPaymentRequestRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.PaymentRequest, int64, error) {
	if err != nil {
		return nil, 0, err
	}

	paymentRequests := make([]*explorer.PaymentRequest, 0)
	for _, hit := range results.Hits.Hits {
		var paymentRequest *explorer.PaymentRequest
		if err := json.Unmarshal(hit.Source, &paymentRequest); err == nil {
			paymentRequests = append(paymentRequests, paymentRequest)
		}
	}

	return paymentRequests, results.TotalHits(), err
}
