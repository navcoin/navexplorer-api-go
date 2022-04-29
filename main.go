package main

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/generated/dic"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/resource"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var container *dic.Container

func main() {
	config.Init()

	container, _ = dic.NewContainer()

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
	r.Use(framework.RR())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to NavExplorer API!")
	})

	authorized := r.Group("/auth", gin.BasicAuth(config.Account()))

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
	r.GET("/addresses", addressResource.GetAddressGroupsTotal)
	authorized.PUT("/address/:hash/meta", addressResource.PutAddressMeta)

	distributionResource := resource.NewDistributionResource(container.GetAddressService(), container.GetBlockService())
	r.GET("/distribution/supply", distributionResource.GetSupply)
	r.GET("/distribution/wealth", distributionResource.GetWealth)

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
	r.GET("/txcount", blockResource.CountTransactions)

	stakingResource := resource.NewStakingResource(container.GetAddressService(), container.GetStakingService())
	r.GET("/staking/blocks", stakingResource.GetBlocks)
	r.GET("/staking/rewards", stakingResource.GetStakingRewardsForAddresses)

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

	supplyResource := resource.NewSupplyResource(container.GetBlockService(), container.GetDaoConsensusService())
	r.GET("/supply", supplyResource.GetSupply)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Resource not found"})
	})

	_ = r.Run(fmt.Sprintf(":%d", config.Get().Server.Port))
}
