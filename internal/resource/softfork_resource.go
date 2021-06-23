package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/softfork"
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
	softForks, err := r.softForkRepository.GetSoftForks(network(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, softForks)
}

func (r *SoftForkResource) GetSoftForkCycle(c *gin.Context) {
	cycle, err := r.softForkService.GetCycle(network(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, cycle)
}
