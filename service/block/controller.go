package block

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"strconv"
	"strings"
)

var service = new(Service)

type Controller struct{}

func (controller *Controller) GetBlocks(c *gin.Context) {
	dir := c.DefaultQuery("dir", "DESC")

	size, sizeErr := strconv.Atoi(c.Query("size"))
	if sizeErr != nil {
		size = 50
	}

	offset := c.DefaultQuery("offset", "")

	blocks, paginator, _ := service.GetBlocks(dir, size, offset)

	if size == 1 {
		c.JSON(200, blocks[0])
	} else {
		c.JSON(200, gin.H{
			"paginator": paginator,
			"content": blocks,
		})
	}
}

func (controller *Controller) GetBlock(c *gin.Context) {
	hash := c.Param("hash")

	block, err := service.GetBlockByHashOrHeight(hash)

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
	hash := c.Param("hash")

	transactions, err := service.GetTransactionsByBlock(hash)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find transactions for block: %s", hash),
		})
		c.Abort()
	} else {
		if transactions == nil {
			transactions = make([]Transaction, 0)
		}

		c.JSON(200, transactions)
	}
}

func (controller *Controller) GetTransactions(c *gin.Context) {
	dir := c.DefaultQuery("dir", "DESC")
	types := strings.Split(c.Query("types"), ",")
	size, sizeErr := strconv.Atoi(c.Query("size"))
	if sizeErr != nil {
		size = 100
	}

	offset := c.DefaultQuery("offset", "")

	transactions, _ := service.GetTransactions(dir, size, offset, types)

	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	if size == 1 {
		c.JSON(200, transactions[0])
	} else {
		c.JSON(200, transactions)
	}
}

func (controller *Controller) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")

	transaction, err := service.GetTransactionByHash(hash)

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
