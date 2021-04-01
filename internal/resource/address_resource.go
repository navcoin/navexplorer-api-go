package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework/paginator"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/address"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type AddressResource struct {
	addressService address.Service
	cache          *cache.Cache
}

func NewAddressResource(addressService address.Service, cache *cache.Cache) *AddressResource {
	return &AddressResource{addressService, cache}
}

func (r *AddressResource) GetAddress(c *gin.Context) {
	a, err := r.addressService.GetAddress(network(c), c.Param("hash"))
	if err != nil {
		if err == repository.ErrAddressNotFound {
			errorNotFound(c, err.Error())
		} else {
			errorInternalServerError(c, err.Error())
		}
		return
	}

	c.JSON(200, a)
}

func (r *AddressResource) GetAddresses(c *gin.Context) {
	addresses, total, err := r.addressService.GetAddresses(network(c), pagination(c))
	if err != nil {
		errorInternalServerError(c, err.Error())
		return
	}

	paginator := paginator.NewPaginator(len(addresses), total, pagination(c))
	paginator.WriteHeader(c)

	c.JSON(200, addresses)
}

func (r *AddressResource) GetSummary(c *gin.Context) {
	summary, err := r.addressService.GetAddressSummary(network(c), c.Param("hash"))
	if err != nil {
		if err == repository.ErrAddressHistoryNotFound {
			errorNotFound(c, err.Error())
		} else {
			errorInternalServerError(c, err.Error())
		}
		return
	}

	c.JSON(200, summary)
}

func (r *AddressResource) GetHistory(c *gin.Context) {
	req := rest(c)

	history, total, err := r.addressService.GetHistory(req.Network(), c.Param("hash"), req)
	if err != nil {
		errorInternalServerError(c, err.Error())
		return
	}

	paginate := paginator.NewPaginator(len(history), total, req.Pagination())
	paginate.WriteHeader(c)

	c.JSON(200, history)
}

func (r *AddressResource) ValidateAddress(c *gin.Context) {
	validateAddress, err := r.addressService.ValidateAddress(network(c), c.Param("hash"))
	if err != nil {
		errorInternalServerError(c, err.Error())
		return
	}

	c.JSON(200, validateAddress)
}

func (r *AddressResource) GetAddressGroups(c *gin.Context) {
	period := group.GetPeriod(c.DefaultQuery("period", "daily"))
	if period == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Invalid period `%s`", c.Query("period")),
			"status":  http.StatusBadRequest,
		})
		return
	}

	count, err := strconv.Atoi(c.DefaultQuery("count", "10"))
	if err != nil || count > 100 {
		count = 10
	}

	groups, err := r.addressService.GetAddressGroups(network(c), period, count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, groups)
}

func (r *AddressResource) GetStakingChart(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	chart, err := r.addressService.GetStakingChart(network(c), period, c.Param("hash"))
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, chart)
}

func (r *AddressResource) GetAssociatedStakingAddresses(c *gin.Context) {
	addresses, err := r.addressService.GetAssociatedStakingAddresses(network(c), c.Param("hash"))
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

	balances, err := r.addressService.GetNamedAddresses(rest(c).Network(), addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, balances)
}
