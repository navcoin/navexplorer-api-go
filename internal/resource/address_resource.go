package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/resource/dto"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/address"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	a, err := r.addressService.GetAddress(n, c.Param("hash"))
	if err != nil {
		if err == repository.ErrAddressNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, a)
}

func (r *AddressResource) GetAddresses(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	cfg, _ := pagination.Bind(c)

	addresses, total, err := r.addressService.GetAddresses(n, cfg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(addresses), total, cfg)
	paginator.WriteHeader(c)

	c.JSON(200, addresses)
}

func (r *AddressResource) GetSummary(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	summary, err := r.addressService.GetAddressSummary(n, c.Param("hash"))
	if err != nil {
		if err == repository.ErrAddressHistoryNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, summary)
}

func (r *AddressResource) GetHistory(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	cfg, _ := pagination.Bind(c)

	var parameters dto.HistoryParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	history, total, err := r.addressService.GetHistory(n, c.Param("hash"), string(parameters.TxType), cfg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(history), total, cfg)
	paginator.WriteHeader(c)

	c.JSON(200, history)
}

func (r *AddressResource) ValidateAddress(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	validateAddress, err := r.addressService.ValidateAddress(n, c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, validateAddress)
}

func (r *AddressResource) GetAddressGroups(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	period := group.GetPeriod(c.DefaultQuery("period", "daily"))
	if period == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Invalid period `%s`", c.Query("period")),
			"status":  http.StatusBadRequest,
		})
		return
	}

	count, err := strconv.Atoi(c.DefaultQuery("count", "10"))
	if err != nil || count > 10 {
		count = 10
	}

	groups, err := r.addressService.GetAddressGroups(n, period, count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, groups)
}

func (r *AddressResource) GetStakingChart(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	period := c.DefaultQuery("period", "daily")

	chart, err := r.addressService.GetStakingChart(n, period, c.Param("hash"))
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, chart)
}

func (r *AddressResource) GetAssociatedStakingAddresses(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	addresses, err := r.addressService.GetAssociatedStakingAddresses(n, c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, addresses)
}

func (r *AddressResource) GetBalancesForAddresses(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	_ = c.Request.ParseForm()

	addresses := make([]string, 0)
	if addressesParam := c.Request.Form.Get("addresses"); addressesParam != "" {
		addresses = strings.Split(addressesParam, ",")
	}

	balances, err := r.addressService.GetNamedAddresses(n, addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, balances)
}
