# navexplorer-api-go
REST API for Navexplorer.com

## To run locally

```
go get -d ./...
go run main.go
```

## Available endpoints

```
GET    /api/address
GET    /api/address/:hash
GET    /api/address/:hash/tx
GET    /api/block
GET    /api/block/:hash
GET    /api/block/:hash/tx
GET    /api/tx/:hash
GET    /api/community-fund/block-cycle
GET    /api/community-fund/proposal
GET    /api/community-fund/proposal/:hash
GET    /api/community-fund/proposal/:hash/vote/:vote
GET    /api/community-fund/proposal/:hash/payment-request
GET    /api/community-fund/payment-request
GET    /api/community-fund/payment-request/:hash
GET    /api/community-fund/payment-request/:hash/vote/:vote
GET    /api/soft-fork

```