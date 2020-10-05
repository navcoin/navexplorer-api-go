package repository

import (
	"context"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao/entity"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/olivere/elastic/v7"
	"math"
)

type DaoVoteRepository struct {
	elastic *elastic_cache.Index
	network string
}

func NewDaoVoteRepository(elastic *elastic_cache.Index) *DaoVoteRepository {
	return &DaoVoteRepository{elastic: elastic}
}

func (r *DaoVoteRepository) Network(network string) *DaoVoteRepository {
	r.network = network

	return r
}

func (r *DaoVoteRepository) GetVotes(voteType explorer.VoteType, hash string, votingCycles []*entity.VotingCycle) ([]*entity.CfundVote, error) {
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

			votesAgg := elastic.NewNestedAggregation().Path("votes")
			votesAgg.SubAggregation("vote", voteAgg)

			addressAgg := elastic.NewTermsAggregation().Field("address.keyword")
			addressAgg.SubAggregation("votes", votesAgg)
			addressAgg.Partition(p).NumPartitions(partitions).Size(10000)

			results, err := r.elastic.Client.Search(elastic_cache.DaoVoteIndex.Get(r.network)).
				Query(elastic.NewRangeQuery("height").Gte(vc.Start).Lte(vc.End)).
				Aggregation("address", addressAgg).
				Size(0).
				Do(context.Background())
			if err != nil {
				return nil, err
			}

			//if height, found := results.Aggregations.Range("height"); found {
			//	if address, found := height.Buckets[0].Terms("address"); found {
			if address, found := results.Aggregations.Terms("address"); found {
				for _, addressBucket := range address.Buckets {
					addressVote := &entity.CfundVoteAddress{Address: addressBucket.Key.(string)}
					if votesBucket, found := addressBucket.Nested("votes"); found {
						if vote, found := votesBucket.Filter("vote"); found {
							if no, found := vote.Filter("no"); found {
								addressVote.No += int(no.DocCount)
							}
							if yes, found := vote.Filter("yes"); found {
								addressVote.Yes += int(yes.DocCount)
							}
							if abstain, found := vote.Filter("abstain"); found {
								addressVote.Abstain += int(abstain.DocCount)
							}
							cfundVote.Addresses = append(cfundVote.Addresses, addressVote)
							cfundVote.Yes += addressVote.Yes
							cfundVote.No += addressVote.No
							cfundVote.Abstain += addressVote.Abstain
						}
					}
				}
			}
			//	}
			//}
		}

		cfundVotes = append(cfundVotes, cfundVote)
	}

	return cfundVotes, nil
}
