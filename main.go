package main

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/generated/dic"
	"github.com/NavExplorer/navexplorer-api-go/internal/config"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource"
	"github.com/gin-contrib/cors"
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
	r.Use(cors.Default())
	r.Use(framework.NetworkSelect)
	r.Use(framework.Options)
	r.Use(framework.ErrorHandler)

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to NavExplorer API!")
	})

	r.GET("/loaderio-4e202b2dc00926a931d50a76aa7fa34c.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "loaderio-4e202b2dc00926a931d50a76aa7fa34c")
	})

	addressResource := resource.NewAddressResource(container.GetAddressRepo(), container.GetAddressTransactionRepo())
	r.GET("/address", addressResource.GetAddresses)
	r.GET("/address/:hash", addressResource.GetAddress)
	r.GET("/address/:hash/tx", addressResource.GetTransactions)
	r.GET("/address/:hash/tx/cold", addressResource.GetColdTransactions)
	r.GET("/address/:hash/validate", addressResource.ValidateAddress)

	blockResource := resource.NewBlockResource(container.GetBlockRepo(), container.GetBlockTransactionRepo())
	r.GET("/bestblock", blockResource.GetBestBlock)
	r.GET("/block", blockResource.GetBlocks)
	r.GET("/block/:hash", blockResource.GetBlock)
	r.GET("/block/:hash/raw", blockResource.GetRawBlock)
	r.GET("/block/:hash/tx", blockResource.GetTransactionsByBlock)
	r.GET("/tx/:hash", blockResource.GetTransactionByHash)
	r.GET("/tx/:hash/raw", blockResource.GetRawTransactionByHash)

	softForkResource := resource.NewSoftForkResource(container.GetSoftforkRepo())
	r.GET("/softfork", softForkResource.GetSoftForks)

	dao := r.Group("/dao")
	daoResource := resource.NewDaoResource(container.GetDaoProposalRepo(), container.GetDaoConsensusRepo())
	dao.GET("/consensus", daoResource.GetConsensus)
	dao.GET("/proposal", daoResource.GetProposals)
	dao.GET("/proposal/:hash", daoResource.GetProposal)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "Resource not found"})
	})

	_ = r.Run(fmt.Sprintf(":%d", config.Get().Server.Port))
}
