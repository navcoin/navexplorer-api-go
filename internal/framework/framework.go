package framework

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/config"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func SetReleaseMode(debug bool) {
	if debug {
		log.Printf("Mode: %s", gin.DebugMode)
		gin.SetMode(gin.DebugMode)
	} else {
		log.Printf("Mode: %s", gin.ReleaseMode)
		gin.SetMode(gin.ReleaseMode)
	}
}

func NetworkSelect(c *gin.Context) {
	network := c.GetHeader("Network")
	if network == "" {
		network = config.Get().DefaultNetwork
	}

	c.Header("X-Network", network)
}

func Options(c *gin.Context) {
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Content-Type", "application/json")
		c.AbortWithStatus(http.StatusOK)
	}
}
