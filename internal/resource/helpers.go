package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/gin-gonic/gin"
	"strings"
)

func getFilters(c *gin.Context) []string {
	filters := make([]string, 0)
	if filtersParam := c.DefaultQuery("filters", ""); filtersParam != "" {
		filters = strings.Split(filtersParam, ",")
	}

	return filters
}

func getNetwork(c *gin.Context) (network.Network, error) {
	return network.GetNetwork(networkHeader(c))
}

func networkHeader(c *gin.Context) string {
	n := c.GetHeader("Network")
	if n == "" {
		n = config.Get().DefaultNetwork
	}

	return n
}

func handleError(c *gin.Context, err error, status int) {
	c.AbortWithStatusJSON(status, gin.H{
		"status":  status,
		"message": err.Error(),
	})
}
