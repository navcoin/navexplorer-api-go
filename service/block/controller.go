package block

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Controller struct{}

func (controller *Controller) GetBestBlock(c *gin.Context) {
	block, err := GetBestBlock()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, block.Height)
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

	offset, err := strconv.Atoi(c.DefaultQuery("offset", ""))
	if err != nil {
		offset = 0
	}

	blocks, total, err := GetBlocks(size, dir == "ASC", offset)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

 	if blocks == nil {
		blocks = make([]Block, 0)
	}

	paginator := pagination.NewPaginator(len(blocks), total, size, dir == "ASC", offset)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, blocks)

}

func (controller *Controller) GetBlock(c *gin.Context) {
	hash := c.Param("hash")

	block, err := GetBlockByHashOrHeight(hash)

	if err != nil {
		if err == ErrBlockNotFound {
			c.Set("error", fmt.Sprintf("The `%s` block could not be found", hash))
			c.AbortWithError(404, err)
		} else {
			c.AbortWithError(500, err)
		}

		return
	}

	c.JSON(200, block)
}

func (controller *Controller) GetBlockTransactions(c *gin.Context) {
	hash := c.Param("hash")
	block, err := GetBlockByHashOrHeight(hash)
	transactions, err := GetTransactionsByHash(block.Hash)

	if err != nil {
		if err == ErrBlockNotFound {
			c.Set("error", fmt.Sprintf("The `%s` block could not be found", hash))
			c.AbortWithError(404, err)
		} else {
			c.AbortWithError(500, err)
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
			c.Set("error", fmt.Sprintf("The `%s` transcaction could not be found", hash))
			c.AbortWithError(404, err)
		} else {
			c.AbortWithError(500, err)
		}

		c.Abort()

		return
	}

	c.JSON(200, transaction)
}
