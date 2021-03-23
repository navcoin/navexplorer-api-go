package framework

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

var ErrInvalidFilterValue = errors.New("The filter parameter is an invalid format")

type Filter interface {
	Options() FilterOptions
	OnlySupportedOptions(supported []string) FilterOptions
	IsEmpty() bool
}

type filter struct {
	options FilterOptions
}

func newFilter(options FilterOptions) Filter {
	return &filter{
		options: options,
	}
}

func newFilterFromContext(c *gin.Context) (Filter, error) {
	sortQuery, exists := c.GetQuery("filter")
	if exists == false {
		return newFilter(nil), nil
	}

	options := make([]FilterOption, 0)
	for _, param := range strings.Split(sortQuery, ",") {
		optionArray := strings.Split(param, ":")
		if len(optionArray) != 2 {
			return newFilter(nil), ErrInvalidFilterValue
		}
		options = append(options, NewFilterOption(optionArray[0], optionArray[1]))
	}

	return newFilter(options), nil
}

func (f *filter) Options() FilterOptions {
	return f.options
}

func (f *filter) OnlySupportedOptions(supportedOptions []string) FilterOptions {
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

func (f *filter) IsEmpty() bool {
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

func (fos *FilterOptions) GetAsBool(field string) (bool, error) {
	option, err := fos.Get(field)
	if err != nil {
		var b bool
		return b, err
	}

	return strconv.ParseBool(option.Value())
}

type FilterOption interface {
	Field() string
	Value() string
}

type filterOption struct {
	field string
	value string
}

func NewFilterOption(field string, value string) FilterOption {
	return &filterOption{
		field: field,
		value: value,
	}
}

func (fo *filterOption) Field() string {
	return fo.field
}

func (fo *filterOption) Value() string {
	return fo.value
}
