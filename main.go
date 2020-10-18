package main

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/generated/dic"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/resource"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sarulabs/dingo/v3"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var container *dic.Container

func main() {
	config.Init()

	container, _ = dic.NewContainer(dingo.App)

	if config.Get().Debug {
		log.SetLevel(log.DebugLevel)
	}

	if config.Get().Subscribe {
		go container.GetBlockSubscriber().Subscribe()
	}

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

	addressResource := resource.NewAddressResource(container.GetAddressService(), container.GetCache())
	r.GET("/address", addressResource.GetAddresses)
	r.GET("/address/:hash", addressResource.GetAddress)
	r.GET("/address/:hash/summary", addressResource.GetSummary)
	r.GET("/address/:hash/history", addressResource.GetHistory)
	r.GET("/address/:hash/validate", addressResource.ValidateAddress)
	r.GET("/address/:hash/staking", addressResource.GetStakingChart)
	r.GET("/address/:hash/assoc/staking", addressResource.GetAssociatedStakingAddresses)
	r.GET("/balance", addressResource.GetBalancesForAddresses)
	r.GET("/addressgroup", addressResource.GetAddressGroups)

	distributionResource := resource.NewDistributionResource(container.GetDistributionService())
	r.GET("/distribution/total-supply", distributionResource.GetTotalSupply)

	blockResource := resource.NewBlockResource(container.GetBlockService(), container.GetDaoService(), container.GetCache())
	r.GET("/bestblock", blockResource.GetBestBlock)
	r.GET("/blockcycle", blockResource.GetBestBlockCycle)
	r.GET("/blockgroup", blockResource.GetBlockGroups)
	r.GET("/block", blockResource.GetBlocks)
	r.GET("/block/:hash", blockResource.GetBlock)
	r.GET("/block/:hash/cycle", blockResource.GetBlockCycle)
	r.GET("/block/:hash/raw", blockResource.GetRawBlock)
	r.GET("/block/:hash/tx", blockResource.GetTransactionsByBlock)
	r.GET("/tx", blockResource.GetTransactions)
	r.GET("/tx/:hash", blockResource.GetTransactionByHash)
	r.GET("/tx/:hash/raw", blockResource.GetRawTransactionByHash)

	//stakingResource := resource.NewStakingResource(container.GetAddressService())
	//r.GET("/staking/blocks", stakingResource.GetBlocks)
	//r.GET("/staking/rewards", stakingResource.GetStakingRewardsForAddresses)

	softForkResource := resource.NewSoftForkResource(container.GetSoftforkService(), container.GetSoftforkRepo())
	r.GET("/softfork", softForkResource.GetSoftForks)
	r.GET("/softfork/cycle", softForkResource.GetSoftForkCycle)

	daoGroup := r.Group("/dao")
	daoResource := resource.NewDaoResource(container.GetDaoService(), container.GetBlockService())
	daoGroup.GET("/consensus/parameters", daoResource.GetConsensusParameters)
	daoGroup.GET("/consensus/parameters/:id", daoResource.GetConsensusParameter)
	daoGroup.GET("/consultation", daoResource.GetConsultations)
	daoGroup.GET("/consultation/:hash", daoResource.GetConsultation)
	daoGroup.GET("/answer/:hash", daoResource.GetAnswer)
	daoGroup.GET("/consultation/:hash/:answer/votes", daoResource.GetAnswerVotes)

	cfundGroup := daoGroup.Group("/cfund")
	cfundGroup.GET("/stats", daoResource.GetCfundStats)
	cfundGroup.GET("/proposal", daoResource.GetProposals)
	cfundGroup.GET("/proposal/:hash", daoResource.GetProposal)
	cfundGroup.GET("/proposal/:hash/votes", daoResource.GetProposalVotes)
	cfundGroup.GET("/proposal/:hash/trend", daoResource.GetProposalTrend)
	cfundGroup.GET("/proposal/:hash/payment-request", daoResource.GetPaymentRequestsForProposal)
	cfundGroup.GET("/payment-request", daoResource.GetPaymentRequests)
	cfundGroup.GET("/payment-request/:hash", daoResource.GetPaymentRequest)
	cfundGroup.GET("/payment-request/:hash/votes", daoResource.GetPaymentRequestVotes)
	cfundGroup.GET("/payment-request/:hash/trend", daoResource.GetPaymentRequestTrend)

	searchResource := resource.NewSearchResource(container.GetAddressService(), container.GetBlockService(), container.GetDaoService())
	r.GET("/search", searchResource.Search)

	if config.Get().Legacy == true {
		includeLegacyApiEndpoints(r)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Resource not found"})
	})

	_ = r.Run(fmt.Sprintf(":%d", config.Get().Server.Port))
}

func includeLegacyApiEndpoints(r *gin.Engine) {
	log.Info("Including Legacy endpoints")
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
	//api.GET("/address/:hash/chart/balance", legacyResource.GetBalanceChart)
	//api.GET("/address/:hash/chart/staking", legacyResource.GetStakingChart)
	api.GET("/address/:hash/assoc/staking", legacyResource.GetAssociatedStakingAddresses)

	api.GET("/balance", legacyResource.GetBalancesForAddresses)

	api.GET("/bestblock", legacyResource.GetBestBlock)
	api.GET("/blockgroup", legacyResource.GetBlockGroups)
	api.GET("/block", legacyResource.GetBlocks)
	api.GET("/block/:hash", legacyResource.GetBlock)
	api.GET("/block/:hash/raw", legacyResource.GetRawBlock)
	api.GET("/tx/:hash", legacyResource.GetTransaction)
	api.GET("/tx/:hash/raw", legacyResource.GetRawTransaction)

	api.GET("/coin/wealth", legacyResource.GetWealthDistribution)

	api.GET("/community-fund/block-cycle", legacyResource.GetBlockCycle)
	api.GET("/community-fund/stats", legacyResource.GetCfundStats)
	api.GET("/community-fund/proposal/:hash", legacyResource.GetProposal)
	api.GET("/community-fund/proposal/:hash/trend", legacyResource.GetProposalVotingTrend)
	api.GET("/community-fund/proposal/:hash/vote/:vote", legacyResource.GetProposalVotes)
	api.GET("/community-fund/proposal/:hash/payment-request", legacyResource.GetPaymentRequestsForProposal)
	api.GET("/community-fund/payment-request/:hash", legacyResource.GetPaymentRequestByHash)
	api.GET("/community-fund/payment-request/:hash/trend", legacyResource.GetPaymentRequestVotingTrend)
	api.GET("/community-fund/payment-request/:hash/vote/:vote", legacyResource.GetPaymentRequestVotes)

	api.GET("/search", legacyResource.Search)

	api.GET("/soft-fork", legacyResource.GetSoftForks)

	//api.GET("/staking/report", legacyResource.GetStakingReport)
	//api.GET("/staking/blocks", legacyResource.GetStakingByBlockCount)
	//api.GET("/staking/rewards", legacyResource.GetStakingRewardsForAddresses)
}
