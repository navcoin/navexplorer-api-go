package framework

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/config"
	networkService "github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const REST string = "rest"

type RestRequest interface {
	Network() networkService.Network
	Pagination() Pagination
	Filters() Filters
	Sort() Sort
	Query() string
}

type restRequest struct {
	network    networkService.Network
	pagination Pagination
	filters    Filters
	sort       Sort
	query      string
}

func RR() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := newRestRequestFromContext(c)
		if err != nil {
			logrus.WithError(err).Error("Failed to create rest request")
		}
	}
}

func newRestRequestFromContext(c *gin.Context) error {
	network, err := networkService.GetNetwork(func(c *gin.Context) string {
		n := c.GetHeader("Network")
		if n == "" {
			n = config.Get().DefaultNetwork
		}
		return n
	}(c))

	if err != nil {
		return err
	}

	pagination, err := newPaginationFromContext(c)
	if err != nil {
		return err
	}

	filters, err := newFiltersFromContext(c)
	if err != nil {
		return err
	}

	sorter, err := newSortFromContext(c, network)
	if err != nil {
		return err
	}

	c.Set(REST, &restRequest{
		network:    network,
		pagination: pagination,
		filters:    filters,
		sort:       sorter,
		query:      c.Request.URL.RawQuery,
	})

	return nil
}

func (rr *restRequest) Network() networkService.Network {
	return rr.network
}

func (rr *restRequest) Pagination() Pagination {
	return rr.pagination
}

func (rr *restRequest) Filters() Filters {
	return rr.filters
}

func (rr *restRequest) Sort() Sort {
	return rr.sort
}

func (rr *restRequest) Query() string {
	return rr.query
}
