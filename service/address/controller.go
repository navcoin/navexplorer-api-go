package address

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"strconv"
	"strings"
)

var service = new(Service)

type Controller struct{}

func (controller *Controller) GetAddresses(c *gin.Context) {
	count, err := strconv.Atoi(c.Request.URL.Query().Get("count"))
	if err != nil {
		count = 100
	}

	addresses, err := service.GetAddresses(count)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Unable to retrieve addresses",
			"status": 500,
			"message": err,
		})
		c.Abort()
	} else {
		c.JSON(200, addresses)
	}
}

func (controller *Controller) GetAddress(c *gin.Context) {
	hash := c.Param("hash")

	address, err := service.GetAddress(hash)

	if err != nil {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"status": 404,
			"message": fmt.Sprintf("Could not find address: %s", hash),
		})
		c.Abort()
	} else {
		c.JSON(200, address)
	}
}

func (controller *Controller) GetTransactions(c *gin.Context) {
	hash := c.Param("hash")

	typesParam := c.DefaultQuery("filters", "")
	types := make([]string, 0)
	if typesParam != "" {
		types = strings.Split(typesParam, ",")
	}

	dir := c.DefaultQuery("dir", "DESC")

	size, sizeErr := strconv.Atoi(c.Query("size"))
	if sizeErr != nil {
		size = 50
	}

	offset := c.DefaultQuery("offset", "")

	transactions, paginator, _ := service.GetTransactions(hash, dir, size, offset, types)

	c.JSON(200, gin.H{
		"paginator": paginator,
		"content": transactions,
	})
}
