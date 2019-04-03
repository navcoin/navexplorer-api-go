package network

import (
	"github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct{}

func (controller *Controller) GetNodes(c *gin.Context) {
	nodes, err := GetNodes()
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, nodes)
}
