package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type BlockResource struct {
	blockService block.Service
	daoService   dao.Service
	cache        *cache.Cache
}

func NewBlockResource(blockService block.Service, daoService dao.Service, cache *cache.Cache) *BlockResource {
	return &BlockResource{blockService, daoService, cache}
}

func (r *BlockResource) GetBestBlock(c *gin.Context) {
	callback := func() (interface{}, error) {
		b, err := r.blockService.GetBestBlock()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
			return nil, err
		}

		return b.Height, nil
	}

	height, err := r.cache.Get("block.bestblock", callback, cache.RefreshingExpiration)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, height)
}

func (r *BlockResource) GetBestBlockCycle(c *gin.Context) {
	b, err := r.blockService.GetBestBlock()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, b.BlockCycle)
}

func (r *BlockResource) GetBlockGroups(c *gin.Context) {
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

	groups, err := r.blockService.GetBlockGroups(period, count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, groups.Items)
}

func (r *BlockResource) GetBlock(c *gin.Context) {
	hash := c.Param("hash")
	callback := func() (interface{}, error) {
		return r.blockService.GetBlock(hash)
	}

	b, err := r.cache.Get(fmt.Sprintf("block.%s", hash), callback, cache.RefreshingExpiration)
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

func (r *BlockResource) GetBlockCycle(c *gin.Context) {
	b, err := r.blockService.GetBlock(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	bc, err := r.daoService.GetBlockCycleByBlock(b)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, bc)
}

func (r *BlockResource) GetBlocks(c *gin.Context) {
	config, _ := pagination.Bind(c)

	callback := func() (interface{}, error) {
		blocks, total, err := r.blockService.GetBlocks(config)
		e := make([]interface{}, len(blocks))
		for i, v := range blocks {
			e[i] = v
		}
		return pagination.Paginated{
			Elements: e,
			Total:    total,
		}, err
	}

	paginated, err := r.cache.Get(fmt.Sprintf("blocks.%s", config.ToString()), callback, cache.RefreshingExpiration)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(paginated.(pagination.Paginated).Elements), paginated.(pagination.Paginated).Total, config)
	paginator.WriteHeader(c)

	c.JSON(200, paginated.(pagination.Paginated).Elements)
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
	tx, err := r.blockService.GetTransactionsByBlockHash(c.Param("hash"))
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

func (r *BlockResource) GetTransactions(c *gin.Context) {
	config, _ := pagination.Bind(c)

	callback := func() (interface{}, error) {
		txs, total, err := r.blockService.GetTransactions(config, true, true)
		e := make([]interface{}, len(txs))
		for i, v := range txs {
			e[i] = v
		}
		return pagination.Paginated{
			Elements: e,
			Total:    total,
		}, err
	}

	paginated, err := r.cache.Get(fmt.Sprintf("txs.%s", config.ToString()), callback, cache.RefreshingExpiration)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(paginated.(pagination.Paginated).Elements), paginated.(pagination.Paginated).Total, config)
	paginator.WriteHeader(c)

	c.JSON(200, paginated.(pagination.Paginated).Elements)
}
