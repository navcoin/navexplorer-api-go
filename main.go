package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/NavExplorer/navexplorer-api-go/address"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to NavExplorer API!")
	})

	api := r.Group("/api")
	addressGroup := api.Group("/address")
	{
		addressController := new (address.Controller)
		addressGroup.GET("/", addressController.GetAddresses)
		addressGroup.GET("/:hash", addressController.GetAddress)
		addressGroup.GET("/:hash/tx", addressController.GetAddressTransactions)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource Not Found"})
	})

	return r
}

func main() {
	r := setupRouter()

	r.Run(":8888")
}
