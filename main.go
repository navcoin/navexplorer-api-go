package main

import (
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/service/address"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"github.com/NavExplorer/navexplorer-api-go/service/communityFund"
	"github.com/NavExplorer/navexplorer-api-go/service/softFork"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"net/http"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to NavExplorer API!")
	})

	api := r.Group("/api")

	addressController := new (address.Controller)
	api.GET("/address", addressController.GetAddresses)
	api.GET("/address/:hash", addressController.GetAddress)
	api.GET("/address/:hash/tx", addressController.GetTransactions)

	blockController := new (block.Controller)
	api.GET("/block", blockController.GetBlocks)
	api.GET("/block/:hash", blockController.GetBlock)
	api.GET("/block/:hash/tx", blockController.GetBlockTransactions)
	api.GET("/tx/:hash", blockController.GetTransaction)

	communityFundController := new (communityFund.Controller)
	api.GET("/community-fund/block-cycle", communityFundController.GetBlockCycle)
	api.GET("/community-fund/proposal", communityFundController.GetProposals)
	api.GET("/community-fund/proposal/:hash", communityFundController.GetProposal)
	api.GET("/community-fund/proposal/:hash/trend", communityFundController.GetProposalVotingTrend)
	api.GET("/community-fund/proposal/:hash/vote/:vote", communityFundController.GetProposalVotes)
	api.GET("/community-fund/proposal/:hash/payment-request", communityFundController.GetProposalPaymentRequests)
	api.GET("/community-fund/payment-request", communityFundController.GetPaymentRequestsByState)
	api.GET("/community-fund/payment-request/:hash", communityFundController.GetPaymentRequestByHash)
	api.GET("/community-fund/payment-request/:hash/trend", communityFundController.GetPaymentRequestVotingTrend)
	api.GET("/community-fund/payment-request/:hash/vote/:vote", communityFundController.GetPaymentRequestVotes)

	//groupController := new (group.Controller)
	//api.GET("/group/block/{category}/{count}", groupController.GetBlockGroup)

	softForkController := new (softFork.Controller)
	api.GET("/soft-fork", softForkController.GetSoftForks)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource Not Found"})
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run(":" + config.Get().Server.Port)
}
