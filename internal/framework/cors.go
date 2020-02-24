package framework

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Network")
	config.ExposeHeaders = append(config.AllowHeaders, "X-Network", "X-Pagination")

	return cors.New(config)
}
