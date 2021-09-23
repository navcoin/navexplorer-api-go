package repository

import (
	"context"
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"math"
)

type DaoVoteRepository interface {
	GetVotes(n network.Network, voteType explorer.VoteType, hash string, votingCycles []*entity.VotingCycle) ([]*entity.CfundVote, error)
	GetExcludedVotes(n network.Network, cycle uint) (uint, error)
}

type daoVoteRepository struct {
	elastic *elastic_cache.Index
}

func NewDaoVoteRepository(elastic *elastic_cache.Index) DaoVoteRepository {
	return &daoVoteRepository{elastic: elastic}
}

func (r *daoVoteRepository) GetVotes(n network.Network, voteType explorer.VoteType, hash string, votingCycles []*entity.VotingCycle) ([]*entity.CfundVote, error) {
	var cfundVotes = make([]*entity.CfundVote, 0)

	for _, vc := range votingCycles {
		cfundVote := entity.NewCfundVote(vc.Index, vc.Start, vc.End)

		size := vc.End - vc.Start + 1
		partitions := int(math.Ceil(float64(size) / 10000))

		for p := 0; p < partitions; p++ {
			voteQuery := elastic.NewBoolQuery()
			voteQuery = voteQuery.Must(elastic.NewTermQuery("votes.type.keyword", voteType))
			voteQuery = voteQuery.Must(elastic.NewTermQuery("votes.hash.keyword", hash))

			voteAgg := elastic.NewFilterAggregation().Filter(voteQuery)
			voteAgg.SubAggregation("yes", elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("votes.vote", 1)))
			voteAgg.SubAggregation("abstain", elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("votes.vote", -1)))
			voteAgg.SubAggregation("no", elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("votes.vote", 0)))

			voteExclusionAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("votes.type.keyword", explorer.ExcludeVote))

			voteTypeAgg := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("votes.type.keyword", voteType))
			voteTypeAgg.SubAggregation("vote", voteAgg)

			votesAgg := elastic.NewNestedAggregation().Path("votes")
			votesAgg.SubAggregation("type", voteTypeAgg)
			votesAgg.SubAggregation("exclusion", voteExclusionAgg)

			addressAgg := elastic.NewTermsAggregation().Field("address.keyword")
			addressAgg.SubAggregation("votes", votesAgg)
			addressAgg.Partition(p).NumPartitions(partitions).Size(10000)

			results, err := r.elastic.Client.Search(elastic_cache.DaoVoteIndex.Get(n)).
				Query(elastic.NewRangeQuery("height").Gte(vc.Start).Lte(vc.End)).
				Aggregation("address", addressAgg).
				Size(0).
				Do(context.Background())
			if err != nil {
				return nil, err
			}

			if address, found := results.Aggregations.Terms("address"); found {
				for _, addressBucket := range address.Buckets {
					addressVote := &entity.CfundVoteAddress{Address: addressBucket.Key.(string)}
					if votesBucket, found := addressBucket.Nested("votes"); found {
						if voteExclusionBucket, found := votesBucket.Filter("exclusion"); found {
							addressVote.Exclude = int(voteExclusionBucket.DocCount)
							cfundVote.Exclude += addressVote.Exclude
						}
						if voteTypeBucket, found := votesBucket.Filter("type"); found {
							if vote, found := voteTypeBucket.Filter("vote"); found {
								if no, found := vote.Filter("no"); found {
									addressVote.No += int(no.DocCount)
								}
								if yes, found := vote.Filter("yes"); found {
									addressVote.Yes += int(yes.DocCount)
								}
								if abstain, found := vote.Filter("abstain"); found {
									addressVote.Abstain += int(abstain.DocCount)
								}
								cfundVote.Yes += addressVote.Yes
								cfundVote.No += addressVote.No
								cfundVote.Abstain += addressVote.Abstain
							}
						}
						if addressVote.Yes != 0 || addressVote.No != 0 || addressVote.Abstain != 0 || addressVote.Exclude != 0 {
							cfundVote.Addresses = append(cfundVote.Addresses, addressVote)
						}
					}
				}
			}
		}

		cfundVotes = append(cfundVotes, cfundVote)
	}

	return cfundVotes, nil
}

func (r *daoVoteRepository) GetExcludedVotes(n network.Network, cycle uint) (uint, error) {
	excludedQuery := elastic.NewTermQuery("votes.type.keyword", explorer.ExcludeVote)
	votesAgg := elastic.NewNestedAggregation().Path("votes").
		SubAggregation("exclusion", elastic.NewFilterAggregation().Filter(excludedQuery))

	results, err := r.elastic.Client.Search(elastic_cache.DaoVoteIndex.Get(n)).
		Query(elastic.NewTermsQuery("cycle", cycle)).
		Aggregation("votes", votesAgg).
		Size(0).
		Do(context.Background())

	if err != nil {
		return 0, err
	}

	if votesBucket, found := results.Aggregations.Nested("votes"); found {
		if voteExclusionBucket, found := votesBucket.Filter("exclusion"); found {
			return uint(voteExclusionBucket.DocCount), nil
		}
	}

	return 0, errors.New("failed to get exclusion count for block cycle")
}
