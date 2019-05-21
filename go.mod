module github.com/NavExplorer/navexplorer-api-go

go 1.12

require (
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-contrib/gzip v0.0.1
	github.com/gin-gonic/gin v1.4.0
	github.com/google/go-cmp v0.3.0 // indirect
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/olivere/elastic v6.2.17+incompatible
	github.com/pkg/errors v0.8.0 // indirect
	github.com/ugorji/go/codec v0.0.0-20181022190402-e5e69e061d4f
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
