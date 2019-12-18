package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DaoProposalResource struct {
	daoProposalRepository *repository.DaoProposalRepository
}

func NewDaoProposalResource(daoProposalRepository *repository.DaoProposalRepository) *DaoProposalResource {
	return &DaoProposalResource{daoProposalRepository}
}

func (r *DaoProposalResource) GetProposals(c *gin.Context) {
	dir, size, page := pagination.GetPaginationParams(c)

	if valid := explorer.ProposalStatusIsValid(c.Query("status")); valid == false {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Invalid Status(%s)", c.Query("status")),
			"status":  http.StatusBadRequest,
		})
		return
	}

	proposals, total, err := r.daoProposalRepository.Proposals(explorer.ProposalStatus(c.Query("status")), dir, size, page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(proposals), total, size, page)
	paginator.WriteHeader(c)

	c.JSON(200, proposals)
}

func (r *DaoProposalResource) GetProposal(c *gin.Context) {
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
