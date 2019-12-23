package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/elastic_cache/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/resource/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BlockResource struct {
	blockService *block.BlockService
}

func NewBlockResource(blockService *block.BlockService) *BlockResource {
	return &BlockResource{blockService}
}

func (r *BlockResource) GetBestBlock(c *gin.Context) {
	b, err := r.blockService.GetBestBlock()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, b.Height)
}

func (r *BlockResource) GetBlock(c *gin.Context) {
	b, err := r.blockService.GetBlock(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, b)
}

func (r *BlockResource) GetBlocks(c *gin.Context) {
	paginationConfig := pagination.GetConfig(c)

	blocks, total, err := r.blockService.GetBlocks(pagination.GetConfig(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(blocks), total, paginationConfig)
	paginator.WriteHeader(c)

	c.JSON(200, blocks)
}

func (r *BlockResource) GetRawBlock(c *gin.Context) {
	b, err := r.blockService.GetRawBlock(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, b)
}

func (r *BlockResource) GetTransactionsByBlock(c *gin.Context) {
	tx, err := r.blockService.GetTransactions(c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetTransactionByHash(c *gin.Context) {
	tx, err := r.blockService.GetTransactionByHash(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetRawTransactionByHash(c *gin.Context) {
	tx, err := r.blockService.GetRawTransactionByHash(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, tx)
}
