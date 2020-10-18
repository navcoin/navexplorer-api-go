package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
)

type DaoConsultationRepository interface {
	GetConsultations(n network.Network, status *explorer.ConsultationStatus, consensus *bool, min *uint, asc bool, size, page int) ([]*explorer.Consultation, int64, error)
	GetConsultation(n network.Network, hash string) (*explorer.Consultation, error)
	GetAnswer(n network.Network, hash string) (*explorer.Answer, error)
	GetConsensusConsultations(n network.Network, dir bool, size, page int) ([]*explorer.Consultation, int64, error)
}

type daoConsultationRepository struct {
	elastic *elastic_cache.Index
}

var (
	ErrConsultationNotFound = errors.New("Consultation not found")
	ErrAnswerNotFound       = errors.New("Answer not found")
)

func NewDaoConsultationRepository(elastic *elastic_cache.Index) DaoConsultationRepository {
	return &daoConsultationRepository{elastic: elastic}
}

func (r *daoConsultationRepository) GetConsultations(n network.Network, status *explorer.ConsultationStatus, consensus *bool, min *uint, asc bool, size, page int) ([]*explorer.Consultation, int64, error) {
	query := elastic.NewBoolQuery()
	if status != nil {
		query = query.Must(elastic.NewTermQuery("state", status.State))
	}
	if min != nil {
		query = query.Must(elastic.NewTermQuery("min", min))
	}
	if consensus != nil {
		query = query.Must(elastic.NewTermQuery("consensusParameter", consensus))
	}

	result, err := r.elastic.Client.Search(elastic_cache.DaoConsultationIndex.Get(n)).
		Query(query).
		Sort("height", asc).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return r.findMany(result, err)
}

func (r *daoConsultationRepository) GetConsultation(n network.Network, hash string) (*explorer.Consultation, error) {
	results, err := r.elastic.Client.Search(elastic_cache.DaoConsultationIndex.Get(n)).
		Query(elastic.NewTermQuery("hash.keyword", hash)).
		Size(1).
		Do(context.Background())

	return r.findOne(results, err)
}

func (r *daoConsultationRepository) GetAnswer(n network.Network, hash string) (*explorer.Answer, error) {
	query := elastic.NewTermQuery("answers.hash.keyword", hash)
	nestedQuery := elastic.NewNestedQuery("answers", query)

	results, err := r.elastic.Client.Search(elastic_cache.DaoConsultationIndex.Get(n)).
		Query(nestedQuery).
		Size(1).
		Do(context.Background())

	c, err := r.findOne(results, err)
	if err != nil {
		return nil, err
	}

	for _, a := range c.Answers {
		if a.Hash == hash {
			return a, nil
		}
	}

	return nil, ErrAnswerNotFound
}

func (r *daoConsultationRepository) GetConsensusConsultations(n network.Network, dir bool, size, page int) ([]*explorer.Consultation, int64, error) {
	result, err := r.elastic.Client.Search(elastic_cache.DaoConsultationIndex.Get(n)).
		Query(elastic.NewTermQuery("consensusParameter", true)).
		Sort("height", dir).
		From((page * size) - size).
		Size(size).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	return r.findMany(result, err)
}

func (r *daoConsultationRepository) findOne(results *elastic.SearchResult, err error) (*explorer.Consultation, error) {
	if err != nil || results.TotalHits() == 0 {
		err = ErrConsultationNotFound
		return nil, err
	}

	var consultation *explorer.Consultation
	hit := results.Hits.Hits[0]
	err = json.Unmarshal(hit.Source, &consultation)
	if err != nil {
		return nil, err
	}

	return consultation, err
}

func (r *daoConsultationRepository) findMany(results *elastic.SearchResult, err error) ([]*explorer.Consultation, int64, error) {
	if err != nil {
		return nil, 0, err
	}

	consultations := make([]*explorer.Consultation, 0)
	for _, hit := range results.Hits.Hits {
		var consultation *explorer.Consultation
		if err := json.Unmarshal(hit.Source, &consultation); err == nil {
			consultations = append(consultations, consultation)
		}
	}

	return consultations, results.TotalHits(), err
}
