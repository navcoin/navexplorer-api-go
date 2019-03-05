package staking

import (
	"github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct{}

func (controller *Controller) GetStakingAddresses(c *gin.Context) {
	stakingEstimate, err := GetStakingAddresses()
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, stakingEstimate)
}
