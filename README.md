# navexplorer-api-go
REST API for Navexplorer.com

## To run locally

```
go get -d ./...
go run main.go
```

## Available endpoints

```
GET    /api/address?size=100
GET    /api/address/:hash
GET    /api/address/:hash/tx?filters=staking,send,receive&size=10&page=1
GET    /api/bestblock
GET    /api/blockgroup?period=[hourly|daily|monthly]&count=10
GET    /api/block?dir=[ASC,DESC]&page=1&size=10
GET    /api/block/:hash
GET    /api/block/:hash/tx
GET    /api/tx/:hash
GET    /api/community-fund/block-cycle
GET    /api/community-fund/proposal?dir=[ASC,DESC]&size=10&page=1&state=[PENDING,ACCEPTED,EXPIRED,...]
GET    /api/community-fund/proposal/:hash
GET    /api/community-fund/proposal/:hash/vote/:vote
GET    /api/community-fund/proposal/:hash/payment-request
GET    /api/community-fund/payment-request?state=[PENDING,ACCEPTED,EXPIRED,...]
GET    /api/community-fund/payment-request/:hash
GET    /api/community-fund/payment-request/:hash/vote/:vote
GET    /api/search?query=[block#,blockHash,txHash,proposalHash,paymentRequestHash]
GET    /api/soft-fork

```

## Network Header

Use the Network header to switch between the available NavCoin networks.

Set `header('Network: mainnet')` for mainnet data

Set `header('Network: testnet')` for testnet data
