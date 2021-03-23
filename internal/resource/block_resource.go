package resource

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/cache"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework/paginator"
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
	b, err := r.blockService.GetBestBlock(network(c))

	if err != nil {
		errorInternalServerError(c, err.Error())
		return
	}

	c.JSON(200, b.Height)
}

func (r *BlockResource) GetBestBlockCycle(c *gin.Context) {

	b, err := r.blockService.GetBestBlock(network(c))
	if err != nil {
		errorInternalServerError(c, err.Error())
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

	groups, err := r.blockService.GetBlockGroups(network(c), period, count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, groups.Items)
}

func (r *BlockResource) GetBlock(c *gin.Context) {
	hash := c.Param("hash")
	b, err := r.blockService.GetBlock(network(c), hash)
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
	b, err := r.blockService.GetBlock(network(c), c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": err, "status": http.StatusNotFound})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	bc, err := r.daoService.GetBlockCycleByBlock(network(c), b)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, bc)
}

func (r *BlockResource) GetBlocks(c *gin.Context) {
	blocks, total, err := r.blockService.GetBlocks(network(c), pagination(c))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	paginate := paginator.NewPaginator(len(blocks), total, pagination(c))
	paginate.WriteHeader(c)

	c.JSON(200, blocks)
}

func (r *BlockResource) GetRawBlock(c *gin.Context) {
	b, err := r.blockService.GetRawBlock(network(c), c.Param("hash"))
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
	tx, err := r.blockService.GetTransactionsByBlockHash(network(c), c.Param("hash"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err, "status": http.StatusInternalServerError})
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetTransactionByHash(c *gin.Context) {
	tx, err := r.blockService.GetTransactionByHash(network(c), c.Param("hash"))
	if err != nil {
		if err == repository.ErrBlockNotFound {
			errorNotFound(c, err.Error())
		} else {
			errorInternalServerError(c, err.Error())
		}
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetRawTransactionByHash(c *gin.Context) {
	tx, err := r.blockService.GetRawTransactionByHash(network(c), c.Param("hash"))
	if err != nil {
		if err == repository.ErrBlockNotFound {
			errorNotFound(c, err.Error())
		} else {
			errorInternalServerError(c, err.Error())
		}
		return
	}

	c.JSON(200, tx)
}

func (r *BlockResource) GetTransactions(c *gin.Context) {
	req := rest(c)

	callback := func() (interface{}, error) {
		txs, total, err := r.blockService.GetTransactions(req.Network(), req)
		e := make([]interface{}, len(txs))
		for i, v := range txs {
			e[i] = v
		}
		return paginator.Paginated{Elements: e, Total: total}, err
	}

	cacheKey := fmt.Sprintf("%s.GetTransactions(%s)", req.Network(), req.Query())
	response, err := r.cache.Get(cacheKey, callback, cache.RefreshingExpiration)
	transactions := response.(paginator.Paginated).Elements

	if err != nil {
		errorInternalServerError(c, err.Error())
		return
	}

	paginate := paginator.NewPaginator(len(transactions), response.(paginator.Paginated).Total, req.Pagination())
	paginate.WriteHeader(c)

	c.JSON(200, transactions)
}
