package main

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/generated/dic"
	"github.com/NavExplorer/navexplorer-api-go/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sarulabs/dingo/v3"
	"net/http"
)

var container *dic.Container

func main() {
	config.Init()
	container, _ = dic.NewContainer(dingo.App)

	framework.SetReleaseMode(config.Get().Debug)

	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(framework.Cors())
	r.Use(framework.NetworkSelect)
	r.Use(framework.Options)
	r.Use(framework.ErrorHandler)

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to NavExplorer API!")
	})

	r.GET("/loaderio-4e202b2dc00926a931d50a76aa7fa34c.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "loaderio-4e202b2dc00926a931d50a76aa7fa34c")
	})

	includeLegacyApiEndpoints(r)

	addressResource := resource.NewAddressResource(container.GetAddressService())
	r.GET("/address", addressResource.GetAddresses)
	r.GET("/address/:hash", addressResource.GetAddress)
	r.GET("/address/:hash/tx", addressResource.GetTransactions)
	r.GET("/address/:hash/tx/cold", addressResource.GetColdTransactions)
	r.GET("/address/:hash/validate", addressResource.ValidateAddress)
	r.GET("/address/:hash/staking", addressResource.GetStakingReport)
	r.GET("/address/:hash/assoc/staking", addressResource.GetAssociatedStakingAddresses)

	blockResource := resource.NewBlockResource(container.GetBlockService(), container.GetDaoService())
	r.GET("/bestblock", blockResource.GetBestBlock)
	r.GET("/blockgroup", blockResource.GetBlockGroups)
	r.GET("/block", blockResource.GetBlocks)
	r.GET("/block/:hash", blockResource.GetBlock)
	r.GET("/block/:hash/cycle", blockResource.GetBlockCycle)
	r.GET("/block/:hash/raw", blockResource.GetRawBlock)
	r.GET("/block/:hash/tx", blockResource.GetTransactionsByBlock)
	r.GET("/tx/:hash", blockResource.GetTransactionByHash)
	r.GET("/tx/:hash/raw", blockResource.GetRawTransactionByHash)

	softForkResource := resource.NewSoftForkResource(container.GetSoftforkService(), container.GetSoftforkRepo())
	r.GET("/softfork", softForkResource.GetSoftForks)
	r.GET("/softfork/cycle", softForkResource.GetSoftForkCycle)

	daoGroup := r.Group("/dao")
	daoResource := resource.NewDaoResource(container.GetDaoService(), container.GetBlockService())
	daoGroup.GET("/cfund/consensus", daoResource.GetConsensus)
	daoGroup.GET("/cfund/stats", daoResource.GetCfundStats)
	daoGroup.GET("/cfund/proposal", daoResource.GetProposals)
	daoGroup.GET("/cfund/proposal/:hash", daoResource.GetProposal)
	daoGroup.GET("/cfund/proposal/:hash/votes", daoResource.GetProposalVotes)
	daoGroup.GET("/cfund/proposal/:hash/votes/:cycle/addresses", daoResource.GetProposalVoteAddresses)
	daoGroup.GET("/cfund/proposal/:hash/trend", daoResource.GetProposalVotes)
	daoGroup.GET("/cfund/proposal/:hash/payment-request", daoResource.GetPaymentRequestsForProposal)
	daoGroup.GET("/cfund/payment-request", daoResource.GetPaymentRequests)
	daoGroup.GET("/cfund/payment-request/:hash", daoResource.GetPaymentRequest)
	daoGroup.GET("/cfund/payment-request/:hash/votes", daoResource.GetPaymentRequestVotes)
	daoGroup.GET("/cfund/payment-request/:hash/votes/:cycle/addresses", daoResource.GetPaymentRequestVoteAddresses)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Resource not found"})
	})

	_ = r.Run(fmt.Sprintf(":%d", config.Get().Server.Port))
}

func includeLegacyApiEndpoints(r *gin.Engine) {
	api := r.Group("/api")

	legacyResource := resource.NewLegacyResource(
		container.GetAddressService(),
		container.GetBlockService(),
		container.GetCoinService(),
		container.GetDaoService(),
		container.GetSoftforkService(),
	)

	api.GET("/address", legacyResource.GetAddresses)
	api.GET("/address/:hash", legacyResource.GetAddress)
	//api.GET("/address/:hash/validate", legacyResource.ValidateAddress)
	api.GET("/address/:hash/tx", legacyResource.GetTransactions)
	api.GET("/address/:hash/coldtx", legacyResource.GetColdTransactions)
	api.GET("/address/:hash/chart/balance", legacyResource.GetBalanceChart)
	api.GET("/address/:hash/chart/staking", legacyResource.GetStakingChart)
	api.GET("/address/:hash/assoc/staking", legacyResource.GetAssociatedStakingAddresses)

	api.GET("/transactions/:type", legacyResource.GetTransactionsForAddresses)
	api.GET("/balance", legacyResource.GetBalancesForAddresses)

	api.GET("/bestblock", legacyResource.GetBestBlock)
	api.GET("/blockgroup", legacyResource.GetBlockGroups)
	api.GET("/block", legacyResource.GetBlocks)
	api.GET("/block/:hash", legacyResource.GetBlock)
	api.GET("/block/:hash/tx", legacyResource.GetBlockTransactions)
	api.GET("/block/:hash/raw", legacyResource.GetRawBlock)
	api.GET("/tx/:hash", legacyResource.GetTransaction)
	api.GET("/tx/:hash/raw", legacyResource.GetRawTransaction)

	api.GET("/coin/wealth", legacyResource.GetWealthDistribution)

	api.GET("/community-fund/block-cycle", legacyResource.GetBlockCycle)
	api.GET("/community-fund/stats", legacyResource.GetCfundStats)
	api.GET("/community-fund/proposal", legacyResource.GetProposals)
	api.GET("/community-fund/proposal/:hash", legacyResource.GetProposal)
	api.GET("/community-fund/proposal/:hash/trend", legacyResource.GetProposalVotingTrend)
	api.GET("/community-fund/proposal/:hash/vote/:vote", legacyResource.GetProposalVotes)
	api.GET("/community-fund/proposal/:hash/payment-request", legacyResource.GetPaymentRequestsForProposal)
	api.GET("/community-fund/payment-request", legacyResource.GetPaymentRequestsByState)
	api.GET("/community-fund/payment-request/:hash", legacyResource.GetPaymentRequestByHash)
	api.GET("/community-fund/payment-request/:hash/trend", legacyResource.GetPaymentRequestVotingTrend)
	api.GET("/community-fund/payment-request/:hash/vote/:vote", legacyResource.GetPaymentRequestVotes)

	api.GET("/search", legacyResource.Search)

	api.GET("/soft-fork", legacyResource.GetSoftForks)

	api.GET("/staking/report", legacyResource.GetStakingReport)
	api.GET("/staking/blocks", legacyResource.GetStakingByBlockCount)
	api.GET("/staking/rewards", legacyResource.GetStakingRewardsForAddresses)
}
