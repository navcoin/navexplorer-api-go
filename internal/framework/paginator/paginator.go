package paginator

import (
	"encoding/json"
	"github.com/navcoin/navexplorer-api-go/v2/internal/framework"
	"github.com/gin-gonic/gin"
	"math"
)

type Paginator struct {
	First            bool  `json:"first"`
	Last             bool  `json:"last"`
	Total            int64 `json:"total"`
	PageSize         int   `json:"size"`
	CurrentPage      int   `json:"current_page"`
	Pages            int   `json:"total_pages"`
	NumberOfElements int   `json:"number_of_elements"`
}

type Paginated struct {
	Elements []interface{}
	Total    int64
}

func NewPaginator(numberOfElements int, total int64, pagination framework.Pagination) Paginator {
	pages := int(math.Ceil(float64(total) / float64(pagination.Size())))
	if pages == 0 {
		pages = 1
	}

	return Paginator{
		CurrentPage:      pagination.Page(),
		Total:            total,
		PageSize:         pagination.Size(),
		Pages:            pages,
		NumberOfElements: numberOfElements,
		First:            pagination.Page() == 1,
		Last:             pagination.Page() == pages,
	}
}

func (p *Paginator) GetHeader() (header []byte) {
	header, _ = json.Marshal(p)

	return header
}

func (p *Paginator) WriteHeader(c *gin.Context) {
	c.Writer.Header().Set("X-Pagination", string(p.GetHeader()))
}
