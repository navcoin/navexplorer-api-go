package softFork

import (
	"github.com/gin-gonic/gin"
)

type Controller struct{}

func (controller *Controller) GetSoftForks(c *gin.Context) {
	softForks, err := GetSoftForks()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, softForks)
}