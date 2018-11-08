package softFork

import (
	"github.com/gin-gonic/gin"
)

var service = new(Service)

type Controller struct{}

func (controller *Controller) GetSoftForks(c *gin.Context) {
	softForks, _ := service.GetSoftForks()

	c.JSON(200, softForks)
}