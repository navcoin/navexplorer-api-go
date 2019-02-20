package pagination

import (
	"encoding/json"
	"math"
)

type Paginator struct {
	CurrentPage int   `json:"currentPage"`
	First       bool  `json:"first"`
	Last        bool  `json:"last"`
	Total       int64 `json:"total"`
	Size        int   `json:"size"`
	Pages       int   `json:"total_pages"`
	Elements    int   `json:"number_of_elements"`
}

func NewPaginator(elements int, total int64, size int, page int) Paginator {
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

func (p *Paginator) GetHeader() (header []byte){
	header, _ = json.Marshal(p)

	return header
}