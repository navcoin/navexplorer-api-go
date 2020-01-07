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

func (r *DaoVoteRepository) GetVotes(
	voteType explorer.VoteType,
	hash string,
	votingCycles []*entity.VotingCycle,
	bestBlockHeight uint64,
) ([]*entity.CfundVote, error) {
	service := r.elastic.Client.Search(elastic_cache.DaoVoteIndex.Get()).Size(0)

	for _, vc := range votingCycles {
		if vc.End > int(bestBlockHeight) {
			vc.End = int(bestBlockHeight)
		}
		voteAgg := elastic.NewTermsAggregation().Field("votes.vote")
		addressAgg := elastic.NewTermsAggregation().Field("address.keyword")
		addressAgg.SubAggregation("vote", voteAgg)

		agg := elastic.NewRangeAggregation().Field("height").AddRange(vc.Start, vc.End+1)
		agg.SubAggregation("vote", voteAgg)

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

			if vote, found := cycles.Buckets[0].Terms("vote"); found {
				for _, voteBucket := range vote.Buckets {
					if voteBucket.Key.(float64) == 1 {
						cfundVote.Yes += int(voteBucket.DocCount)
					}
					if voteBucket.Key.(float64) == -1 {
						cfundVote.No += int(voteBucket.DocCount)
					}
				}
				cfundVote.Abstain = votingCycles[i].End + 1 - votingCycles[i].Start - cfundVote.Yes - cfundVote.No
			}
			cfundVotes = append(cfundVotes, cfundVote)
			i++
		} else {
			break
		}
	}

	return cfundVotes, nil
}
