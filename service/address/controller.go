package address

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type Controller struct{}

func (controller *Controller) GetAddresses(c *gin.Context) {
	size, err := strconv.Atoi(c.Request.URL.Query().Get("size"))
	if err != nil {
		size = 100
	}

	addresses, err := GetAddresses(size)

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

	address, err := GetAddress(hash)

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

	filtersParam := c.DefaultQuery("filters", "")
	filters := make([]string, 0)
	if filtersParam != "" {
		filters = strings.Split(filtersParam, ",")
	}

	dir := c.DefaultQuery("dir", "DESC")

	size, sizeErr := strconv.Atoi(c.Query("size"))
	if sizeErr != nil {
		size = 50
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", ""))
	if err != nil {
		offset = 0
	}

	transactions, total, _ := GetTransactions(hash, strings.Join(filters, " "), size, dir == "ASC", offset)
	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	paginator := pagination.NewPaginator(len(transactions), total, size, dir == "ASC", offset)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, transactions)
}
