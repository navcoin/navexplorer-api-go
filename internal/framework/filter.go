package framework

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strings"
)

var ErrInvalidFilterValue = errors.New("The filter parameter is an invalid format")

type Filters interface {
	Options() FilterOptions
	OnlySupportedOptions(supported []string) FilterOptions
	IsEmpty() bool
}

type filters struct {
	options FilterOptions
}

func newFilters(options FilterOptions) Filters {
	for _, o := range options {
		log.Infof("%s = %s", o.Field(), o.Values())
	}
	return &filters{
		options: options,
	}
}

func newFiltersFromContext(c *gin.Context) (Filters, error) {
	sortQuery, exists := c.GetQuery("filters")
	if exists == false {
		return newFilters(nil), nil
	}

	options := make([]FilterOption, 0)
	for _, param := range strings.Split(sortQuery, ",") {
		optionArray := strings.Split(param, ":")
		if len(optionArray) != 2 {
			return newFilters(nil), ErrInvalidFilterValue
		}

		values := make([]interface{}, 0)
		for _, v := range strings.Split(optionArray[1], "|") {
			values = append(values, v)
		}
		options = append(options, NewFilterOption(optionArray[0], values))
	}

	return newFilters(options), nil
}

func (f *filters) Options() FilterOptions {
	return f.options
}

func (f *filters) OnlySupportedOptions(supportedOptions []string) FilterOptions {
	result := make(FilterOptions, 0)

	for _, option := range f.options {
		for _, supported := range supportedOptions {
			if option.Field() == supported {
				result = append(result, option)
			}
		}
	}
	return result
}

func (f *filters) IsEmpty() bool {
	return len(f.options) == 0
}

type FilterOptions []FilterOption

func (fos *FilterOptions) Get(field string) (FilterOption, error) {
	for _, fo := range *fos {
		if fo.Field() == field {
			return fo, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Filter Option %s not found", field))
}

type FilterOption interface {
	Field() string
	Values() []interface{}
}

type filterOption struct {
	field  string
	values []interface{}
}

func NewFilterOption(field string, values []interface{}) FilterOption {
	return &filterOption{
		field:  field,
		values: values,
	}
}

func (fo *filterOption) Field() string {
	return fo.field
}

func (fo *filterOption) Values() []interface{} {
	return fo.values
}
