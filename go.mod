module github.com/NavExplorer/navexplorer-api-go

go 1.12

require (
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-contrib/gzip v0.0.1
	github.com/gin-gonic/gin v1.4.0
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/olivere/elastic v6.2.21+incompatible
	github.com/pkg/errors v0.8.1
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
