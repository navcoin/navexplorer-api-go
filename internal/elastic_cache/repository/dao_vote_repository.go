package repository

import (
	"context"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/dto"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/voting_cycle"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
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
	votingCycles []*voting_cycle.Cycle,
	bestBlockHeight uint64,
) ([]*dto.CfundVote, error) {
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

	var cfundVotes = make([]*dto.CfundVote, 0)
	cycle := 0
	for {
		if cycles, found := results.Aggregations.Range(fmt.Sprintf("%d", cycle)); found {
			cfundVote := &dto.CfundVote{Cycle: cycle, Start: votingCycles[cycle].Start, End: votingCycles[cycle].End}

			if vote, found := cycles.Buckets[0].Terms("vote"); found {
				for _, voteBucket := range vote.Buckets {
					if voteBucket.Key.(float64) == 1 {
						cfundVote.Vote.Yes += int(voteBucket.DocCount)
					}
					if voteBucket.Key.(float64) == -1 {
						cfundVote.Vote.No += int(voteBucket.DocCount)
					}
				}
				cfundVote.Vote.Abstain = votingCycles[cycle].End + 1 - votingCycles[cycle].Start - cfundVote.Vote.Yes - cfundVote.Vote.No
			}
			cfundVotes = append(cfundVotes, cfundVote)
			cycle++
		} else {
			break
		}
	}

	return cfundVotes, nil
}

func (r *DaoVoteRepository) GetTrend(
	voteType explorer.VoteType,
	hash string,
	votingCycles []*voting_cycle.Cycle,
	bestBlockHeight uint64,
) ([]*dto.CfundVote, error) {
	cfundVotes, err := r.GetVotes(voteType, hash, votingCycles, bestBlockHeight)
	if err != nil {
		return nil, err
	}

	for _, cfundVote := range cfundVotes {
		cfundVote.Vote.Yes = int(float64(cfundVote.Vote.Yes)/10) * 100
		cfundVote.Vote.No = int(float64(cfundVote.Vote.No)/10) * 100
		cfundVote.Vote.Abstain = int(float64(cfundVote.Vote.Abstain)/10) * 100
	}

	return cfundVotes, nil
}
