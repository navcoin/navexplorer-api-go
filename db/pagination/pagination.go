package pagination

import "math"

type Paginator struct {
	First    bool          `json:"first"`
	Last     bool          `json:"last"`
	Total    int           `json:"total"`
	Size     int           `json:"size"`
	Pages    int           `json:"total_pages"`
	Elements int           `json:"number_of_elements"`
}

func NewPaginator(elements int, total int, size int, dir string, offset string) Paginator {
	paginator := Paginator{}

	paginator.Total = total
	paginator.Size = size
	paginator.Pages = int(math.Ceil(float64(total) / float64(size)))
	paginator.Elements = elements

	if dir == "DESC" {
		if offset == "" {
			paginator.First = true
		} else {
			paginator.First = false
		}
		if total <= size {
			paginator.Last = true
		} else {
			paginator.Last = false
		}
	}

	if dir == "ASC" {
		if offset == "" {
			paginator.Last = true
		} else {
			paginator.Last = false
		}
		if total <= size {
			paginator.First = true
		} else {
			paginator.First = false
		}
	}

	return paginator
}
