package resource

import (
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/block"
	"github.com/navcoin/navexplorer-api-go/v2/internal/service/dao/consensus"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type SupplyResource struct {
	blockService     block.Service
	consensusService consensus.Service
}

func NewSupplyResource(blockService block.Service, consensusService consensus.Service) *SupplyResource {
	return &SupplyResource{blockService, consensusService}
}

func (r *SupplyResource) GetSupply(c *gin.Context) {
	blocks, err := strconv.Atoi(c.Query("blocks"))
	if err != nil {
		blocks = r.consensusService.GetParameter(network(c), consensus.VOTING_CYCLE_LENGTH).Value
	}
	fillEmpty, err := strconv.ParseBool(c.Query("fill"))
	if err != nil {
		fillEmpty = true
	}

	supply, err := r.blockService.GetSupply(network(c), blocks, fillEmpty)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, supply)
}
