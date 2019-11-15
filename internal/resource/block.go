package resource

import (
	"github.com/NavExplorer/navexplorer-api-go/internal/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type BlockResource struct {
	blockRepo       *repository.BlockRepository
	transactionRepo *repository.BlockTransactionRepository
}

func NewBlockResource(blockRepo *repository.BlockRepository, blockTransactionRepo *repository.BlockTransactionRepository) *BlockResource {
	return &BlockResource{blockRepo, blockTransactionRepo}
}

func (r *BlockResource) GetBestBlock(c *gin.Context) {
	block, err := r.blockRepo.BestBlock()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, block.Height)
}

func (r *BlockResource) GetBlock(c *gin.Context) {
	block, err := r.blockRepo.BlockByHashOrHeight(c.Param("hash"))
	if err != nil {
		if err == repository.ErrBlockNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, block)
}

func (r *BlockResource) GetBlocks(c *gin.Context) {
	dir, size, page := pagination.GetPaginationParams(c)

	blocks, total, err := r.blockRepo.Blocks(size, dir, page)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(blocks), total, size, page)
	paginator.WriteHeader(c)

	c.JSON(200, blocks)
}

func (r *BlockResource) GetBlockGroups(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	count, err := strconv.Atoi(c.DefaultQuery("count", "10"))
	if err != nil || count < 10 {
		count = 10
	}

	groups, err := r.blockRepo.BlockGroups(period, count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, groups)
}

func (r *BlockResource) GetRawBlock(c *gin.Context) {
	block, err := r.blockRepo.RawBlockByHashOrHeight(c.Param("hash"))
	if err != nil {
		if err == repository.ErrBlockNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, block)
}

func (r *BlockResource) GetTransactionsByBlock(c *gin.Context) {
	block, err := r.blockRepo.BlockByHashOrHeight(c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
	}

	tx, err := r.transactionRepo.TransactionsByBlock(block)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetTransactionByHash(c *gin.Context) {
	tx, err := r.transactionRepo.TransactionByHash(c.Param("hash"))
	if err != nil {
		if err == repository.ErrBlockNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetRawTransactionByHash(c *gin.Context) {
	tx, err := r.transactionRepo.RawTransactionByHash(c.Param("hash"))
	if err != nil {
		if err == repository.ErrBlockNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		}
		return
	}

	c.JSON(200, tx)
}
