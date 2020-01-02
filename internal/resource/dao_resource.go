package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DaoResource struct {
	daoService *dao.DaoService
}

func NewDaoResource(daoService *dao.DaoService) *DaoResource {
	return &DaoResource{daoService}
}

func (r *DaoResource) GetBlockCycle(c *gin.Context) {
	blockCycle, err := r.daoService.GetBlockCycle(&explorer.Block{})
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

	c.JSON(200, consensus)
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

	statusString := c.DefaultQuery("status", "")
	if statusString != "" {
		if valid := explorer.ProposalStatusIsValid(statusString); valid == false {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("Invalid Status(%s)", statusString),
				"status":  http.StatusBadRequest,
			})
			return
		}
	}

	status := explorer.ProposalStatus(statusString)
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

	if err != nil {
		if err == repository.ErrProposalNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, proposal)
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
