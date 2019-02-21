package main

import (
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/service/address"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"github.com/NavExplorer/navexplorer-api-go/service/coin"
	"github.com/NavExplorer/navexplorer-api-go/service/communityFund"
	"github.com/NavExplorer/navexplorer-api-go/service/search"
	"github.com/NavExplorer/navexplorer-api-go/service/softFork"
	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func init() {
	if config.Get().Sentry.Active == true {
		dsn := config.Get().Sentry.DSN
		log.Println("Sentry logging to ", dsn)
		raven.SetDSN(dsn)
	}
}

func main() {
	r := setupRouter()

	if config.Get().Ssl == false {
		r.Run(":" + config.Get().Server.Port)
	} else {
		log.Fatal(autotls.Run(r, config.Get().Server.Domain))
	}

	if config.Get().Sentry.Active == true {
		r.Use(sentry.Recovery(raven.DefaultClient, false))
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(networkSelect)
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(errorHandler)

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to NavExplorer API!")
	})

	api := r.Group("/api")

	addressController := new (address.Controller)
	api.GET("/address", addressController.GetAddresses)
	api.GET("/address/:hash", addressController.GetAddress)
	api.GET("/address/:hash/tx", addressController.GetTransactions)

	blockController := new (block.Controller)
	api.GET("/bestblock", blockController.GetBestBlock)
	api.GET("/blockgroup", blockController.GetBlockGroups)
	api.GET("/block", blockController.GetBlocks)
	api.GET("/block/:hash", blockController.GetBlock)
	api.GET("/block/:hash/tx", blockController.GetBlockTransactions)
	api.GET("/tx/:hash", blockController.GetTransaction)

	coinController := new (coin.Controller)
	api.GET("/coin/wealth", coinController.GetWealthDistribution)

	communityFundController := new (communityFund.Controller)
	api.GET("/community-fund/block-cycle", communityFundController.GetBlockCycle)
	api.GET("/community-fund/stats", communityFundController.GetStats)
	api.GET("/community-fund/proposal", communityFundController.GetProposals)
	api.GET("/community-fund/proposal/:hash", communityFundController.GetProposal)
	api.GET("/community-fund/proposal/:hash/trend", communityFundController.GetProposalVotingTrend)
	api.GET("/community-fund/proposal/:hash/vote/:vote", communityFundController.GetProposalVotes)
	api.GET("/community-fund/proposal/:hash/payment-request", communityFundController.GetProposalPaymentRequests)
	api.GET("/community-fund/payment-request", communityFundController.GetPaymentRequestsByState)
	api.GET("/community-fund/payment-request/:hash", communityFundController.GetPaymentRequestByHash)
	api.GET("/community-fund/payment-request/:hash/trend", communityFundController.GetPaymentRequestVotingTrend)
	api.GET("/community-fund/payment-request/:hash/vote/:vote", communityFundController.GetPaymentRequestVotes)

	searchController := new(search.Controller)
	api.GET("/search", searchController.Search)

	softForkController := new (softFork.Controller)
	api.GET("/soft-fork", softForkController.GetSoftForks)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource Not Found"})
	})

	return r
}

func networkSelect(c *gin.Context) {
	switch network := c.GetHeader("Network"); network {
	case "testnet":
		config.Get().SelectedNetwork = network
		break
	case "mainnet":
		config.Get().SelectedNetwork = network
		break
	default:
		config.Get().SelectedNetwork = "mainnet"
	}

	c.Header("X-Network", config.Get().SelectedNetwork)
	log.Printf("Using Network %s", config.Get().SelectedNetwork)
}

func errorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, c.Errors)
}