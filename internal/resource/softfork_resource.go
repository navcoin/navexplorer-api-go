package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SoftForkResource struct {
	softForkRepository *repository.SoftForkRepository
}

func NewSoftForkResource(softForkRepository *repository.SoftForkRepository) *SoftForkResource {
	return &SoftForkResource{softForkRepository}
}

func (r *SoftForkResource) GetSoftForks(c *gin.Context) {
	softForks, err := r.softForkRepository.SoftForks()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, softForks)
}
