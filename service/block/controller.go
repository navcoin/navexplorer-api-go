package block

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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

	pagination, _ := json.Marshal(paginator)
	c.Writer.Header().Set("X-Pagination", string(pagination))

	c.JSON(200, blocks)

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
	transactions, err := service.GetTransactionsByHash(hash)

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
	types := strings.Split(c.Query("filters"), ",")
	size, sizeErr := strconv.Atoi(c.Query("size"))
	if sizeErr != nil {
		size = 100
	}

	offset := c.DefaultQuery("offset", "")

	transactions, paginator, _ := service.GetTransactions(dir, size, offset, types)

	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	pagination, _ := json.Marshal(paginator)
	c.Writer.Header().Set("X-Pagination", string(pagination))

	c.JSON(200, transactions)
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
