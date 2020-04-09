package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DaoResource struct {
	daoService   *dao.Service
	blockService *block.Service
}

func NewDaoResource(daoService *dao.Service, blockService *block.Service) *DaoResource {
	return &DaoResource{daoService, blockService}
}

func (r *DaoResource) GetBlockCycle(c *gin.Context) {
	b, err := r.blockService.GetBestBlock()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	blockCycle, err := r.daoService.GetBlockCycleByBlock(b)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, blockCycle)
}

func (r *DaoResource) GetConsensus(c *gin.Context) {
	consensus, err := r.daoService.GetConsensus()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, consensus.All())
}

func (r *DaoResource) GetCfundStats(c *gin.Context) {
	cfundStats, err := r.daoService.GetCfundStats()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, cfundStats)
}

func (r *DaoResource) GetProposals(c *gin.Context) {
	config := pagination.GetConfig(c)

	var status explorer.ProposalStatus
	statusString := c.DefaultQuery("status", "")
	if statusString != "" {
		if valid := explorer.ProposalStatusIsValid(statusString); valid == false {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Invalid Status(%s)", statusString),
				"status":  http.StatusBadRequest,
			})
			return
		}
		status = explorer.GetStatusByStatus(statusString)
	}

	proposals, total, err := r.daoService.GetProposals(&status, config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(proposals), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, proposals)
}

func (r *DaoResource) GetProposal(c *gin.Context) {
	proposal, err := r.daoService.GetProposal(c.Param("hash"))

	if err == repository.ErrProposalNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, proposal)
}

func (r *DaoResource) GetProposalVotes(c *gin.Context) {
	votes, err := r.daoService.GetProposalVotes(c.Param("hash"))
	if err == repository.ErrProposalNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, votes)
}

func (r *DaoResource) GetProposalTrend(c *gin.Context) {
	trend, err := r.daoService.GetProposalTrend(c.Param("hash"))
	if err == repository.ErrProposalNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, trend)
}

func (r *DaoResource) GetPaymentRequests(c *gin.Context) {
	config := pagination.GetConfig(c)

	statusString := c.DefaultQuery("status", "")
	if statusString != "" {
		if valid := explorer.PaymentRequestStatusIsValid(statusString); valid == false {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Invalid Status(%s)", statusString),
				"status":  http.StatusBadRequest,
			})
			return
		}
	}

	status := explorer.PaymentRequestStatus(statusString)
	paymentRequests, total, err := r.daoService.GetPaymentRequests(&status, config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(paymentRequests), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, paymentRequests)
}

func (r *DaoResource) GetPaymentRequestsForProposal(c *gin.Context) {
	proposal, err := r.daoService.GetProposal(c.Param("hash"))

	if err == repository.ErrProposalNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}

	paymentRequests, err := r.daoService.GetPaymentRequestsForProposal(proposal)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, paymentRequests)
}

func (r *DaoResource) GetPaymentRequest(c *gin.Context) {
	paymentRequest, err := r.daoService.GetPaymentRequest(c.Param("hash"))

	if err != nil {
		if err == repository.ErrPaymentRequestNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, paymentRequest)
}

func (r *DaoResource) GetPaymentRequestVotes(c *gin.Context) {
	votes, err := r.daoService.GetPaymentRequestVotes(c.Param("hash"))
	if err == repository.ErrPaymentRequestNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, votes)
}

func (r *DaoResource) GetPaymentRequestTrend(c *gin.Context) {
	trend, err := r.daoService.GetPaymentRequestTrend(c.Param("hash"))
	if err == repository.ErrProposalNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, trend)
}
