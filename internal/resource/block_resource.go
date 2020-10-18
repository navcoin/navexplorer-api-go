package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/dao"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/group"
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
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	b, err := r.blockService.GetBestBlock(n)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, b.Height)
}

func (r *BlockResource) GetBestBlockCycle(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	b, err := r.blockService.GetBestBlock(n)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, b.BlockCycle)
}

func (r *BlockResource) GetBlockGroups(c *gin.Context) {
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

	groups, err := r.blockService.GetBlockGroups(n, period, count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, groups.Items)
}

func (r *BlockResource) GetBlock(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	hash := c.Param("hash")
	b, err := r.blockService.GetBlock(n, hash)
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
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	b, err := r.blockService.GetBlock(n, c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	bc, err := r.daoService.GetBlockCycleByBlock(n, b)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, bc)
}

func (r *BlockResource) GetBlocks(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	cfg, _ := pagination.Bind(c)

	blocks, total, err := r.blockService.GetBlocks(n, cfg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(blocks), total, cfg)
	paginator.WriteHeader(c)

	c.JSON(200, blocks)
}

func (r *BlockResource) GetRawBlock(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	b, err := r.blockService.GetRawBlock(n, c.Param("hash"))
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
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	tx, err := r.blockService.GetTransactionsByBlockHash(n, c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetTransactionByHash(c *gin.Context) {
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	tx, err := r.blockService.GetTransactionByHash(n, c.Param("hash"))
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
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	tx, err := r.blockService.GetRawTransactionByHash(n, c.Param("hash"))
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
	n, err := getNetwork(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Network not available", "status": http.StatusNotFound})
		return
	}

	cfg, _ := pagination.Bind(c)

	callback := func() (interface{}, error) {
		txs, total, err := r.blockService.GetTransactions(n, cfg, true, true)
		e := make([]interface{}, len(txs))
		for i, v := range txs {
			e[i] = v
		}
		return pagination.Paginated{
			Elements: e,
			Total:    total,
		}, err
	}

	paginated, err := r.cache.Get(
		fmt.Sprintf("%s.txs.%s", n.ToString(), cfg.ToString()),
		callback,
		cache.RefreshingExpiration,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginator := pagination.NewPaginator(len(paginated.(pagination.Paginated).Elements), paginated.(pagination.Paginated).Total, cfg)
	paginator.WriteHeader(c)

	c.JSON(200, paginated.(pagination.Paginated).Elements)
}
