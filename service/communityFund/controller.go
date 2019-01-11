package communityFund

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Controller struct{}

func (controller *Controller) GetBlockCycle(c *gin.Context) {
	c.JSON(200, GetBlockCycle())
}

func (controller *Controller) GetProposals(c *gin.Context) {
	dir := c.DefaultQuery("dir", "DESC")

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		size = 10
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", ""))
	if err != nil {
		offset = 0
	}

	proposals, total, _ := GetProposalsByState(c.Query("state"), size, dir == "ASC", offset)
	if proposals == nil {
		proposals = make([]Proposal, 0)
	}

	paginator := pagination.NewPaginator(len(proposals), total, size, dir == "ASC", offset)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, proposals)
}

func (controller *Controller) GetProposal(c *gin.Context) {
	proposal, err := GetProposalByHash(c.Param("hash"))

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find proposal: %s", c.Param("hash")),
		})
		c.Abort()
	} else {
		c.JSON(200, proposal)
	}
}

func (controller *Controller) GetProposalVotes(c *gin.Context) {
	vote, err := strconv.ParseBool(c.Param("vote"))

	votes, err := GetProposalVotes(c.Param("hash"), vote)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find proposal: %s", c.Param("hash")),
		})
		c.Abort()
	} else {
		c.JSON(200, votes)
	}
}

func (controller *Controller) GetProposalPaymentRequests(c *gin.Context) {
	paymentRequests, _ := GetProposalPaymentRequests(c.Query("hash"))

	c.JSON(200, paymentRequests)
}

func (controller *Controller) GetPaymentRequestsByState(c *gin.Context) {
	paymentRequests, _ := GetPaymentRequestsByState(c.Query("state"))

	c.JSON(200, paymentRequests)
}

func (controller *Controller) GetPaymentRequestByHash(c *gin.Context) {
	paymentRequests, _ := GetPaymentRequestByHash(c.Query("hash"))

	c.JSON(200, paymentRequests)
}

func (controller *Controller) GetPaymentRequestVotes(c *gin.Context) {
	vote, err := strconv.ParseBool(c.Param("vote"))

	votes, err := GetPaymentRequestVotes(c.Param("hash"), vote)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find proposal: %s", c.Param("hash")),
		})
		c.Abort()
	} else {
		c.JSON(200, votes)
	}
}