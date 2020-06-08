module github.com/NavExplorer/navexplorer-api-go

go 1.13

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43

require (
	github.com/NavExplorer/navexplorer-indexer-go v0.0.0-20200531215233-a3474653be41
	github.com/getsentry/raven-go v0.2.0
	github.com/gin-contrib/cache v1.1.0 // indirect
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-contrib/gzip v0.0.1
	github.com/gin-gonic/gin v1.5.0
	github.com/joho/godotenv v1.3.0
	github.com/olivere/elastic/v7 v7.0.9
	github.com/sarulabs/dingo/v3 v3.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	gopkg.in/go-playground/validator.v8 v8.18.2
)
