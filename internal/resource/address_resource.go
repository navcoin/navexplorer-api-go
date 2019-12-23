package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AddressResource struct {
	addressService *address.AddressService
}

func NewAddressResource(addressService *address.AddressService) *AddressResource {
	return &AddressResource{addressService}
}

func (r *AddressResource) GetAddresses(c *gin.Context) {
	config := pagination.GetConfig(c)

	addresses, total, err := r.addressService.GetAddresses(pagination.GetConfig(c))
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
	config := pagination.GetConfig(c)

	txs, total, err := r.addressService.GetTransactions(c.Param("hash"), false, config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(txs), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, txs)
}

func (r *AddressResource) GetColdTransactions(c *gin.Context) {
	config := pagination.GetConfig(c)

	txs, total, err := r.addressService.GetTransactions(c.Param("hash"), true, config)
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
