package resource

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func getFilters(c *gin.Context) []string {
	filters := make([]string, 0)
	if filtersParam := c.DefaultQuery("filters", ""); filtersParam != "" {
		filters = strings.Split(filtersParam, ",")
	}

	return filters
}
