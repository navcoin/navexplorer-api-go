package repository

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/olivere/elastic/v7"
)

type defaultSort struct {
	Field     string
	Ascending bool
}

func sort(service *elastic.SearchService, sorter framework.Sort, defaultSort *defaultSort) {
	for _, so := range sorter.Options() {
		service.Sort(so.Field(), so.Direction().Value())
	}
	if sorter.IsEmpty() && defaultSort != nil {
		service.Sort(defaultSort.Field, defaultSort.Ascending)
	}
}
