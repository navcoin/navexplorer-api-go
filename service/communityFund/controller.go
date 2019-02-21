package communityFund

import (
	"github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Controller struct{}

func (controller *Controller) GetBlockCycle(c *gin.Context) {
	c.JSON(200, GetBlockCycle())
}

func (controller *Controller) GetStats(c *gin.Context) {
	c.JSON(200, GetStats())
}

func (controller *Controller) GetProposals(c *gin.Context) {
	dir := c.DefaultQuery("dir", "DESC")

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		size = 10
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	proposals, total, err := GetProposalsByState(c.Query("state"), size, dir == "ASC", page)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}
	if proposals == nil {
		proposals = make([]Proposal, 0)
	}

	paginator := pagination.NewPaginator(len(proposals), total, size, page)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, proposals)
}

func (controller *Controller) GetProposal(c *gin.Context) {
	proposal, err := GetProposalByHash(c.Param("hash"))

	if err != nil {
		if err == ErrProposalNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}
		return
	}

	c.JSON(200, proposal)
}

func (controller *Controller) GetProposalVotes(c *gin.Context) {
	vote, err := strconv.ParseBool(c.Param("vote"))
	votes, err := GetProposalVotes(c.Param("hash"), vote)

	if err != nil {
		if err == ErrProposalNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}
		return
	}

	if votes == nil {
		votes = make([]Votes, 0)
	}

	c.JSON(200, votes)
}

func (controller *Controller) GetProposalVotingTrend(c *gin.Context) {
	proposal, err := GetProposalByHash(c.Param("hash"))

	if err != nil {
		if err == ErrProposalNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}
		return
	}

	trend, err := GetProposalTrend(proposal.Hash)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, trend)
}

func (controller *Controller) GetProposalPaymentRequests(c *gin.Context) {
	paymentRequests, err := GetProposalPaymentRequests(c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	if paymentRequests == nil {
		paymentRequests = make([]PaymentRequest, 0)
	}

	c.JSON(200, paymentRequests)
}

func (controller *Controller) GetPaymentRequestsByState(c *gin.Context) {
	paymentRequests, err := GetPaymentRequestsByState(c.Query("state"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	if paymentRequests == nil {
		paymentRequests = make([]PaymentRequest, 0)
	}

	c.JSON(200, paymentRequests)
}

func (controller *Controller) GetPaymentRequestByHash(c *gin.Context) {
	paymentRequests, err := GetPaymentRequestByHash(c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, paymentRequests)
}

func (controller *Controller) GetPaymentRequestVotes(c *gin.Context) {
	vote, err := strconv.ParseBool(c.Param("vote"))
	votes, err := GetPaymentRequestVotes(c.Param("hash"), vote)

	if err != nil {
		if err == ErrPaymentRequestNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}

		return
	}

	c.JSON(200, votes)
}

func (controller *Controller) GetPaymentRequestVotingTrend(c *gin.Context) {
	paymentRequest, err := GetPaymentRequestByHash(c.Param("hash"))

	if err != nil {
		if err == ErrPaymentRequestNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}

		return
	}

	trend, err := GetPaymentRequestTrend(paymentRequest.Hash)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, trend)
}
