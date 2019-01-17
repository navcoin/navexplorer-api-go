package coin

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type Controller struct{}

func (controller *Controller) GetWealthDistribution(c *gin.Context) {
	groupsQuery := c.DefaultQuery("groups", "10,100,1000")

	groups := make([]string, 0)
	groups = strings.Split(groupsQuery, ",")

	b := make([]int, len(groups))
	for i, v := range groups {
		b[i], _ = strconv.Atoi(v)
	}

	GetWealthDistribution(b)

	c.JSON(200, make([]string, 0))
}