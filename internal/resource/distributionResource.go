package resource

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/address"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/block"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type DistributionResource struct {
	addressService address.Service
	blockService   block.Service
}

type DistributionSupplyResponse struct {
	Total   float64 `json:"total"`
	Public  float64 `json:"public"`
	Private float64 `json:"private"`
	Wrapped float64 `json:"wrapped"`
}

func NewDistributionResource(addressService address.Service, blockService block.Service) *DistributionResource {
	return &DistributionResource{addressService, blockService}
}

func (r *DistributionResource) GetSupply(c *gin.Context) {
	bestBlock, err := r.blockService.GetBestBlock(network(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	c.JSON(200, &DistributionSupplyResponse{
		Total:   float64(bestBlock.SupplyBalance.Public+bestBlock.SupplyBalance.Private+bestBlock.SupplyBalance.Wrapped) / 100000000,
		Public:  float64(bestBlock.SupplyBalance.Public) / 100000000,
		Private: float64(bestBlock.SupplyBalance.Private) / 100000000,
		Wrapped: float64(bestBlock.SupplyBalance.Wrapped) / 100000000,
	})
}

func (r *DistributionResource) GetWealth(c *gin.Context) {
	groupsQuery := c.DefaultQuery("groups", "10,100,1000")
	if groupsQuery == "" {
		groupsQuery = "10,100,1000"
	}

	groups := make([]string, 0)
	groups = strings.Split(groupsQuery, ",")

	b := make([]int, len(groups))
	for i, v := range groups {
		b[i], _ = strconv.Atoi(v)
	}

	distribution, err := r.addressService.GetPublicWealthDistribution(network(c), b)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, distribution)
}
