package pagination

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"math"
)

type Paginator struct {
	First       bool  `json:"first"`
	Last        bool  `json:"last"`
	Total       int64 `json:"total"`
	PageSize    int   `json:"size"`
	CurrentPage int   `json:"current_page"`
	Pages       int   `json:"total_pages"`
	Elements    int   `json:"number_of_elements"`
}

type Config struct {
	Ascending bool `form:"ascending,default=false"`
	Size      int  `form:"size,default=10"`
	Page      int  `form:"page,default=1"`
}

func Bind(c *gin.Context) (*Config, error) {
	var config Config
	if err := c.BindQuery(&config); err != nil {
		log.WithError(err).Error("Failed to Bind pagination")
		return nil, err
	}

	return &config, nil
}

func NewPaginator(elements int, total int64, config *Config) Paginator {
	paginator := Paginator{}

	paginator.CurrentPage = config.Page
	paginator.Total = total
	paginator.PageSize = config.Size
	pages := int(math.Ceil(float64(total) / float64(config.Size)))
	if pages == 0 {
		pages = 1
	}
	paginator.Pages = pages
	paginator.Elements = elements
	paginator.First = config.Page == 1
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
