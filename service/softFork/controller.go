package softFork

import (
	"github.com/gin-gonic/gin"
)

type Controller struct{}

func (controller *Controller) GetSoftForks(c *gin.Context) {
	softForks, _ := GetSoftForks()

	c.JSON(200, softForks)
}