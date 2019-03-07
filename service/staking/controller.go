package staking

import (
	"github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct{}

func (controller *Controller) GetStakingReport(c *gin.Context) {
	stakingEstimate, err := GetStakingReport()
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, stakingEstimate)
}
