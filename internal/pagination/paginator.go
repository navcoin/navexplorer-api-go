package pagination

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"math"
	"strconv"
)

type Paginator struct {
	CurrentPage int  `json:"currentPage"`
	First       bool `json:"first"`
	Last        bool `json:"last"`
	Total       int  `json:"total"`
	Size        int  `json:"size"`
	Pages       int  `json:"total_pages"`
	Elements    int  `json:"number_of_elements"`
}

func GetPaginationParams(c *gin.Context) (dir bool, size int, page int) {
	dir = c.DefaultQuery("dir", "DESC") == "ASC"

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		size = 10
	}

	page, err = strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	return
}

func NewPaginator(elements int, total int, size int, page int) Paginator {
	paginator := Paginator{}

	paginator.CurrentPage = page
	paginator.Total = total
	paginator.Size = size
	paginator.Pages = int(math.Ceil(float64(total) / float64(size)))
	paginator.Elements = elements
	paginator.First = page == 1
	paginator.Last = paginator.CurrentPage == paginator.Pages

	return paginator
}

func (p *Paginator) GetHeader() (header []byte) {
	header, _ = json.Marshal(p)

	return header
}

func (p *Paginator) WriteHeader(ctx *gin.Context) {
	ctx.Writer.Header().Set("X-Pagination", string(p.GetHeader()))
}
