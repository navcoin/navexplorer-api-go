package communityFund

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

var service = new(Service)

type Controller struct{}

func (controller *Controller) GetBlockCycle(c *gin.Context) {
	c.JSON(200, service.GetBlockCycle())
}

func (controller *Controller) GetProposals(c *gin.Context) {
	proposals, _ := service.GetProposalsByState(c.Query("state"))

	c.JSON(200, proposals)
}

func (controller *Controller) GetProposal(c *gin.Context) {
	proposal, err := service.GetProposalByHash(c.Param("hash"))

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

func (controller *Controller) GetPaymentRequests(c *gin.Context) {
	paymentRequests, _ := service.GetPaymentRequests(c.Param("hash"))

	c.JSON(200, paymentRequests)
}