package address

import (
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/NavExplorer/navexplorer-api-go/pagination"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
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
		} else if err == ErrAddressNotValid {
			error.HandleError(c, err, http.StatusBadRequest)
		} else {
			error.HandleError(c, err, http.StatusInternalServerError)
		}

		return
	}

	c.JSON(200, address)
}

func (controller *Controller) ValidateAddress(c *gin.Context) {
	validateAddress, err := ValidateAddress(c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
	}

	c.JSON(200, validateAddress)
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

func (controller *Controller) GetStakingChart(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

	chart, err := GetStakingChart(period, c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, chart)
}

func (controller *Controller) GetBalancesForAddresses(c *gin.Context) {
	c.Request.ParseForm()

	addresses := make([]string, 0)
	if addressesParam := c.Request.Form.Get("addresses"); addressesParam != "" {
		addresses = strings.Split(addressesParam, ",")
	}

	balances, err := GetBalancesForAddresses(addresses)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, balances)
}

func (controller *Controller) GetTransactionsForAddresses(c *gin.Context) {
	addresses := strings.Split(c.Query("addresses"), ",")
	if len(addresses) == 0 {
		error.HandleError(c, errors.New("No addresses provided"), http.StatusBadRequest)
		return
	}

	endTimestamp, err := strconv.ParseInt(c.Query("end"), 10, 64)
	endTime := time.Now()
	if err != nil && endTimestamp != 0 {
		endTime = time.Unix(endTimestamp, 0)
	}

	startTimestamp, err := strconv.ParseInt(c.Query("start"), 10, 64)
	startTime := time.Now().Add(- (time.Hour * 24))
	if err != nil && startTimestamp != 0 {
		startTime = time.Unix(startTimestamp, 0)
	}

	transactions, err := GetTransactionsForAddresses(addresses, urlDecodeType(c.Param("type")), &startTime, &endTime)
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, transactions)
}

func (controller *Controller) GetAssociatedStakingAddresses(c *gin.Context) {
	addresses, err := GetAssociatedStakingAddresses(c.Param("hash"))
	if err != nil {
		error.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, addresses)
}

func urlDecodeType(txType string) string {
	txType = strings.ReplaceAll(txType, "-", "_")
	txType = strings.ToUpper(txType)

	return txType
}