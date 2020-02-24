package repository

import (
	"context"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
)

type DaoVoteRepository struct {
	elastic *elastic_cache.Index
}

func NewDaoVoteRepository(elastic *elastic_cache.Index) *DaoVoteRepository {
	return &DaoVoteRepository{elastic}
}

func (r *DaoVoteRepository) GetVotes(voteType explorer.VoteType, hash string, votingCycles []*entity.VotingCycle) ([]*entity.CfundVote, error) {
	service := r.elastic.Client.Search(elastic_cache.DaoVoteIndex.Get()).Size(0)

	for _, vc := range votingCycles {
		agg := elastic.NewRangeAggregation().Field("height").AddRange(vc.Start-1, vc.End)
		voteAgg := elastic.NewTermsAggregation().Field("votes.vote")
		addressAgg := elastic.NewTermsAggregation().Field("address.keyword")

		agg.SubAggregation("address", addressAgg)
		addressAgg.SubAggregation("vote", voteAgg)

		service.Aggregation(fmt.Sprintf("%d", vc.Index), agg)
	}

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("votes.type.keyword", voteType))
	query = query.Must(elastic.NewTermQuery("votes.hash.keyword", hash))

	results, err := service.Query(query).Do(context.Background())

	if err != nil {
		return nil, err
	}

	var cfundVotes = make([]*entity.CfundVote, 0)
	i := 0
	for {
		if cycles, found := results.Aggregations.Range(fmt.Sprintf("%d", i)); found {
			cfundVote := entity.NewCfundVote(i, votingCycles[i].Start, votingCycles[i].End)

			if address, found := cycles.Buckets[0].Terms("address"); found {
				for _, addressBucket := range address.Buckets {
					addressVote := &entity.CfundVoteAddress{Address: addressBucket.Key.(string)}
					if vote, found := cycles.Buckets[0].Terms("vote"); found {
						for _, voteBucket := range vote.Buckets {
							if voteBucket.Key.(float64) == 1 {
								addressVote.Yes = int(voteBucket.DocCount)

							}
							if voteBucket.Key.(float64) == -1 {
								addressVote.No += int(voteBucket.DocCount)
							}
						}
						cfundVote.Yes += addressVote.Yes
						cfundVote.No += addressVote.No
						cfundVote.Abstain = votingCycles[i].End + 1 - votingCycles[i].Start - cfundVote.Yes - cfundVote.No
					}
				}
			}

			cfundVotes = append(cfundVotes, cfundVote)
			i++
		} else {
			break
		}
	}

	return cfundVotes, nil
}

func (r *DaoVoteRepository) GetVotingAddresses(voteType explorer.VoteType, hash string, votingCycle *entity.VotingCycle) (*entity.CfundVoteAddresses, error) {
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("votes.type.keyword", voteType))
	query = query.Must(elastic.NewTermQuery("votes.hash.keyword", hash))
	query = query.Must(elastic.NewRangeQuery("height").Gte(votingCycle.Start).Lt(votingCycle.End))

	agg := elastic.NewTermsAggregation().Field("votes.vote")
	agg.SubAggregation("address", elastic.NewTermsAggregation().Field("address.keyword").OrderByCountDesc().Size(2147483647))

	results, err := r.elastic.Client.
		Search(elastic_cache.DaoVoteIndex.Get()).
		Query(query).
		Aggregation("vote", agg).
		Size(0).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	voteAddresses := &entity.CfundVoteAddresses{
		Cycle: votingCycle.Index,
		Yes:   make([]*entity.CfundVoteAddressesElement, 0),
		No:    make([]*entity.CfundVoteAddressesElement, 0),
	}

	if agg, found := results.Aggregations.Terms("vote"); found {
		for _, voteBucket := range agg.Buckets {
			if voteBucket.Key.(float64) == 1 {
				if addressAgg, found := voteBucket.Terms("address"); found {
					for _, bucket := range addressAgg.Buckets {
						voteAddresses.Yes = append(voteAddresses.Yes, &entity.CfundVoteAddressesElement{
							Address: bucket.Key.(string),
							Votes:   int(bucket.DocCount),
						})
					}
				}
			}
			if voteBucket.Key.(float64) == -1 {
				if addressAgg, found := voteBucket.Terms("address"); found {
					for _, bucket := range addressAgg.Buckets {
						voteAddresses.No = append(voteAddresses.No, &entity.CfundVoteAddressesElement{
							Address: bucket.Key.(string),
							Votes:   int(bucket.DocCount),
						})
					}
				}
			}
		}
	}

	return voteAddresses, nil
}
