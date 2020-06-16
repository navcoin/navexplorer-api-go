package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AddressResource struct {
	addressService address.Service
}

func NewAddressResource(addressService address.Service) *AddressResource {
	return &AddressResource{addressService}
}

func (r *AddressResource) GetAddresses(c *gin.Context) {
	config, _ := pagination.Bind(c)

	addresses, total, err := r.addressService.GetAddresses(config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(addresses), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, addresses)
}

func (r *AddressResource) GetAddress(c *gin.Context) {
	address, err := r.addressService.GetAddress(c.Param("hash"))
	if err != nil {
		if err == repository.ErrAddressNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, address)
}

func (r *AddressResource) GetTransactions(c *gin.Context) {
	config, _ := pagination.Bind(c)

	txs, total, err := r.addressService.GetTransactions(c.Param("hash"), strings.Join(getFilters(c), " "), false, config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(txs), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, txs)
}

func (r *AddressResource) GetColdTransactions(c *gin.Context) {
	config, _ := pagination.Bind(c)

	filters := make([]string, 0)
	if filtersParam := c.DefaultQuery("filters", ""); filtersParam != "" {
		filters = strings.Split(filtersParam, ",")
	}

	txs, total, err := r.addressService.GetTransactions(c.Param("hash"), strings.Join(filters, " "), true, config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(txs), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, txs)
}

func (r *AddressResource) ValidateAddress(c *gin.Context) {
	validateAddress, err := r.addressService.ValidateAddress(c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, validateAddress)
}

func (r *AddressResource) GetStakingChart(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

	chart, err := r.addressService.GetStakingChart(period, c.Param("hash"))
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, chart)
}

func (r *AddressResource) GetAssociatedStakingAddresses(c *gin.Context) {
	addresses, err := r.addressService.GetAssociatedStakingAddresses(c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, addresses)
}

func (r *AddressResource) GetBalancesForAddresses(c *gin.Context) {
	_ = c.Request.ParseForm()

	addresses := make([]string, 0)
	if addressesParam := c.Request.Form.Get("addresses"); addressesParam != "" {
		addresses = strings.Split(addressesParam, ",")
	}

	balances, err := r.addressService.GetBalancesForAddresses(addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, balances)
}
