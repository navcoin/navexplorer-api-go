package framework

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
)

var ErrInvalidSortValue = errors.New("The sort parameter is an invalid format")
var ErrInvalidSortDirection = errors.New("The sort direction is invalid format")

type Sort interface {
	Options() []SortOption
	IsEmpty() bool
}

type sort struct {
	options []SortOption
}

func newSort(options []SortOption) Sort {
	return &sort{
		options: options,
	}
}

func newSortFromContext(c *gin.Context) (Sort, error) {
	sortQuery, exists := c.GetQuery("sort")
	if exists == false {
		return newSort(nil), nil
	}

	options := make([]SortOption, 0)
	for _, param := range strings.Split(sortQuery, ",") {
		optionArray := strings.Split(param, ":")
		if len(optionArray) != 2 {
			return newSort(nil), ErrInvalidSortValue
		}
		direction, err := SortDirectionByName(optionArray[1])
		if err != nil {
			return newSort(nil), err
		}
		options = append(options, SortOption{
			field:     optionArray[0],
			direction: direction,
		})
	}

	return newSort(options), nil
}

func (s *sort) Options() []SortOption {
	return s.options
}

func (s *sort) IsEmpty() bool {
	return len(s.options) == 0
}

type SortOption struct {
	field     string
	direction SortDirection
}

func (so *SortOption) Field() string {
	return so.field
}

func (so *SortOption) Direction() SortDirection {
	return so.direction
}

type SortDirection interface {
	Name() string
	Value() bool
}

type sortDirection struct {
	name  string
	value bool
}

func (sd *sortDirection) Name() string {
	return sd.name
}

func (sd *sortDirection) Value() bool {
	return sd.value
}

func SortDirections() []SortDirection {
	return []SortDirection{
		&sortDirection{"asc", true},
		&sortDirection{"desc", false},
	}
}

func SortDirectionByName(name string) (SortDirection, error) {
	for _, v := range SortDirections() {
		if v.Name() == name {
			return v, nil
		}
	}

	return nil, ErrInvalidSortDirection
}
