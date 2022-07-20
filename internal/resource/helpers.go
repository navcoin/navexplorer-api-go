package resource

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/config"
	"github.com/navcoin/navexplorer-api-go/v2/internal/framework"
	networkService "github.com/navcoin/navexplorer-api-go/v2/internal/service/network"
	"github.com/gin-gonic/gin"
	"net/http"
)

func rest(c *gin.Context) framework.RestRequest {
	return c.MustGet(framework.REST).(framework.RestRequest)
}

func network(c *gin.Context) networkService.Network {
	return rest(c).Network()
}

func pagination(c *gin.Context) framework.Pagination {
	return rest(c).Pagination()
}

func networkHeader(c *gin.Context) string {
	n := c.GetHeader("Network")
	if n == "" {
		n = config.Get().DefaultNetwork
	}

	return n
}

func errorNetworkNotAvailable(c *gin.Context) {
	c.AbortWithStatusJSON(
		http.StatusNotFound,
		gin.H{"message": "Network not available", "status": http.StatusNotFound},
	)
}

func errorNotFound(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(
		http.StatusNotFound,
		gin.H{"message": msg, "status": http.StatusNotFound},
	)
}

func ErrorBadRequest(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{"message": msg, "status": http.StatusBadRequest},
	)
}

func errorRequestError(c *gin.Context, err error) {
	errorInternalServerError(c, "Failed to process request:"+err.Error())
}

func errorInternalServerError(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(
		http.StatusInternalServerError,
		gin.H{"message": msg, "status": http.StatusInternalServerError})
}

func handleError(c *gin.Context, err error, status int) {
	c.AbortWithStatusJSON(status, gin.H{
		"status":  status,
		"message": err.Error(),
	})
}
