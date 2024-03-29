module github.com/navcoin/navexplorer-api-go/v2

go 1.14

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43

require (
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/gzip v0.0.3
	github.com/gin-gonic/gin v1.7.0
	github.com/joho/godotenv v1.3.0
	github.com/mattn/go-colorable v0.1.8
	github.com/navcoin/navexplorer-indexer-go/v2 v2.2.10
	github.com/olivere/elastic/v7 v7.0.21
	github.com/sarulabs/di/v2 v2.4.2
	github.com/sarulabs/dingo/v4 v4.2.0
	github.com/sirupsen/logrus v1.6.0
	go.uber.org/zap v1.17.0
	gopkg.in/go-playground/validator.v8 v8.18.2
)
