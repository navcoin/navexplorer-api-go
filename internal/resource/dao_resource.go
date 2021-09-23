package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework/paginator"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/dao"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type DaoResource struct {
	daoService   dao.Service
	blockService block.Service
}

func NewDaoResource(daoService dao.Service, blockService block.Service) *DaoResource {
	return &DaoResource{daoService, blockService}
}

func (r *DaoResource) GetBlockCycle(c *gin.Context) {
	b, err := r.blockService.GetBestBlock(network(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	blockCycle, err := r.daoService.GetBlockCycleByBlock(network(c), b)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, blockCycle)
}

func (r *DaoResource) GetConsensusParameters(c *gin.Context) {
	consensus, err := r.daoService.GetConsensus(network(c))
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

	consensus, err := r.daoService.GetConsensus(network(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	for _, parameter := range consensus.All() {
		if parameter.Id == id {
			c.JSON(200, parameter)
			return
		}
	}

	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
		"message": "Consensus parameter not found", "status": http.StatusInternalServerError,
	})
}

func (r *DaoResource) GetCfundStats(c *gin.Context) {
	cfundStats, err := r.daoService.GetCfundStats(network(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, cfundStats)
}

func (r *DaoResource) GetProposals(c *gin.Context) {
	var parameters dao.ProposalParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	proposals, total, err := r.daoService.GetProposals(network(c), parameters, pagination(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginate := paginator.NewPaginator(len(proposals), total, pagination(c))
	paginate.WriteHeader(c)

	c.JSON(200, proposals)
}

func (r *DaoResource) GetProposal(c *gin.Context) {
	proposal, err := r.daoService.GetProposal(network(c), c.Param("hash"))

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
	votes, _, err := r.daoService.GetProposalVotes(network(c), c.Param("hash"))
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
	trend, err := r.daoService.GetProposalTrend(network(c), c.Param("hash"))
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
	var parameters dao.PaymentRequestParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	paymentRequests, total, err := r.daoService.GetPaymentRequests(network(c), parameters, pagination(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginate := paginator.NewPaginator(len(paymentRequests), total, pagination(c))
	paginate.WriteHeader(c)

	c.JSON(200, paymentRequests)
}

func (r *DaoResource) GetPaymentRequestsForProposal(c *gin.Context) {
	proposal, err := r.daoService.GetProposal(network(c), c.Param("hash"))

	if err == repository.ErrProposalNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}

	paymentRequests, err := r.daoService.GetPaymentRequestsForProposal(network(c), proposal)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, paymentRequests)
}

func (r *DaoResource) GetPaymentRequest(c *gin.Context) {
	paymentRequest, err := r.daoService.GetPaymentRequest(network(c), c.Param("hash"))

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
	votes, _, err := r.daoService.GetPaymentRequestVotes(network(c), c.Param("hash"))
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
	trend, err := r.daoService.GetPaymentRequestTrend(network(c), c.Param("hash"))
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
	var parameters dao.ConsultationParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	consultations, total, err := r.daoService.GetConsultations(network(c), parameters, pagination(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginate := paginator.NewPaginator(len(consultations), total, pagination(c))
	paginate.WriteHeader(c)

	c.JSON(200, consultations)
}

func (r *DaoResource) GetConsultation(c *gin.Context) {
	proposal, err := r.daoService.GetConsultation(network(c), c.Param("hash"))

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

func (r *DaoResource) GetAnswer(c *gin.Context) {
	proposal, err := r.daoService.GetAnswer(network(c), c.Param("hash"))

	if err == repository.ErrAnswerNotFound {
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
	votes, _, err := r.daoService.GetAnswerVotes(network(c), c.Param("hash"), c.Param("answer"))
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

func (r *DaoResource) GetExcludedVotesForCycle(c *gin.Context) {
	cycle, err := strconv.Atoi(c.DefaultQuery("cycle", "0"))
	if err != nil || cycle == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid block cycle", "status": http.StatusBadRequest,
		})
		return
	}

	votes, err := r.daoService.GetExcludedVotes(network(c), uint(cycle))

	c.JSON(200, votes)
}
