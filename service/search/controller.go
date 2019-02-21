package search

import (
	"errors"
	error1 "github.com/NavExplorer/navexplorer-api-go/error"
	"github.com/NavExplorer/navexplorer-api-go/service/address"
	"github.com/NavExplorer/navexplorer-api-go/service/block"
	"github.com/NavExplorer/navexplorer-api-go/service/communityFund"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct{}

func (controller *Controller) Search(c *gin.Context) {
	query := c.Query("query")

	var result Result
	var err error

	_, err = communityFund.GetProposalByHash(query)
	if err == nil {
		result.Type = "proposal"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = communityFund.GetPaymentRequestByHash(query)
	if err == nil {
		result.Type = "paymentRequest"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = block.GetBlockByHashOrHeight(query)
	if err == nil {
		result.Type = "block"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = block.GetTransactionByHash(query)
	if err == nil {
		result.Type = "transaction"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = address.GetAddress(query)
	if err == nil {
		result.Type = "address"
		result.Value = query
		c.JSON(200, result)
		return
	}

	error1.HandleError(c, errors.New("no search result"), http.StatusNotFound)
}
