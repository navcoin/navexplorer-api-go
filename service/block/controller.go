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

	offset, err := strconv.Atoi(c.DefaultQuery("offset", ""))
	if err != nil {
		offset = 0
	}

	blocks, total, _ := GetBlocks(size, dir == "ASC", offset)

	paginator := pagination.NewPaginator(len(blocks), total, size, dir == "ASC", offset)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, blocks)

}

func (controller *Controller) GetBlock(c *gin.Context) {
	hash := c.Param("hash")

	block, err := GetBlockByHashOrHeight(hash)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find block: %s", hash),
		})
		c.Abort()
	} else {
		c.JSON(200, block)
	}
}

func (controller *Controller) GetBlockTransactions(c *gin.Context) {
	block, err := GetBlockByHashOrHeight(c.Param("hash"))
	transactions, err := GetTransactionsByHash(block.Hash)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find transactions for block: %s", block.Hash),
		})
		c.Abort()
	} else {
		if transactions == nil {
			transactions = make([]Transaction, 0)
		}

		c.JSON(200, transactions)
	}
}

func (controller *Controller) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")

	transaction, err := GetTransactionByHash(hash)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find transaction: %s", hash),
		})
		c.Abort()
	} else {
		c.JSON(200, transaction)
	}
}
