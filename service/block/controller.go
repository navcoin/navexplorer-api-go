package block

import (
	"github.com/NavExplorer/navexplorer-api-go/config"
	"github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/NavExplorer/navexplorer-api-go/navcoind"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Controller struct{}

func (controller *Controller) GetBestBlock(c *gin.Context) {
	block, err := GetBestBlock()

	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, block.Height)
}

func (controller *Controller) GetBlockGroups(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	count, err := strconv.Atoi(c.DefaultQuery("count", "10"))
	if err != nil || count < 10 {
		count = 10
	}

	groups, err := GetBlockGroups(period, count)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, groups)
}

func (controller *Controller) GetBlocks(c *gin.Context) {
	dir := c.DefaultQuery("dir", "DESC")

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		size = 10
	}
	if size > 1000 {
		size = 1000
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	blocks, total, err := GetBlocks(size, dir == "ASC", page)

	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

 	if blocks == nil {
		blocks = make([]Block, 0)
	}

	paginator := pagination.NewPaginator(len(blocks), total, size, page)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, blocks)
}

func (controller *Controller) GetBlock(c *gin.Context) {
	block, err := GetBlockByHashOrHeight(c.Param("hash"))

	if err != nil {
		if err == ErrBlockNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}

		return
	}

	c.JSON(200, block)
}


func (controller *Controller) GetRawBlock(c *gin.Context) {
	nav, err := navcoind.New(config.Get().SelectedNetwork)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	data, err := nav.GetBlock(c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.String(200, data)
}


func (controller *Controller) GetBlockTransactions(c *gin.Context) {
	hash := c.Param("hash")
	block, err := GetBlockByHashOrHeight(hash)
	transactions, err := GetTransactionsByHash(block.Hash)

	if err != nil {
		if err == ErrBlockNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}

		return
	}

	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	c.JSON(200, transactions)
}

func (controller *Controller) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")

	transaction, err := GetTransactionByHash(hash)

	if err != nil {
		if err == ErrTransactionNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}

		return
	}

	c.JSON(200, transaction)
}

func (controller *Controller) GetRawTransaction(c *gin.Context) {
	nav, err := navcoind.New(config.Get().SelectedNetwork)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	data, err := nav.GetRawTransaction(c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.String(200, data)
}
