package address

import (
	"github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type Controller struct{}

func (controller *Controller) GetAddresses(c *gin.Context) {
	size, err := strconv.Atoi(c.Request.URL.Query().Get("size"))
	if err != nil {
		size = 100
	} else if size > 1000 {
		size = 1000
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	addresses, total, err := GetAddresses(size, page)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	paginator := pagination.NewPaginator(len(addresses), total, size, page)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, addresses)
}

func (controller *Controller) GetAddress(c *gin.Context) {
	address, err := GetAddress(c.Param("hash"))
	if err != nil {
		if err == ErrAddressNotFound {
			error.HandleError(c, err, http.StatusNotFound)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}

		return
	}

	c.JSON(200, address)
}

func (controller *Controller) GetTransactions(c *gin.Context) {
	hash := c.Param("hash")

	filters := make([]string, 0)
	if filtersParam := c.DefaultQuery("filters", ""); filtersParam != "" {
		filters = strings.Split(filtersParam, ",")
	}

	size, sizeErr := strconv.Atoi(c.Query("size"))
	if sizeErr != nil {
		size = 50
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	transactions, total, err := GetTransactions(hash, strings.Join(filters, " "), size, page)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	paginator := pagination.NewPaginator(len(transactions), total, size, page)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, transactions)
}

func (controller *Controller) GetColdTransactions(c *gin.Context) {
	hash := c.Param("hash")

	filters := make([]string, 0)
	if filtersParam := c.DefaultQuery("filters", ""); filtersParam != "" {
		filters = strings.Split(filtersParam, ",")
	}

	size, sizeErr := strconv.Atoi(c.Query("size"))
	if sizeErr != nil {
		size = 50
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	transactions, total, err := GetColdTransactions(hash, strings.Join(filters, " "), size, page)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	if transactions == nil {
		transactions = make([]Transaction, 0)
	}

	paginator := pagination.NewPaginator(len(transactions), total, size, page)
	c.Writer.Header().Set("X-Pagination", string(paginator.GetHeader()))

	c.JSON(200, transactions)
}

func (controller *Controller) GetBalanceChart(c *gin.Context) {
	chart, err := GetBalanceChart(c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, chart)
}