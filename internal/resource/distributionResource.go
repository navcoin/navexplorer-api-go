package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/service/distribution"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DistributionResource struct {
	distributionService distribution.Service
}

func NewDistributionResource(distributionService distribution.Service) *DistributionResource {
	return &DistributionResource{distributionService}
}

func (r *DistributionResource) GetTotalSupply(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	totalSupply, err := r.distributionService.GetTotalSupply(n)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, totalSupply)
}
