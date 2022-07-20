package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/navcoin/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/network"
	"github.com/navcoin/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type DaoPaymentRequestRepository interface {
	GetPaymentRequests(n network.Network, hash string, status *explorer.PaymentRequestStatus, dir bool, size int, page int) ([]*explorer.PaymentRequest, int64, error)
	GetPaymentRequestsForProposal(n network.Network, proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error)
	GetPaymentRequest(n network.Network, hash string) (*explorer.PaymentRequest, error)
	GetValuePaid(n network.Network) (*float64, error)
}

type daoPaymentRequestRepository struct {
	elastic *elastic_cache.Index
}

var (
	ErrPaymentRequestNotFound = errors.New("Payment request not found")
)

func NewDaoPaymentRequestRepository(elastic *elastic_cache.Index) DaoPaymentRequestRepository {
	return &daoPaymentRequestRepository{elastic: elastic}
}

func (r *daoPaymentRequestRepository) GetPaymentRequests(n network.Network, hash string, status *explorer.PaymentRequestStatus, dir bool, size int, page int) ([]*explorer.PaymentRequest, int64, error) {
	query := elastic.NewBoolQuery()
	if hash != "" {
		query = query.Must(elastic.NewTermQuery("proposalHash.keyword", hash))
	}
	if status != nil {
		query = query.Must(elastic.NewTermQuery("status.keyword", status.Status))
	}

	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get(n)).
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

func (r *daoPaymentRequestRepository) GetPaymentRequestsForProposal(n network.Network, proposal *explorer.Proposal) ([]*explorer.PaymentRequest, error) {
	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get(n)).
		Query(elastic.NewTermQuery("proposalHash.keyword", proposal.Hash)).
		Size(999).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	paymentRequests, _, err := r.findMany(results, err)

	return paymentRequests, err
}

func (r *daoPaymentRequestRepository) GetPaymentRequest(n network.Network, hash string) (*explorer.PaymentRequest, error) {
	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get(n)).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *daoPaymentRequestRepository) GetValuePaid(n network.Network) (*float64, error) {
	paidAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("state.keyword", explorer.PaymentRequestPaid.State))
	paidAgg.SubAggregation("requestedAmount", elastic.NewSumAggregation().Field("requestedAmount"))

	results, err := r.elastic.Client.Search(elastic_cache.PaymentRequestIndex.Get(n)).
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

func (r *daoPaymentRequestRepository) findOne(results *elastic.SearchResult, err error) (*explorer.PaymentRequest, error) {
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

func (r *daoPaymentRequestRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.PaymentRequest, int64, error) {
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
