package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type DaoResource struct {
	daoService   service.DaoService
	blockService *block.Service
}

func NewDaoResource(daoService service.DaoService, blockService *block.Service) *DaoResource {
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

func (r *DaoResource) GetConsensusParameters(c *gin.Context) {
	consensus, err := r.daoService.GetConsensus()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, consensus.All())
}

func (r *DaoResource) GetConsensusParameter(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.WithError(err).Error("Invalid consensus parameter")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Consensus Parameter not provided", "status": http.StatusBadRequest,
		})
		return
	}

	consensus, err := r.daoService.GetConsensus()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	parameter := consensus.Get(id)
	if parameter == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "Consensus parameter not found", "status": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(200, parameter)
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
	config, _ := pagination.Bind(c)

	var parameters dao.ProposalParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	proposals, total, err := r.daoService.GetProposals(parameters, config)
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
	config, _ := pagination.Bind(c)

	var parameters dao.PaymentRequestParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	paymentRequests, total, err := r.daoService.GetPaymentRequests(parameters, config)
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
	if err == repository.ErrPaymentRequestNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, trend)
}

func (r *DaoResource) GetConsultations(c *gin.Context) {
	config, _ := pagination.Bind(c)

	var parameters dao.ConsultationParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	consultations, total, err := r.daoService.GetConsultations(parameters, config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(consultations), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, consultations)
}

func (r *DaoResource) GetConsultation(c *gin.Context) {
	proposal, err := r.daoService.GetConsultation(c.Param("hash"))

	if err == repository.ErrConsultationNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, proposal)
}

func (r *DaoResource) GetAnswerVotes(c *gin.Context) {
	votes, err := r.daoService.GetAnswerVotes(c.Param("hash"), c.Param("answer"))
	if err == repository.ErrAnswerNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, votes)
}
