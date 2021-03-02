package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DistributionResource struct {
	supplyService service.SupplyService
}

type DistributionSupplyResponse struct {
	Public  float64 `json:"public"`
	Private float64 `json:"private"`
}

func NewDistributionResource(supplyService service.SupplyService) *DistributionResource {
	return &DistributionResource{supplyService}
}

func (r *DistributionResource) GetTotalSupply(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	publicSupply, err := r.supplyService.GetPublicSupply(n)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	privateSupply, err := r.supplyService.GetPrivateSupply(n)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, &DistributionSupplyResponse{
		Public:  publicSupply,
		Private: privateSupply,
	})
}
