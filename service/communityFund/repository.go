package communityFund

import (
	"github.com/NavExplorer/navexplorer-api-go/db"
	"github.com/globalsign/mgo/bson"
)

type Repository struct{}

func (r *Repository) FindAllProposals() (proposals []Proposal, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use("communityFundProposal")
	defer dbConnection.Close()

	err = c.Find(bson.M{}).Sort("-id").All(&proposals)

	return proposals, err
}

func (r *Repository) FindProposalsByState(state string) (proposals []Proposal, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use("communityFundProposal")
	defer dbConnection.Close()

	err = c.Find(bson.M{"state": state}).Sort("-height").All(&proposals)

	return proposals, err
}

func (r *Repository) FindProposalByHash(hash string) (proposal Proposal, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use("communityFundProposal")
	defer dbConnection.Close()

	err = c.Find(bson.M{"hash": hash}).One(&proposal)

	return proposal, err
}

func (r *Repository) FindPaymentRequestsByProposalHash(proposalHash string) (paymentRequests []PaymentRequest, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use("communityFundPaymentRequest")
	defer dbConnection.Close()

	err = c.Find(bson.M{"proposalHash": proposalHash}).Sort("-height").All(&paymentRequests)

	return paymentRequests, err
}