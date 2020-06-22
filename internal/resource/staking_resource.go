package resource

import (
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type StakingResource struct {
	addressService address.Service
}

func NewStakingResource(addressService address.Service) *StakingResource {
	return &StakingResource{addressService}
}

func (r *StakingResource) GetBlocks(c *gin.Context) {
	blockCount, err := strconv.Atoi(c.DefaultQuery("blocks", "1000"))
	if err != nil {
		blockCount = 1000
	}
	if blockCount > 100000 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "100,000 blocks is the maximum", "status": http.StatusBadRequest})
		return
	}

	extended, err := strconv.ParseBool(c.DefaultQuery("extended", "false"))
	if err != nil {
		extended = false
	}

	staking, err := r.addressService.GetStakingByBlockCount(blockCount, extended)
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

	rewards, err := r.addressService.GetStakingRewardsForAddresses(addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, rewards)
}
