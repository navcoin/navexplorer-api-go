package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
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

	state, err := r.daoProposalRepository.StateFromString(c.Query("state"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err, "status": http.StatusBadRequest})
		return
	}

	proposals, total, err := r.daoProposalRepository.Proposals(*state, dir, size, page)
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
