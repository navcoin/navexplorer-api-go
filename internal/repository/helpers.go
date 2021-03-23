package repository

import (
	"github.com/NavExplorer/navexplorer-api-go/v2/internal/framework"
	"github.com/olivere/elastic/v7"
)

func sort(service *elastic.SearchService, sorter framework.Sort) {
	for _, so := range sorter.Options() {
		service.Sort(so.Field(), so.Direction().Value())
	}
	if sorter.IsEmpty() {
		service.Sort("txheight", false)
	}

}
