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

	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("votes.type.keyword", voteType))
	query = query.Must(elastic.NewTermQuery("votes.hash.keyword", hash))

	for _, vc := range votingCycles {
		agg := elastic.NewRangeAggregation().Field("height").AddRange(vc.Start, vc.End)
		addressAgg := elastic.NewTermsAggregation().Field("address.keyword")
		votesAgg := elastic.NewNestedAggregation().Path("votes")
		typeAgg := elastic.NewFilterAggregation().Filter(query)
		voteAgg := elastic.NewTermsAggregation().Field("votes.vote")

		agg.SubAggregation("address", addressAgg)
		addressAgg.SubAggregation("votes", votesAgg)
		votesAgg.SubAggregation("type", typeAgg)
		typeAgg.SubAggregation("vote", voteAgg)

		service.Aggregation(fmt.Sprintf("cycle-%d", vc.Index), agg)
	}

	results, err := service.Do(context.Background())

	if err != nil {
		return nil, err
	}

	var cfundVotes = make([]*entity.CfundVote, 0)
	i := 0
	for {
		if cycles, found := results.Aggregations.Range(fmt.Sprintf("cycle-%d", i)); found {
			cfundVote := entity.NewCfundVote(i, votingCycles[i].Start, votingCycles[i].End)

			if address, found := cycles.Buckets[0].Terms("address"); found {
				for _, addressBucket := range address.Buckets {
					addressVote := &entity.CfundVoteAddress{Address: addressBucket.Key.(string)}
					if votesBucket, found := addressBucket.Nested("votes"); found {
						if typeBucket, found := votesBucket.Filter("type"); found {
							if vote, found := typeBucket.Terms("vote"); found {
								for _, voteBucket := range vote.Buckets {
									if voteBucket.Key.(float64) == 1 {
										addressVote.Yes = int(voteBucket.DocCount)
									}
									if voteBucket.Key.(float64) == 0 {
										addressVote.No += int(voteBucket.DocCount)
									}
									if voteBucket.Key.(float64) == -1 {
										addressVote.Abstain += int(voteBucket.DocCount)
									}
								}
								cfundVote.Addresses = append(cfundVote.Addresses, addressVote)
								cfundVote.Yes += addressVote.Yes
								cfundVote.No += addressVote.No
								cfundVote.Abstain += addressVote.Abstain
							}
						}
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
