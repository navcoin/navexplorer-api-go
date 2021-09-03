package resource

import (
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/address"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type StakingResource struct {
	addressService address.Service
	stakingService service.StakingService
}

func NewStakingResource(addressService address.Service, stakingService service.StakingService) *StakingResource {
	return &StakingResource{addressService, stakingService}
}

func (r *StakingResource) GetBlocks(c *gin.Context) {
	blockCount, err := strconv.Atoi(c.DefaultQuery("blocks", "1000"))
	if err != nil {
		blockCount = 1000
	}
	zap.S().Infof("Staking: GetBlocks(%d)", blockCount)

	if blockCount > 100000 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "100,000 blocks is the maximum", "status": http.StatusBadRequest})
		return
	}

	staking, err := r.addressService.GetStakingByBlockCount(network(c), blockCount)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, staking)
}

func (r *StakingResource) GetStakingRewardsForAddresses(c *gin.Context) {
	addresses := strings.Split(c.Query("addresses"), ",")
	if len(addresses) == 0 {
		handleError(c, errors.New("No addresses provided"), http.StatusBadRequest)
		return
	}

	rewards, err := r.stakingService.GetStakingRewardsForAddresses(network(c), addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, rewards)
}
