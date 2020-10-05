package resource

import (
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/param"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SearchResource struct {
	addressService address.Service
	blockService   block.Service
	daoService     dao.Service
}

type Result struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func NewSearchResource(addressService address.Service, blockService block.Service, daoService dao.Service) *SearchResource {
	return &SearchResource{
		addressService,
		blockService,
		daoService,
	}
}

func (r *SearchResource) Search(c *gin.Context) {
	query := c.Query("query")
	network := param.GetNetwork()

	if _, err := r.daoService.GetProposal(network, query); err == nil {
		c.JSON(200, &Result{"proposal", query})
		return
	}

	if _, err := r.daoService.GetPaymentRequest(network, query); err == nil {
		c.JSON(200, &Result{"paymentRequest", query})
		return
	}

	if _, err := r.blockService.GetBlock(network, query); err == nil {
		c.JSON(200, &Result{"block", query})
		return
	}

	if _, err := r.blockService.GetTransactionByHash(network, query); err == nil {
		c.JSON(200, &Result{"transaction", query})
		return
	}

	if _, err := r.addressService.GetAddress(network, query); err == nil {
		c.JSON(200, &Result{"address", query})
		return
	}

	handleError(c, errors.New("no search result"), http.StatusNotFound)
}
