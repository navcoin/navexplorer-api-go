package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AddressResource struct {
	addressRepo *repository.AddressRepository
}

func NewAddressResource(addressRepo *repository.AddressRepository) *AddressResource {
	return &AddressResource{addressRepo}
}

func (r *AddressResource) GetAddresses(c *gin.Context) {
	_, size, page := pagination.GetPaginationParams(c)

	addresses, total, err := r.addressRepo.Addresses(size, page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(addresses), total, size, page)
	paginator.WriteHeader(c)

	c.JSON(200, addresses)
}

func (r *AddressResource) GetAddress(c *gin.Context) {
	address, err := r.addressRepo.AddressByHash(c.Param("hash"))
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

func (r *AddressResource) ValidateAddress(c *gin.Context) {
	validateAddress, err := r.addressRepo.Validate(c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, validateAddress)
}
