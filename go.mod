module github.com/NavExplorer/navexplorer-api-go/v2

go 1.14

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43

require (
	github.com/NavExplorer/navexplorer-indexer-go/v2 v2.1.4-0.20210401083341-9fb6d923c666
	github.com/getsentry/raven-go v0.2.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/gzip v0.0.3
	github.com/gin-gonic/gin v1.6.3
	github.com/joho/godotenv v1.3.0
	github.com/olivere/elastic/v7 v7.0.21
	github.com/sarulabs/dingo/v3 v3.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/streadway/amqp v1.0.0
	gopkg.in/go-playground/validator.v8 v8.18.2
)
