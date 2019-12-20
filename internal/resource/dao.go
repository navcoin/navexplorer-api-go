package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DaoResource struct {
	daoProposalRepository  *repository.DaoProposalRepository
	daoConsensusRepository *repository.DaoConsensusRepository
}

func NewDaoResource(
	daoProposalRepository *repository.DaoProposalRepository,
	daoConsensusRepository *repository.DaoConsensusRepository,
) *DaoResource {
	return &DaoResource{daoProposalRepository, daoConsensusRepository}
}

func (r *DaoResource) GetConsensus(c *gin.Context) {
	consensus, err := r.daoConsensusRepository.GetConsensus()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, consensus)
}

func (r *DaoResource) GetProposals(c *gin.Context) {
	dir, size, page := pagination.GetPaginationParams(c)

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
	proposals, total, err := r.daoProposalRepository.Proposals(&status, dir, size, page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(proposals), total, size, page)
	paginator.WriteHeader(c)

	c.JSON(200, proposals)
}

func (r *DaoResource) GetProposal(c *gin.Context) {
	proposal, err := r.daoProposalRepository.Proposal(c.Param("hash"))

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
