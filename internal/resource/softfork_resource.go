package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	network "github.com/NavExplorer/navexplorer-api-go/internal/service/network"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/softfork"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SoftForkResource struct {
	softForkService    softfork.Service
	softForkRepository repository.SoftForkRepository
}

func NewSoftForkResource(softForkService softfork.Service, softForkRepository repository.SoftForkRepository) *SoftForkResource {
	return &SoftForkResource{softForkService, softForkRepository}
}

func (r *SoftForkResource) GetSoftForks(c *gin.Context) {
	n, err := network.GetNetwork(networkHeader(c))
	softForks, err := r.softForkRepository.GetSoftForks(n)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, softForks)
}

func (r *SoftForkResource) GetSoftForkCycle(c *gin.Context) {
	n, err := network.GetNetwork(networkHeader(c))
	cycle, err := r.softForkService.GetCycle(n)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, cycle)
}
