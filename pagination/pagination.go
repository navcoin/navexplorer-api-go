package pagination

import (
	"encoding/json"
	"math"
)

type Paginator struct {
	First    bool          `json:"first"`
	Last     bool          `json:"last"`
	Total    int64         `json:"total"`
	Size     int           `json:"size"`
	Pages    int           `json:"total_pages"`
	Elements int           `json:"number_of_elements"`
}

func NewPaginator(elements int, total int64, size int, ascending bool, offset int) Paginator {
	paginator := Paginator{}

	paginator.Total = total
	paginator.Size = size
	paginator.Pages = int(math.Ceil(float64(total) / float64(size)))
	paginator.Elements = elements

	if ascending == false {
		paginator.First = offset == 0
		paginator.Last = total <= int64(size)
	}

	if ascending == true {
		paginator.Last = offset == 0
		paginator.First = total <= int64(size)
	}

	return paginator
}

func (p *Paginator) GetHeader() (header []byte){
	header, _ = json.Marshal(p)

	return header
}