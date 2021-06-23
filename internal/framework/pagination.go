package framework

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

const (
	defaultSize int = 10
	defaultPage int = 1
	maxSize     int = 1000
)

type Pagination interface {
	Page() int
	Size() int
	From() int
}

type pagination struct {
	page int
	size int
}

func newPagination(page, size int) Pagination {
	return &pagination{
		page: page,
		size: size,
	}
}

func newPaginationFromContext(c *gin.Context) (Pagination, error) {
	page := defaultPage
	pageParam, exists := c.GetQuery("page")
	if exists == true {
		p, err := strconv.Atoi(pageParam)
		if err != nil {
			return newPagination(defaultPage, defaultSize), err
		}
		page = p
	}

	size := defaultSize
	sizeParam, exists := c.GetQuery("size")
	if exists == true {
		s, err := strconv.Atoi(sizeParam)
		if err != nil {
			return newPagination(defaultSize, defaultPage), err
		}
		if s > maxSize {
			s = maxSize
		}
		size = s
	}

	return newPagination(page, size), nil
}

func (p *pagination) Page() int {
	return p.page
}

func (p *pagination) Size() int {
	return p.size
}

func (p *pagination) From() int {
	return (p.page * p.size) - p.size
}
