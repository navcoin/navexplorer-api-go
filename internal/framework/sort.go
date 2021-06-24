package framework

import (
	"errors"
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/service/network"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strings"
)

var ErrInvalidSortValue = errors.New("The sort parameter is an invalid format")
var ErrInvalidSortDirection = errors.New("The sort direction is invalid format")

type Sort interface {
	Options() []SortOption
	IsEmpty() bool
	HasOption(field string) bool
}

type sort struct {
	options []SortOption
}

func NewSort(options []SortOption) Sort {
	return &sort{
		options: options,
	}
}

func newSortFromContext(c *gin.Context, n network.Network) (Sort, error) {
	sortQuery, exists := c.GetQuery("sort")
	if exists == false {
		return NewSort(nil), nil
	}

	options := make([]SortOption, 0)
	for _, param := range strings.Split(sortQuery, ",") {
		optionArray := strings.Split(param, ":")
		if len(optionArray) != 2 {
			return NewSort(nil), ErrInvalidSortValue
		}
		direction, err := SortDirectionByName(optionArray[1])
		if err != nil {
			return NewSort(nil), err
		}

		log.Info(optionArray[0])
		log.Info(optionArray[0])
		if n.NetworkNeedsPolyfill() {
			if optionArray[0] == "txheight" {
				optionArray[0] = "height"
			}
		}
		options = append(options, SortOption{
			field:     optionArray[0],
			direction: direction,
		})
	}

	return NewSort(options), nil
}

func (s *sort) Options() []SortOption {
	return s.options
}

func (s *sort) HasOption(field string) bool {
	for _, option := range s.options {
		if option.field == field {
			return true
		}
	}

	return false
}

func (s *sort) IsEmpty() bool {
	return len(s.options) == 0
}

type SortOption struct {
	field     string
	direction SortDirection
}

func NewSortOption(field string, direction SortDirection) SortOption {
	return SortOption{
		field:     field,
		direction: direction,
	}
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

func NewSortDirection(name string, value bool) SortDirection {
	return &sortDirection{
		name:  name,
		value: value,
	}
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
