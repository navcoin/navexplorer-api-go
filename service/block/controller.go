package block

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Controller struct{}

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
		if err.Error() == "Unable to connect to elastic search" {
			c.JSON(500, gin.H{
				"error": "Unable to get block",
				"status": 500,
				"message": err.Error(),
			})
		} else {
			c.JSON(404, gin.H{
				"error":   "Not Found",
				"status":  404,
				"message": fmt.Sprintf("Could not find block: %s", hash),
			})
		}

		c.Abort()

		return
	}

	c.JSON(200, block)
}

func (controller *Controller) GetBlockTransactions(c *gin.Context) {
	block, err := GetBlockByHashOrHeight(c.Param("hash"))
	transactions, err := GetTransactionsByHash(block.Hash)

	if err != nil {
		if err.Error() == "Unable to connect to elastic search" {
			c.JSON(500, gin.H{
				"error": "Unable to get transactions",
				"status": 500,
				"message": err.Error(),
			})
		} else {
			c.JSON(404, gin.H{
				"error": "Not Found",
				"status": 404,
				"message": fmt.Sprintf("Could not find transactions for block: %s", block.Hash),
			})
		}
		c.Abort()

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
		if err.Error() == "Unable to connect to elastic search" {
			c.JSON(500, gin.H{
				"error":   "Unable to get transaction",
				"status":  500,
				"message": err.Error(),
			})
		} else {
			c.JSON(404, gin.H{
				"error":   "Not Found",
				"status":  404,
				"message": fmt.Sprintf("Could not find transaction: %s", hash),
			})
		}

		c.Abort()

		return
	}

	c.JSON(200, transaction)
}
