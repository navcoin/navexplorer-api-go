package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/dto"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type AddressResource struct {
	addressService address.Service
}

func NewAddressResource(addressService address.Service) *AddressResource {
	return &AddressResource{addressService}
}

func (r *AddressResource) GetAddress(c *gin.Context) {
	a, err := r.addressService.GetAddress(c.Param("hash"))
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

func (r *AddressResource) GetSummary(c *gin.Context) {
	summary, err := r.addressService.GetAddressSummary(c.Param("hash"))
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
	config, _ := pagination.Bind(c)

	var parameters dto.HistoryParameters
	if err := c.BindQuery(&parameters); err != nil {
		log.WithError(err).Error("Failed to bind query")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request", "status": http.StatusBadRequest,
		})
		return
	}

	history, total, err := r.addressService.GetHistory(c.Param("hash"), string(parameters.TxType), config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(history), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, history)
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

	balances, err := r.addressService.GetNamedAddresses(addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, balances)
}
